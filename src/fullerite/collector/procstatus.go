package collector

import (
	"fullerite/metric"

	"github.com/Sirupsen/logrus"
)

// ProcStatus collector type
type ProcStatus struct {
	BaseCollector
	processName string
}

// ProcessName returns ProcStatus collectors process name
func (ps ProcStatus) ProcessName() string {
	return ps.processName
}

// NewProcStatus creates a new Test collector.
func NewProcStatus() *ProcStatus {
	ps := new(ProcStatus)
	ps.name = "ProcStatus"
	ps.log = logrus.WithFields(logrus.Fields{"app": "fullerite", "pkg": "collector", "collector": "ProcStatus"})
	ps.channel = make(chan metric.Metric)
	ps.interval = DefaultCollectionInterval
	ps.processName = ""
	return ps
}

// Configure this takes a dictionary of values with which the handler can configure itself
func (ps *ProcStatus) Configure(configMap map[string]interface{}) {
	if processName, exists := configMap["processName"]; exists == true {
		ps.processName = processName.(string)
	}
	ps.configureCommonParams(configMap)
}
