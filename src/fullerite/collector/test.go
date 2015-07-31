package collector

import (
	"fullerite/metric"
	"math/rand"
)

// Test collector type
type Test struct {
	BaseCollector
	metricName string
}

// NewTest creates a new Test collector.
func NewTest() *Test {
	t := new(Test)
	t.name = "Test"
	t.channel = make(chan metric.Metric)
	return t
}

// Configure : this takes a dictionary of values with which the handler can configure itself
func (t *Test) Configure(config *map[string]string) {
	var exists bool
	t.metricName, exists = (*config)["metricName"]
	if !exists {
		t.metricName = "TestMetric"
	}
}

// Collect produces some random test metrics.
func (t Test) Collect() {
	metric := metric.New(t.metricName)
	metric.Value = rand.Float64()
	metric.AddDimension("testing", "yes")
	t.Channel() <- metric
}
