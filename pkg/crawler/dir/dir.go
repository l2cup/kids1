package dir

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/l2cup/kids1/pkg/crawler"
	"github.com/l2cup/kids1/pkg/dispatcher"
	"github.com/l2cup/kids1/pkg/errors"
)

type Config struct {
	Crawler     *crawler.Crawler
	SleepTimeMS uint64
	Prefix      string
	Dispatcher  *dispatcher.Dispatcher
}

type crawlerImplementation struct {
	*crawler.Crawler

	dispatcher        *dispatcher.Dispatcher
	prefix            string
	sleepTime         time.Duration
	lastModifiedCache map[string]time.Time
	mutex             sync.Mutex
	directories       []string

	done chan struct{}
}

var _ crawler.DirCrawler = (*crawlerImplementation)(nil)

func NewCrawlerImplementation(c *Config) crawler.DirCrawler {
	sleepTime, err := time.ParseDuration(fmt.Sprintf("%dms", c.SleepTimeMS))
	if err != nil {
		c.Crawler.Logger.Fatal("couldn't parse dir sleep time duration", "err", err, "duration", c.SleepTimeMS)
	}

	return &crawlerImplementation{
		Crawler:           c.Crawler,
		lastModifiedCache: make(map[string]time.Time),
		done:              make(chan struct{}),
		sleepTime:         sleepTime,
		directories:       make([]string, 0),
		mutex:             sync.Mutex{},
		prefix:            c.Prefix,
	}
}

func (ci *crawlerImplementation) Start() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ticker := time.NewTicker(ci.sleepTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go ci.crawl()
		case <-interrupt:
			return
		case <-ci.done:
			return
		}
	}
}

func (ci *crawlerImplementation) AddDirectoryPath(path string) errors.Error {
	defer ci.mutex.Unlock()
	ci.mutex.Lock()

	return errors.Nil()
}

func (ci *crawlerImplementation) Stop() {
	ci.done <- struct{}{}
}

func (ci *crawlerImplementation) crawl() {
	defer ci.mutex.Unlock()
	ci.mutex.Lock()

	if len(ci.directories) == 0 {
		return
	}

	for _, dir := range ci.directories {
		ci.crawlDir(dir)
	}
}

func (ci *crawlerImplementation) crawlDir(path string) {

	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if (f.IsDir() && !strings.HasPrefix(f.Name(), ci.prefix)) || !f.IsDir() {
			return nil
		}

		lastMod, exists := ci.lastModifiedCache[path]
		if !exists {
			ci.lastModifiedCache[path] = f.ModTime()
			ci.pushJob(f.Name(), path, f.Size())
			return filepath.SkipDir
		}

		if lastMod != f.ModTime() {
			ci.pushJob(f.Name(), path, f.Size())
		}

		return filepath.SkipDir
	})

	if err != nil {
		ci.Logger.Error("error walking path", "err", err, "path", path)
	}
}

func (ci *crawlerImplementation) pushJob(corpusName, path string, size int64) {
	ci.dispatcher.Push(&dispatcher.Job{
		Type: dispatcher.DirectoryJobType,
		Payload: &dispatcher.DirectoryCrawlerPayload{
			CorpusName: corpusName,
			Path:       path,
			Size:       size,
		},
	})
}
