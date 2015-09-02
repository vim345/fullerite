package collector

import (
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

// Collect produces some random test metrics.
func (ps ProcStatus) Collect() {
	for _, m := range ps.procStatusMetrics() {
		ps.Channel() <- m
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

func (ps ProcStatus) getMetrics(proc procfs.Proc) []metric.Metric {
	stat, err := proc.NewStat()
	if err != nil {
		ps.log.Warn("Error getting stats: ", err)
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

func (ps ProcStatus) procStatusMetrics() []metric.Metric {
	procs, err := procfs.AllProcs()
	if err != nil {
		ps.log.Warn("Error getting processes: ", err)
		return nil
	}

	ret := []metric.Metric{}

	for _, proc := range procs {
		cmd, err := proc.CmdLine()
		if err != nil {
			ps.log.Warn("Error getting command line: ", err)
			continue
		}

		if len(ps.processName) == 0 || len(cmd) > 0 && strings.Contains(cmd[0], ps.processName) {
			ret = append(ret, ps.getMetrics(proc)...)
		}
	}

	return ret
}
