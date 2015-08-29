package handler_test

import (
	"fullerite/handler"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignalfxConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	s := handler.NewSignalFx()
	s.Configure(config)

	assert.Equal(t,
		s.Interval(),
		handler.DefaultInterval,
		"should be the default interval",
	)
}

func TestSignalfxConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = "10"
	config["timeout"] = "10"
	config["max_buffer_size"] = "100"
	config["endpoint"] = "signalfx.server"

	s := handler.NewSignalFx()
	s.Configure(config)

	assert := assert.New(t)
	assert.Equal(
		s.Interval(),
		10,
		"should be the set value",
	)
	assert.Equal(
		s.MaxBufferSize(),
		100,
		"should be the set value",
	)
	assert.Equal(
		s.Endpoint(),
		config["endpoint"],
		"should be the set value",
	)
}
