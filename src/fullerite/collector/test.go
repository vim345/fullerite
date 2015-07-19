package collector

import (
	"fullerite/metric"
)

type Test struct {
	interval int
	channel  chan metric.Metric
}

func (t Test) Collect() {
	// TODO: implement
}

func (t Test) Name() string {
	return "Test"
}

func (t Test) Interval() int {
	return t.interval
}

func (t Test) Channel() chan metric.Metric {
	return t.channel
}

func (t Test) String() string {
	return t.Name() + "Collector"
}
