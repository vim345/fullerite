package collector

import (
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
// Currently only supports top-level keys
// Converts truthy/falsey values to 1/0 respectively
// Otherwise, emits any keys with numeric (or stringy numeric) values
//   config:
//     yamlSource       - location of YAML/JSON file to read
//     yamlKeyWhitelist - array of regexps to filter keys processed,
//                        useful in particular for facter or similar output
//
type YamlMetrics struct {
	baseCollector
	metricPrefix     string
	yamlSource       string
	YamlKeyWhitelist []string
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
		switch val := v.(type) {
		case []interface{}:
			for _, v := range val {
				c.YamlKeyWhitelist = append(c.YamlKeyWhitelist, v.(string))
			}
		case []string:
			c.YamlKeyWhitelist = val
		}
	}
	if metricPrefix, exists := configMap["metricPrefix"]; exists {
		c.metricPrefix = metricPrefix.(string)
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
	for _, w := range c.YamlKeyWhitelist {
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

// GetMetrics Get metrics from the YAML supplied
func (c *YamlMetrics) GetMetrics(yamlData []byte) map[string]float64 {
	m := make(map[string]interface{})
	metrics := make(map[string]float64)
	if len(yamlData) == 0 {
		return metrics
	}
	err := yaml.Unmarshal(yamlData, &m)
	if err != nil {
		c.log.Error("Could not unmarshall YAML: ", err.Error())
	}
	// only keep values that convert to float
	for k, v := range m {
		if len(c.YamlKeyWhitelist) > 0 && !c.yamlKeyMatchesWhitelist(k) {
			continue
		}
		if newVal, ok := processYamlAcceptedValues(v); ok == true {
			metrics[k] = newVal
		}
	}
	return metrics
}

// Collect Compares box IP against leader IP and if true, sends data.
func (c *YamlMetrics) Collect() {
	go c.sendMetrics()
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
func (c *YamlMetrics) sendMetrics() {
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
	for k, v := range c.GetMetrics(y) {
		s := c.buildMetric(k, v)
		c.Channel() <- s
	}
}

// buildMetric Takes a k/v, adds common dimensions, and returns a metric to send
func (c *YamlMetrics) buildMetric(k string, v float64) metric.Metric {
	metricName := k
	if c.metricPrefix != "" {
		metricName = c.metricPrefix + "." + k
	}
	m := metric.New(metricName)
	m.Value = v
	return m
}
