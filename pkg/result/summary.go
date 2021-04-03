package result

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/l2cup/kids1/pkg/dispatcher"
)

type Summaries map[string]*Summary

type Summary struct {
	wg      sync.WaitGroup
	counter int64

	mutex   sync.Mutex
	results map[string]int64
	ttl     time.Time
}

type Results struct {
	JobType    dispatcher.JobType
	CorpusName string
	Results    map[string]int64
}

func (s *Summary) GetResults() map[string]int64 {
	s.wg.Wait()
	return s.results
}

func (s *Summary) QueryResults() map[string]int64 {
	if atomic.LoadInt64(&s.counter) == 0 {
		return s.results
	}
	return nil
}

func (s *Summary) AddResults(results map[string]int64) {
	defer s.mutex.Unlock()
	s.mutex.Lock()

	for k, v := range results {
		if existing, ok := s.results[k]; ok {
			s.results[k] = existing + v
		}
		s.results[k] = v
	}

	s.wg.Done()
	atomic.AddInt64(&s.counter, -1)
}
