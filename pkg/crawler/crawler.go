package crawler

import (
	"github.com/l2cup/kids1/pkg/errors"
	"github.com/l2cup/kids1/pkg/log"
)

type Runner interface {
	Start()
	Stop()
}

type DirCrawler interface {
	Runner
	AddDirectoryPath(path string) errors.Error
}

type FileCrawler interface {
	Runner
}

type WebCrawler interface {
	Runner
}

type Crawler struct {
	Logger *log.Logger
}

func New(logger *log.Logger) *Crawler {
	return &Crawler{
		Logger: logger,
	}
}
