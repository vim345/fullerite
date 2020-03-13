package util

import (
	"fmt"
	"io"
	"strings"

	l "github.com/Sirupsen/logrus"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/textparse"

	grpcMetrics "fullerite/collector/metrics"
	"fullerite/metric"
)

func trimSuffix(name string, suffix string) (string, bool) {
	nameLen := len(name)
	suffixLen := len(suffix)
	if nameLen > suffixLen && name[nameLen-suffixLen:] == suffix {
		return name[:nameLen-suffixLen], true
	}
	return name, false
}

// ExtractPrometheusMetrics returns an array of metrics extracted from the
// given Prometheus endpoint.
func ExtractPrometheusMetrics(
	body []byte,
	contentType string,
	metricsWhitelist map[string]bool,
	metricsBlacklist map[string]bool,
	prefix string,
	generatedDimensions map[string]string,
	log *l.Entry,
) (metrics []metric.Metric, err error) {
	metrics = []metric.Metric{}

	var metricType textparse.MetricType

	parser := textparse.New(body, contentType)
	for {
		var et textparse.Entry
		if et, err = parser.Next(); err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		switch et {
		case textparse.EntryType:
			_, metricType = parser.Type()
			continue
		case textparse.EntryHelp:
			continue
		case textparse.EntryUnit:
			continue
		case textparse.EntryComment:
			continue
		default:
		}

		var ls labels.Labels
		parser.Metric(&ls)
		entryLabels := ls.Map()

		metricName := entryLabels[labels.MetricName]
		delete(entryLabels, labels.MetricName)

		var isSum, isCount, isBucket bool
		if metricType == textparse.MetricTypeSummary || metricType == textparse.MetricTypeHistogram {
			metricName, isSum = trimSuffix(metricName, "_sum")
			if !isSum {
				metricName, isCount = trimSuffix(metricName, "_count")
				if !isCount && metricType == textparse.MetricTypeHistogram {
					metricName, isBucket = trimSuffix(metricName, "_bucket")
				}
			}
		}

		if metricsWhitelist != nil {
			if _, ok := metricsWhitelist[metricName]; !ok {
				continue
			}
		} else if metricsBlacklist != nil {
			if _, ok := metricsBlacklist[metricName]; ok {
				continue
			}
		}

		var fulleriteMetricType string
		switch metricType {
		case textparse.MetricTypeGauge:
			fulleriteMetricType = metric.Gauge
		case textparse.MetricTypeCounter:
			fulleriteMetricType = metric.CumulativeCounter
		case textparse.MetricTypeSummary:
			if isCount {
				fulleriteMetricType = metric.CumulativeCounter
				metricName = fmt.Sprintf("%s_count", metricName)
			} else if isSum {
				fulleriteMetricType = metric.CumulativeCounter
			} else {
				fulleriteMetricType = metric.Gauge
				metricName = fmt.Sprintf("%s_quantile", metricName)
			}
		case textparse.MetricTypeHistogram:
			fulleriteMetricType = metric.CumulativeCounter
			if isCount {
				metricName = fmt.Sprintf("%s_count", metricName)
			} else if isBucket {
				metricName = fmt.Sprintf("%s_bucket", metricName)
			}
		default:
			continue
		}

		var fulleriteMetricName string
		if prefix != "" {
			fulleriteMetricName = fmt.Sprintf("%s%s", prefix, metricName)
		} else {
			fulleriteMetricName = metricName
		}

		_, _, value := parser.Series()

		metric := metric.New(fulleriteMetricName)
		metric.MetricType = fulleriteMetricType
		metric.Value = value
		for labelName, labelValue := range entryLabels {
			metric.AddDimension(labelName, labelValue)
		}
		for dimensionName, dimensionValue := range generatedDimensions {
			metric.AddDimension(dimensionName, dimensionValue)
		}
		metrics = append(metrics, metric)
	}

	return metrics, err
}

// ExtractPrometheusMetricsFromSamples returns an array of metrics extracted from the
// given grpc metric sample.
func ExtractPrometheusMetricsFromSamples(serviceName string, familySamples []*grpcMetrics.MetricFamilySamples) (metrics []metric.Metric) {
	allMetrics := []metric.Metric{}

	for _, familySample := range familySamples {
		for _, sample := range familySample.Samples {
			metricName := fmt.Sprintf("%s_%s", serviceName, sample.Name)
			metric := metric.Metric{
				Name:       metricName,
				MetricType: strings.ToLower(familySample.GetType().String()),
				Value:      sample.Value,
				Dimensions: make(map[string]string),
			}
			// It is assumed label names and label values always have the same length.
			for i := 0; i < len(sample.LabelNames); i++ {
				metric.AddDimension(sample.LabelNames[i], sample.LabelValues[i])
			}
			allMetrics = append(allMetrics, metric)
		}
	}
	return allMetrics
}
