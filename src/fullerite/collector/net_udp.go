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

// The full /proc/net/udp struct contains all the following fields,
// however we only load the ones we care about for now
//
// sl            kernel hash slot for the socket
// localAddress  local address and port number pair -- hex encoded
// remoteAddress remote address and port number pair (if connected) -- hex encoded
// st            internal status of the socket
// queues        outgoing and incoming data queue in terms of kernel memory usage
// trRexmits     not used by UDP
// tmWhen        not used by UDP
// uid           effective UID of the creator of the socket
// timeout       socket timeout
// inode         inode
// ref           internal kernel field
// pointer       internal kernel field
// drops         packets dropped since the socket was created

type procNetUpdLine struct {
	localAddress  string // hex encoded
	remoteAddress string // hex encoded
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
				localAddress:  parts[1],
				remoteAddress: parts[2],
				drops:         parts[12],
			})
		}
	}

	return stats
}
