package dropwizard

import (
	"encoding/json"
	"fullerite/metric"
)

// UWSGIMetric parser for UWSGI metrics
type UWSGIMetric struct {
	BaseParser
}

func NewUWSGIMetric(data []byte, schemaVer string, ccEnabled bool) *UWSGIMetric {
	parser := new(UWSGIMetric)
	parser.data = data
	parser.schemaVer = schemaVer
	parser.ccEnabled = ccEnabled
	return parser
}

func (parser *UWSGIMetric) convertToMetrics(metricMap map[string]map[string]interface{}, metricType string) []metric.Metric {
	results := []metric.Metric{}

	for metricName, metricData := range metricMap {
		tempResults := parser.metricFromMap(metricData, metricName, metricType)
		results = append(results, tempResults...)
	}
	return results
}

// Parse method parses metrics and returns
func (parser *UWSGIMetric) Parse() ([]metric.Metric, error) {
	if parser.schemaVer == "uwsgi-1.1" {
		return parser.parseUWSGIMetrics11()
	} else {
		return parser.parseUWSGIMetrics10()
	}
}

// parseUWSGIMetrics10 takes the json returned from the endpoint and converts
// it into raw metrics. We first check that the metrics returned have a float value
// otherwise we skip the metric.
func (parser *UWSGIMetric) parseUWSGIMetrics10() ([]metric.Metric, error) {
	parsed := new(Format)

	err := json.Unmarshal(parser.data, parsed)
	if err != nil {
		return []metric.Metric{}, err
	}

	results := parser.extractParsedMetric(parsed)

	return results, nil
}

// parseUWSGIMetrics11 will parse UWSGI metrics under the assumption of
// the response header containing a Metrics-Schema version 'uwsgi.1.1'.
func (parser *UWSGIMetric) parseUWSGIMetrics11() ([]metric.Metric, error) {
	parsed := new(Format)

	err := json.Unmarshal(parser.data, parsed)
	if err != nil {
		return []metric.Metric{}, err
	}

	results := parser.extractParsedMetric(parsed)

	// This is necessary as Go doesn't allow us to type assert
	// map[string]interface{} as map[string]string.
	// Basically go doesn't allow type assertions for interface{}'s nested
	// inside data structures across the entire structure since it is a linearly
	// complex action
	for k, v := range parsed.ServiceDims {
		metric.AddToAll(&results, map[string]string{k: v.(string)})
	}
	return results, nil
}
