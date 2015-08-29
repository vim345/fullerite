package handler_test

import (
	"fullerite/handler"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKairosConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	k := handler.NewKairos()
	k.Configure(config)

	assert.Equal(t,
		k.Interval(),
		handler.DefaultInterval,
		"should be the default interval",
	)
}

func TestKairosConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = "10"
	config["timeout"] = "10"
	config["max_buffer_size"] = "100"
	config["server"] = "kairos.server"
	config["port"] = "10101"

	k := handler.NewKairos()
	k.Configure(config)

	assert := assert.New(t)
	assert.Equal(
		k.Interval(),
		10,
		"should be the set value",
	)
	assert.Equal(
		k.MaxBufferSize(),
		100,
		"should be the set value",
	)
	assert.Equal(
		k.Server(),
		config["server"],
		"should be the set value",
	)
	assert.Equal(
		k.Port(),
		config["port"],
		"should be the set value",
	)
}
