package dropwizard

import (
	"encoding/json"
	"fullerite/metric"
	"strings"
)

// LegacyMetric is for parsing metrics without version information
type LegacyMetric struct {
	BaseParser
}

// For parsing Dropwizard json output
type nestedMetricMap struct {
	metricSegments []string
	metricMap      map[string]interface{}
}

// NewLegacyMetric new parser for legacy metrics
func NewLegacyMetric(data []byte, schemaVer string, ccEnabled bool) *LegacyMetric {
	parser := new(LegacyMetric)
	parser.data = data
	parser.schemaVer = schemaVer
	parser.ccEnabled = ccEnabled
	return parser
}

// Parse parses metric using legacy format. It is usable when schema version is
// missing from metrics
func (parser *LegacyMetric) Parse() ([]metric.Metric, error) {
	uwsgiMetric := NewUWSGIMetric(parser.data, parser.schemaVer, parser.ccEnabled)

	results, err := uwsgiMetric.Parse()

	if err != nil {
		return results, err
	}

	if len(results) == 0 {
		return parser.parseLegacyJavaMetric()
	}
	return results, nil
}

func (parser *LegacyMetric) parseLegacyJavaMetric() ([]metric.Metric, error) {
	var parsed map[string]interface{}

	err := json.Unmarshal(parser.data, &parsed)

	if err != nil {
		return []metric.Metric{}, err
	}

	return parser.parseNestedMetricMaps(parsed), nil
}

// parseNestedMetricMaps takes in arbitrarily nested map of following format::
//        "jetty": {
//            "trace-requests": {
//                "put-requests": {
//                    "duration": {
//                        "30x-response": {
//                            "count": 0,
//                            "type": "counter"
//                        }
//                    }
//                }
//            }
//        }
// and returns list of metrices by unrolling the map until it finds flattened map
// and then it combines keys encountered so far to emit metrices. For above sample
// data - emitted metrices will look like:
//	metric.Metric(
//		MetricName=jetty.trace-requests.put-request.duration.30x-response,
//		MetricType=COUNTER,
//		Value=0,
//		Dimenstions={rollup:count}
//		)
func (parser *LegacyMetric) parseNestedMetricMaps(
	jsonMap map[string]interface{}) []metric.Metric {

	results := []metric.Metric{}
	unvisitedMetricMaps := []nestedMetricMap{}

	startMetricMap := nestedMetricMap{
		metricSegments: []string{},
		metricMap:      jsonMap,
	}

	unvisitedMetricMaps = append(unvisitedMetricMaps, startMetricMap)

	for len(unvisitedMetricMaps) > 0 {
		nodeToVisit := unvisitedMetricMaps[0]
		unvisitedMetricMaps = unvisitedMetricMaps[1:]

		currentMetricSegment := nodeToVisit.metricSegments

	nodeVisitorLoop:
		for k, v := range nodeToVisit.metricMap {
			switch t := v.(type) {
			case map[string]interface{}:
				unvistedNode := nestedMetricMap{
					metricSegments: append(currentMetricSegment, k),
					metricMap:      t,
				}
				unvisitedMetricMaps = append(unvisitedMetricMaps, unvistedNode)
			default:
				tempResults := parser.parseFlattenedMetricMap(nodeToVisit.metricMap,
					currentMetricSegment)
				if len(tempResults) > 0 {
					results = append(results, tempResults...)
					break nodeVisitorLoop
				}
				m, ok := parser.extractGaugeValue(k, v, currentMetricSegment)
				if ok {
					results = append(results, m)
				}
			}
		}
	}

	return results
}

func (parser *LegacyMetric) parseFlattenedMetricMap(
	jsonMap map[string]interface{}, metricName []string) []metric.Metric {
	if t, ok := jsonMap["type"]; ok {
		metricType := t.(string)
		if metricType == "gauge" {
			return parser.collectGauge(jsonMap, metricName, "gauge")
		} else if metricType == "histogram" {
			return parser.collectHistogram(jsonMap, metricName, "histogram")
		} else if metricType == "counter" {
			return parser.collectCounter(jsonMap, metricName, "counter")
		} else if metricType == "meter" {
			return parser.collectMeter(jsonMap, metricName)
		}
	}

	// if nothing else works try for rate
	return parser.collectRate(jsonMap, metricName)
}

// extractGaugeValue emits metric for Map data which otherwise did not conform to
// any of predefined schemas as GAUGEs. For example::
//  "jvm": {
//    "garbage-collectors": {
//      "ConcurrentMarkSweep": {
//        "runs": 13,
//        "time": 1531
//      },
//      "ParNew": {
//        "runs": 45146,
//        "time": 1324093
//      }
//    },
//    "memory": {
//      "heap_usage": 0.24599579808405247,
//      "totalInit": 12887457792,
//      "memory_pool_usages": {
//        "Par Survivor Space": 0.11678684852358097,
//        "CMS Old Gen": 0.2679682979112999,
//        "Metaspace": 0.9466757034141924
//      },
//      "totalMax": 12727877631
//    },
//    "buffers": {
//      "direct": {
//        "count": 410,
//        "memoryUsed": 23328227,
//        "totalCapacity": 23328227
//      },
//      "mapped": {
//        "count": 1,
//        "memoryUsed": 18421396,
//        "totalCapacity": 18421396
//      }
//    }
//  }
func (parser *LegacyMetric) extractGaugeValue(key string, value interface{}, metricName []string) (metric.Metric, bool) {
	compositeMetricName := strings.Join(append(metricName, key), ".")
	return parser.createMetricFromDatam("value", value, compositeMetricName, "GAUGE")
}

// collectGauge emits metric array for maps that contain guage values:
//    "percent-idle": {
//      "value": 0.985,
//      "type": "gauge"
//    }
func (parser *LegacyMetric) collectGauge(jsonMap map[string]interface{}, metricName []string,
	metricType string) []metric.Metric {

	if _, ok := jsonMap["value"]; ok {
		compositeMetricName := strings.Join(metricName, ".")
		return parser.metricFromMap(jsonMap, compositeMetricName, metricType)
	}
	return []metric.Metric{}
}

// collectHistogram returns metrics list for maps that contain following data::
//    "prefix-length": {
//        "type": "histogram",
//        "count": 1,
//        "min": 2,
//        "max": 2,
//        "mean": 2,
//        "std_dev": 0,
//        "median": 2,
//        "p75": 2,
//        "p95": 2,
//        "p98": 2,
//        "p99": 2,
//        "p999": 2
//    }
func (parser *LegacyMetric) collectHistogram(jsonMap map[string]interface{},
	metricName []string, metricType string) []metric.Metric {

	results := []metric.Metric{}

	if _, ok := jsonMap["count"]; ok {
		for key, value := range jsonMap {
			if key == "type" {
				continue
			}

			metricType := MetricTypeGauge
			if key == "count" {
				metricType = MetricTypeCounter
			}

			compositeMetricName := strings.Join(metricName, ".")
			m, ok := parser.createMetricFromDatam(key, value, compositeMetricName, metricType)
			if ok {
				results = append(results, m)
			}
		}
	}
	return results
}

// collectCounter returns metric list for data that looks like:
//    "active-suspended-requests": {
//       "count": 0,
//       "type": "counter"
//    }
func (parser *LegacyMetric) collectCounter(jsonMap map[string]interface{}, metricName []string,
	metricType string) []metric.Metric {

	if _, ok := jsonMap["count"]; ok {
		compositeMetricName := strings.Join(metricName, ".")
		return parser.metricFromMap(jsonMap, compositeMetricName, metricType)
	}
	return []metric.Metric{}
}

// collectRate returns metric list for data that looks like:
//    "rate": {
//      "m15": 0,
//      "m5": 0,
//      "m1": 0,
//      "mean": 0,
//      "count": 0,
//      "unit": "seconds"
//    }
func (parser *LegacyMetric) collectRate(jsonMap map[string]interface{}, metricName []string) []metric.Metric {
	results := []metric.Metric{}
	if unit, ok := jsonMap["unit"]; ok && (unit == "seconds" || unit == "milliseconds") {
		for key, value := range jsonMap {
			if key == "unit" {
				continue
			}
			metricType := MetricTypeGauge

			if key == "count" {
				metricType = MetricTypeCounter
			}

			compositeMetricName := strings.Join(metricName, ".")
			m, ok := parser.createMetricFromDatam(key, value, compositeMetricName, metricType)
			if ok {
				results = append(results, m)
			}
		}

	}
	return results
}

// collectMeter returns metric list for data that looks like:
//    "suspends": {
//      "m15": 0,
//      "m5": 0,
//      "m1": 0,
//      "mean": 0,
//      "count": 0,
//      "unit": "seconds",
//      "event_type": "requests",
//      "type": "meter"
//    }
func (parser *LegacyMetric) collectMeter(jsonMap map[string]interface{}, metricName []string) []metric.Metric {
	results := []metric.Metric{}

	if checkForMeterUnits(jsonMap) {
		for key, value := range jsonMap {
			if key == "unit" || key == "event_type" || key == "type" {
				continue
			}

			metricType := MetricTypeGauge
			if key == "count" {
				metricType = MetricTypeCounter
			}

			compositeMetricName := strings.Join(metricName, ".")
			m, ok := parser.createMetricFromDatam(key, value, compositeMetricName, metricType)
			if ok {
				results = append(results, m)
			}
		}
	}

	return results
}

func checkForMeterUnits(jsonMap map[string]interface{}) bool {
	if _, ok := jsonMap["event_type"]; ok {
		if unit, ok := jsonMap["unit"]; ok &&
			(unit == "seconds" || unit == "milliseconds" || unit == "minutes") {
			return true
		}
	}
	return false
}
