package handler

import (
	"fullerite/metric"
	"log"
)

// Graphite type
type Graphite struct {
	server        string
	interval      int
	maxBufferSize int
	channel       chan metric.Metric
}

// NewGraphite returns a new Graphite handler.
func NewGraphite() *Graphite {
	g := new(Graphite)
	g.channel = make(chan metric.Metric)
	return g
}

// Run sends metrics in the channel to the graphite server.
func (g Graphite) Run() {
	// TODO: check interval and queue size and metrics.
	for metric := range g.channel {
		// TODO: Actually send to graphite server
		log.Println("Sending metric to Graphite:", metric)
	}

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

// SetInterval sets the flush rate of the handler.
func (g *Graphite) SetInterval(interval int) {
	g.interval = interval
}

// SetMaxBufferSize sets the buffer size for flush to be called.
func (g *Graphite) SetMaxBufferSize(maxBufferSize int) {
	g.maxBufferSize = maxBufferSize
}
