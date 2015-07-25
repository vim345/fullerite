package handler

import (
	"fullerite/metric"
	"log"
)

// Some sane values to default things to
const (
	DefaultInterval   = 10
	DefaultBufferSize = 100
)

// New creates a new Handler based on the requested handler name.
func New(name string) Handler {
	log.Println("Building handler " + name)

	var handler Handler
	switch name {
	case "Graphite":
		handler = NewGraphite()
	case "SignalFx":
		handler = NewSignalFx()
	default:
		log.Fatal("Cannot create handler ", name)
		return nil
	}

	return handler
}

// Handler defines the interface of a generic handler.
type Handler interface {
	Run()
	Name() string
	Interval() int
	MaxBufferSize() int
	SetInterval(int)
	SetMaxBufferSize(int)
	Channel() chan metric.Metric
}

// BaseHandler is class to handle the boiler plate parts of the handlers
type BaseHandler struct {
	channel       chan metric.Metric
	name          string
	interval      int
	maxBufferSize int
}

func (handler BaseHandler) Channel() chan metric.Metric {
	return handler.channel
}

func (handler BaseHandler) Name() string {
	return handler.name
}

func (handler BaseHandler) Interval() int {
	return handler.interval
}

func (handler BaseHandler) SetInterval(interval int) {
	handler.interval = interval
}

func (handler BaseHandler) MaxBufferSize() int {
	return handler.maxBufferSize
}

func (handler BaseHandler) SetMaxBufferSize(size int) {
	handler.maxBufferSize = size
}

// String returns the handler name in a printable format.
func (handler BaseHandler) String() string {
	return handler.Name() + "Handler"
}
