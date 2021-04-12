package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Jeffail/tunny"
	"github.com/l2cup/kids1/pkg/crawler"
	"github.com/l2cup/kids1/pkg/dispatcher"
	"github.com/l2cup/kids1/pkg/result"
)

type Config struct {
	Crawler              *crawler.Crawler
	Dispatcher           *dispatcher.Dispatcher
	ResultRetriever      result.Retriever
	Keywords             []string
	QueuedFilesSizeLimit uint64
}

var _ crawler.FileCrawler = (*crawlerImplementation)(nil)

type crawlerImplementation struct {
	*crawler.Crawler

	keywords             []string
	dispatcher           *dispatcher.Dispatcher
	resultRetriever      result.Retriever
	pool                 *tunny.Pool
	queuedFilesSizeLimit uint64

	done chan struct{}
}

func NewCrawlerImplementation(c *Config) crawler.FileCrawler {
	ci := &crawlerImplementation{
		Crawler:              c.Crawler,
		dispatcher:           c.Dispatcher,
		resultRetriever:      c.ResultRetriever,
		keywords:             c.Keywords,
		done:                 make(chan struct{}),
		queuedFilesSizeLimit: c.QueuedFilesSizeLimit,
	}

	ci.pool = tunny.NewFunc(200, ci.wordCounterWorker)
	return ci
}

func (ci *crawlerImplementation) Start() {
	for {
		select {
		case msg := <-ci.dispatcher.Stream(dispatcher.DirectoryJobType):
			go ci.handleDirectory(msg.Payload)
		case <-ci.done:
			ci.pool.Close()
			return
		}
	}
}

func (ci *crawlerImplementation) handleDirectory(payload dispatcher.JobPayload) {
	dirPayload, ok := payload.(*dispatcher.DirectoryCrawlerPayload)
	if !ok {
		ci.Logger.Error("payload not of type directory crawler payload", "type", fmt.Sprintf("%T", payload))
		return
	}

	filePayloads := make([]*dispatcher.FileCrawlerPayload, 0)

	err := filepath.Walk(dirPayload.Path, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}

		filePayloads = append(filePayloads, &dispatcher.FileCrawlerPayload{
			CorpusName: dirPayload.CorpusName,
			Path:       path,
			Size:       f.Size(),
		})
		ci.Logger.Debug("appended file payload", "payload", filePayloads)
		return nil
	})

	if err != nil {
		ci.Logger.Error("couldn't handle directory", "err", err)
		return
	}

	ci.resultRetriever.InitializeSummary(dispatcher.FileJobType, dirPayload.CorpusName, len(filePayloads), time.Time{})

	minimumJobCount := uint64(dirPayload.Size) / ci.queuedFilesSizeLimit
	if minimumJobCount == 0 {
		minimumJobCount = 1
	}

	filesPerBatch := len(filePayloads) / int(minimumJobCount)
	if len(filePayloads)%int(minimumJobCount) > 0 {
		filesPerBatch += 1
	}

	ci.Logger.Debug("finished calculating", "min_job_count", minimumJobCount, "files_per_batch", filesPerBatch)

	for i := 0; i < int(minimumJobCount); i++ {
		queuedFiles := make([]*dispatcher.FileCrawlerPayload, 0, filesPerBatch)

		for j := 0; j < filesPerBatch; j++ {
			if (i*int(minimumJobCount) + j) == len(filePayloads) {
				break
			}
			queuedFiles = append(queuedFiles, filePayloads[i*int(minimumJobCount)+j])
		}

		go ci.startWCWorker(append(make([]*dispatcher.FileCrawlerPayload, 0, len(queuedFiles)), queuedFiles...))
	}
}

func (ci *crawlerImplementation) startWCWorker(payload []*dispatcher.FileCrawlerPayload) {
	if ci.pool.GetSize() == 0 {
		return
	}

	_, err := ci.pool.ProcessTimed(payload, 60*time.Second)
	ci.Logger.Debug("started timed file process with payload", "payload", payload)

	if err == tunny.ErrJobTimedOut {
		ci.Logger.Error("goroutine timed out", "err", err)
		return
	}

	if err != nil {
		ci.Logger.Error("there was an error while counting words", "err", err)
	}
}

func (ci *crawlerImplementation) wordCounterWorker(payload interface{}) interface{} {
	filePayloads, ok := payload.([]*dispatcher.FileCrawlerPayload)
	if !ok {
		return errors.New("couldn't cast job payload to file payload")
	}

	for _, fp := range filePayloads {
		err := ci.countWords(fp)
		if err != nil {
			ci.Logger.Error("couldn't count words for file", "err", err)
		}
	}
	return nil
}

func (ci *crawlerImplementation) countWords(filePayload *dispatcher.FileCrawlerPayload) error {
	data, err := os.ReadFile(filePayload.Path)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("couldn't read file, path %s", filePayload.Path))
	}

	ci.Logger.Debug("starting word count for file", "file", filePayload.Path)
	results := make(map[string]int64)
	for _, word := range ci.keywords {
		results[word] = 0
	}

	words := strings.Fields(string(data))
	for _, w := range words {
		if result, ok := results[w]; ok {
			results[w] = result + 1
		}
	}

	ci.Logger.Debug("ended word count for file", "file", filePayload.Path, "results", results)

	ci.resultRetriever.UpdateSummary(&result.Results{
		JobType:    dispatcher.FileJobType,
		CorpusName: filePayload.CorpusName,
		Results:    results,
	})

	return nil
}

func (ci *crawlerImplementation) Stop() {
	<-ci.done
}
