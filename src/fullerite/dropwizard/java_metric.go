package dropwizard

import (
	"encoding/json"
	"fullerite/metric"
	"regexp"
	"strings"
)

type JavaMetric struct {
	BaseParser
}

func (parser *JavaMetric) Parse() ([]metric.Metric, error) {
	parsed := new(Format)
	err := json.Unmarshal(parser.data, parsed)

	if err != nil {
		return []metric.Metric{}, err
	}

	results := parser.extractParsedMetric(parsed)
	return results, nil
}

func (parser *JavaMetric) convertToMetrics(
	metricMap map[string]map[string]interface{},
	metricType string) []metric.Metric {
	results := []metric.Metric{}
	var values []string

	for metricName, metricData := range metricMap {
		values = strings.Split(metricName, ",")
		for rollup, value := range metricData {
			mName := values[0]
			mType := metricType
			matched, _ := regexp.MatchString("m[0-9]+_rate", rollup)

			// If cumulCounterEnabled is true:
			//		1. change metric type meter.count and timer.count moving them to cumulative counter
			//		2. don't send back metered metrics (rollup == 'mXX_rate')
			if parser.ccEnabled && matched {
				continue
			}
			if parser.ccEnabled && rollup != "value" {
				mName = mName + "." + rollup
				if rollup == "count" {
					mType = metric.CumulativeCounter
				}
			}
			tmpMetric, ok := parser.createMetricFromDatam(rollup, value, mName, mType)
			if ok {
				addDimensionsFromName(&tmpMetric, values)
				results = append(results, tmpMetric)
			}
		}
	}

	return results
}

func addDimensionsFromName(m *metric.Metric, dimensions []string) {
	var dimension []string
	for i := 1; i < len(dimensions); i++ {
		dimension = strings.Split(dimensions[i], "=")
		m.AddDimension(dimension[0], dimension[1])
	}

}
