package collector

import (
	"fullerite/metric"
)

// CPU collector type.
type CPU struct {
	interval int
	channel  chan metric.Metric
}

// Collect CPU metrics.
func (c CPU) Collect() {
	// TODO: implement
}

// Name of the collector.
func (c CPU) Name() string {
	return "CPU"
}

// Interval returns the collect rate of the collector.
func (c CPU) Interval() int {
	return c.interval
}

// Channel returns the internal metrics channel. fullerite reads from
// this channel to pass metrics to the handlers.
func (c CPU) Channel() chan metric.Metric {
	return c.channel
}

// String returns the collector name in a printable format.
func (c CPU) String() string {
	return c.Name() + "Collector"
}
