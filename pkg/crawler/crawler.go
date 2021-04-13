package crawler

import (
	"github.com/l2cup/kids1/pkg/errors"
	"github.com/l2cup/kids1/pkg/log"
	"github.com/l2cup/kids1/pkg/runner"
)

type DirCrawler interface {
	runner.Runner
	AddDirectoryPath(path string) errors.Error
}

type FileCrawler interface {
	runner.Runner
}

type WebCrawler interface {
	runner.Runner
	AddWebPage(url string)
}

type Crawler struct {
	Logger *log.Logger
}

func New(logger *log.Logger) *Crawler {
	return &Crawler{
		Logger: logger,
	}
}
