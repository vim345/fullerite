package handler

import (
	"fullerite/config"
	"fullerite/metric"

	"container/list"
	"time"

	l "github.com/Sirupsen/logrus"
)

// Some sane values to default things to
const (
	DefaultBufferSize = 100
	DefaultTimeoutSec = 2
	DefaultInterval   = 10
)

var defaultLog = l.WithFields(l.Fields{"app": "fullerite", "pkg": "handler"})

// New creates a new Handler based on the requested handler name.
func New(name string) Handler {
	var base Handler

	channel := make(chan metric.Metric)
	handlerLog := defaultLog.WithFields(l.Fields{"handler": name})
	timeout := time.Duration(DefaultTimeoutSec * time.Second)

	switch name {
	case "Graphite":
		base = NewGraphite(channel, DefaultInterval, DefaultBufferSize, timeout, handlerLog)
	case "SignalFx":
		base = NewSignalFx(channel, DefaultInterval, DefaultBufferSize, timeout, handlerLog)
	case "Datadog":
		base = NewDatadog(channel, DefaultInterval, DefaultBufferSize, timeout, handlerLog)
	case "Kairos":
		base = NewKairos(channel, DefaultInterval, DefaultBufferSize, timeout, handlerLog)
	default:
		defaultLog.Error("Cannot create handler ", name)
		return nil
	}
	return base
}

// InternalMetrics holds the key:value pairs for counters/gauges
type InternalMetrics struct {
	Counters map[string]float64
	Gauges   map[string]float64
}

// NewInternalMetrics initializes the internal components of InternalMetrics
func NewInternalMetrics() *InternalMetrics {
	inst := new(InternalMetrics)
	inst.Counters = make(map[string]float64)
	inst.Gauges = make(map[string]float64)
	return inst
}

// Handler defines the interface of a generic handler.
type Handler interface {
	Run()
	Configure(map[string]interface{})

	// InternalMetrics is to publish a set of values
	// that are relevant to the handler itself.
	InternalMetrics() InternalMetrics

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

type emissionTiming struct {
	timestamp time.Time
	duration  time.Duration
}

// BaseHandler is class to handle the boiler plate parts of the handlers
type BaseHandler struct {
	channel           chan metric.Metric
	name              string
	prefix            string
	defaultDimensions map[string]string
	log               *l.Entry

	interval      int
	maxBufferSize int
	timeout       time.Duration

	// for tracking
	emissionTimes  list.List
	totalEmissions uint64
	metricsSent    uint64
	metricsDropped uint64
}

// SetMaxBufferSize : set the buffer size
func (base *BaseHandler) SetMaxBufferSize(size int) {
	base.maxBufferSize = size
}

// SetInterval : set the interval
func (base *BaseHandler) SetInterval(val int) {
	base.interval = val
}

// SetPrefix : any prefix that should be applied to the metrics name as they're sent
// it is appended without any punctuation, include your own
func (base *BaseHandler) SetPrefix(prefix string) {
	base.prefix = prefix
}

// SetDefaultDimensions : set the defautl dimensions
func (base *BaseHandler) SetDefaultDimensions(defaults map[string]string) {
	base.defaultDimensions = make(map[string]string)
	for name, value := range defaults {
		base.defaultDimensions[name] = value
	}
}

// Channel : the channel to handler listens for metrics on
func (base BaseHandler) Channel() chan metric.Metric {
	return base.channel
}

// Name : the name of the handler
func (base BaseHandler) Name() string {
	return base.name
}

// MaxBufferSize : the maximum number of metrics that should be buffered before sending
func (base BaseHandler) MaxBufferSize() int {
	return base.maxBufferSize
}

// Prefix : the prefix (with punctuation) to use on each emitted metric
func (base BaseHandler) Prefix() string {
	return base.prefix
}

// DefaultDimensions : dimensions that should be included in any metric
func (base BaseHandler) DefaultDimensions() map[string]string {
	return base.defaultDimensions
}

// Interval : the maximum interval that the handler should buffer stats for
func (base BaseHandler) Interval() int {
	return base.interval
}

// String returns the handler name in a printable format.
func (base BaseHandler) String() string {
	return base.name + "Handler"
}

// InternalMetrics : Returns the internal metrics that are being collected by this handler
func (base BaseHandler) InternalMetrics() InternalMetrics {
	// now we calculate the average emission seconds for
	var totalTime float64
	for e := base.emissionTimes.Front(); e != nil; e = e.Next() {
		totalTime += e.Value.(emissionTiming).duration.Seconds()
	}
	avg := totalTime / float64(base.emissionTimes.Len())

	counters := map[string]float64{
		"totalEmissions": float64(base.totalEmissions),
		"metricsDropped": float64(base.metricsDropped),
		"metricsSent":    float64(base.metricsSent),
	}
	gauges := map[string]float64{
		"averageEmissionTiming": avg,
		"intervalLength":        float64(base.interval),
		"emissionsInWindow":     float64(base.emissionTimes.Len()),
	}

	return InternalMetrics{
		Counters: counters,
		Gauges:   gauges,
	}
}

// manages the rolling window of emissions
// the emissions are a timesorted list, and we purge things older than
// the base handler's interval
func (base *BaseHandler) recordEmission(emissionDuration time.Duration) {
	base.totalEmissions++
	now := time.Now()

	timing := emissionTiming{now, emissionDuration}
	base.emissionTimes.PushBack(timing)

	// now kull the list of old times, iterate through the list until we find
	// a timestamp that is within the interval
	minTime := now.Add(time.Duration(-1*base.interval) * time.Second)
	toRemove := []*list.Element{}
	for e := base.emissionTimes.Front(); e.Value.(emissionTiming).timestamp.Before(minTime); e = e.Next() {
		toRemove = append(toRemove, e)
	}

	for _, entry := range toRemove {
		base.emissionTimes.Remove(entry)
	}
	base.log.Debug("We removed ", len(toRemove), " entries and now have ", base.emissionTimes.Len())
}

// configureCommonParams will extract the common parameters that are used and set them in the handler
func (base *BaseHandler) configureCommonParams(configMap map[string]interface{}) {
	if asInterface, exists := configMap["timeout"]; exists == true {
		timeout := config.GetAsFloat(asInterface, DefaultTimeoutSec)
		base.timeout = time.Duration(timeout) * time.Second
	}

	if asInterface, exists := configMap["max_buffer_size"]; exists == true {
		base.maxBufferSize = config.GetAsInt(asInterface, DefaultBufferSize)
	}

	if asInterface, exists := configMap["interval"]; exists == true {
		base.interval = config.GetAsInt(asInterface, DefaultInterval)
	}
}

func (base *BaseHandler) run(emitFunc func([]metric.Metric) bool) {
	metrics := make([]metric.Metric, 0, base.maxBufferSize)

	lastEmission := time.Now()
	for incomingMetric := range base.Channel() {
		base.log.Debug(base.name, " metric: ", incomingMetric)
		metrics = append(metrics, incomingMetric)

		emitIntervalPassed := time.Since(lastEmission).Seconds() >= float64(base.interval)
		bufferSizeLimitReached := len(metrics) >= base.maxBufferSize

		if emitIntervalPassed || bufferSizeLimitReached {
			beforeEmission := time.Now()
			result := emitFunc(metrics)
			lastEmission = time.Now()

			emissionDuration := lastEmission.Sub(beforeEmission)

			base.log.Info("POST to ", base.name, " took ", emissionDuration.Seconds(), " seconds")
			base.recordEmission(emissionDuration)

			if result {
				base.metricsSent += uint64(len(metrics))
			} else {
				base.metricsDropped += uint64(len(metrics))
			}

			// reset metrics
			metrics = make([]metric.Metric, 0, base.maxBufferSize)
		}
	}
}
