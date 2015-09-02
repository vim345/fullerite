package collector_test

import (
	"fullerite/collector"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcStatusConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	f := collector.NewProcStatus()
	f.Configure(config)

	assert.Equal(t,
		f.Interval(),
		collector.DefaultCollectionInterval,
		"should be the default collection interval",
	)
	assert.Equal(t,
		f.ProcessName(),
		"",
		"should be the default process name",
	)
}

func TestProcStatusConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999
	config["processName"] = "fullerite"
	f := collector.NewProcStatus()
	f.Configure(config)

	assert.Equal(t,
		f.Interval(),
		9999,
		"should be the defined interval",
	)
	assert.Equal(t,
		f.ProcessName(),
		"fullerite",
		"should be the defined process name",
	)
}

func TestProcStatusCollect(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999

	f := collector.NewProcStatus()
	f.Configure(config)

	go f.Collect()

	select {
	case <-f.Channel():
		return
	case <-time.After(2 * time.Second):
		t.Fail()
	}
}
