package collector

import (
	"encoding/json"
	"fmt"
	"fullerite/config"
	"fullerite/metric"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"

	l "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
)

const (
	defaultYamlCollectorName = "YamlMetrics"
	defaultYamlSource        = "/var/lib/fullerite/yaml_metrics.yaml"
	defaultYamlMetricPrefix  = "YamlMetrics"
)

// YamlMetrics collector extracts metrics from YAML data files
//
// Supports two modes:
//   - 'simple' (default):
//     reads YAML and trys to extract float64 gauges from top level keys
//       - keys with numeric values
//       - keys with stringy numeric values (eg "123", "123.01")
//       - booleans and stringy bools: converts to 1/0 for true/false
//       - does not support adding dimensions
//       - designed to read arbrary data, eg facter.yaml
//   - 'fullerite':
//     enabled by 'format: "fullerite"'
//     reads YAML 'metrics' key as an array of metrics
//     for each of those, convert to a metrics.Metric object if possible
//     (via remarshalling to JSON and using standard JSON unmarshal into object)
//
//  Config:
//     metricPrefix     - prefix to add to Metrics, default 'yamlMetrics'.
//     yamlSource       - location of YAML/JSON file to read
//     yamlKeyWhitelist - array of regexps to filter keys processed,
//                        useful in particular for facter or similar output
//                        (only used in simple mode)
//
type YamlMetrics struct {
	baseCollector
	metricPrefix     string
	yamlSource       string
	yamlKeyWhitelist []string
}

func init() {
	RegisterCollector("YamlMetrics", NewYamlMetrics)
}

// NewYamlMetrics returns a initial collector, to be configured
func NewYamlMetrics(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	c := new(YamlMetrics)

	c.log = log
	c.channel = channel
	c.interval = initialInterval

	c.name = defaultYamlCollectorName
	c.yamlSource = defaultYamlSource
	c.metricPrefix = defaultYamlMetricPrefix
	return c
}

// Configure takes a dictionary of values with which the handler can configure itself
func (c *YamlMetrics) Configure(configMap map[string]interface{}) {
	if yamlSource, exists := configMap["yamlSource"]; exists {
		c.yamlSource = yamlSource.(string)
	}
	if v, exists := configMap["yamlKeyWhitelist"]; exists {
		c.yamlKeyWhitelist = config.GetAsSlice(v)
	}
	if metricPrefix, exists := configMap["metricPrefix"]; exists {
		if metricPrefix == "" {
			c.log.Error("metricPrefix cannot be an empty string")
		} else {
			c.metricPrefix = metricPrefix.(string)
		}
	}
	c.configureCommonParams(configMap)
}

func processYamlAcceptedValues(v interface{}) (float64, bool) {
	switch value := v.(type) {
	case string:
		f, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return f, true
		}
		b, err := strconv.ParseBool(value)
		if err == nil {
			if b {
				return 1, true
			}
			return 0, true
		}
	case float64:
		return value, true
	case bool:
		if value {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

func (c *YamlMetrics) yamlKeyMatchesWhitelist(k string) bool {
	for _, w := range c.yamlKeyWhitelist {
		ok, err := regexp.Match(w, []byte(k))
		if err != nil {
			c.log.Error("Invalid regexp in yamlKeyWhitelist: ", k)
			continue
		}
		if ok {
			return true
		}
	}
	return false
}

// 'advanced' format - YAML representation of metric.Metric objects
func (c *YamlMetrics) getFulleriteFormatMetrics(m []interface{}) (metrics []metric.Metric) {
	var metric metric.Metric
	for _, v := range m {
		j, err := json.Marshal(v)
		if err != nil {
			c.log.Error("getFulleriteFormatMetrics: Skipping, could not Marshal '%s': %s", v, err.Error())
			continue
		}
		c.log.Debug(fmt.Sprintf("Got metric defn: %s", j))
		if err := json.Unmarshal(j, &metric); err != nil {
			c.log.Error("getFulleriteFormatMetrics: Skipping, could not Unmarshal '%s': %s", string(j), err.Error())
			continue
		}
		metric.Name = c.metricPrefix + "." + metric.Name
		metrics = append(metrics, metric)
	}
	return metrics
}

// simple k/v format; only keep values that convert to float
func (c *YamlMetrics) getSimpleFormatMetrics(m map[string]interface{}) (metrics []metric.Metric) {
	if len(c.yamlKeyWhitelist) == 0 {
		c.log.Error("Must specify yamlKeyWhitelist for simple format metrics")
		return metrics
	}
	for k, v := range m {
		if !c.yamlKeyMatchesWhitelist(k) {
			continue
		}
		if newVal, ok := processYamlAcceptedValues(v); ok {
			s := c.buildMetric(k, newVal)
			metrics = append(metrics, s)
		}
	}
	return metrics
}

// GetMetrics Get metrics from the YAML supplied
func (c *YamlMetrics) GetMetrics(yamlData []byte) (metrics []metric.Metric) {
	m := make(map[string]interface{})
	if len(yamlData) == 0 {
		return metrics
	}
	err := yaml.Unmarshal(yamlData, &m)
	if err != nil {
		c.log.Error("Could not unmarshal YAML: ", err.Error())
		c.log.Debug(fmt.Sprintf("%s", yamlData))
		return metrics
	}
	if f, ok := m["format"]; ok && f.(string) == "fullerite" {
		switch val := m["metrics"].(type) {
		case []interface{}:
			return c.getFulleriteFormatMetrics(val)
		default:
			return metrics
		}
	}
	return c.getSimpleFormatMetrics(m)
}

// Collect Compares box IP against leader IP and if true, sends data.
func (c *YamlMetrics) Collect() {
	method, source := extractYamlSourceMethodAndSource(c.yamlSource)
	var y []byte
	var err error
	if method == "file" {
		y, err = c.getYamlFromFile(source)
	} else if method == "shell" {
		y, err = c.getYamlFromShell(source)
	} else if method == "exec" {
		y, err = c.getYamlFromExec(source)
	}
	if err != nil {
		c.log.Error("Could not get YAML from source: ", err.Error())
		return
	}
	go c.sendMetrics(c.GetMetrics(y))
}

func (c *YamlMetrics) getYamlFromFile(file string) ([]byte, error) {
	y, err := ioutil.ReadFile(c.yamlSource)
	return y, err
}

func (c *YamlMetrics) getYamlFromExec(command string) ([]byte, error) {
	y, err := exec.Command(command).Output()
	return y, err
}

func (c *YamlMetrics) getYamlFromShell(pipeline string) ([]byte, error) {
	y, err := exec.Command("sh", "-c", pipeline).Output()
	return y, err
}

func extractYamlSourceMethodAndSource(s string) (string, string) {
	if ok, _ := regexp.MatchString("^file:", s); ok {
		return "file", s[len("file:"):len(s)]
	}
	if ok, _ := regexp.MatchString("^exec:", s); ok {
		return "exec", s[len("exec:"):len(s)]
	}
	if ok, _ := regexp.MatchString("^shell:", s); ok {
		return "shell", s[len("shell:"):len(s)]
	}
	return "file", s
}

// sendMetrics Send to baseCollector channel.
func (c *YamlMetrics) sendMetrics(m []metric.Metric) {
	for _, s := range m {
		c.Channel() <- s
	}
}

// buildMetric Takes a k/v, adds common dimensions, and returns a metric to send
func (c *YamlMetrics) buildMetric(k string, v float64) metric.Metric {
	metricName := c.metricPrefix + "." + k
	m := metric.New(metricName)
	m.Value = v
	return m
}
