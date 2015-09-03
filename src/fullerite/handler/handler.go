package handler

import (
	"fullerite/config"
	"fullerite/metric"
	"time"

	"github.com/Sirupsen/logrus"
)

// Some sane values to default things to
const (
	DefaultBufferSize = 100
	DefaultTimeoutSec = 2
	DefaultInterval   = 10
)

var defaultLog = logrus.WithFields(logrus.Fields{"app": "fullerite", "pkg": "handler"})

// New creates a new Handler based on the requested handler name.
func New(name string) Handler {
	var handler Handler
	switch name {
	case "Graphite":
		handler = NewGraphite()
	case "SignalFx":
		handler = NewSignalFx()
	case "Datadog":
		handler = NewDatadog()
	case "Kairos":
		handler = NewKairos()
	default:
		defaultLog.Error("Cannot create handler ", name)
		return nil
	}
	return handler
}

// Handler defines the interface of a generic handler.
type Handler interface {
	Run()
	Configure(map[string]interface{})

	// taken care of by the base
	Name() string
	String() string
	Channel() chan metric.Metric

	Interval() int
	SetInterval(int)

	MaxBufferSize() int
	SetMaxBufferSize(int)

	Prefix() string
	SetPrefix(string)

	DefaultDimensions() map[string]string
	SetDefaultDimensions(map[string]string)
}

// BaseHandler is class to handle the boiler plate parts of the handlers
type BaseHandler struct {
	channel           chan metric.Metric
	name              string
	maxBufferSize     int
	prefix            string
	timeout           time.Duration
	interval          int
	source            string
	defaultDimensions map[string]string
	emissionTimes     []float64
	metricsSent       int
	metricsDropped    int
	log               *logrus.Entry
}

// configureCommonParams will extract the common parameters that are used and set them in the handler
func (handler *BaseHandler) configureCommonParams(configMap map[string]interface{}) {
	if asInterface, exists := configMap["timeout"]; exists == true {
		timeout := config.GetAsFloat(asInterface, DefaultTimeoutSec)
		handler.timeout = time.Duration(timeout) * time.Second
	}

	if asInterface, exists := configMap["max_buffer_size"]; exists == true {
		handler.maxBufferSize = config.GetAsInt(asInterface, DefaultBufferSize)
	}

	if asInterface, exists := configMap["interval"]; exists == true {
		handler.interval = config.GetAsInt(asInterface, DefaultInterval)
	}
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
func (handler *BaseHandler) DefaultDimensions() map[string]string {
	return handler.defaultDimensions
}

// SetDefaultDimensions : set the defautl dimensions
func (handler *BaseHandler) SetDefaultDimensions(defaults map[string]string) {
	handler.defaultDimensions = make(map[string]string)
	for name, value := range defaults {
		handler.defaultDimensions[name] = value
	}
}

// Interval : the maximum interval that the handler should buffer stats for
func (handler *BaseHandler) Interval() int {
	return handler.interval
}

// SetInterval : set the interval
func (handler *BaseHandler) SetInterval(val int) {
	handler.interval = val
}

// String returns the handler name in a printable format.
func (handler *BaseHandler) String() string {
	return handler.name + "Handler"
}

func (handler *BaseHandler) makeEmissionTimeMetric() metric.Metric {
	value := 0.0
	for _, v := range handler.emissionTimes {
		value += v
	}
	m := metric.New("HandlerEmitTiming")
	m.Value = value / float64(len(handler.emissionTimes))
	m.AddDimension("handler", handler.name)
	return m
}

func (handler *BaseHandler) resetEmissionTimes() {
	handler.emissionTimes = make([]float64, 0)
}

func (handler *BaseHandler) makeMetricsSentMetric() metric.Metric {
	m := metric.New("MetricsSent")
	m.Value = float64(handler.metricsSent)
	m.AddDimension("handler", handler.name)
	return m
}

func (handler *BaseHandler) resetMetricsSent() {
	handler.metricsSent = 0
}

func (handler *BaseHandler) makeMetricsDroppedMetric() metric.Metric {
	m := metric.New("MetricsDropped")
	m.Value = float64(handler.metricsDropped)
	m.AddDimension("handler", handler.name)
	return m
}

func (handler *BaseHandler) resetMetricsDropped() {
	handler.metricsDropped = 0
}
