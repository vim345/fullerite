package metric

// The different types of metrics that are supported
const (
	Gauge             = "gauge"
	Counter           = "counter"
	CumulativeCounter = "cumcounter"
)

// Metric type holds all the information for a single metric data
// point. Metrics are generated in collectors and passed to handlers.
type Metric struct {
	Name       string      `json:"name"`
	MetricType string      `json:"type"`
	Value      float64     `json:"value"`
	Dimensions []Dimension `json:"dimensions"`
}

// Dimension is a name:value pair. Each Metric have a list of
// dimensions.
type Dimension struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// New returns a new metric with name. Default metric type is "gauge"
// and timestamp is set to now. Value is initialized to 0.0.
func New(name string) Metric {
	return Metric{
		Name:       name,
		MetricType: "gauge",
		Value:      0.0,
	}
}

// AddDimension adds a new dimension to the Metric.
func (m *Metric) AddDimension(name, value string) {
	m.Dimensions = append(m.Dimensions, Dimension{Name: name, Value: value})
}
