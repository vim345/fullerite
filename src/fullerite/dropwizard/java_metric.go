package dropwizard

import (
	"encoding/json"
	"fullerite/metric"
	"regexp"
	"strings"
)

// JavaMetric is a parser for parsing java-1.1 metrics
type JavaMetric struct {
	BaseParser
}

// NewJavaMetric new parser for java metrics
func NewJavaMetric(data []byte, schemaVer string, ccEnabled bool) *JavaMetric {
	parser := new(JavaMetric)
	parser.data = data
	parser.schemaVer = schemaVer
	parser.ccEnabled = ccEnabled
	return parser
}

// Parse returns dimensionalized parsed metrics
func (parser *JavaMetric) Parse() ([]metric.Metric, error) {
	parsed := new(Format)
	err := json.Unmarshal(parser.data, parsed)

	if err != nil {
		return []metric.Metric{}, err
	}

	results := extractParsedMetric(parser, parsed)
	return results, nil
}

func (parser *JavaMetric) parseMapOfMap(
	metricMap map[string]map[string]interface{},
	metricType string) []metric.Metric {

	results := []metric.Metric{}
	var values []string

	for metricName, metricData := range metricMap {
		values = strings.Split(metricName, ",")
		for rollup, value := range metricData {
			mName := values[0]
			mNameWithSuffix := mName
			mType := metricType
			matched, _ := regexp.MatchString("m[0-9]+_rate", rollup)

			// If cumulCounterEnabled is true:
			//		1. change metric type meter.count and timer.count moving them to cumulative counter
			//		2. don't send back metered metrics (rollup == 'mXX_rate')
			if parser.ccEnabled && matched {
				continue
			}

			if parser.ccEnabled && rollup != "value" {
			// For legacy reasons we append the rollup to the metric name
				mNameWithSuffix = mName + "." + rollup
				if rollup == "count" {
					mType = metric.CumulativeCounter
				}
			}

			tmpMetric, ok := parser.createMetricFromDatam(rollup, value, mNameWithSuffix, mType)
			if ok {
				addNameDimension(&tmpMetric, mName)
				addDimensionsFromName(&tmpMetric, values)
				results = append(results, tmpMetric)
			}
		}
	}

	return results
}

// If we want to be able to do operations the same metric but different rollup (eg. mean * count)
// we need to have a common dimension to link them. Name isn't because we happened a suffix to it
// It would be better to use only one metric name with many values for rollup dimensions
// but that involve updating all our dashbords. Hence the gross hack of adding the metric name as
// a dimension.
func addNameDimension(m *metric.Metric, mName string) {
	m.AddDimension("java_metric", mName)
}

func addDimensionsFromName(m *metric.Metric, dimensions []string) {
	var dimension []string
	for i := 1; i < len(dimensions); i++ {
		dimension = strings.Split(dimensions[i], "=")
		m.AddDimension(dimension[0], dimension[1])
	}

}
