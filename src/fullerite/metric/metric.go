package metric

type Metric struct {
	name        string
	metric_type string
	value       float64
	timestamp   int
	dimensions  []Dimension
}

type Dimension struct {
	name  string
	value string
}
