package dropwizard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUWSGIMetricConversion(t *testing.T) {
	testMeters := make(map[string]map[string]interface{})
	testMeters["pyramid_uwsgi_metrics.tweens.5xx-responses"] = map[string]interface{}{
		"count":     957,
		"mean_rate": 0.0006172935981330262,
		"m15_rate":  2.8984757611832113e-41,
		"m5_rate":   1.8870959302511822e-119,
		"m1_rate":   3e-323,

		// this will not create a metric
		"units": "events/second",
	}
	testMeters["pyramid_uwsgi_metrics.tweens.4xx-responses"] = map[string]interface{}{
		"count":     366116,
		"mean_rate": 0.2333071157843687,
		"m15_rate":  0.22693345170298124,
		"m5_rate":   0.21433439128223822,
		"m1_rate":   0.14771304656654516,

		// this will not create a metric
		"units": "events/second",
	}
	parser := NewUWSGIMetric([]byte(``), "", false)

	actual := parser.convertToMetrics(testMeters, "metricType")

	// only the numbers are made
	assert.Equal(t, 10, len(actual))
	for _, m := range actual {
		assert.Equal(t, "metricType", m.MetricType)

		// the other dims are applied at a higher level
		assert.Equal(t, 1, len(m.Dimensions))

		rollup, exists := m.GetDimensionValue("rollup")
		assert.True(t, exists)

		switch m.Name {
		case "pyramid_uwsgi_metrics.tweens.5xx-responses":
			val, exists := map[string]float64{
				"mean_rate": 0.0006172935981330262,
				"m15_rate":  2.8984757611832113e-41,
				"m5_rate":   1.8870959302511822e-119,
				"m1_rate":   3e-323,
				"count":     957,
			}[rollup]
			assert.True(t, exists, "unknown rollup "+rollup)
			assert.Equal(t, val, m.Value)
		case "pyramid_uwsgi_metrics.tweens.4xx-responses":
			val, exists := map[string]float64{
				"count":     366116,
				"mean_rate": 0.2333071157843687,
				"m15_rate":  0.22693345170298124,
				"m5_rate":   0.21433439128223822,
				"m1_rate":   0.14771304656654516,
			}[rollup]
			assert.True(t, exists, "unknown rollup "+rollup)
			assert.Equal(t, val, m.Value, "mismatching value on rollup "+rollup)
		default:
			t.Fatalf("unknown metric name %s", m.Name)
		}
	}
}

func TestUWSGIMetricConversionCumulativeCountersEnabled(t *testing.T) {
	testMeters := make(map[string]map[string]interface{})
	testMeters["pyramid_uwsgi_metrics.tweens.5xx-responses"] = map[string]interface{}{
		"count":     957,
		"mean_rate": 0.0006172935981330262,
		"m15_rate":  2.8984757611832113e-41,
		"m5_rate":   1.8870959302511822e-119,
		"m1_rate":   3e-323,

		// this will not create a metric
		"units": "events/second",
	}
	testMeters["pyramid_uwsgi_metrics.tweens.4xx-responses"] = map[string]interface{}{
		"count":     366116,
		"mean_rate": 0.2333071157843687,
		"m15_rate":  0.22693345170298124,
		"m5_rate":   0.21433439128223822,
		"m1_rate":   0.14771304656654516,

		// this will not create a metric
		"units": "events/second",
	}

	parser := NewUWSGIMetric([]byte(``), "", true)

	actual := parser.convertToMetrics(testMeters, "metricType")

	for _, m := range actual {
		switch m.Name {
		case "pyramid_uwsgi_metrics.tweens.5xx-responses.mean_rate":
			assert.Equal(t, 0.0006172935981330262, m.Value)
		case "pyramid_uwsgi_metrics.tweens.4xx-responses.mean_rate":
			assert.Equal(t, 0.2333071157843687, m.Value)
		case "pyramid_uwsgi_metrics.tweens.5xx-responses.count":
			assert.Equal(t, 957.0, m.Value)
		case "pyramid_uwsgi_metrics.tweens.4xx-responses.count":
			assert.Equal(t, 366116.0, m.Value)
		default:
			t.Fatalf("unknown metric name %s", m.Name)
		}
	}
}
