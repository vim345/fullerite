package collector

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"fullerite/metric"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestChronosStatsNewChronosStats(t *testing.T) {
	c := make(chan metric.Metric)
	i := 10
	l := defaultLog.WithFields(l.Fields{"collector": "Chronos"})

	sut := newChronosStats(c, i, l).(*ChronosStats)

	assert.Equal(t, c, sut.channel)
	assert.Equal(t, i, sut.interval)
	assert.Equal(t, l, sut.log)
	assert.Equal(t, http.Client{Timeout: getTimeout}, sut.client)
}

func TestChronosStatsGetChronosMetrics(t *testing.T) {
	oldGetChronosMetricsURL := getChronosMetricsURL
	defer func() { getChronosMetricsURL = oldGetChronosMetricsURL }()

	tests := []struct {
		rawResponse string
		expected    []struct {
			Name  string
			Value float64
			T     string
		}
		err bool
		msg string
	}{
		{
			`{"gauges": {"foo.bar": {"value": 10}}}`,
			[]struct {
				Name  string
				Value float64
				T     string
			}{{"foo.bar", 10.0, metric.Gauge}},
			false,
			"Should parse a simple input",
		},
		{
			"", nil, true, "Should return an error on bad input",
		},
		{
			`{"version": "3.0.0", "gauges": {"bar.foo": {"value": 20}}}`,
			[]struct {
				Name  string
				Value float64
				T     string
			}{{"bar.foo", 20.0, metric.Gauge}},
			false,
			"Should ignore the version field",
		},
		{
			`{"version": "3.0.0", "gauges": {"bar.foo": {"value": 20}}, "counters": {"foo.bar": {"count": 30}}}`,
			[]struct {
				Name  string
				Value float64
				T     string
			}{{"bar.foo", 20.0, metric.Gauge}, {"foo.bar.count", 30.0, metric.CumulativeCounter}},
			false,
			"Should work with multiple metrics",
		},
	}

	for _, test := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, test.rawResponse)
		}))
		defer ts.Close()

		getChronosMetricsURL = func(ip string) string { return ts.URL }

		sut := newChronosStats(nil, 10, defaultLog).(*ChronosStats)
		actual := getChronosMetrics(sut)

		if test.err {
			assert.True(t, actual == nil, test.msg)
		} else {
			for i, v := range test.expected {
				assert.Equal(t, v.Name, actual[i].Name)
				assert.Equal(t, v.Value, actual[i].Value)
				assert.Equal(t, v.T, actual[i].MetricType)
			}
		}
	}
}

func TestChronosStatsConfigure(t *testing.T) {
	sut := newChronosStats(nil, 10, defaultLog).(*ChronosStats)
	sut.Configure(map[string]interface{}{
		"chronosHost":    "foobar",
		"extraDimensions": "{\"cluster\": \"bar\"}"})

	assert.Equal(t, sut.extraDimensions["cluster"], "bar")
}
