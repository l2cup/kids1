package dispatcher

import "time"

type JobType string
type JobPayload = interface{}

const (
	WebJobType         JobType = "WEB_JOB_TYPE"
	FileJobType        JobType = "FILE_JOB_TYPE"
	DirectoryJobType   JobType = "DIRECTORY_JOB_TYPE"
	UpdateCacheJobType JobType = "DIRECTORY_UPDATE_CACHE"
)

type Job struct {
	Type    JobType
	Payload JobPayload
}

type DirectoryCrawlerPayload struct {
	CorpusName string
	Path       string
	Size       int64
}

type FileCrawlerPayload struct {
	CorpusName string
	Path       string
	Size       int64
}

type WebCrawlerPayload struct {
	CorpusName string
	HopCount   int
	URL        string
}

type UpdateCachePayload struct {
	CorpusName string
	Time       time.Time
}
