package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// Config type holds the global Fullerite configuration.
type Config struct {
	Collectors []string `json:"collectors"`
	Handlers   []string `json:"handlers"`
}

func readConfig(configFile string) (c Config) {
	log.Println("Reading configuration file at", configFile)
	contents, e := ioutil.ReadFile(configFile)
	if e != nil {
		log.Fatal("Config file error:", e)
	}
	json.Unmarshal(contents, &c)
	return c
}
