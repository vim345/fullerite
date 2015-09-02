package collector

import (
	"fullerite/config"
	"fullerite/metric"

	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/procfs"
)

// ProcStatus collector type
type ProcStatus struct {
	BaseCollector
	processName string
}

// ProcessName returns ProcStatus collectors process name
func (f ProcStatus) ProcessName() string {
	return f.processName
}

// NewProcStatus creates a new Test collector.
func NewProcStatus() *ProcStatus {
	f := new(ProcStatus)
	f.name = "ProcStatus"
	f.log = logrus.WithFields(logrus.Fields{"app": "fullerite", "pkg": "collector", "collector": "ProcStatus"})
	f.channel = make(chan metric.Metric)
	f.interval = DefaultCollectionInterval
	f.processName = ""
	return f
}

// Configure this takes a dictionary of values with which the handler can configure itself
func (f *ProcStatus) Configure(configMap map[string]interface{}) {
	if interval, exists := configMap["interval"]; exists == true {
		f.interval = config.GetAsInt(interval, DefaultCollectionInterval)
	}
	if processName, exists := configMap["processName"]; exists == true {
		f.processName = processName.(string)
	}
}

// Collect produces some random test metrics.
func (f ProcStatus) Collect() {
	for _, m := range f.procStatusMetrics() {
		f.Channel() <- m
	}
}

func procStatusPoint(name string, value float64, dimensions map[string]string) (m metric.Metric) {
	m = metric.New(name)
	m.Value = value
	m.AddDimension("collector", "fullerite")
	for k, v := range dimensions {
		m.AddDimension(k, v)
	}
	return m
}

func (f ProcStatus) getMetrics(proc procfs.Proc) []metric.Metric {
	stat, err := proc.NewStat()
	if err != nil {
		f.log.Warn("Error getting stats: ", err)
		return nil
	}

	dim := map[string]string{
		"processName": stat.Comm,
		"pid":         strconv.Itoa(stat.PID),
	}

	ret := []metric.Metric{}

	m := procStatusPoint("VirtualMemory", float64(stat.VirtualMemory()), dim)
	ret = append(ret, m)

	return ret
}

func (f ProcStatus) procStatusMetrics() []metric.Metric {
	procs, err := procfs.AllProcs()
	if err != nil {
		f.log.Warn("Error getting processes: ", err)
		return nil
	}

	ret := []metric.Metric{}

	for _, proc := range procs {
		cmd, err := proc.CmdLine()
		if err != nil {
			f.log.Warn("Error getting command line: ", err)
			continue
		}

		if len(f.processName) == 0 || len(cmd) > 0 && strings.Contains(cmd[0], f.processName) {
			ret = append(ret, f.getMetrics(proc)...)
		}
	}

	return ret
}
