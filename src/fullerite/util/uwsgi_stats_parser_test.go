package util

import (
	"fullerite/metric"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getArtificialUWSGIStatsResponse() []byte {
	return []byte(`{
        "workers":[
		{"status":"idle"},
		{"status":"pause"},
		{"status":"cheap"},
		{"status":"sigsig"},
		{"status":"sig255"},
		{"status":"invalid"},
		{"status":"idle"}
	]
	}`)
}

// Validate Results
func TestParseUWSGIWorkersStatsShort(t *testing.T) {
	outMetrics, _ := ParseUWSGIWorkersStats(getArtificialUWSGIStatsResponse())
	assert.Equal(t, 6, len(outMetrics))

	for _, m := range outMetrics {

		switch m.Name {
		case "IdleWorkers":
			assert.Equal(t, 2.0, m.Value)
			assert.Equal(t, metric.Gauge, m.MetricType)
			// Even if there are no busy workers, we want to emit 0
		case "BusyWorkers":
			assert.Equal(t, 0.0, m.Value)
			assert.Equal(t, metric.Gauge, m.MetricType)
		case "SigWorkers":
			assert.Equal(t, 2.0, m.Value)
			assert.Equal(t, metric.Gauge, m.MetricType)
		case "CheapWorkers":
			assert.Equal(t, 1.0, m.Value)
			assert.Equal(t, metric.Gauge, m.MetricType)
		case "InvalidWorkers":
			assert.Equal(t, 1.0, m.Value)
			assert.Equal(t, metric.Gauge, m.MetricType)
		case "PauseWorkers":
			assert.Equal(t, 1.0, m.Value)
			assert.Equal(t, metric.Gauge, m.MetricType)
		default:
			t.Fatal("Unexpected metric name: " + m.Name)
		}
	}
}

// Returns nothing if the value of a worker isn't a string
func TestParseUWSGIWorkersStatsWrongVal(t *testing.T) {
	outMetrics, _ := ParseUWSGIWorkersStats([]byte(`{
        "workers":[
		{"status":1},
		{"status":"cheap255"}
	]
	}`))
	assert.Equal(t, 0, len(outMetrics))
}

// Returns nothing if iJSON is unparsable
func TestParseUWSGIWorkersStatsWeirdJSON(t *testing.T) {
	outMetrics, _ := ParseUWSGIWorkersStats([]byte(`{
        foo/*&^%%$bar
	}`))
	assert.Equal(t, 0, len(outMetrics))
}
