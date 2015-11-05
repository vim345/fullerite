package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockerStatsConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})

	d := NewDockerStats(nil, 123, nil)
	d.Configure(config)

	assert.Equal(t, 123, d.Interval())
}

func TestDockerStatsConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999

	d := NewDockerStats(nil, 123, nil)
	d.Configure(config)

	assert.Equal(t, 9999, d.Interval())
}
