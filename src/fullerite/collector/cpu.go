package collector

import (
	"fullerite/metric"
)

type CPU struct {
	interval int
	channel  chan metric.Metric
}

func (c CPU) Collect() {
	// TODO: implement
}

func (c CPU) Name() string {
	return "CPU"
}

func (c CPU) Interval() int {
	return c.interval
}

func (c CPU) Channel() chan metric.Metric {
	return c.channel
}

func (c CPU) String() string {
	return c.Name() + "Collector"
}
