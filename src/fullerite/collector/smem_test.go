package collector

import (
	"errors"
	"fullerite/metric"
	"os/exec"
	"testing"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var smemOutput = `    5     343     864    2442180 apache2 1234
`

func TestNewSmemStats(t *testing.T) {
	c := make(chan metric.Metric)
	i := 10
	l := defaultLog.WithFields(l.Fields{"collector": "Mesos"})

	actual := newSmemStats(c, i, l).(*SmemStats)

	assert.Equal(t, "SmemStats", actual.Name())
	assert.Equal(t, c, actual.Channel())
	assert.Equal(t, i, actual.Interval())
	assert.Equal(t, l, actual.log)
}

func TestSmemStatsConfigure(t *testing.T) {
	tests := []struct {
		config              map[string]interface{}
		expectedWhitelist   string
		expectedUser        string
		expectedSmemPath    string
		expectedMetricslist []string
		msg                 string
	}{
		{
			config: map[string]interface{}{
				"user":             "fullerite",
				"procsWhitelist":   "apache2|tmux",
				"smemPath":         "/path/to/smem",
				"metricsBlacklist": []string{"rss", "vss"},
			},
			expectedWhitelist:   "apache2|tmux",
			expectedUser:        "fullerite",
			expectedSmemPath:    "/path/to/smem",
			expectedMetricslist: []string{"pss", "uss"},
			msg:                 "All configs are valid, so no errors",
		},
		{
			config:              map[string]interface{}{},
			expectedWhitelist:   "",
			expectedUser:        "",
			expectedSmemPath:    "",
			expectedMetricslist: []string{"rss", "vss", "pss", "uss"},
			msg:                 "Required configs missing",
		},
	}

	l := defaultLog.WithFields(l.Fields{"collector": "SmemStats"})

	for _, test := range tests {
		sut := newSmemStats(nil, 0, l).(*SmemStats)
		sut.Configure(test.config)

		assert.Equal(t, test.expectedUser, sut.user, test.msg)
		assert.Equal(t, test.expectedWhitelist, sut.whitelistedProcs, test.msg)
		assert.Equal(t, test.expectedSmemPath, sut.smemPath, test.msg)
		assert.Equal(t, test.expectedMetricslist, sut.whitelistedMetrics, test.msg)
	}
}

func TestSmemStatsCollect(t *testing.T) {
	oldExecCommand := execCommand
	oldCommandOutput := commandOutput
	oldGetCmdLineDimensions := getCmdLineDimensions
	oldGetEnvDimensions := getEnvDimensions

	defer func() {
		execCommand = oldExecCommand
		commandOutput = oldCommandOutput
		getCmdLineDimensions = oldGetCmdLineDimensions
		getEnvDimensions = oldGetEnvDimensions
	}()

	execCommand = func(string, ...string) *exec.Cmd {
		return &exec.Cmd{}
	}

	commandOutput = func(*exec.Cmd) ([]byte, error) {
		return []byte(smemOutput), nil
	}

	getCmdLineDimensions = func(*SmemStats, int) map[string]string {
		return map[string]string{
			"dim1": "val1",
		}
	}

	getEnvDimensions = func(*SmemStats, int) map[string]string {
		return map[string]string{
			"dim2": "val2",
		}
	}

	expectedDims := map[string]string{"dim2": "val2", "dim1": "val1"}
	actual := []metric.Metric{}
	expected := []metric.Metric{
		metric.Metric{Name: "apache2.smem.pss", MetricType: "gauge", Value: 5, Dimensions: expectedDims},
		metric.Metric{Name: "apache2.smem.uss", MetricType: "gauge", Value: 343, Dimensions: expectedDims},
		metric.Metric{Name: "apache2.smem.vss", MetricType: "gauge", Value: 2.44218e+06, Dimensions: expectedDims},
		metric.Metric{Name: "apache2.smem.rss", MetricType: "gauge", Value: 864, Dimensions: expectedDims},
	}

	c := make(chan metric.Metric)
	sut := newSmemStats(c, 0, defaultLog).(*SmemStats)
	sut.user = "user"
	sut.whitelistedProcs = "some|whitelist"
	sut.smemPath = "/path/to/smem"
	sut.whitelistedMetrics = []string{"pss", "uss", "vss", "rss"}
	go sut.Collect()

	for i := 0; i < len(expected); i++ {
		actual = append(actual, <-c)
	}

	assert.Equal(t, expected, actual)
}

func TestSmemStatsCollectNotCalled(t *testing.T) {
	oldGetSmemStats := getSmemStats
	defer func() { getSmemStats = oldGetSmemStats }()

	getSmemStatsCalled := false
	getSmemStats = func(*SmemStats) []smemStatLine {
		getSmemStatsCalled = true
		return nil
	}

	tests := []struct {
		user               string
		whitelistedProcs   string
		smemPath           string
		whitelistedMetrics []string
	}{
		{
			user: "user",
		},
		{
			whitelistedProcs: "apache2",
		},
		{
			smemPath: "/path/to/smem",
		},
		{
			whitelistedMetrics: []string{"pss", "uss"},
		},
	}

	for _, test := range tests {
		sut := newSmemStats(nil, 0, nil).(*SmemStats)
		sut.user = test.user
		sut.whitelistedProcs = test.whitelistedProcs
		sut.smemPath = test.smemPath
		sut.whitelistedMetrics = test.whitelistedMetrics

		sut.Collect()

		assert.False(t, getSmemStatsCalled)
	}
}

// func TestGetCmdLineDimensions(t *testing.T) {
// 	oldReadCmdline := readCmdline
// 	defer func() { readCmdline = oldReadCmdline }()

// 	readCmdline = func(*SmemStats, string) ([]byte, error) {
// 		return []byte("apache worker 1"), nil
// 	}

// 	s := newSmemStats(nil, 0, nil).(*SmemStats)
// 	s.dimensionsFromCmdline = map[string]string{
// 		"worker_id": ".* worker ([0-9]+)$",
// 	}

// 	expectedDims := map[string]string{
// 		"worker_id": "1",
// 	}

// 	assert.Equal(t, getCmdLineDimensions(s, 123), expectedDims)
// }

func TestGetCmdLineDimensions(t *testing.T) {
	oldExecCommand := execCommand
	oldCommandOutput := commandOutput

	defer func() {
		execCommand = oldExecCommand
		commandOutput = oldCommandOutput
	}()

	tests := []struct {
		pid                   int
		cmdLineData           string
		cmdLineReadError      error
		dimensionsFromCmdLine map[string]string
		expectedDimensions    map[string]string
		msg                   string
	}{
		{
			pid: 0,
			dimensionsFromCmdLine: map[string]string{},
			expectedDimensions:    map[string]string{},
			msg:                   "PID is 0; so no dimensions should be reported",
		},
		{
			pid: 1234,
			dimensionsFromCmdLine: map[string]string{},
			expectedDimensions:    map[string]string{},
			msg:                   "Although PID is not 0, the dimensionsFromEnv is empty; so no dimensions should be reported",
		},
		{
			pid:                   1234,
			cmdLineData:           "",
			cmdLineReadError:      errors.New("Error reading cmdLine"),
			dimensionsFromCmdLine: map[string]string{"worker_id": ".* worker ([0-9]+)$"},
			expectedDimensions:    map[string]string{},
			msg:                   "/proc/PID/cmdline could not be read; so no dimensions should be reported",
		},
		{
			pid:                   1234,
			cmdLineData:           "apache worker 1",
			cmdLineReadError:      nil,
			dimensionsFromCmdLine: map[string]string{"worker_id": ".* worker ([0-9]+)$"},
			expectedDimensions:    map[string]string{"worker_id": "1"},
			msg:                   "/proc/PID/cmdline can be read; so report the proper dimensions",
		},
		{
			pid:                   1234,
			cmdLineData:           "apache woker 1",
			cmdLineReadError:      nil,
			dimensionsFromCmdLine: map[string]string{"worker_id": ".* worker ([0-9]+)$"},
			expectedDimensions:    map[string]string{},
			msg:                   "/proc/PID/cmdline can be read but does not contain the matching regex; so report no dimensions",
		},
	}

	for _, test := range tests {
		s := newSmemStats(nil, 0, defaultLog.WithFields(l.Fields{"collector": "SmemStats"})).(*SmemStats)
		s.dimensionsFromCmdline = test.dimensionsFromCmdLine

		execCommand = func(string, ...string) *exec.Cmd {
			return &exec.Cmd{}
		}

		commandOutput = func(*exec.Cmd) ([]byte, error) {
			return []byte(test.cmdLineData), test.cmdLineReadError
		}

		assert.Equal(t, test.expectedDimensions, getCmdLineDimensions(s, test.pid), test.msg)
	}
}

func TestGetEnvDimensions(t *testing.T) {
	oldExecCommand := execCommand
	oldCommandOutput := commandOutput

	defer func() {
		execCommand = oldExecCommand
		commandOutput = oldCommandOutput
	}()

	tests := []struct {
		pid                int
		environ            string
		environReadError   error
		dimensionsFromEnv  map[string]string
		expectedDimensions map[string]string
		msg                string
	}{
		{
			pid:                0,
			dimensionsFromEnv:  map[string]string{},
			expectedDimensions: map[string]string{},
			msg:                "PID is 0; so no dimensions should be reported",
		},
		{
			pid:                1234,
			dimensionsFromEnv:  map[string]string{},
			expectedDimensions: map[string]string{},
			msg:                "Although PID is not 0, the dimensionsFromEnv is empty; so no dimensions should be reported",
		},
		{
			pid:                1234,
			environ:            "",
			environReadError:   errors.New("Error reading environ"),
			dimensionsFromEnv:  map[string]string{"paasta_cluster": "PAASTA_CLUSTER_NAME"},
			expectedDimensions: map[string]string{},
			msg:                "The environ for the PID could not be read; so no dimensions should be reported",
		},
		{
			pid:                1234,
			environ:            "PAASTA_INSTANCE_NAME=instance-name\000PAASTA_CLUSTER_NAME=cluster-name\000",
			environReadError:   nil,
			dimensionsFromEnv:  map[string]string{"paasta_cluster": "PAASTA_CLUSTER_NAME"},
			expectedDimensions: map[string]string{"paasta_cluster": "cluster-name"},
			msg:                "environ can be read; so report the proper dimensions",
		},
		{
			pid:                1234,
			environ:            "PAASTA_INSTANCE_NAME=instance-name\000HOSTNAME=127.0.0.1\000",
			environReadError:   nil,
			dimensionsFromEnv:  map[string]string{"paasta_cluster": "PAASTA_CLUSTER_NAME"},
			expectedDimensions: map[string]string{},
			msg:                "environ can be read but does not contain the required environ variable; so report no dimensions",
		},
	}

	for _, test := range tests {
		s := newSmemStats(nil, 0, defaultLog.WithFields(l.Fields{"collector": "SmemStats"})).(*SmemStats)
		s.dimensionsFromEnv = test.dimensionsFromEnv

		execCommand = func(string, ...string) *exec.Cmd {
			return &exec.Cmd{}
		}

		commandOutput = func(*exec.Cmd) ([]byte, error) {
			return []byte(test.environ), test.environReadError
		}

		assert.Equal(t, test.expectedDimensions, getEnvDimensions(s, test.pid), test.msg)
	}
}
