package handler_test

import (
	"fullerite/handler"
	"fullerite/metric"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKairosConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	k := handler.NewKairos()
	k.Configure(config)

	assert.Equal(t,
		k.Interval(),
		handler.DefaultInterval,
		"should be the default interval",
	)
}

func TestKairosConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = "10"
	config["timeout"] = "10"
	config["max_buffer_size"] = "100"
	config["server"] = "kairos.server"
	config["port"] = "10101"

	k := handler.NewKairos()
	k.Configure(config)

	assert := assert.New(t)
	assert.Equal(
		k.Interval(),
		10,
		"should be the set value",
	)
	assert.Equal(
		k.MaxBufferSize(),
		100,
		"should be the set value",
	)
	assert.Equal(
		k.Server(),
		config["server"],
		"should be the set value",
	)
	assert.Equal(
		k.Port(),
		config["port"],
		"should be the set value",
	)
}

func TestKairosRun(t *testing.T) {
	assert := assert.New(t)

	wait := make(chan bool)
	// Mock Kairos server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.Nil(err)

		kairosMetrics := make([]handler.KairosMetric, 1)
		err = json.Unmarshal(body, &kairosMetrics)
		assert.Nil(err)

		assert.Equal(kairosMetrics[0].Name, "Test")
		assert.Equal(r.Header["Content-Type"], []string{"application/json"})

		wait <- true
	}))
	defer ts.Close()

	url, _ := url.Parse(ts.URL)
	urlParts := strings.Split(url.Host, ":")

	config := make(map[string]interface{})
	config["interval"] = "1"
	config["timeout"] = "1"
	config["max_buffer_size"] = "1"
	config["server"] = urlParts[0]
	config["port"] = urlParts[1]
	k := handler.NewKairos()
	k.Configure(config)

	go k.Run()

	m := metric.New("Test")
	k.Channel() <- m

	<-wait
}
