package web

import (
	"net/http"
	"strings"

	"github.com/Jeffail/tunny"
	"github.com/gocolly/colly/v2"
	"github.com/l2cup/kids1/pkg/crawler"
	"github.com/l2cup/kids1/pkg/dispatcher"
	"github.com/l2cup/kids1/pkg/result"
)

type Config struct {
	Crawler         *crawler.Crawler
	Dispatcher      *dispatcher.Dispatcher
	ResultRetriever result.Retriever
	InitialHopCount int
	Keywords        []string
}

var _ crawler.WebCrawler = (*crawlerImplementation)(nil)

type crawlerImplementation struct {
	*crawler.Crawler
	dispatcher      *dispatcher.Dispatcher
	resultRetriever result.Retriever
	pool            *tunny.Pool
	initialHopCount int
	keywords        []string
	done            chan struct{}
}

func NewCrawlerImplementation(c *Config) crawler.WebCrawler {
	ci := &crawlerImplementation{
		Crawler:         c.Crawler,
		dispatcher:      c.Dispatcher,
		resultRetriever: c.ResultRetriever,
		initialHopCount: c.InitialHopCount,
	}

	ci.pool = tunny.NewFunc(200, ci.crawlPage)
	return ci
}

func (ci *crawlerImplementation) AddWebPage(url string, ttl int) {
}

func (ci *crawlerImplementation) onResponse(jobName string, hopCount int) colly.ResponseCallback {
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

		ci.resultRetriever.UpdateSummary(&result.Results{
			CorpusName: jobName,
			JobType:    dispatcher.WebJobType,
			Results:    results,
		})
	}
}

func (ci *crawlerImplementation) onHtml(jobName string, hopCount int) colly.HTMLCallback {
	return func(e *colly.HTMLElement) {
		url := e.Attr("href")
		payload := &dispatcher.WebCrawlerPayload{
			CorpusName: jobName,
			HopCount:   hopCount - 1,
			URL:        url,
		}

		job := &dispatcher.Job{
			Type:    dispatcher.WebJobType,
			Payload: payload,
		}

		err := ci.resultRetriever.IncrementResultCount(dispatcher.WebJobType, jobName)
		if err != nil {
			ci.Logger.Error("couldn't increment result count for web jobs",
				"err", err,
				"job_name", jobName,
			)
		}

		ci.dispatcher.Push(job)
	}
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

}

func (ci *crawlerImplementation) crawlPage(payload interface{}) interface{} {
	webPayload, ok := payload.(dispatcher.WebCrawlerPayload)
	if !ok {
		ci.Logger.Error("payload not of type web crawler payload")
	}

	c := colly.NewCollector()

	if webPayload.HopCount > 0 {
		c.OnHTML("a[href]", ci.onHtml(webPayload.CorpusName, webPayload.HopCount))
	}

	c.OnResponse(ci.onResponse(webPayload.CorpusName, webPayload.HopCount))
	c.IgnoreRobotsTxt = true
	err := c.Visit(webPayload.URL)
	if err != nil {
		ci.Logger.Error("error visiting url", "err", err)
	}

	return nil
}
