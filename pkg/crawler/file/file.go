package file

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	QueuedFilesSizeLimit int64
}

type crawlerImplementation struct {
	*crawler.Crawler

	keywords        []string
	dispatcher      *dispatcher.Dispatcher
	resultRetriever result.Retriever
	pool            *tunny.Pool

	queuedFilesMutex     sync.Mutex
	queuedFiles          []*dispatcher.FileCrawlerPayload
	queuedFilesSize      int64
	queuedFilesSizeLimit int64

	done chan struct{}
}

var _ crawler.FileCrawler = (*crawlerImplementation)(nil)

func NewCrawlerImplementation(c *Config) crawler.FileCrawler {
	ci := &crawlerImplementation{
		Crawler:    c.Crawler,
		dispatcher: c.Dispatcher,
		keywords:   c.Keywords,
		done:       make(chan struct{}),

		queuedFilesMutex:     sync.Mutex{},
		queuedFiles:          make([]*dispatcher.FileCrawlerPayload, 0),
		queuedFilesSize:      0,
		queuedFilesSizeLimit: c.QueuedFilesSizeLimit,
	}

	pool := tunny.NewFunc(200, ci.countWordsMultipleFiles)
	ci.pool = pool

	return ci
}

func (ci *crawlerImplementation) Start() {
	for {
		select {
		case msg := <-ci.dispatcher.Stream(dispatcher.DirectoryJobType):
			go ci.handleDirectory(msg.Payload)
		case msg := <-ci.dispatcher.Stream(dispatcher.FileJobType):
			go ci.handleFile(msg.Payload)
		case <-ci.done:
			ci.pool.Close()
			return
		}
	}
}

func (ci *crawlerImplementation) handleDirectory(payload dispatcher.JobPayload) {
	dirPayload, ok := payload.(dispatcher.DirectoryCrawlerPayload)
	if !ok {
		ci.Logger.Error("payload not of type directory crawler payload")
	}

	filePayloads := make([]*dispatcher.FileCrawlerPayload, 0)

	filepath.Walk(dirPayload.Path, func(path string, f os.FileInfo, err error) error {
		filePayloads = append(filePayloads, &dispatcher.FileCrawlerPayload{
			CorpusName: dirPayload.CorpusName,
			Path:       path,
			Size:       f.Size(),
		})
		return nil
	})

	ci.resultRetriever.InitializeSummary(dispatcher.FileJobType, dirPayload.CorpusName, len(filePayloads), time.Time{})

	for _, fp := range filePayloads {
		ci.dispatcher.Push(&dispatcher.Job{
			Type:    dispatcher.FileJobType,
			Payload: fp,
		})
	}
}

func (ci *crawlerImplementation) handleFile(payload dispatcher.JobPayload) {
	defer ci.queuedFilesMutex.Unlock()
	ci.queuedFilesMutex.Lock()

	filePayload, ok := payload.(*dispatcher.FileCrawlerPayload)
	if !ok {
		ci.Logger.Error("couldn't cast file payload to file crawler")
		return
	}

	if filePayload.Size+ci.queuedFilesSize > ci.queuedFilesSizeLimit {
		ci.dispatcher.Push(&dispatcher.Job{
			Payload: payload,
			Type:    dispatcher.DirectoryJobType,
		})

		queuedFiles := append([]*dispatcher.FileCrawlerPayload{}, ci.queuedFiles...)
		ci.queuedFiles = make([]*dispatcher.FileCrawlerPayload, 0, 0)
		ci.queuedFilesSize = 0

		go ci.startCount(queuedFiles)
		return
	}

	ci.queuedFiles = append(ci.queuedFiles, filePayload)
	ci.queuedFilesSize += filePayload.Size
}

func (ci *crawlerImplementation) startCount(payload []*dispatcher.FileCrawlerPayload) {
	if ci.pool.GetSize() == 0 {
		return
	}

	_, err := ci.pool.ProcessTimed(payload, 60*time.Second)

	if err == tunny.ErrJobTimedOut {
		ci.Logger.Error("goroutine timed out", "err", err)
		return
	}

	if err != nil {
		ci.Logger.Error("there was an error while counting words", "err", err)
	}
}

func (ci *crawlerImplementation) countWordsMultipleFiles(payload interface{}) interface{} {
	filePayloads, ok := payload.([]*dispatcher.FileCrawlerPayload)
	if !ok {
		return errors.New("couldn't cast job payload to file payload")
	}

	for _, fp := range filePayloads {
		err := ci.countWords(fp)
		if err != nil {
			ci.Logger.Error("couldn't count words for file")
		}
	}
	return nil
}

func (ci *crawlerImplementation) countWords(filePayload *dispatcher.FileCrawlerPayload) error {
	data, err := os.ReadFile(filePayload.Path)
	if err != nil {
		return errors.Wrap(err, "couldn't read file")
	}

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
