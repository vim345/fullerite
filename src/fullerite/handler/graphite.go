package handler

import (
	"fullerite/metric"
)

// Graphite type
type Graphite struct {
	server        string
	interval      int
	maxBufferSize int
	channel       chan metric.Metric
}

// Send metrics in the channel to the graphite server.
func (g Graphite) Send() {
	// TODO: implement
}

// Name of the handler.
func (g Graphite) Name() string {
	return "Graphite"
}

// Interval returns the flush rate. Once the interval time is reached
// Send will get called no matter how full is the channel.
func (g Graphite) Interval() int {
	return g.interval
}

// MaxBufferSize defines the maximum size of the metric channel.
func (g Graphite) MaxBufferSize() int {
	return g.maxBufferSize
}

// Channel returns the internal metric channel. fullerite writes
// metrics to this channel. When Send() is called metrics will be
// flushed.
func (g Graphite) Channel() chan metric.Metric {
	return g.channel
}

// String returns the handler name in a printable format.
func (g Graphite) String() string {
	return g.Name() + "Handler"
}
