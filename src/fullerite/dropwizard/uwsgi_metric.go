package dropwizard

import (
	"encoding/json"
	"fullerite/metric"
	"strconv"
	"strings"
)

// UWSGIMetric parser for UWSGI metrics
type UWSGIMetric struct {
	BaseParser
	UWSGIVersion string `json:"version"`
}

// Non-exported specific format for uWSGI 1.3.0 version metrics
// Input metrics are expected in the form of:
// {
// 	"gauges": [],
// 	"histograms": [],
// 	"version": "xxx",
// 	"timers": [
// 		{
//          "name": "pyramid_uwsgi_metrics.tweens.status.metrics",
// 			"count": ###,
// 			"p98": ###,
// 			...
// 		},
// 		{
//          "name": "pyramid_uwsgi_metrics.tweens.lookup",
// 			"count": ###,
// 			...
// 		}
// 	],
// 	"meters": [
// 		{
//          "name": "pyramid_uwsgi_metrics.tweens.XXX",
//			"count": ###,
//			"mean_rate": ###,
// 			"m1_rate": ###
// 		}
// 	],
// 	"counters": {
//		{
//          "name": "myname",
//			"count": ###,
// 	]
// }
//
// The logic behind this structural change is to support metric-specific
// dimensions without requiring key modification to differentiate the same
// metric name with different dimension sets. This way all metric objects under
// a given type are supplied as a list of objects, rather than a map keyed by object
// name. Thus objects with the same name can have different dimension sets.
type uwsgiFormat struct {
	ServiceDims map[string]interface{} `json:"service_dims"`
	Counters    []map[string]interface{}
	Gauges      []map[string]interface{}
	Histograms  []map[string]interface{}
	Meters      []map[string]interface{}
	Timers      []map[string]interface{}
}

// NewUWSGIMetric creates new parser for uwsgi metrics
func NewUWSGIMetric(data []byte, schemaVer string, ccEnabled bool) *UWSGIMetric {
	parser := new(UWSGIMetric)
	parser.data = data
	parser.schemaVer = schemaVer
	parser.ccEnabled = ccEnabled
	err := json.Unmarshal(data, parser)
	if err != nil {
		// Assign a nil version that's easy to manage
		parser.UWSGIVersion = "0.0.0"
	}
	return parser
}

func (parser *UWSGIMetric) parseArrOfMap(metricArray []map[string]interface{}, metricType string) []metric.Metric {
	results := []metric.Metric{}

	for _, metricData := range metricArray {
		if name, ok := metricData["name"]; ok {
			delete(metricData, "name")
			tempResults := parser.metricFromMap(metricData, name.(string), metricType)
			results = append(results, tempResults...)
		}
	}
	return results
}

func (parser *UWSGIMetric) parseMapOfMap(metricMap map[string]map[string]interface{}, metricType string) []metric.Metric {
	results := []metric.Metric{}

	for metricName, metricData := range metricMap {
		tempResults := parser.metricFromMap(metricData, metricName, metricType)
		results = append(results, tempResults...)
	}
	return results
}

// Parse method parses metrics and returns
func (parser *UWSGIMetric) Parse() ([]metric.Metric, error) {
	if parser.schemaVer == "uwsgi.1.1" {
		return parser.parseUWSGIMetrics11()
	}
	return parser.parseUWSGIMetrics10()
}

// parseUWSGIMetrics10 takes the json returned from the endpoint and converts
// it into raw metrics. We first check that the metrics returned have a float value
// otherwise we skip the metric.
func (parser *UWSGIMetric) parseUWSGIMetrics10() ([]metric.Metric, error) {
	results := []metric.Metric{}
	if isNewUWSGI(parser.UWSGIVersion) {
		parsed := new(uwsgiFormat)

		err := json.Unmarshal(parser.data, parsed)
		if err != nil {
			return []metric.Metric{}, err
		}

		results = extractUWSGIParsedMetric(parser, parsed)

	} else {
		parsed := new(Format)

		err := json.Unmarshal(parser.data, parsed)
		if err != nil {
			return []metric.Metric{}, err
		}

		results = extractParsedMetric(parser, parsed)
	}

	return results, nil
}

// parseUWSGIMetrics11 will parse UWSGI metrics under the assumption of
// the response header containing a Metrics-Schema version 'uwsgi.1.1'.
func (parser *UWSGIMetric) parseUWSGIMetrics11() ([]metric.Metric, error) {
	results := []metric.Metric{}
	if isNewUWSGI(parser.UWSGIVersion) {
		parsed := new(uwsgiFormat)

		err := json.Unmarshal(parser.data, parsed)
		if err != nil {
			return []metric.Metric{}, err
		}

		results = extractUWSGIParsedMetric(parser, parsed)
		for k, v := range parsed.ServiceDims {
			metric.AddToAll(&results, map[string]string{k: v.(string)})
		}

	} else {
		parsed := new(Format)

		err := json.Unmarshal(parser.data, parsed)
		if err != nil {
			return []metric.Metric{}, err
		}

		results = extractParsedMetric(parser, parsed)
		for k, v := range parsed.ServiceDims {
			metric.AddToAll(&results, map[string]string{k: v.(string)})
		}
	}

	return results, nil
}

func isNewUWSGI(s string) bool {
	var base_ver = []int{1, 3, 0}

	for idx, elem := range strings.Split(s, ".") {

		v, err := strconv.Atoi(elem)
		if err != nil || v < base_ver[idx] {
			return false
		}
		if v > base_ver[idx] {
			return true
		}
	}
	return true
}

func extractUWSGIParsedMetric(parser *UWSGIMetric, parsed *uwsgiFormat) []metric.Metric {
	results := []metric.Metric{}
	appendIt := func(metrics []metric.Metric, typeDimVal string) {
		if !parser.isCCEnabled() {
			metric.AddToAll(&metrics, map[string]string{"type": typeDimVal})
		}
		results = append(results, metrics...)
	}

	appendIt(parser.parseArrOfMap(parsed.Gauges, metric.Gauge), "gauge")
	appendIt(parser.parseArrOfMap(parsed.Counters, metric.Counter), "counter")
	appendIt(parser.parseArrOfMap(parsed.Histograms, metric.Gauge), "histogram")
	appendIt(parser.parseArrOfMap(parsed.Meters, metric.Gauge), "meter")
	appendIt(parser.parseArrOfMap(parsed.Timers, metric.Gauge), "timer")

	return results
}
