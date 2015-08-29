package handler_test

import (
	"fullerite/handler"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphiteConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	g := handler.NewGraphite()
	g.Configure(config)

	assert.Equal(t,
		g.Interval(),
		handler.DefaultInterval,
		"should be the default interval",
	)
}

func TestGraphiteConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = "10"
	config["timeout"] = "10"
	config["max_buffer_size"] = "100"
	config["server"] = "test_server"
	config["port"] = "10101"

	g := handler.NewGraphite()
	g.Configure(config)

	assert := assert.New(t)
	assert.Equal(
		g.Interval(),
		10,
		"should be the set value",
	)
	assert.Equal(
		g.MaxBufferSize(),
		100,
		"should be the set value",
	)
	assert.Equal(
		g.Server(),
		config["server"],
		"should be the set value",
	)
	assert.Equal(
		g.Port(),
		config["port"],
		"should be the set value",
	)
}
