package handler

import (
	"fullerite/metric"
)

type Graphite struct {
	interval      int
	maxBufferSize int
	channel       chan metric.Metric
}

func (h Graphite) Send() {
	// TODO: implement
}

func (g Graphite) Name() string {
	return "Graphite"
}

func (g Graphite) Interval() int {
	return g.interval
}

func (g Graphite) MaxBufferSize() int {
	return g.maxBufferSize
}

func (g Graphite) Channel() chan metric.Metric {
	return g.channel
}

func (g Graphite) String() string {
	return g.Name() + "Handler"
}
