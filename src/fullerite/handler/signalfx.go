package handler

import (
	"fullerite/metric"
	"log"
)

// SignalFx type.
type SignalFx struct {
	BaseHandler
	endpoint  string
	authToken string
}

// NewSignalFx returns a new SignalFx handler.
func NewSignalFx() *SignalFx {

	s := new(SignalFx)
	s.name = "SignalFx"
	s.interval = DefaultInterval
	s.maxBufferSize = DefaultBufferSize
	s.channel = make(chan metric.Metric)
	return s
}

func (s SignalFx) Configure(config *map[string]string) {
	asmap := *config
	s.authToken = asmap["authToken"]
	s.endpoint = asmap["endpoint"]
}

// Run send metrics in the channel to SignalFx.
func (s SignalFx) Run() {
	// TODO: check interval and queue size and metrics.
	for metric := range s.Channel() {
		// TODO: Actually send to signalfx.
		log.Println("Sending metric to SignalFx:", metric)
	}
}
