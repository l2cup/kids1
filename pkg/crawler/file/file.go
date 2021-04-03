package file

import (
	"github.com/Jeffail/tunny"
	"github.com/l2cup/kids1/pkg/crawler"
	"github.com/l2cup/kids1/pkg/dispatcher"
	"github.com/l2cup/kids1/pkg/result"
)

type Config struct {
	Crawler         *crawler.Crawler
	Dispatcher      *dispatcher.Dispatcher
	ResultRetriever *result.Retriever
}

type crawlerImplementation struct {
	*crawler.Crawler

	dispatcher      *dispatcher.Dispatcher
	resultRetriever *result.Retriever
	pool            *tunny.Pool

	done chan struct{}
}

var _ crawler.FileCrawler = (*crawlerImplementation)(nil)

func NewCrawlerImplementation(c *Config) crawler.FileCrawler {
	return &crawlerImplementation{
		Crawler:    c.Crawler,
		dispatcher: c.Dispatcher,
		done:       make(chan struct{}),
	}
}

func (ci *crawlerImplementation) Start() {
	for {
		select {
		case msg := <-ci.dispatcher.Stream(dispatcher.DirectoryJobType):
			go ci.handleDirectory(msg.Payload)
		case msg := <-ci.dispatcher.Stream(dispatcher.FileJobType):
			go ci.handleFile(msg.Payload)
		case <-ci.done:
			return
		}
	}
}

func (ci *crawlerImplementation) handleDirectory(payload dispatcher.JobPayload) {

}

func (ci *crawlerImplementation) handleFile(payload dispatcher.JobPayload) {

}

func (ci *crawlerImplementation) Stop() {

}
