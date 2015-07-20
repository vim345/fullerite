package handler

import (
	"fullerite/metric"
)

// SignalFx type.
type SignalFx struct {
	api           string
	interval      int
	maxBufferSize int
	channel       chan metric.Metric
}

// Send metrics in the channel to SignalFx.
func (s SignalFx) Send() {
	// TODO: implement
}

// Name of the handler.
func (s SignalFx) Name() string {
	return "SignalFx"
}

// Interval returns the flush rate. Once the interval time is reached
// Send will get called no matter how full is the channel.
func (s SignalFx) Interval() int {
	return s.interval
}

// MaxBufferSize defines the maximum size of the metric channel.
func (s SignalFx) MaxBufferSize() int {
	return s.maxBufferSize
}

// Channel returns the internal metric channel. fullerite writes
// metrics to this channel. When Send() is called metrics will be
// flushed.
func (s SignalFx) Channel() chan metric.Metric {
	return s.channel
}

// String returns the handler name in a printable format.
func (s SignalFx) String() string {
	return s.Name() + "Handler"
}
