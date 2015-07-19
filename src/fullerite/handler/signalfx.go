package handler

import (
	"fullerite/metric"
)

type SignalFx struct {
	interval      int
	maxBufferSize int
	channel       chan metric.Metric
}

func (h SignalFx) Send() {
	// TODO: implement
}

func (s SignalFx) Name() string {
	return "SignalFx"
}

func (s SignalFx) Interval() int {
	return s.interval
}

func (s SignalFx) MaxBufferSize() int {
	return s.maxBufferSize
}
func (s SignalFx) Channel() chan metric.Metric {
	return s.channel
}

func (s SignalFx) String() string {
	return s.Name() + "Handler"
}
