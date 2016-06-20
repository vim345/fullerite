package dropwizard

import "fullerite/metric"

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

func (parser *UWSGIMetric) extractParsedMetric(parsed *Format) []metric.Metric {
	results := []metric.Metric{}
	appendIt := func(metrics []metric.Metric, typeDimVal string) {
		if !parser.ccEnabled {
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

}
