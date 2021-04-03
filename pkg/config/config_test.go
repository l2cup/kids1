package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var systemConfig = &SystemConfig{
	Prefix:                "corpus_",
	DirCrawlerSleepTimeMS: 1000,
	URLRefreshTimeMS:      1000000,
	HopCount:              10,
	FileScanningSizeLimit: 12345,
	Keywords:              []string{"test1", "test2", "test3"},
}

func TestJsonConfig(t *testing.T) {
	sc, err := loadJsonFile("test/config.json")
	assert.NoError(t, err)

	assert.Equal(t, systemConfig.Prefix, sc.Prefix)
	assert.Equal(t, systemConfig.DirCrawlerSleepTimeMS, sc.DirCrawlerSleepTimeMS)
	assert.Equal(t, systemConfig.URLRefreshTimeMS, sc.URLRefreshTimeMS)
	assert.Equal(t, systemConfig.HopCount, sc.HopCount)
	assert.Equal(t, systemConfig.FileScanningSizeLimit, sc.FileScanningSizeLimit)
	assert.Equal(t, systemConfig.Keywords, sc.Keywords)
}

func TestPropertiesConfig(t *testing.T) {
	sc, err := loadPropertiesFile("test/config.properties")
	assert.NoError(t, err)

	assert.Equal(t, systemConfig.Prefix, sc.Prefix)
	assert.Equal(t, systemConfig.DirCrawlerSleepTimeMS, sc.DirCrawlerSleepTimeMS)
	assert.Equal(t, systemConfig.URLRefreshTimeMS, sc.URLRefreshTimeMS)
	assert.Equal(t, systemConfig.HopCount, sc.HopCount)
	assert.Equal(t, systemConfig.FileScanningSizeLimit, sc.FileScanningSizeLimit)
	assert.Equal(t, systemConfig.Keywords, sc.Keywords)
}
