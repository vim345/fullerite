package metric

import (
	"time"
)

// Metric type holds all the information for a single metric data
// point. Metrics are generated in collectors and passed to handlers.
type Metric struct {
	name       string
	metricType string
	value      float64
	timestamp  int64
	dimensions []Dimension
}

// Dimension is a name:value pair. Each Metric have a list of
// dimensions.
type Dimension struct {
	name  string
	value string
}

// New returns a new metric with name. Default metric type is "gauge"
// and timestamp is set to now. Value is initialized to 0.0.
func New(name string) Metric {
	return Metric{
		name:       "TestMetric",
		metricType: "gauge",
		value:      0.0,
		timestamp:  time.Now().Unix(),
	}
}

// SetTimestamp sets the timestamp of a Metric.
func (m *Metric) SetTimestamp(timestamp int64) {
	m.timestamp = timestamp
}

// SetType sets the metric type of a Metric.
func (m *Metric) SetType(metricType string) {
	m.metricType = metricType
}

// SetValue sets the value of a Metric.
func (m *Metric) SetValue(value float64) {
	m.value = value
}

// AddDimension adds a new dimension to the Metric.
func (m *Metric) AddDimension(name, value string) {
	m.dimensions = append(m.dimensions, Dimension{name: name, value: value})
}
