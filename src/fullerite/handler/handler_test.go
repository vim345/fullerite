package handler

import (
	"fullerite/metric"

	"fmt"
	"testing"
	"time"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	names := []string{"Graphite", "Kairos", "SignalFx", "Datadog"}
	for _, name := range names {
		h := New(name)
		assert.NotNil(t, h, "should create a Handler for "+name)
		assert.NotNil(t, h.Channel(), "should create a channel")
		assert.Equal(t, name, h.Name())
		assert.Equal(t, "", h.Prefix(), "")
		assert.Equal(t, 0, len(h.DefaultDimensions()))
		assert.Equal(t, DefaultBufferSize, h.MaxBufferSize())
		assert.Equal(t, DefaultInterval, h.Interval())
		assert.Equal(t, name+"Handler", fmt.Sprintf("%s", h), "String() should append Handler to the name for "+name)

		// Test Set* functions
		h.SetInterval(999)
		assert.Equal(t, 999, h.Interval())

		h.SetMaxBufferSize(999)
		assert.Equal(t, 999, h.MaxBufferSize())

		dims := map[string]string{"test": "test value"}
		h.SetDefaultDimensions(dims)
		assert.Equal(t, 1, len(h.DefaultDimensions()))
	}
}

func TestRecordTimings(t *testing.T) {
	base := BaseHandler{}
	base.log = l.WithField("testing", "basehandler")
	base.interval = 1

	minusTwoSec := -1 * 2 * time.Second
	minusThreeSec := -1 * 3 * time.Second
	someDur := time.Duration(5)

	// create a list of emissions in order with some older than 1 second
	base.emissionTimes.PushBack(emissionTiming{time.Now().Add(minusThreeSec), someDur})
	base.emissionTimes.PushBack(emissionTiming{time.Now().Add(minusTwoSec), someDur})
	base.emissionTimes.PushBack(emissionTiming{time.Now(), someDur})

	base.recordEmission(someDur)

	assert.Equal(t, 2, base.emissionTimes.Len())
}

func TestHandlerRun(t *testing.T) {
	base := BaseHandler{}
	base.log = l.WithField("testing", "basehandler")
	base.interval = 1
	base.maxBufferSize = 1
	base.channel = make(chan metric.Metric)

	emitCalled := false
	emitFunc := func([]metric.Metric) bool {
		emitCalled = true
		return true
	}

	// now we are waiting for some metrics
	go base.run(emitFunc)

	base.channel <- metric.New("testMetric")
	close(base.channel)

	assert.True(t, emitCalled)
	assert.Equal(t, 1, base.emissionTimes.Len())
	assert.Equal(t, uint64(1), base.metricsSent)
	assert.Equal(t, uint64(0), base.metricsDropped)
	assert.Equal(t, uint64(1), base.totalEmissions)
}

func TestInternalMetrics(t *testing.T) {
	base := BaseHandler{}
	base.totalEmissions = 10
	base.metricsDropped = 100
	base.metricsSent = 2
	base.interval = 4

	timing := emissionTiming{time.Now(), 5 * time.Second}
	base.emissionTimes.PushBack(timing)
	timing = emissionTiming{time.Now(), 10 * time.Second}
	base.emissionTimes.PushBack(timing)
	timing = emissionTiming{time.Now(), 6 * time.Second}
	base.emissionTimes.PushBack(timing)

	results := base.InternalMetrics()
	expected := InternalMetrics{
		Counters: map[string]float64{
			"metricsDropped": 100,
			"metricsSent":    2,
			"totalEmissions": 10,
		},
		Gauges: map[string]float64{
			"averageEmissionTiming": 7,
			"emissionsInWindow":     3,
			"intervalLength":        4,
		},
	}
	assert.Equal(t, expected, results)
}

func TestInternalMetricsWithNan(t *testing.T) {
	base := BaseHandler{}
	fmt.Println(base.InternalMetrics())

	expected := InternalMetrics{
		Counters: map[string]float64{
			"metricsDropped": 0,
			"metricsSent":    0,
			"totalEmissions": 0,
		},
		// specifically missing the averageEmissionTiming
		// because we have no emissions yet
		Gauges: map[string]float64{
			"emissionsInWindow": 0,
			"intervalLength":    0,
		},
	}
	im := base.InternalMetrics()
	assert.Equal(t, expected, im)

}
