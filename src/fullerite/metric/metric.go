package metric

// Metric type holds all the information for a single metric data
// point. Metrics are generated in collectors and passed to handlers.
type Metric struct {
	name       string
	metricType string
	value      float64
	timestamp  int
	dimensions []Dimension
}

// Dimension is a name:value pair. Each Metric have a list of
// dimensions.
type Dimension struct {
	name  string
	value string
}
