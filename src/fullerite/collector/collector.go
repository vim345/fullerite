package collector

import (
	"fullerite/metric"
	"log"
)

const (
	// DefaultCollectionInterval the interval to collect on unless overridden by a collectors config
	DefaultCollectionInterval = 10
)

// Collector defines the interface of a generic collector.
type Collector interface {
	Collect()
	Configure(*map[string]string)

	// taken care of by the base class
	Name() string
	Channel() chan metric.Metric
	Interval() int64
	SetInterval(int64)
}

// New creates a new Collector based on the requested collector name.
func New(name string) Collector {
	var collector Collector
	switch name {
	case "Test":
		collector = NewTest()
	case "CPU":
		collector = NewCPU()
	case "Diamond":
		collector = NewDiamond()
	default:
		log.Fatal("Cannot create collector", name)
		return nil
	}
	return collector
}

// BaseCollector is to handle the common components used in a collector
type BaseCollector struct {
	channel  chan metric.Metric
	name     string
	interval int64
}

// Channel : the channel on which the collector should send metrics
func (collector *BaseCollector) Channel() chan metric.Metric {
	return collector.channel
}

// Name : the name of the collector
func (collector *BaseCollector) Name() string {
	return collector.name
}

// Interval : the interval to collect the metrics on
func (collector *BaseCollector) Interval() int64 {
	return collector.interval
}

// SetInterval : set the interval to collect on
func (collector *BaseCollector) SetInterval(interval int64) {
	collector.interval = interval
}

// Configure is a placeholder for other collectors
func (collector *BaseCollector) Configure(*map[string]string) {
	// noop
}

// String returns the collector name in printable format.
func (collector *BaseCollector) String() string {
	return collector.Name() + "Collector"
}
