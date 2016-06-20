package dropwizard

import (
	"fmt"
	"fullerite/metric"
	"regexp"

	l "github.com/Sirupsen/logrus"
)

var defaultLog = l.WithFields(l.Fields{"app": "fullerite", "pkg": "dropwizard"})

const (
	// MetricTypeCounter String for counter metric type
	MetricTypeCounter string = "COUNTER"
	// MetricTypeGauge String for Gauge metric type
	MetricTypeGauge string = "GAUGE"
)

// Parser is an interface for dropwizard parsers
type Parser interface {
	Parse() ([]metric.Metric, error)
	createMetricFromDatam(string, interface{}, string, string) (metric.Metric, bool)
	metricFromMap(map[string]interface{}, string, string) []metric.Metric
	convertToMetrics(map[string]map[string]interface{}, string) []metric.Metric
	isCCEnabled() bool
}

// Format defines format in which dropwizard metrics are emitted
type Format struct {
	ServiceDims map[string]interface{} `json:"service_dims"`
	Counters    map[string]map[string]interface{}
	Gauges      map[string]map[string]interface{}
	Histograms  map[string]map[string]interface{}
	Meters      map[string]map[string]interface{}
	Timers      map[string]map[string]interface{}
}

// BaseParser is a base struct for real parsers
type BaseParser struct {
	data      []byte
	log       *l.Entry
	ccEnabled bool // Enable cumulative counters
	schemaVer string
}

// Parse can be called from collector code to parse results
func Parse(raw []byte, schemaVer string, ccEnabled bool) ([]metric.Metric, error) {
	var parser Parser
	if schemaVer == "uwsgi.1.0" || schemaVer == "uwsg.1.1" {
		parser = NewUWSGIMetric(raw, schemaVer, ccEnabled)
	} else if schemaVer == "java-1.1" {
		parser = NewJavaMetric(raw, schemaVer, ccEnabled)
	} else {
		parser = NewLegacyMetric(raw, schemaVer, ccEnabled)
	}
	return parser.Parse()
}

// metricFromMap takes in flattened maps formatted like this::
// {
//    "count":      3443,
//    "mean_rate": 100
// }
// and metricname and metrictype and returns metrics for each name:rollup pair
func (parser *BaseParser) metricFromMap(metricMap map[string]interface{},
	metricName string,
	metricType string) []metric.Metric {
	results := []metric.Metric{}

	for rollup, value := range metricMap {
		mName := metricName
		mType := metricType
		matched, _ := regexp.MatchString("m[0-9]+_rate", rollup)

		// If cumulCounterEnabled is true:
		//		1. change metric type meter.count and timer.count moving them to cumulative counter
		//		2. don't send back metered metrics (rollup == 'mXX_rate')
		if parser.ccEnabled && matched {
			continue
		}
		if parser.ccEnabled && rollup != "value" {
			mName = metricName + "." + rollup
			if rollup == "count" {
				mType = metric.CumulativeCounter
			}
		}
		tempMetric, ok := parser.createMetricFromDatam(rollup, value, mName, mType)
		if ok {
			results = append(results, tempMetric)
		}
	}

	return results
}

func (parser *BaseParser) isCCEnabled() bool {
	return parser.ccEnabled
}

// createMetricFromDatam takes in rollup, value, metricName, metricType and returns metric only if
// value was numeric
func (parser *BaseParser) createMetricFromDatam(rollup string,
	value interface{},
	metricName string, metricType string) (metric.Metric, bool) {
	m := metric.New(metricName)
	m.MetricType = metricType
	if parser.ccEnabled {
		m.AddDimension("rollup", rollup)
	}
	// only add things that have a numeric base
	switch value.(type) {
	case float64:
		m.Value = value.(float64)
	case int:
		m.Value = float64(value.(int))
	default:
		return m, false
	}
	return m, true
}

func extractParsedMetric(parser Parser, parsed *Format) []metric.Metric {
	results := []metric.Metric{}
	appendIt := func(metrics []metric.Metric, typeDimVal string) {
		if !parser.isCCEnabled() {
			metric.AddToAll(&metrics, map[string]string{"type": typeDimVal})
		}
		results = append(results, metrics...)
	}

	appendIt(parser.convertToMetrics(parsed.Gauges, metric.Gauge), "gauge")
	appendIt(parser.convertToMetrics(parsed.Counters, metric.Counter), "counter")
	appendIt(parser.convertToMetrics(parsed.Histograms, metric.Gauge), "histogram")
	appendIt(parser.convertToMetrics(parsed.Meters, metric.Gauge), "meter")
	appendIt(parser.convertToMetrics(parsed.Timers, metric.Gauge), "timer")

	return results
}

func (parser *BaseParser) convertToMetrics(
	metricMap map[string]map[string]interface{},
	metricType string) []metric.Metric {

	fmt.Println("^^^^ I am being called")
	return []metric.Metric{}
}

// Parse is just a placehoder function
func (parser *BaseParser) Parse() ([]metric.Metric, error) {
	return []metric.Metric{}, nil
}
