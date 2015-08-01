package main

import (
	"encoding/json"
	"io/ioutil"
)

// Config type holds the global Fullerite configuration.
type Config struct {
	Prefix                string                            `json:"prefix"`
	Interval              int                               `json:"interval"`
	DiamondCollectorsPath string                            `json:"diamond_collectors_path"`
	DiamondCollectors     map[string]map[string]interface{} `json:"diamond_collectors"`
	Handlers              map[string]map[string]interface{} `json:"handlers"`
	Collectors            map[string]map[string]interface{} `json:"collectors"`
	DefaultDimensions     map[string]string                 `json:"defaultDimensions"`
}

func readConfig(configFile string) (c Config, e error) {
	log.Info("Reading configuration file at ", configFile)
	contents, e := ioutil.ReadFile(configFile)
	if e != nil {
		log.Error("Config file error: ", e)
		return c, e
	}
	json.Unmarshal(contents, &c)
	return c, nil
}
