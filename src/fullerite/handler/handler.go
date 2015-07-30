package handler

import (
	"fullerite/metric"
	"log"
)

// Some sane values to default things to
const (
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
	Configure(*map[string]string)
	Name() string
	Channel() chan metric.Metric

	MaxBufferSize() int
	SetMaxBufferSize(int)

	Prefix() string
	SetPrefix(string)

	DefaultDimensions() *[]metric.Dimension
	SetDefaultDimensions(*[]metric.Dimension)
}

// BaseHandler is class to handle the boiler plate parts of the handlers
type BaseHandler struct {
	channel           chan metric.Metric
	name              string
	maxBufferSize     int
	prefix            string
	defaultDimensions []metric.Dimension
}

// Channel : the channel to handler listens for metrics on
func (handler *BaseHandler) Channel() chan metric.Metric {
	return handler.channel
}

// Name : the name of the handler
func (handler *BaseHandler) Name() string {
	return handler.name
}

// MaxBufferSize : the maximum number of metrics that should be buffered before sending
func (handler *BaseHandler) MaxBufferSize() int {
	return handler.maxBufferSize
}

// SetMaxBufferSize : set the buffer size
func (handler *BaseHandler) SetMaxBufferSize(size int) {
	handler.maxBufferSize = size
}

// Prefix : any prefix that should be applied to the metrics name as they're sent
// it is appended without any punctuation, include your own
func (handler *BaseHandler) Prefix() string {
	return handler.prefix
}

// SetPrefix : set the prefix
func (handler *BaseHandler) SetPrefix(prefix string) {
	handler.prefix = prefix
}

// DefaultDimensions : dimensions that should be included in any metric
func (handler *BaseHandler) DefaultDimensions() *[]metric.Dimension {
	return &handler.defaultDimensions
}

// SetDefaultDimensions : set the defautl dimensions
func (handler *BaseHandler) SetDefaultDimensions(defaults *[]metric.Dimension) {
	handler.defaultDimensions = *defaults
}

// Configure : this takes a dictionary of values with which the handler can configure itself
func (handler *BaseHandler) Configure(*map[string]string) {
	// noop
}

// String returns the handler name in a printable format.
func (handler *BaseHandler) String() string {
	return handler.Name() + "Handler"
}
