package main

import (
	"fullerite/config"

	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestStartHandlersEmptyConfig(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)

	handlers := startHandlers(config.Config{})
	assert.Equal(t, len(handlers), 0)
}

func TestStartHandlerUnknownHandler(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)

	c := make(map[string]interface{})
	handler := startHandler("unknown handler", config.Config{}, c)
	assert.Equal(t, handler, nil)
}
