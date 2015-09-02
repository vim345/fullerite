// +build linux

package collector

import (
	"fullerite/metric"

	"strconv"
	"strings"

	"github.com/prometheus/procfs"
)

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

	ret = append(ret, procStatusPoint("VirtualMemory", float64(stat.VirtualMemory()), dim))
	ret = append(ret, procStatusPoint("ResidentMemory", float64(stat.ResidentMemory()), dim))

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
