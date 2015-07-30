package collector

import (
	"fullerite/metric"
	"math/rand"
)

// Test collector type
type Test struct {
	interval int
	channel  chan metric.Metric
}

// NewTest creates a new Test collector.
func NewTest() *Test {
	t := new(Test)
	t.channel = make(chan metric.Metric)
	return t
}

// Collect produces some random test metrics.
func (t Test) Collect() {
	metric := metric.New("TestMetric")
	metric.Value = rand.Float64()
	metric.AddDimension("testing", "yes")
	t.Channel() <- metric
}

// Name of the collector.
func (t Test) Name() string {
	return "Test"
}

// Interval returns the collect rate of the collector.
func (t Test) Interval() int {
	return t.interval
}

// Channel returns the internal metrics channel. fullerite reads from
// this channel to pass metrics to the handlers.
func (t Test) Channel() chan metric.Metric {
	return t.channel
}

// String returns the collector name in printable format.
func (t Test) String() string {
	return t.Name() + "Collector"
}

// SetInterval sets the collect rate of the collector.
func (t *Test) SetInterval(interval int) {
	t.interval = interval
}
