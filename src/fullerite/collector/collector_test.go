package collector_test

import (
	"fullerite/collector"

	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	names := []string{"Test", "Diamond", "Fullerite", "ProcStatus"}
	for _, name := range names {
		c := collector.New(name)
		assert := assert.New(t)
		assert.NotNil(c, "should create a Collector for "+name)
		assert.Equal(c.Name(), name)
		assert.Equal(
			c.Interval(),
			collector.DefaultCollectionInterval,
			"should be the default collection interval for "+name,
		)
		assert.Equal(
			fmt.Sprintf("%s", c),
			name+"Collector",
			"String() should append Collector to the name for "+name,
		)

		c.SetInterval(999)
		assert.Equal(c.Interval(), 999)
	}
}

func TestNewInvalidCollector(t *testing.T) {
	c := collector.New("INVALID COLLECTOR")
	assert.Nil(t, c, "should not create a Collector")
}
