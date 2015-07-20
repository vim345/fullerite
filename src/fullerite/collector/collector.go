package collector

import (
	"fullerite/metric"
	"log"
)

// Collector defines the interface of a generic collector.
type Collector interface {
	Collect()
	Name() string
	Interval() int
	SetInterval(int)
	Channel() chan metric.Metric
}

// New creates a new Collector based on the requested collector name.
func New(name string) Collector {
	var collector Collector
	switch name {
	case "Test":
		collector = NewTest()
	case "CPU":
		collector = NewCPU()
	default:
		log.Fatal("Cannot create collector", name)
		return nil
	}
	return collector
}
