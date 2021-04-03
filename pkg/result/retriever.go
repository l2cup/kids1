package result

import (
	"github.com/pkg/errors"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/l2cup/kids1/pkg/dispatcher"
	"github.com/l2cup/kids1/pkg/log"
	cmap "github.com/orcaman/concurrent-map"
)

type Retriever interface {
	Start()
	Stop()
	InitializeSummary(jobType dispatcher.JobType, corpusName string, jobs int, ttl time.Time)
	GetSummary(jobType dispatcher.JobType, corpusName string) (map[string]int64, error)
	QuerySummary(jobType dispatcher.JobType, corpusName string) (map[string]int64, error)
	UpdateSummary(results *Results)
}

type retrieverImplementation struct {
	logger       *log.Logger
	summariesMap cmap.ConcurrentMap
	resultsChan  chan *Results
	pool         *tunny.Pool

	done chan struct{}
}

func NewRetrieverImplementation(bufferSize int, logger *log.Logger) Retriever {
	summariesMap := cmap.New()
	summariesMap.Set(string(dispatcher.FileJobType), cmap.New())
	summariesMap.Set(string(dispatcher.WebJobType), cmap.New())

	ri := &retrieverImplementation{
		logger:       logger,
		summariesMap: summariesMap,
		resultsChan:  make(chan *Results, bufferSize),
		done:         make(chan struct{}),
	}

	pool := tunny.NewFunc(bufferSize, ri.poolAddResults)
	ri.pool = pool

	return ri
}

func (ri *retrieverImplementation) Start() {
	for {
		select {
		case results := <-ri.resultsChan:
			go ri.addResults(results)
		case <-ri.done:
			ri.pool.Close()
			return
		}
	}
}

func (ri *retrieverImplementation) Stop() {
	ri.done <- struct{}{}
}

func (ri *retrieverImplementation) InitializeSummary(
	jobType dispatcher.JobType,
	corpusName string,
	jobs int,
	ttl time.Time,
) {
	summary := &Summary{
		wg:      sync.WaitGroup{},
		counter: int64(jobs),
		mutex:   sync.Mutex{},
		results: make(map[string]int64),
		ttl:     ttl,
	}

	summary.wg.Add(jobs)

	summaries, ok := ri.summariesMap.Get(string(jobType))
	if !ok {
		ri.logger.Error("summaries map for job type doesn't exist")
		return
	}

	summariesMap, ok := summaries.(cmap.ConcurrentMap)
	if !ok {
		ri.logger.Fatal("couldn't cast summaries to concurrent map")
	}

	summariesMap.Set(corpusName, summary)
}

func (ri *retrieverImplementation) GetSummary(
	summaryType dispatcher.JobType,
	corpusName string,
) (map[string]int64, error) {

	summary, err := ri.getSummary(summaryType, corpusName)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get summary: ")
	}

	if !summary.ttl.IsZero() && summary.ttl.Before(time.Now()) {
		return nil, errors.New("summary expired")
	}

	return summary.GetResults(), nil
}

func (ri *retrieverImplementation) QuerySummary(
	summaryType dispatcher.JobType,
	corpusName string,
) (map[string]int64, error) {

	summary, err := ri.getSummary(summaryType, corpusName)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get summary: ")
	}

	return summary.QueryResults(), nil
}

func (ri *retrieverImplementation) UpdateSummary(results *Results) {
	ri.resultsChan <- results
}

func (ri *retrieverImplementation) addResults(results *Results) {
	if ri.pool.GetSize() == 0 {
		return
	}

	_, err := ri.pool.ProcessTimed(results, 60*time.Second)

	if err == tunny.ErrJobTimedOut {
		ri.logger.Error("goroutine timed out", "err", err)
		return
	}

	if err != nil {
		ri.logger.Error("there was an error while adding results", "err", err)
	}
}

func (ri *retrieverImplementation) poolAddResults(payload interface{}) interface{} {
	results, ok := payload.(Results)
	if !ok {
		return errors.New("couldn't convert payload")
	}

	summary, err := ri.getSummary(results.JobType, results.CorpusName)
	if err != nil {
		return errors.Wrap(err, "couldn't get summary")
	}

	summary.AddResults(results.Results)
	return nil
}

func (ri *retrieverImplementation) getSummary(
	summaryType dispatcher.JobType,
	corpusName string,
) (*Summary, error) {
	summaries, ok := ri.summariesMap.Get(string(summaryType))
	if !ok {
		ri.logger.Error("summaries map for job type doesn't exist")
		return nil, errors.New("summaries map for job type doens't exist")
	}

	summariesMap, ok := summaries.(cmap.ConcurrentMap)
	if !ok {
		ri.logger.Fatal("couldn't cast summaries to concurrent map")
		return nil, errors.New("couldn't cast summaries to concurrent map")
	}

	isummary, ok := summariesMap.Get(corpusName)
	if !ok {
		return nil, errors.New("corpus with that name doesn't exist, couldn't update")
	}

	summary, ok := isummary.(*Summary)
	if !ok {
		return nil, errors.New("map value couldn't be cast as summary")
	}

	return summary, nil
}
