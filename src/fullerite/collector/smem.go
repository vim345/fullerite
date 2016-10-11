// Requirements: smem (https://www.selenic.com/smem/).
//
// Config file: SmemStats.conf
// Example: {
//   "user": "some-user", <-- This user should be able to access the /proc/<pid>/smaps files listed in the whitelist below
//   "procsWhitelist": "apache2|tmux",
//   "smemPath": "/usr/bin/smem", <-- Path to the smem executable
//   "dimensionsFromCmdline": {"worker_id": "apache worker ([0-9]+)"},
//   "dimensionsFromEnv": {"env_var_1": "ENV_VAR_1", "env_var_2": "ENV_VAR_2"}, <-- Environment variables gotten from /proc/<pid>/environ
// }

package collector

import (
	"fmt"
	"fullerite/config"
	"fullerite/metric"
	"fullerite/util"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	l "github.com/Sirupsen/logrus"
)

type smemStatLine struct {
	proc string
	pss  float64
	uss  float64
	rss  float64
	vss  float64
	pid  int
}

// SmemStats Collector to record smem stats
type SmemStats struct {
	baseCollector
	user                  string
	whitelistedProcs      string
	smemPath              string
	whitelistedMetrics    []string
	dimensionsFromCmdline map[string]string
	dimensionsFromEnv     map[string]string
}

var (
	requiredConfigs      = []string{"user", "procsWhitelist"}
	execCommand          = exec.Command
	commandOutput        = (*exec.Cmd).Output
	getSmemStats         = (*SmemStats).getSmemStats
	getCmdLineDimensions = (*SmemStats).getCmdLineDimensions
	getEnvDimensions     = (*SmemStats).getEnvDimensions
	readCmdline          = (*SmemStats).readCmdline
	allMetrics           = []string{"rss", "vss", "pss", "uss"}
)

func init() {
	RegisterCollector("SmemStats", newSmemStats)
}

func newSmemStats(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	s := new(SmemStats)
	s.log = log
	s.channel = channel
	s.interval = initialInterval
	s.name = "SmemStats"

	return s
}

// Configure Override *baseCollector.Configure(); will fetch the whitelisted processes
func (s *SmemStats) Configure(configMap map[string]interface{}) {
	s.configureCommonParams(configMap)

	if user, exists := configMap["user"]; exists {
		s.user = user.(string)
	} else {
		s.log.Warn("Required config does not exist for SmemStats: user")
	}

	if whitelist, exists := configMap["procsWhitelist"]; exists {
		s.whitelistedProcs = whitelist.(string)
	} else {
		s.log.Warn("Required config does not exist for SmemStats: procsWhitelist")
	}

	if smemPath, exists := configMap["smemPath"]; exists {
		s.smemPath = smemPath.(string)
	} else {
		s.log.Warn("Required config does not exist for SmemStats: smemPath")
	}

	if blacklist, exists := configMap["metricsBlacklist"]; exists {
		s.whitelistedMetrics = getWhitelistedMetrics(config.GetAsSlice(blacklist))
	} else {
		s.whitelistedMetrics = allMetrics
	}

	if dimensionsFromCmdline, exists := configMap["dimensionsFromCmdline"]; exists {
		s.dimensionsFromCmdline = config.GetAsMap(dimensionsFromCmdline)
	}

	if dimensionsFromEnv, exists := configMap["dimensionsFromEnv"]; exists {
		s.dimensionsFromEnv = config.GetAsMap(dimensionsFromEnv)
	}
}

// Collect calls smem periodically
func (s *SmemStats) Collect() {
	if s.whitelistedProcs == "" || s.user == "" || s.smemPath == "" {
		return
	}

	for _, stat := range getSmemStats(s) {
		dims := s.getCustomDimensions(stat.pid)
		for _, element := range s.whitelistedMetrics {
			var m metric.Metric
			switch element {
			case "pss":
				m = metric.WithValue(stat.proc+".smem.pss", stat.pss)
			case "uss":
				m = metric.WithValue(stat.proc+".smem.uss", stat.uss)
			case "vss":
				m = metric.WithValue(stat.proc+".smem.vss", stat.vss)
			case "rss":
				m = metric.WithValue(stat.proc+".smem.rss", stat.rss)
			}
			m.AddDimensions(dims)
			s.Channel() <- m
		}
	}
}

func (s *SmemStats) getCustomDimensions(pid int) map[string]string {
	dims := getEnvDimensions(s, pid)

	for k, v := range getCmdLineDimensions(s, pid) {
		dims[k] = v
	}

	return dims
}

func (s *SmemStats) readCmdline(fileName string) ([]byte, error) {
	return ioutil.ReadFile(fileName)
}

func (s *SmemStats) getEnvDimensions(pid int) map[string]string {
	dims := make(map[string]string)

	if pid == 0 || len(s.dimensionsFromEnv) == 0 {
		return dims
	}

	environ := s.getEnviron(pid)

	if environ == "" {
		return dims
	}

	for dimName, envVar := range s.dimensionsFromEnv {
		pattern, _ := regexp.Compile(fmt.Sprintf(`(%s=)(.*?)(\000)`, envVar))
		matches := pattern.FindStringSubmatch(string(environ))

		if len(matches) > 1 {
			dims[dimName] = string(matches[2])
		}
	}

	return dims
}

func (s *SmemStats) getCmdLineDimensions(pid int) map[string]string {
	dimensions := make(map[string]string)
	if pid != 0 {
		for name, rexStr := range s.dimensionsFromCmdline {
			data, err := readCmdline(s, fmt.Sprintf("/proc/%d/cmdline", pid))
			if err != nil {
				s.log.Warn(err.Error())
			} else {
				rex, _ := regexp.Compile(rexStr)
				match := rex.FindStringSubmatch(string(data))
				if len(match) > 1 {
					dimensions[name] = string(match[1])
				}
			}
		}
	}
	return dimensions
}

func (s *SmemStats) getSmemStats() []smemStatLine {
	cmdLine := []string{
		"/usr/bin/sudo",
		"-u", s.user,
		s.smemPath,
		"-P", s.whitelistedProcs,
		"-c", "pss uss rss vss name pid"}

	out := s.runCommand(cmdLine)

	if out == nil {
		return nil
	}

	return s.parseSmemLines(string(out))
}

func (s *SmemStats) getEnviron(pid int) string {
	cmdLine := []string{
		"/usr/bin/sudo",
		"-u", s.user,
		"/bin/cat",
		fmt.Sprintf("/proc/%d/environ", pid),
	}

	environ := s.runCommand(cmdLine)

	if environ == nil {
		return ""
	}

	return string(environ)
}

func (s *SmemStats) runCommand(cmdLine []string) []byte {
	var out []byte
	var err error

	cmd := execCommand(cmdLine[0], cmdLine[1:]...)

	if out, err = commandOutput(cmd); err != nil {
		s.log.Error(err.Error())
		return nil
	}

	return out
}

func (s *SmemStats) parseSmemLines(out string) []smemStatLine {
	raw := strings.Trim(out, "\n")
	lines := strings.Split(raw, "\n")
	stats := []smemStatLine{}

	for _, line := range lines {
		parts := strings.Fields(line)
		pid, _ := strconv.Atoi(parts[5])
		stats = append(stats, smemStatLine{
			proc: parts[4],
			pss:  util.StrToFloat(parts[0]),
			uss:  util.StrToFloat(parts[1]),
			rss:  util.StrToFloat(parts[2]),
			vss:  util.StrToFloat(parts[3]),
			pid:  pid,
		})
	}

	return stats
}

func getWhitelistedMetrics(blacklist []string) []string {
	var diff []string
	for _, s1 := range allMetrics {
		found := false
		for _, s2 := range blacklist {
			if s1 == s2 {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, s1)
		}
	}
	return diff
}
