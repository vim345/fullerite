package collector

import (
	"fullerite/metric"
)

// CPU collector type.
type CPU struct {
	BaseCollector
}

// NewCPU creates a new CPU collector.
func NewCPU() *CPU {
	c := new(CPU)
	c.channel = make(chan metric.Metric)
	return c
}

// Collect currently a noop
func (c *CPU) Collect() {
	// TODO make this do something
}
