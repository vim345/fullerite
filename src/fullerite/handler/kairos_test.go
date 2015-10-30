package handler

import (
	"fullerite/metric"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func getTestKairosHandler(interval, buffsize, timeoutsec int) *Kairos {
	testChannel := make(chan metric.Metric)
	testLog := l.WithField("testing", "kairos_handler")
	timeout := time.Duration(timeoutsec) * time.Second

	return NewKairos(testChannel, interval, buffsize, timeout, testLog)
}

func TestKairosConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	k := getTestKairosHandler(12, 13, 14)
	k.Configure(config)

	assert.Equal(t, 12, k.Interval())
}

func TestKairosConfigure(t *testing.T) {
	config := map[string]interface{}{
		"interval":        "10",
		"timeout":         "10",
		"max_buffer_size": "100",
		"server":          "kairos.server",
		"port":            "10101",
	}

	k := getTestKairosHandler(12, 13, 14)
	k.Configure(config)

	assert.Equal(t, 10, k.Interval())
	assert.Equal(t, 100, k.MaxBufferSize())
	assert.Equal(t, "kairos.server", k.Server())
	assert.Equal(t, "10101", k.Port())
}

func TestKairosConfigureIntPort(t *testing.T) {
	config := map[string]interface{}{
		"interval":        "10",
		"timeout":         "10",
		"max_buffer_size": "100",
		"server":          "kairos.server",
		"port":            10101,
	}

	k := getTestKairosHandler(12, 13, 14)
	k.Configure(config)

	assert.Equal(t, 10, k.Interval())
	assert.Equal(t, 100, k.MaxBufferSize())
	assert.Equal(t, "kairos.server", k.Server())
	assert.Equal(t, "10101", k.Port())
}

func TestKairosRun(t *testing.T) {
	wait := make(chan bool)
	// Mock Kairos server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)

		kairosMetrics := make([]KairosMetric, 1)
		err = json.Unmarshal(body, &kairosMetrics)
		assert.Nil(t, err)

		assert.Equal(t, kairosMetrics[0].Name, "Test")
		assert.Equal(t, r.Header["Content-Type"], []string{"application/json"})

		wait <- true
	}))
	defer ts.Close()

	url, _ := url.Parse(ts.URL)
	urlParts := strings.Split(url.Host, ":")

	config := map[string]interface{}{
		"interval":        "1",
		"timeout":         "1",
		"max_buffer_size": "1",
		"server":          urlParts[0],
		"port":            urlParts[1],
	}

	k := getTestKairosHandler(12, 13, 14)
	k.Configure(config)

	go k.Run()

	m := metric.New("Test")
	k.Channel() <- m

	select {
	case <-wait:
		// noop
	case <-time.After(2 * time.Second):
		t.Fatal("Failed to post and handle after 2 seconds")
	}
}
