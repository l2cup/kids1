package dispatcher

import (
	"github.com/l2cup/kids1/pkg/log"
	"github.com/orcaman/concurrent-map"
)

type JobType string
type JobPayload = interface{}

const (
	WebJobType       JobType = "WEB_JOB_TYPE"
	FileJobType      JobType = "FILE_JOB_TYPE"
	DirectoryJobType JobType = "DIRECTORY_JOB_TYPE"
)

type Job struct {
	Type    JobType
	Payload JobPayload
}

type Config struct {
	Logger     log.Logger
	BufferSize int
}

type Dispatcher struct {
	logger      log.Logger
	dispatchMap cmap.ConcurrentMap
	bufferSize  int
}

func New(c *Config) *Dispatcher {
	return &Dispatcher{
		logger:      c.Logger,
		bufferSize:  c.BufferSize,
		dispatchMap: cmap.New(),
	}
}

func (d *Dispatcher) Push(job *Job) {
	ich, ok := d.dispatchMap.Get(string(job.Type))
	if !ok {
		d.logger.Info("[dispatcher] push registered new job type", "type", job.Type)
		ich = make(chan *Job, d.bufferSize)
		d.dispatchMap.Set(string(job.Type), ich)
	}

	ch, ok := ich.(chan *Job)
	if !ok {
		d.logger.Fatal("[fatal] push couldn't cast channel to job channel")
	}

	ch <- job
}

func (d *Dispatcher) Stream(jobType JobType) <-chan *Job {
	ich, ok := d.dispatchMap.Get(string(jobType))
	if !ok {
		d.logger.Info("[dispatcher] stream registered new job type", "type", jobType)
		ich = make(chan *Job, d.bufferSize)
		d.dispatchMap.Set(string(jobType), ich)
	}

	ch, ok := ich.(chan *Job)
	if !ok {
		d.logger.Fatal("[fatal] push couldn't cast channel to job channel")
	}

	return ch
}

func (d *Dispatcher) Pop(jobType JobType) *Job {
	ich, ok := d.dispatchMap.Get(string(jobType))
	if !ok {
		d.logger.Info("[dispatcher] pop registered new job type", "type", jobType)
		ich = make(chan *Job, d.bufferSize)
		d.dispatchMap.Set(string(jobType), ich)
	}

	ch, ok := ich.(chan *Job)
	if !ok {
		d.logger.Fatal("[fatal] push couldn't cast channel to job channel")
	}
	return <-ch
}
