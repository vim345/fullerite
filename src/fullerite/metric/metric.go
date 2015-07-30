package metric

import (
	"time"
)

// The different types of metrics that are supported
const (
	Gauge             = "gauge"
	Counter           = "counter"
	CumulativeCounter = "cumcounter"
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
	Name  string
	Value string
}

// New returns a new metric with name. Default metric type is "gauge"
// and timestamp is set to now. Value is initialized to 0.0.
func New(name string) Metric {
	return Metric{
		name:       name,
		metricType: "gauge",
		value:      0.0,
		timestamp:  time.Now().Unix(),
	}
}

// Value : the floating value of the metric
func (m *Metric) Value() float64 {
	return m.value
}

// Name : the name of the metric
func (m *Metric) Name() string {
	return m.name
}

// Type : the type of the metric: Gauge, Counter or CumulativeCounter
func (m *Metric) Type() string {
	return m.metricType
}

// Dimensions : the list of dimensions that the metric has
func (m *Metric) Dimensions() *[]Dimension {
	return &m.dimensions
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
	m.dimensions = append(m.dimensions, Dimension{Name: name, Value: value})
}
