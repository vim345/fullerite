package collector

import (
	"fullerite/config"
	"fullerite/metric"

	"os/exec"
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

func (f ProcStatus) getMetrics(pid string) []metric.Metric {
	i, err := strconv.Atoi(pid)
	if err != nil {
		f.log.Warnf("Error parsing pid %s: %v", pid, err)
		return nil
	}

	proc, err := procfs.NewProc(i)
	if err != nil {
		f.log.Warn("Error creating Proc: ", err)
		return nil
	}

	stat, err := proc.NewStat()
	if err != nil {
		f.log.Warn("Error getting stats: ", err)
		return nil
	}

	dim := map[string]string{
		"processName": stat.Comm,
		"pid":         pid,
	}
	ret := []metric.Metric{}

	m := procStatusPoint("VirtualMemory", float64(stat.VirtualMemory()), dim)
	ret = append(ret, m)

	return ret
}

func (f ProcStatus) procStatusMetrics() []metric.Metric {
	// Get pids
	c := exec.Command("pgrep", f.processName)
	out, err := c.Output()
	if err != nil {
		f.log.Warn("Error while getting process ids: ", err)
		return nil
	}

	ret := []metric.Metric{}

	pids := strings.Split(string(out), "\n")
	for _, pid := range pids {
		if len(pid) == 0 {
			continue
		}

		ret = append(ret, f.getMetrics(pid)...)
	}
	return ret
}
