package collector

import (
	"fullerite/config"
	"fullerite/metric"

	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
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

func procStatusPoint(name string, value float64, pid string, processName string) (m metric.Metric) {
	m = metric.New(name)
	m.Value = value
	m.AddDimension("collector", "fullerite")
	m.AddDimension("pid", pid)
	m.AddDimension("processName", processName)
	return m
}

func (f ProcStatus) getMetrics(pid string) []metric.Metric {
	// Read from /proc/<pid>/status
	contents, err := ioutil.ReadFile(path.Join("/proc", pid, "status"))
	if err != nil {
		f.log.Warn("Error while getting process stats: ", err)
		return nil
	}

	ret := []metric.Metric{}

	// Parse file into fields
	fields := make(map[string][]string)
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		field := strings.Fields(line)
		fields[field[0]] = field
	}

	// Gather dimensions
	processName := fields["Name:"][1]

	if field, ok := fields["VmSize:"]; ok && len(field) > 1 {
		value, err := strconv.ParseFloat(field[1], 64)
		if err != nil {
			f.log.Warn("Error while reading VmSize: ", err)
		} else {
			// VmSize is in hardcoded to be kiB
			// http://unix.stackexchange.com/questions/199482/does-proc-pid-status-always-use-kb
			m := procStatusPoint("VmSize", value*1024, pid, processName)
			ret = append(ret, m)
		}
	}
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
