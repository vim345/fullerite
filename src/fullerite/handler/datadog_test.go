package handler_test

import (
	"fullerite/handler"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatadogConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	d := handler.NewDatadog()
	d.Configure(config)

	assert.Equal(t,
		d.Interval(),
		handler.DefaultInterval,
		"should be the default interval",
	)
}

func TestDatadogConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = "10"
	config["timeout"] = "10"
	config["max_buffer_size"] = "100"
	config["endpoint"] = "datadog.server"

	d := handler.NewDatadog()
	d.Configure(config)

	assert := assert.New(t)
	assert.Equal(
		d.Interval(),
		10,
		"should be the set value",
	)
	assert.Equal(
		d.MaxBufferSize(),
		100,
		"should be the set value",
	)
	assert.Equal(
		d.Endpoint(),
		config["endpoint"],
		"should be the set value",
	)
}
