package app

import (
	golog "log"

	"github.com/l2cup/kids1/pkg/config"
	"github.com/l2cup/kids1/pkg/crawler"
	"github.com/l2cup/kids1/pkg/crawler/dir"
	"github.com/l2cup/kids1/pkg/crawler/file"
	"github.com/l2cup/kids1/pkg/crawler/web"
	"github.com/l2cup/kids1/pkg/dispatcher"
	"github.com/l2cup/kids1/pkg/log"
	"github.com/l2cup/kids1/pkg/result"
)

const (
	EnvLogVerbosity = "LOG_VERBOSITY"
	EnvNotSet       = "ENV_NOT_SET"
)

type App struct {
	Logger           *log.Logger
	Configuration    *config.SystemConfig
	Dispatcher       *dispatcher.Dispatcher
	ResultRetriever  result.Retriever
	DirectoryCrawler crawler.DirCrawler
	FileCrawler      crawler.FileCrawler
	WebCrawler       crawler.WebCrawler
}

func New() *App {
	logger, err := log.NewLogger(&log.Config{
		LogVerbosity: config.GetEnv(EnvLogVerbosity, EnvNotSet),
	})

	if err != nil {
		golog.Fatalf("couldn't initialize logger, err : %s", err)
	}

	syscfg, err := config.LoadConfigFile("./config.properties")
	if err != nil {
		logger.Error("[syscfg]couldn't load properties syscfg, trying json",
			"err", err)
		syscfg, err = config.LoadConfigFile("./config.json")
		if err != nil {
			logger.Fatal("[syscfg]couldn't load json syscfg, exiting.",
				"err", err)
		}
	}

	dispatcher := dispatcher.New(&dispatcher.Config{
		Logger:     logger,
		BufferSize: 50,
	})

	resultRetriever := result.NewRetrieverImplementation(50, logger)

	directoryCrawler := dir.NewCrawlerImplementation(&dir.Config{
		Crawler:     crawler.New(logger),
		Dispatcher:  dispatcher,
		SleepTimeMS: syscfg.DirCrawlerSleepTimeMS,
		Prefix:      syscfg.Prefix,
	})

	fileCrawler := file.NewCrawlerImplementation(&file.Config{
		Crawler:              crawler.New(logger),
		Dispatcher:           dispatcher,
		ResultRetriever:      resultRetriever,
		Keywords:             syscfg.Keywords,
		QueuedFilesSizeLimit: syscfg.FileScanningSizeLimit,
	})

	webCrawler := web.NewCrawlerImplementation(&web.Config{
		Crawler:         crawler.New(logger),
		Dispatcher:      dispatcher,
		ResultRetriever: resultRetriever,
		InitialHopCount: syscfg.HopCount,
		Keywords:        syscfg.Keywords,
		TTLMS:           syscfg.URLRefreshTimeMS,
	})

	return &App{
		Logger:           logger,
		Configuration:    syscfg,
		Dispatcher:       dispatcher,
		ResultRetriever:  resultRetriever,
		DirectoryCrawler: directoryCrawler,
		FileCrawler:      fileCrawler,
		WebCrawler:       webCrawler,
	}
}
