package handler_test

import (
	"fullerite/handler"
	"fullerite/metric"

	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestSignalfxConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	s := handler.NewSignalFx()
	s.Configure(config)

	assert.Equal(t,
		s.Interval(),
		handler.DefaultInterval,
		"should be the default interval",
	)
}

func TestSignalfxConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = "10"
	config["timeout"] = "10"
	config["max_buffer_size"] = "100"
	config["authToken"] = "secret"
	config["endpoint"] = "signalfx.server"

	s := handler.NewSignalFx()
	s.Configure(config)

	assert := assert.New(t)
	assert.Equal(
		s.Interval(),
		10,
		"should be the set value",
	)
	assert.Equal(
		s.MaxBufferSize(),
		100,
		"should be the set value",
	)
	assert.Equal(
		s.Endpoint(),
		config["endpoint"],
		"should be the set value",
	)
}

func TestSignalFxRun(t *testing.T) {
	assert := assert.New(t)

	wait := make(chan bool)
	// Mock SignalFx server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		message := &handler.DataPointUploadMessage{}
		proto.Unmarshal(body, message)
		datapoint := message.Datapoints[0]

		assert.Nil(err)
		assert.Equal(*datapoint.Metric, "Test")
		assert.Equal(r.Header["Content-Type"], []string{"application/x-protobuf"})
		assert.Equal(r.Header["X-Sf-Token"], []string{"secret"})

		wait <- true
	}))
	defer ts.Close()

	config := make(map[string]interface{})
	config["interval"] = "1"
	config["timeout"] = "1"
	config["max_buffer_size"] = "1"
	config["authToken"] = "secret"
	config["endpoint"] = ts.URL
	s := handler.NewSignalFx()
	s.Configure(config)

	go s.Run()

	m := metric.New("Test")
	s.Channel() <- m

	<-wait
}
