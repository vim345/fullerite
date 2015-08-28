package collector_test

import (
	"fullerite/collector"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTestConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	test := collector.NewTest()
	test.Configure(config)

	assert.Equal(t,
		test.Interval(),
		collector.DefaultCollectionInterval,
		"should be the default collection interval",
	)
}

func TestTestConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999
	test := collector.NewTest()
	test.Configure(config)

	assert.Equal(t,
		test.Interval(),
		9999,
		"should be the defined interval",
	)
}

func TestTestCollect(t *testing.T) {
	config := make(map[string]interface{})
	test := collector.NewTest()
	test.Configure(config)

	go test.Collect()

	select {
	case <-test.Channel():
		return
	case <-time.After(2 * time.Second):
		t.Fail()
	}
}
