package collector_test

import (
	"fullerite/collector"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFulleriteConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	f := collector.NewFullerite()
	f.Configure(config)

	assert.Equal(t,
		f.Interval(),
		collector.DefaultCollectionInterval,
		"should be the default collection interval",
	)
}

func TestFulleriteConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999
	f := collector.NewFullerite()
	f.Configure(config)

	assert.Equal(t,
		f.Interval(),
		9999,
		"should be the defined interval",
	)
}

func TestFulleriteCollect(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999

	f := collector.NewFullerite()
	f.Configure(config)

	go f.Collect()

	select {
	case <-f.Channel():
		return
	case <-time.After(2 * time.Second):
		t.Fail()
	}
}
