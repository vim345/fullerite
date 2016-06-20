package dropwizard

import (
	"fullerite/metric"
	"regexp"

	l "github.com/Sirupsen/logrus"
)

var defaultLog = l.WithFields(l.Fields{"app": "fullerite", "pkg": "dropwizard"})

type Parser interface {
	Parse() []metric.Metric
}

// DropwizardFormat defines format in which dropwizard metrics are emitted
type Format struct {
	ServiceDims map[string]interface{} `json:"service_dims"`
	Counters    map[string]map[string]interface{}
	Gauges      map[string]map[string]interface{}
	Histograms  map[string]map[string]interface{}
	Meters      map[string]map[string]interface{}
	Timers      map[string]map[string]interface{}
}

type BaseParser struct {
	data      []byte
	log       *l.Entry
	ccEnabled bool // Enable cumulative counters
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
