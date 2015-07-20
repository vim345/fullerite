package handler

import (
	"fullerite/metric"
	"log"
)

// SignalFx type.
type SignalFx struct {
	api           string
	interval      int
	maxBufferSize int
	channel       chan metric.Metric
}

// NewSignalFx returns a new SignalFx handler.
func NewSignalFx() *SignalFx {
	s := new(SignalFx)
	s.channel = make(chan metric.Metric)
	return s
}

// Run send metrics in the channel to SignalFx.
func (s SignalFx) Run() {
	// TODO: check interval and queue size and metrics.
	for metric := range s.channel {
		// TODO: Actually send to signalfx.
		log.Println("Sending metric to SignalFx:", metric)
	}
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

// SetInterval sets the flush rate of the handler.
func (s *SignalFx) SetInterval(interval int) {
	s.interval = interval
}

// SetMaxBufferSize sets the buffer size for flush to be called.
func (s *SignalFx) SetMaxBufferSize(maxBufferSize int) {
	s.maxBufferSize = maxBufferSize
}
