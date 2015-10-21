// +build linux

package collector

import (
	"fullerite/metric"

	"regexp"
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
	m.AddDimensions(dimensions)
	return m
}

func (ps ProcStatus) getMetrics(proc procfs.Proc) []metric.Metric {
	stat, err := proc.NewStat()
	if err != nil {
		ps.log.Warn("Error getting stats: ", err)
		return nil
	}

	pid := strconv.Itoa(stat.PID)
	dim := map[string]string{
		"processName": stat.Comm,
		"pid":         pid,
	}

	for dimension, generator := range ps.generatedDimensions {
		if len(generator) != 2 {
			continue
		}

		procAttribute := generator[0]
		procRegex := generator[1]

		var err interface{}
		var cmdOutput []string

		switch procAttribute {

		case "cmdline":
			cmdOutput, err = proc.CmdLine()
		case "filedescriptors":
			cmdOutput, err = proc.FileDescriptorTargets()
		}

		if err != nil {
			ps.log.Warn("Error getting generated dimensions: ", dimension, generator, err)
			continue
		}

		if len(cmdOutput) > 0 {
			//don't use MustCompile otherwise program will panic due to misformated regex
			re, err := regexp.Compile(procRegex)
			if err != nil {
				continue
			}

			subMatch := re.FindStringSubmatch(cmdOutput[0])
			if len(subMatch) > 1 {
				dim[dimension] = subMatch[1]
			}
		}
	}

	ret := []metric.Metric{
		procStatusPoint("VirtualMemory", float64(stat.VirtualMemory()), dim),
		procStatusPoint("ResidentMemory", float64(stat.ResidentMemory()), dim),
		procStatusPoint("CPUTime", float64(stat.CPUTime()), dim),
	}

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
