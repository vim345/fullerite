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
		assert.Equal(h.Name(), name)
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
	}
}
