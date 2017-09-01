package collector

import (
	"fmt"
	"strings"
	"testing"

	"fullerite/metric"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	names := []string{"Test", "Diamond", "Fullerite", "ProcStatus", "ProcStatus Instance2"}
	for _, name := range names {
		c := New(name)
		name = strings.Split(name, " ")[0]
		assert.NotNil(t, c, "should create a Collector for "+name)
		assert.Equal(t, name, c.Name())
		assert.Equal(
			t,
			DefaultCollectionInterval,
			c.Interval(),
			"should be the default collection interval for "+name,
		)
		assert.Equal(
			t,
			name+"Collector",
			fmt.Sprintf("%s", c),
			"String() should append Collector to the name for "+name,
		)

		c.SetInterval(999)
		assert.Equal(t, 999, c.Interval(), "should have set the interval")
	}
}

func TestNewInvalidCollector(t *testing.T) {
	c := New("INVALID COLLECTOR")
	assert.Nil(t, c, "should not create a Collector")
}

func TestRemoveBlacklistedDimensions(t *testing.T) {
	c := make(map[string]interface{})
	c["dimensions_blacklist"] = map[string]string{"rollup": "p9[0-9]+"}
	col := New("Test")
	col.Configure(c)

	// Remove p95 rollup
	m := metric.Metric{Name: "test_gauge", MetricType: "gauge", Value: 10, Dimensions: map[string]string{"rollup": "p95"}}
	result := col.ContainsBlacklistedDimension(m.Dimensions)
	assert.True(t, result)

	// Accept p50 rollup
	m = metric.Metric{Name: "test_gauge", MetricType: "gauge", Value: 10, Dimensions: map[string]string{"rollup": "p50"}}
	result = col.ContainsBlacklistedDimension(m.Dimensions)
	assert.False(t, result)

	// Dimension set is empty
	m = metric.Metric{Name: "test_gauge", MetricType: "gauge", Value: 10}
	assert.Equal(t, len(m.Dimensions), 0)
	result = col.ContainsBlacklistedDimension(m.Dimensions)
	assert.False(t, result)
}

func TestDimensionsBlacklistNotSet(t *testing.T) {
	col := New("Test")
	m := metric.Metric{Name: "test_gauge", MetricType: "gauge", Value: 10, Dimensions: map[string]string{"rollup": "p95"}}
	result := col.ContainsBlacklistedDimension(m.Dimensions)
	assert.False(t, result)
}
