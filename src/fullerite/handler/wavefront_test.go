package handler

import (
	"fullerite/metric"

	"testing"
	"time"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func getTestWavefrontHandler(interval, buffsize, timeoutsec int) *Wavefront {
	testChannel := make(chan metric.Metric)
	testLog := l.WithField("testing", "wavefront_handler")
	timeout := time.Duration(timeoutsec) * time.Second
	w := newWavefront(testChannel, interval, buffsize, timeout, testLog).(*Wavefront)
	w.proxyFlag = false
	return w
}

func TestWavefrontConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})

	w := getTestWavefrontHandler(12, 13, 14)
	w.Configure(config)

	assert.Equal(t, 12, w.Interval())
	assert.Equal(t, 13, w.MaxBufferSize())
}

func TestWavefrontConfigure(t *testing.T) {
	config := map[string]interface{}{
		"interval":        "10",
		"timeout":         "10",
		"max_buffer_size": "100",
		"endpoint":        "wavefront.server",
		"proxyFlag":       "true",
		"port":            "2878",
		"proxyServer":     "wavefront.proxy",
	}

	w := getTestWavefrontHandler(40, 50, 60)
	w.Configure(config)

	assert.Equal(t, 10, w.Interval())
	assert.Equal(t, 100, w.MaxBufferSize())
	assert.Equal(t, true, w.proxyFlag)
	assert.Equal(t, "wavefront.proxy", w.proxyServer)
	assert.Equal(t, "2878", w.port)
}

func TestWavefrontSanitation(t *testing.T) {
	w := getTestWavefrontHandler(12, 12, 12)

	m1 := metric.New(" Test= .me$tric ")
	var host = []byte{260: 'x'}
	m1.AddDimension("host", string(host))
	m1.AddDimension("With_quotes", "_Value_with-\"quotes++\"_")
	datapoint1 := w.convertToWavefront(m1)

	m2 := metric.New("Test-_.metric")
	var tag = []byte{1030: 'x'}
	m2.AddDimension("long_tag", string(tag))
	datapoint2 := w.convertToWavefront(m2)

	assert.Equal(t, datapoint1.Name, datapoint2.Name, "the metric name should be the same")
	assert.Equal(t, len(datapoint1.PointTags), len(datapoint2.PointTags))
}
