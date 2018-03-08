// Config file: ProcNetUDPStats.conf
// Example: {
//	 "localAddressWhitelist": "7F000001:613|*:4ED6",
//	 "remoteAddressWhitelist": "0A000005:50"
// }

package collector

import (
	"fmt"
	"fullerite/metric"
	"fullerite/util"
	"os/exec"
	"regexp"
	"strings"

	l "github.com/Sirupsen/logrus"
)

// Most of these fields are internal kernel values and we don't care about them
type procNetUpdLine struct {
	sl            string
	localAddress  string // hex encoded
	remoteAddress string // hex encoded
	st            string
	queues        string // "tx_queue:rx_queue": outgoing and incoming data queue in terms of kernel memory usage
	trRexmits     string // not used by UDP
	tmWhen        string // not used by UDP
	uid           string
	timeout       string
	inode         string
	ref           string
	pointer       string
	drops         string
}

// ProcNetUDPStats Collector to record udp stats
type ProcNetUDPStats struct {
	baseCollector
	localAddressWhitelist  *regexp.Regexp
	remoteAddressWhitelist *regexp.Regexp
}

func init() {
	RegisterCollector("ProcNetUDPStats", newProcNetUDPStats)
}

func newProcNetUDPStats(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	s := new(ProcNetUDPStats)
	s.log = log
	s.channel = channel
	s.interval = initialInterval
	s.name = "ProcNetUDPStats"

	return s
}

// Configure Override *baseCollector.Configure()
func (s *ProcNetUDPStats) Configure(configMap map[string]interface{}) {
	s.configureCommonParams(configMap)

	if whitelist, exists := configMap["localAddressWhitelist"]; exists {
		localRex, err := regexp.Compile(whitelist.(string))
		if err == nil {
			s.localAddressWhitelist = localRex
		} else {
			s.log.Warn(fmt.Sprintf("Failed to compile regex %s. Error: ", whitelist.(string), err))
		}
	}

	if whitelist, exists := configMap["remoteAddressWhitelist"]; exists {
		remoteRex, err := regexp.Compile(whitelist.(string))
		if err == nil {
			s.remoteAddressWhitelist = remoteRex
		} else {
			s.log.Warn(fmt.Sprintf("Failed to compile regex %s. Error: ", whitelist.(string), err))
		}
	}

	if s.localAddressWhitelist == nil && s.remoteAddressWhitelist == nil {
		s.log.Warn("No whitelist provided, no metric will be emitted.")
	}
}

func (s *ProcNetUDPStats) Collect() {
	if s.localAddressWhitelist == nil && s.remoteAddressWhitelist == nil {
		return
	}

	for _, stat := range s.getProcNetUDPStats() {
		if s.localAddressWhitelist != nil && s.localAddressWhitelist.MatchString(stat.localAddress) {
			s.Channel() <- s.createMetric(
				"upd.drops",
				util.StrToFloat(stat.drops),
				map[string]string{"local_address": stat.localAddress},
			)
		}
		if s.remoteAddressWhitelist != nil && s.remoteAddressWhitelist.MatchString(stat.remoteAddress) {
			s.Channel() <- s.createMetric(
				"upd.drops",
				util.StrToFloat(stat.drops),
				map[string]string{"remote_address": stat.remoteAddress},
			)
		}
	}
}

func (s *ProcNetUDPStats) createMetric(name string, value float64, dims map[string]string) metric.Metric {
	m := metric.New("udp.drops")
	m.Value = value
	m.MetricType = metric.CumulativeCounter
	m.AddDimensions(dims)
	return m
}

func (s *ProcNetUDPStats) getProcNetUDPStats() []procNetUpdLine {
	cmdLine := []string{
		"cat",
		"/proc/net/udp",
	}

	out := s.runCommand(cmdLine)

	if out == nil {
		return nil
	}

	return s.parseProcNetUDPLines(string(out))
}

func (s *ProcNetUDPStats) runCommand(cmdLine []string) []byte {
	var out []byte
	var err error

	cmd := exec.Command(cmdLine[0], cmdLine[1:]...)

	if out, err = (*exec.Cmd).Output(cmd); err != nil {
		s.log.Error(err.Error())
		return nil
	}

	return out
}

func (s *ProcNetUDPStats) parseProcNetUDPLines(out string) []procNetUpdLine {
	raw := strings.Trim(out, "\n")
	lines := strings.Split(raw, "\n")
	stats := []procNetUpdLine{}

	for idx, line := range lines {
		// The first line contains the column titles
		if idx > 0 {
			parts := strings.Fields(line)
			stats = append(stats, procNetUpdLine{
				sl:            parts[0],
				localAddress:  parts[1],
				remoteAddress: parts[2],
				st:            parts[3],
				queues:        parts[4],
				trRexmits:     parts[5],
				tmWhen:        parts[6],
				uid:           parts[7],
				timeout:       parts[8],
				inode:         parts[9],
				ref:           parts[10],
				pointer:       parts[11],
				drops:         parts[12],
			})
		}
	}

	return stats
}
