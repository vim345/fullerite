package collector

import (
	"fullerite/metric"

	"math/rand"

	"github.com/Sirupsen/logrus"
)

// Test collector type
type Test struct {
	BaseCollector
	metricName string
}

// NewTest creates a new Test collector.
func NewTest() *Test {
	t := new(Test)
	t.name = "Test"
	t.log = logrus.WithFields(logrus.Fields{"app": "fullerite", "pkg": "collector", "collector": "Test"})
	t.channel = make(chan metric.Metric)
	t.interval = DefaultCollectionInterval
	t.metricName = "TestMetric"
	return t
}

// Configure this takes a dictionary of values with which the handler can configure itself
func (t *Test) Configure(configMap map[string]interface{}) {
	if metricName, exists := configMap["metricName"]; exists == true {
		t.metricName = metricName.(string)
	}
	t.configureCommonParams(configMap)
}

// Collect produces some random test metrics.
func (t Test) Collect() {
	metric := metric.New(t.metricName)
	metric.Value = rand.Float64()
	metric.AddDimension("testing", "yes")
	t.Channel() <- metric
	t.log.Debug(metric)
}
