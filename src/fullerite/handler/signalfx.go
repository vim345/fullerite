package handler

import (
	"bytes"
	"fullerite/metric"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// SignalFx Handler
type SignalFx struct {
	BaseHandler
	endpoint  string
	authToken string
}

// NewSignalFx returns a new SignalFx handler.
func NewSignalFx() *SignalFx {
	s := new(SignalFx)
	s.name = "SignalFx"
	s.maxBufferSize = DefaultBufferSize
	s.channel = make(chan metric.Metric)
	return s
}

// Configure : accepts the different configuration options for the signalfx handler
func (s *SignalFx) Configure(config *map[string]string) {
	asmap := *config
	var exists bool
	s.authToken, exists = asmap["authToken"]
	if !exists {
		log.Println("There was no auth key specified for the SignalFx Handler, there won't be any emissions")
	}
	s.endpoint, exists = asmap["endpoint"]
	if !exists {
		log.Println("There was no endpoint specified for the SignalFx Handler, there won't be any emissions")
	}
}

// Run send metrics in the channel to SignalFx.
func (s *SignalFx) Run() {
	log.Println("starting signalfx handler")
	lastEmission := time.Now()

	datapoints := make([]*DataPoint, 0, s.maxBufferSize)

	for incomingMetric := range s.Channel() {
		log.Println("Processing metric to SignalFx:", incomingMetric)

		datapoint := s.convertToProto(&incomingMetric)
		datapoints = append(datapoints, datapoint)

		if time.Since(lastEmission).Seconds() >= float64(s.interval) || len(datapoints) >= s.maxBufferSize {
			s.emitMetrics(&datapoints)
			datapoints = make([]*DataPoint, 0, s.maxBufferSize)
		}
	}
}

func (s *SignalFx) convertToProto(incomingMetric *metric.Metric) *DataPoint {

	datapoint := new(DataPoint)
	outname := s.Prefix() + (*incomingMetric).Name

	datapoint.Metric = &outname
	datapoint.Value = &Datum{
		DoubleValue: &(*incomingMetric).Value,
	}
	datapoint.Source = new(string)
	*datapoint.Source = "fullerite"

	switch incomingMetric.MetricType {
	case metric.Gauge:
		datapoint.MetricType = MetricType_GAUGE.Enum()
	case metric.Counter:
		datapoint.MetricType = MetricType_COUNTER.Enum()
	case metric.CumulativeCounter:
		datapoint.MetricType = MetricType_CUMULATIVE_COUNTER.Enum()
	}

	if s.DefaultDimensions() != nil {
		for _, dimension := range *s.DefaultDimensions() {
			dim := Dimension{
				Key:   &dimension.Name,
				Value: &dimension.Value,
			}
			datapoint.Dimensions = append(datapoint.Dimensions, &dim)
		}
	}
	for _, dimension := range incomingMetric.Dimensions {
		dim := Dimension{
			Key:   &dimension.Name,
			Value: &dimension.Value,
		}
		datapoint.Dimensions = append(datapoint.Dimensions, &dim)
	}

	return datapoint
}

func (s *SignalFx) emitMetrics(datapoints *[]*DataPoint) {
	log.Println("Starting to emit", len(*datapoints), "datapoints")

	if len(*datapoints) == 0 {
		log.Println("Skipping send because of an empty payload")
		return
	}

	payload := new(DataPointUploadMessage)
	payload.Datapoints = *datapoints
	log.Println("payload", payload)
	if s.authToken == "" || s.endpoint == "" {
		log.Println("Skipping emission because we're missing the auth token ",
			"or the endpoint, payload would have been", payload.String())
		return
	}
	serialized, err := proto.Marshal(payload)
	if err != nil {
		log.Println("Failed to serailize payload", *payload)
		return
	}

	req, err := http.NewRequest("POST", s.endpoint, bytes.NewBuffer(serialized))
	if err != nil {
		log.Println("Failed to create a request to endpoint", s.endpoint)
		return
	}
	req.Header.Set("X-SF-TOKEN", s.authToken)
	req.Header.Set("Content-Type", "application/x-protobuf")

	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to complete POST", err)
		return
	}

	defer rsp.Body.Close()
	log.Println("status", rsp.Status)
	log.Println("headers", rsp.Header)
	body, _ := ioutil.ReadAll(rsp.Body)
	log.Println("body", string(body))
}
