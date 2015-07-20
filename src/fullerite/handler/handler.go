package handler

import (
	"fullerite/metric"
	"log"
)

// Handler defines the interface of a generic handler.
type Handler interface {
	Send()
	Name() string
	Interval() int
	MaxBufferSize() int
	Channel() chan metric.Metric
}

// New creates a new Handler based on the requested handler name.
func New(name string) Handler {
	var handler Handler
	switch name {
	case "Graphite":
		handler = new(Graphite)
	case "SignalFx":
		handler = new(SignalFx)
	default:
		log.Fatal("Cannot create handler ", name)
		return nil
	}
	return handler
}
