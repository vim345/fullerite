package handler

import (
	"encoding/json"
	"fullerite/metric"
	"log"
	"time"
)

type signalfxPayload struct {
	Gauges      []signalfxMetric `json:"gauges"`
	Counters    []signalfxMetric `json:"counter"`
	Cumulatives []signalfxMetric `json:"cumulativeCounter"`
}

type signalfxMetric struct {
	Name       string            `json:"metric"`
	Value      float64           `json:"value"`
	Dimensions map[string]string `json:"dimensions"`
}

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

	gauges := make([]signalfxMetric, 0, s.maxBufferSize)
	cumCounters := make([]signalfxMetric, 0, s.maxBufferSize)
	counters := make([]signalfxMetric, 0, s.maxBufferSize)

	for incomingMetric := range s.Channel() {
		log.Println("Processing metric to SignalFx:", incomingMetric)
		sfxVersion := *s.convertToSignalFx(&incomingMetric)

		switch incomingMetric.Type() {
		case metric.Gauge:
			gauges = append(gauges, sfxVersion)
		case metric.CumulativeCounter:
			cumCounters = append(cumCounters, sfxVersion)
		case metric.Counter:
			counters = append(counters, sfxVersion)
		}

		numMetrics := len(gauges) + len(cumCounters) + len(counters)
		if time.Since(lastEmission).Seconds() >= float64(s.interval) || numMetrics >= s.maxBufferSize {
			s.emitMetrics(&gauges, &cumCounters, &counters)
			gauges = make([]signalfxMetric, 0, s.maxBufferSize)
			cumCounters = make([]signalfxMetric, 0, s.maxBufferSize)
			counters = make([]signalfxMetric, 0, s.maxBufferSize)
		}
	}
}

func (s *SignalFx) convertToSignalFx(metric *metric.Metric) *signalfxMetric {
	sfx := new(signalfxMetric)
	sfx.Name = s.Prefix() + metric.Name()
	sfx.Value = metric.Value()
	sfx.Dimensions = make(map[string]string)

	if s.DefaultDimensions() != nil {
		for _, dimension := range *s.DefaultDimensions() {
			sfx.Dimensions[dimension.Name()] = dimension.Value()
		}
	}

	for _, dimension := range *metric.Dimensions() {
		sfx.Dimensions[dimension.Name()] = dimension.Value()
	}

	return sfx
}

func (s *SignalFx) emitMetrics(gauges *[]signalfxMetric, cumCounters *[]signalfxMetric, counters *[]signalfxMetric) {
	log.Println("Starting to emit ", len(*gauges), "gauges",
		len(*counters), "counters", len(*cumCounters), "cumulative counters")

	payload := new(signalfxPayload)
	payload.Gauges = *gauges
	payload.Counters = *counters
	payload.Cumulatives = *cumCounters

	asjson, err := json.Marshal(*payload)
	if err != nil {
		log.Println("error occurred while marshaling counters", *counters,
			"gauges", *gauges, "cum counters", *cumCounters)
	}
	log.Println("going to send", string(asjson))
}
