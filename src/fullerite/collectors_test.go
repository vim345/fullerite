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
	assert.Equal(t, len(collectors), 0)
}

func TestStartCollectorUnknownCollector(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)

	c := make(map[string]interface{})
	collector := startCollector("unknown collector", config.Config{}, c)
	assert.Equal(t, collector, nil)
}
