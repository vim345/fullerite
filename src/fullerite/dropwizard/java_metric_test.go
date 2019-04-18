package dropwizard

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJavaMetric(t *testing.T) {
	testMeters := make(map[string]map[string]interface{})
	testMeters["com.yelp.service.endpoint"] = map[string]interface{}{
		"count":   957,
		"p50":     0.5,
		"m1_rate": 0.1,
		"units":   "events/second",
	}
	parser := NewJavaMetric([]byte(``), "", true)
	actual := parser.parseMapOfMap(testMeters, "metricType")

	for _, m := range actual {
		switch m.Name {
		case "com.yelp.service.endpoint.count":
			assert.Equal(t, 957.0, m.Value)
		case "com.yelp.service.endpoint.p50":
			assert.Equal(t, 0.5, m.Value)
			assert.Equal(t, "p50", m.Dimensions["rollup"])
			assert.Equal(t, "com.yelp.service.endpoint", m.Dimensions["java_metric"])
			assert.Equal(t, 2, len(m.Dimensions))
		case "com.yelp.service.endpoint.m1_rate":
			t.Fatalf("m*_rate metrics should be discarded, found: %s", m.Name)
		default:
			t.Fatalf("unknown metric name %s", m.Name)
		}
	}
}

func TestNoServiceDimsWithJavaMetrics(t *testing.T) {
	var jsonBlob = []byte(`{
  "version": "4.0.0",
  "gauges": {
    "jvm.attribute.uptime": {
      "value": 252892259
    }
  }
}`)

	parser := NewJavaMetric(jsonBlob, "4.0.0", false)

	actual, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse input json: %s", err)
	}

	for _, m := range actual {
		assert.Equal(t, "gauge", m.MetricType)
		assert.Equal(t, 3, len(m.Dimensions))
		assert.Equal(t, "jvm.attribute.uptime", m.Name)
	}
}

func TestServiceDimsWithJavaMetrics(t *testing.T) {
	var jsonBlob = []byte(`{
  "version": "4.0.0",
  "service_dims": {
    "git_sha": "aabbcc",
    "deploy_group": "canary"
  },
  "gauges": {
    "jvm.attribute.uptime": {
      "value": 252892259
    }
  }
}`)

	parser := NewJavaMetric(jsonBlob, "4.0.0", false)

	actual, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse input json: %s", err)
	}

	for _, m := range actual {
		assert.Equal(t, "gauge", m.MetricType)
		assert.Equal(t, 5, len(m.Dimensions))
		assert.Equal(t, "jvm.attribute.uptime", m.Name)
		assert.Equal(t, m.Dimensions["git_sha"], "aabbcc")
		assert.Equal(t, m.Dimensions["deploy_group"], "canary")
	}
}
