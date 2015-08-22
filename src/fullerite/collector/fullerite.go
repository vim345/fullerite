package collector

import (
	"fullerite/config"
	"fullerite/metric"

	"runtime"

	"github.com/Sirupsen/logrus"
)

// Fullerite collector type
type Fullerite struct {
	BaseCollector
}

// NewTest creates a new Test collector.
func NewFullerite() *Fullerite {
	f := new(Fullerite)
	f.name = "Fullerite"
	f.log = logrus.WithFields(logrus.Fields{"app": "fullerite", "pkg": "collector", "collector": "fullerite"})
	f.channel = make(chan metric.Metric)
	f.interval = DefaultCollectionInterval
	return f
}

// Configure this takes a dictionary of values with which the handler can configure itself
func (f *Fullerite) Configure(configMap map[string]interface{}) {
	if interval, exists := configMap["interval"]; exists == true {
		f.interval = config.GetAsInt(interval, DefaultCollectionInterval)
	}
}

// Collect produces some random test metrics.
func (f Fullerite) Collect() {
	for _, m := range golangMetrics() {
		f.Channel() <- m
	}
}

func point(name string, value float64) (m metric.Metric) {
	m = metric.New(name)
	m.Value = value
	m.AddDimension("collector", "fullerite")
	return m
}

func pointc(name string, value float64) (m metric.Metric) {
	m = metric.New(name)
	m.MetricType = metric.Counter
	m.Value = value
	m.AddDimension("collector", "fullerite")
	return m
}

func golangMetrics() []metric.Metric {
	ret := []metric.Metric{}

	// See https://golang.org/src/runtime/mstats.go?s=3251:5102#L82
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	ret = append(
		ret,
		point("Alloc", float64(m.Alloc)))
	ret = append(
		ret,
		pointc("TotalAlloc", float64(m.TotalAlloc)))
	ret = append(
		ret,
		point("Sys", float64(m.Sys)))
	ret = append(
		ret,
		pointc("Lookups", float64(m.Lookups)))
	ret = append(
		ret,
		pointc("Mallocs", float64(m.Mallocs)))
	ret = append(
		ret,
		pointc("Frees", float64(m.Frees)))
	ret = append(
		ret,
		point("HeapAlloc", float64(m.HeapAlloc)))
	ret = append(
		ret,
		point("HeapSys", float64(m.HeapSys)))
	ret = append(
		ret,
		point("HeapIdle", float64(m.HeapIdle)))
	ret = append(
		ret,
		point("HeapInuse", float64(m.HeapInuse)))
	ret = append(
		ret,
		point("HeapReleased", float64(m.HeapReleased)))
	ret = append(
		ret,
		point("HeapObjects", float64(m.HeapObjects)))
	ret = append(
		ret,
		point("StackInuse", float64(m.StackInuse)))
	ret = append(
		ret,
		point("StackSys", float64(m.StackSys)))
	ret = append(
		ret,
		point("MSpanInuse", float64(m.MSpanInuse)))
	ret = append(
		ret,
		point("MSpanSys", float64(m.MSpanSys)))
	ret = append(
		ret,
		point("MCacheInuse", float64(m.MCacheInuse)))
	ret = append(
		ret,
		point("MCacheSys", float64(m.MCacheSys)))
	ret = append(
		ret,
		point("BuckHashSys", float64(m.BuckHashSys)))
	ret = append(
		ret,
		point("GCSys", float64(m.GCSys)))
	ret = append(
		ret,
		point("OtherSys", float64(m.OtherSys)))
	ret = append(
		ret,
		point("NextGC", float64(m.NextGC)))
	ret = append(
		ret,
		point("LastGC", float64(m.LastGC)))
	ret = append(
		ret,
		pointc("PauseTotalNs", float64(m.PauseTotalNs)))
	ret = append(
		ret,
		pointc("NumGC", float64(m.NumGC)))
	return ret

}
