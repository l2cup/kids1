package web

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/gocolly/colly/v2"
	"github.com/l2cup/kids1/pkg/crawler"
	"github.com/l2cup/kids1/pkg/dispatcher"
	"github.com/l2cup/kids1/pkg/result"
	"github.com/l2cup/kids1/pkg/runner"
)

type Config struct {
	RunnerRegistrator runner.Registrator
	Crawler           *crawler.Crawler
	Dispatcher        *dispatcher.Dispatcher
	ResultRetriever   result.Retriever
	InitialHopCount   int
	Keywords          []string
	TTLMS             uint64
}

var _ crawler.WebCrawler = (*crawlerImplementation)(nil)
var _ runner.Runner = (*crawlerImplementation)(nil)

type crawlerImplementation struct {
	*crawler.Crawler
	dispatcher      *dispatcher.Dispatcher
	resultRetriever result.Retriever
	pool            *tunny.Pool
	initialHopCount int
	keywords        []string
	done            chan struct{}
	ttl             time.Duration
}

func NewCrawlerImplementation(c *Config) crawler.WebCrawler {
	ttl, err := time.ParseDuration(fmt.Sprintf("%dms", c.TTLMS))
	if err != nil {
		c.Crawler.Logger.Fatal("couldn't parse web page ttl time duration", "err", err, "duration", c.TTLMS)
	}

	ci := &crawlerImplementation{
		Crawler:         c.Crawler,
		dispatcher:      c.Dispatcher,
		resultRetriever: c.ResultRetriever,
		initialHopCount: c.InitialHopCount,
		keywords:        c.Keywords,
		done:            make(chan struct{}),
		ttl:             ttl,
	}

	c.RunnerRegistrator.Register(ci)
	ci.pool = tunny.NewFunc(200, ci.crawlPage)
	return ci
}

func (ci *crawlerImplementation) AddWebPage(url string) {
	//ci.resultRetriever.InitializeSummary(
	//	dispatcher.WebJobType, url, 1, time.Now().Add(ci.ttl))

	ci.dispatcher.Push(&dispatcher.Job{
		Type: dispatcher.WebJobType,
		Payload: &dispatcher.WebCrawlerPayload{
			CorpusName: url,
			HopCount:   ci.initialHopCount,
			URL:        url,
		},
	})
}

func (ci *crawlerImplementation) Start() {
	for {
		select {
		case job := <-ci.dispatcher.Stream(dispatcher.WebJobType):
			go ci.startJob(job.Payload)
		case <-ci.done:
			ci.pool.Close()
			return
		}
	}
}

func (ci *crawlerImplementation) Stop() {
	ci.done <- struct{}{}
}

func (ci *crawlerImplementation) startJob(payload dispatcher.JobPayload) {
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

func (ci *crawlerImplementation) crawlPage(payload interface{}) interface{} {
	webPayload, ok := payload.(*dispatcher.WebCrawlerPayload)
	if !ok {
		ci.Logger.Error("payload not of type web crawler payload")
	}

	c := colly.NewCollector()

	if webPayload.HopCount > 0 {
		c.OnHTML("a[href]", ci.onHtml(webPayload.CorpusName, webPayload.HopCount))
	}

	c.OnScraped(ci.onScraped(webPayload.CorpusName, webPayload.HopCount))
	c.IgnoreRobotsTxt = true
	err := c.Visit(webPayload.URL)
	if err != nil {
		ci.Logger.Error("error visiting url", "err", err, "url", webPayload.URL)
		ci.resultRetriever.UpdateSummary(&result.Results{
			JobType:    dispatcher.WebJobType,
			CorpusName: webPayload.URL,
		})
	}

	return nil
}

func (ci *crawlerImplementation) onScraped(jobName string, hopCount int) colly.ScrapedCallback {
	return func(r *colly.Response) {
		if r.StatusCode != http.StatusOK {
			ci.Logger.Error("couldn't scrape web page and it's children",
				"url", r.Request.URL,
				"code", r.StatusCode,
				"hops_left", hopCount)
		}

		results := make(map[string]int64)
		for _, word := range ci.keywords {
			results[word] = 0
		}

		words := strings.Fields(string(r.Body))
		for _, w := range words {
			if result, ok := results[w]; ok {
				results[w] = result + 1
			}
		}

		ci.Logger.Debug("web job finished, updating summary", "results", results)

		ci.resultRetriever.UpdateSummary(&result.Results{
			CorpusName: jobName,
			JobType:    dispatcher.WebJobType,
			Results:    results,
		})
	}
}

func (ci *crawlerImplementation) onHtml(jobName string, hopCount int) colly.HTMLCallback {
	return func(e *colly.HTMLElement) {
		URL := e.Attr("href")

		if strings.HasPrefix(URL, "/") {
			URL = "http://" + e.Request.URL.Host + URL
		}

		if _, err := url.ParseRequestURI(URL); err != nil {
			return
		}

		if e.Request.URL.Scheme != "http" && e.Request.URL.Scheme != "https" {
			return
		}

		payload := &dispatcher.WebCrawlerPayload{
			CorpusName: URL,
			HopCount:   hopCount - 1,
			URL:        URL,
		}

		job := &dispatcher.Job{
			Type:    dispatcher.WebJobType,
			Payload: payload,
		}

		if _, err := ci.resultRetriever.QuerySummary(dispatcher.WebJobType, URL); err != nil {
			ci.dispatcher.Push(job)
			ci.resultRetriever.InitializeSummary(
				dispatcher.WebJobType, URL, 1, time.Now().Add(ci.ttl))
			return
		}
	}
}
