package main

import (
	"fullerite/config"

	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestStartCollectorsEmptyConfig(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	collectors := startCollectors(config.Config{})

	assert.NotEqual(t, len(collectors), 1, "should create a Collector")
}

func TestStartCollectorUnknownCollector(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	c := make(map[string]interface{})
	collector := startCollector("unknown collector", config.Config{}, c)

	assert.Nil(t, collector, "should NOT create a Collector")
}
