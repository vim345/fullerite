package main

import (
	"encoding/json"
	"io/ioutil"
)

// Config type holds the global Fullerite configuration.
type Config struct {
	Prefix                string                       `json:"prefix"`
	Interval              int                          `json:"interval"`
	DiamondCollectorsPath string                       `json:"diamond_collectors_path"`
	DiamondCollectors     map[string]map[string]string `json:"diamond_collectors"`
	Handlers              map[string]map[string]string `json:"handlers"`
	Collectors            map[string]map[string]string `json:"collectors"`
	DefaultDimensions     map[string]string            `json:"defaultDimensions"`
}

func readConfig(configFile string) (c Config) {
	log.Info("Reading configuration file at ", configFile)
	contents, e := ioutil.ReadFile(configFile)
	if e != nil {
		log.Fatal("Config file error: ", e)
	}
	json.Unmarshal(contents, &c)
	return c
}
