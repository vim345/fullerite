package collector

import (
	"fullerite/metric"

	"regexp"

	l "github.com/Sirupsen/logrus"
)

// ProcStatus collector type
type ProcStatus struct {
	baseCollector
	compiledRegex map[string]*regexp.Regexp
	processName   string
}

// ProcessName returns ProcStatus collectors process name
func (ps ProcStatus) ProcessName() string {
	return ps.processName
}

// NewProcStatus creates a new Test collector.
func NewProcStatus(channel chan metric.Metric, initialInterval int, log *l.Entry) *ProcStatus {
	ps := new(ProcStatus)

	ps.log = log
	ps.channel = channel
	ps.interval = initialInterval

	ps.name = "ProcStatus"
	ps.processName = ""
	ps.compiledRegex = make(map[string]*regexp.Regexp)

	return ps
}

// Configure this takes a dictionary of values with which the handler can configure itself
func (ps *ProcStatus) Configure(configMap map[string]interface{}) {
	if processName, exists := configMap["processName"]; exists == true {
		ps.processName = processName.(string)
	}

	if generatedDimensions, exists := configMap["generatedDimensions"]; exists {
		for dimension, generator := range generatedDimensions.(map[string]string) {
			//don't use MustCompile otherwise program will panic due to misformated regex
			re, err := regexp.Compile(generator)
			if err != nil {
				ps.log.Warn("Failed to compile regex: ", generator, err)
			} else {
				ps.compiledRegex[dimension] = re
			}
		}
	}

	ps.configureCommonParams(configMap)
}
