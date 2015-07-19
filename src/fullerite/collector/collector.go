package collector

import (
	"fullerite/metric"
	"log"
	"time"
)

type Collector interface {
	Collect()
	Name() string
	Interval() int
	Channel() chan metric.Metric
}

func New(name string) Collector {
	var collector Collector
	switch name {
	case "Test":
		collector = new(Test)
	case "CPU":
		collector = new(CPU)
	default:
		log.Fatal("Cannot create collector", name)
		return nil
	}
	return collector
}

// TODO: do we need this?
func Run(c Collector) {
	for {
		c.Collect()
		time.Sleep(time.Duration(c.Interval()) * time.Second)
	}
}
