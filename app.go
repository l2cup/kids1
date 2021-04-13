package kids1

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
	"github.com/l2cup/kids1/pkg/runner"
)

const (
	EnvLogVerbosity         = "LOG_VERBOSITY"
	EnvPropertiesConfigPath = "PROP_CONFIG_PATH"
	EnvJsonCofigPath        = "JSON_CONFIG_PATH"
	EnvNotSet               = "ENV_NOT_SET"

	ConfigPathDefaultProp = "./config.properties"
	ConfigPathDefaultJson = "./config.json"
)

var _ runner.Runner = (*App)(nil)
var _ runner.Registrator = (*App)(nil)

type App struct {
	Logger           *log.Logger
	Configuration    *config.SystemConfig
	Dispatcher       *dispatcher.Dispatcher
	ResultRetriever  result.Retriever
	DirectoryCrawler crawler.DirCrawler
	FileCrawler      crawler.FileCrawler
	WebCrawler       crawler.WebCrawler

	runners []runner.Runner
}

func New() *App {
	logger, err := log.NewLogger(&log.Config{
		//LogVerbosity: config.GetEnv(EnvLogVerbosity, EnvNotSet),
		LogVerbosity: log.DebugVerbosity,
	})

	if err != nil {
		golog.Fatalf("couldn't initialize logger, err : %s", err)
	}

	syscfg, err := config.LoadConfigFile(config.GetEnv(EnvPropertiesConfigPath, ConfigPathDefaultProp))
	if err != nil {
		logger.Error("[syscfg]couldn't load properties syscfg, trying json",
			"err", err)
		syscfg, err = config.LoadConfigFile(config.GetEnv(EnvJsonCofigPath, ConfigPathDefaultJson))
		if err != nil {
			logger.Error("[syscfg]couldn't load json syscfg, using defaults.",
				"err", err)
			syscfg = &config.SystemConfig{
				Keywords:              []string{"one", "two", "three", "Core"},
				Prefix:                "corpus_",
				URLRefreshTimeMS:      86400000,
				FileScanningSizeLimit: 1048576,
				DirCrawlerSleepTimeMS: 1000,
				HopCount:              1,
			}
		}
	}

	dispatcher := dispatcher.New(&dispatcher.Config{
		Logger:     logger,
		BufferSize: 50,
	})

	app := &App{
		Logger:        logger,
		Configuration: syscfg,
		Dispatcher:    dispatcher,
		runners:       make([]runner.Runner, 0),
	}

	app.ResultRetriever = result.NewRetrieverImplementation(50, logger, app)

	app.DirectoryCrawler = dir.NewCrawlerImplementation(&dir.Config{
		Crawler:           crawler.New(logger),
		Dispatcher:        dispatcher,
		SleepTimeMS:       syscfg.DirCrawlerSleepTimeMS,
		Prefix:            syscfg.Prefix,
		RunnerRegistrator: app,
	})

	app.FileCrawler = file.NewCrawlerImplementation(&file.Config{
		Crawler:              crawler.New(logger),
		Dispatcher:           dispatcher,
		ResultRetriever:      app.ResultRetriever,
		Keywords:             syscfg.Keywords,
		QueuedFilesSizeLimit: syscfg.FileScanningSizeLimit,
		RunnerRegistrator:    app,
	})

	app.WebCrawler = web.NewCrawlerImplementation(&web.Config{
		Crawler:           crawler.New(logger),
		Dispatcher:        dispatcher,
		ResultRetriever:   app.ResultRetriever,
		InitialHopCount:   syscfg.HopCount,
		Keywords:          syscfg.Keywords,
		TTLMS:             syscfg.URLRefreshTimeMS,
		RunnerRegistrator: app,
	})

	return app
}

func (a *App) Stop() {
	for _, r := range a.runners {
		r.Stop()
	}
}

func (a *App) Start() {
	for _, r := range a.runners {
		go r.Start()
	}
}

func (a *App) Register(r runner.Runner) {
	a.runners = append(a.runners, r)
}
