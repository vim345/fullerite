package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	Collectors []string `json:"collectors"`
	Handlers   []string `json:"handlers"`
}

func readConfig(config_file string) (c Config) {
	log.Println("Reading configuration file at", config_file)
	config_contents, e := ioutil.ReadFile(config_file)
	if e != nil {
		log.Fatal("Config file error: %v\n", e)
	}
	json.Unmarshal(config_contents, &c)
	return c
}
