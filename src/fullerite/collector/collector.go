package collector

import (
	"fullerite/metric"
	"log"
	"time"
)

// Collector defines the interface of a generic collector.
type Collector interface {
	Collect()
	Name() string
	Interval() int
	Channel() chan metric.Metric
}

// New creates a new Collector based on the requested collector name.
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

// Run runs the collector forever with its defined interval.
func Run(c Collector) {
	// TODO: do we need this?
	for {
		c.Collect()
		time.Sleep(time.Duration(c.Interval()) * time.Second)
	}
}
