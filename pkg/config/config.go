package config

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
)

type AppConfigProperties map[string]interface{}

type SystemConfig struct {
	Prefix                string   `properties:"file_corpus_prefix" json:"file_corpus_prefix"`
	DirCrawlerSleepTimeMS uint64   `properties:"dir_crawler_sleep_time" json:"dir_crawler_sleep_time"`
	URLRefreshTimeMS      uint64   `properties:"url_refresh_time" json:"url_refresh_time"`
	FileScanningSizeLimit uint64   `properties:"file_scanning_size_limit" json:"file_scanning_size_limit"`
	HopCount              int      `properties:"hop_count" json:"hop_count"`
	Keywords              []string `json:"keywords"`
}

func LoadEnvFile(path string) error {

	if path == "" {
		path = ".env"
	}

	return godotenv.Load(path)
}

func GetEnv(key, fallback string) string {

	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func LoadConfigFile(path string) (*SystemConfig, error) {
	switch filepath.Ext(path) {
	case ".prop", ".properties":
		return loadPropertiesFile(path)
	case ".json":
		return loadJsonFile(path)
	default:
		return nil, errors.New("unsuported extension")
	}
}

func loadPropertiesFile(path string) (*SystemConfig, error) {
	sc := &SystemConfig{}

	properties, err := readPropertiesFile(path)
	if err != nil {
		return nil, err
	}

	keywords, ok := properties["keywords"]
	if !ok {
		return nil, errors.New("no keywords given in config file")
	}

	keywordsStr, ok := keywords.(string)
	if !ok {
		return nil, errors.New("keywords not of type string")
	}

	err = unmarshal(properties, sc)
	if err != nil {
		return nil, err
	}

	keywordsArr := strings.Split(keywordsStr, ",")
	sc.Keywords = keywordsArr

	return sc, nil
}

func loadJsonFile(path string) (*SystemConfig, error) {
	if len(path) == 0 {
		return nil, errors.New("no path given")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't read json config: ")
	}

	sc := &SystemConfig{}
	err = json.Unmarshal(data, sc)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't unmarshal json config: ")
	}

	return sc, nil
}

func readPropertiesFile(path string) (AppConfigProperties, error) {
	config := AppConfigProperties{}

	if len(path) == 0 {
		return config, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				config[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func unmarshal(input AppConfigProperties, output interface{}) error {
	decoderConfig := &mapstructure.DecoderConfig{
		TagName:          "properties",
		Result:           output,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return err
	}

	return decoder.Decode(map[string]interface{}(input))
}
