package config

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/Sirupsen/logrus"
)

var log = logrus.WithFields(logrus.Fields{"app": "fullerite", "pkg": "config"})

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

// ReadConfig reads a fullerite configuration file
func ReadConfig(configFile string) (c Config, e error) {
	log.Info("Reading configuration file at ", configFile)
	contents, e := ioutil.ReadFile(configFile)
	if e != nil {
		log.Error("Config file error: ", e)
		return c, e
	}
	err := json.Unmarshal(contents, &c)
	if err != nil {
    		log.Error("Invalid JSON in config: ", err)
		return c, err
	}
	return c, nil
}

// GetAsFloat parses a string to a float or returns the float if float is passed in
func GetAsFloat(value interface{}) (result float64, err error) {
	err = nil

	switch value.(type) {
	case string:
		result, err = strconv.ParseFloat(value.(string), 64)
	case float64:
		result = value.(float64)
	}

	return
}

// GetAsInt parses a string/float to an int or returns the int if int is passed in
func GetAsInt(value interface{}) (result int, err error) {
	err = nil
	var somethingelse int64
	switch value.(type) {
	case string:
		somethingelse, err = strconv.ParseInt(value.(string), 10, 64)
	case int64:
		somethingelse = value.(int64)
	case float64:
		somethingelse = int64(value.(float64))
	}
	result = int(somethingelse)
	return
}
