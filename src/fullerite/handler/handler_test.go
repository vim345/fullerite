package handler_test

import (
	"fullerite/handler"

	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	names := []string{"Graphite", "Kairos", "SignalFx", "Datadog"}
	for _, name := range names {
		h := handler.New(name)
		assert := assert.New(t)
		assert.NotNil(h, "should create a Handler for "+name)
		assert.NotNil(h.Channel(), "should create a channel")
		assert.Equal(h.Name(), name)
		assert.Equal(h.Prefix(), "", "should be empty string")
		assert.Equal(len(h.DefaultDimensions()), 0, "should be empty")
		assert.Equal(h.MaxBufferSize(),
			handler.DefaultBufferSize,
			"should be the default buffer size",
		)
		assert.Equal(
			h.Interval(),
			handler.DefaultInterval,
			"should be the default interval for "+name,
		)
		assert.Equal(
			fmt.Sprintf("%s", h),
			name+"Handler",
			"String() should append Handler to the name for "+name,
		)

		// Test Set* functions
		h.SetInterval(999)
		assert.Equal(h.Interval(), 999)

		h.SetMaxBufferSize(999)
		assert.Equal(h.MaxBufferSize(), 999)

		dims := make(map[string]string)
		dims["test"] = "test value"
		h.SetDefaultDimensions(dims)
		assert.Equal(len(h.DefaultDimensions()), 1)
	}
}
