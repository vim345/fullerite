package handler

import (
	"fullerite/metric"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

// Kairos handler
type Kairos struct {
	BaseHandler
	server string
	port   string
}

type KairosMetric struct {
	Name       string            `json:"name"`
	Timestamp  int64             `json:"timestamp"`
	MetricType string            `json:"type"`
	Value      float64           `json:"value"`
	Tags       map[string]string `json:"tags"`
}

// NewKairos returns a new Kairos handler
func NewKairos() *Kairos {
	k := new(Kairos)
	k.name = "Kairos"
	k.interval = DefaultInterval
	k.maxBufferSize = DefaultBufferSize
	k.timeout = time.Duration(DefaultTimeoutSec * time.Second)
	k.log = logrus.WithFields(logrus.Fields{"app": "fullerite", "pkg": "handler", "handler": "Kairos"})
	k.channel = make(chan metric.Metric)
	k.emissionTimes = make([]float64, 0)
	return k
}

// Configure the Kairos handler
func (k *Kairos) Configure(configMap map[string]interface{}) {
	if server, exists := configMap["server"]; exists == true {
		k.server = server.(string)
	} else {
		k.log.Error("There was no server specified for the Kairos Handler, there won't be any emissions")
	}

	if port, exists := configMap["port"]; exists == true {
		k.port = port.(string)
	} else {
		k.log.Error("There was no port specified for the Kairos Handler, there won't be any emissions")
	}
	k.configureCommonParams(configMap)
}

// Server returns the Kairos server's hostname or IP address
func (k *Kairos) Server() string {
	return k.server
}

// Port returns the Kairos server's port number
func (k *Kairos) Port() string {
	return k.port
}

// Run runs the Kairos handler
func (k *Kairos) Run() {
	datapoints := make([]KairosMetric, 0, k.maxBufferSize)

	lastEmission := time.Now()
	lastHandlerMetricsEmission := lastEmission
	for incomingMetric := range k.Channel() {
		datapoint := k.convertToKairos(incomingMetric)
		k.log.Debug("Kairos datapoint: ", datapoint)
		datapoints = append(datapoints, datapoint)

		emitIntervalPassed := time.Since(lastEmission).Seconds() >= float64(k.interval)
		emitHandlerIntervalPassed := time.Since(lastHandlerMetricsEmission).Seconds() >= float64(k.interval)
		bufferSizeLimitReached := len(datapoints) >= k.maxBufferSize
		doEmit := emitIntervalPassed || bufferSizeLimitReached

		if emitHandlerIntervalPassed {
			lastHandlerMetricsEmission = time.Now()
			m := k.makeEmissionTimeMetric()
			k.resetEmissionTimes()
			m.AddDimension("handler", "Kairos")
			datapoints = append(datapoints, k.convertToKairos(m))
		}

		if doEmit {
			// emit datapoints
			beforeEmission := time.Now()
			k.emitMetrics(datapoints)
			lastEmission = time.Now()

			emissionTimeInSeconds := lastEmission.Sub(beforeEmission).Seconds()
			k.log.Info("POST to Kairos took ", emissionTimeInSeconds, " seconds")
			k.emissionTimes = append(k.emissionTimes, emissionTimeInSeconds)

			// reset datapoints
			datapoints = make([]KairosMetric, 0, k.maxBufferSize)
		}

	}
}

func (k *Kairos) convertToKairos(incomingMetric metric.Metric) (datapoint KairosMetric) {
	km := new(KairosMetric)
	km.Name = k.Prefix() + incomingMetric.Name
	km.Value = incomingMetric.Value
	km.MetricType = "double"
	km.Timestamp = time.Now().Unix() * 1000 // Kairos require timestamps to be milliseconds
	km.Tags = incomingMetric.GetDimensions(k.DefaultDimensions())
	return *km
}

func (k *Kairos) emitMetrics(series []KairosMetric) {
	k.log.Info("Starting to emit ", len(series), " datapoints")

	if len(series) == 0 {
		k.log.Warn("Skipping send because of an empty payload")
		return
	}

	payload, err := json.Marshal(series)
	if err != nil {
		k.log.Error("Failed marshaling datapoints to Kairos format")
		k.log.Error("Dropping Kairos datapoints ", series)
		return
	}

	apiURL := fmt.Sprintf("http://%s:%s/api/v1/datapoints", k.server, k.port)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		k.log.Error("Failed to create a request to API url ", apiURL)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	transport := http.Transport{
		Dial: k.dialTimeout,
	}
	client := &http.Client{
		Transport: &transport,
	}
	rsp, err := client.Do(req)
	if err != nil {
		k.log.Error("Failed to complete POST ", err)
		return
	}

	defer rsp.Body.Close()
	if rsp.StatusCode == http.StatusNoContent {
		k.log.Info("Successfully sent ", len(series), " datapoints to Kairos")
	} else {
		body, _ := ioutil.ReadAll(rsp.Body)
		k.log.Error("Failed to post to Kairos @", apiURL,
			" status was ", rsp.Status,
			" rsp body was ", string(body),
			" payload was ", string(payload))
		return
	}

}

func (k *Kairos) dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, k.timeout)
}
