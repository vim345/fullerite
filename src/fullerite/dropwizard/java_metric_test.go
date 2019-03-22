package dropwizard

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestJavaMetric(t *testing.T) {
    testMeters := make(map[string]map[string]interface{})
    testMeters["com.yelp.service.endpoint"] = map[string]interface{}{
        "count":    957,
        "p50":      0.5,
        "m1_rate":  0.1,
        "units":    "events/second",
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
