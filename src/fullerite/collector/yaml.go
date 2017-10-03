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

	"github.com/ghodss/yaml"

	l "github.com/Sirupsen/logrus"
)

const (
	defaultYamlCollectorName = "YamlMetrics"
	defaultYamlSource        = "/var/lib/fullerite/yaml_metrics.yaml"
	defaultYamlSourceMethod  = "file"
	defaultYamlFormat        = "fullerite"
	defaultYamlMetricPrefix  = "YamlMetrics"
)

// YamlMetrics collector extracts metrics from YAML data files
//
// Supports two modes:
//   - 'fullerite' (default):
//     Recommended, as it allows full control of metric type, value, and dimensions,
//     and is compatible with the Adhoc collector.
//     Enabled explicitly by config option yamlFormat: "fullerite"
//     Reads YAML 'metrics' key as an array of metrics, and
//     for each of those, convert to a metrics.Metric object if possible
//     (via remarshalling to JSON and using standard JSON unmarshal into object)
//   - 'simple':
//     Enabled by config option yamlFormat: "simple"
//     Useful for reading snippets of data from preexisting sources (eg simple APIs, or Facter),
//     or for very simple tools
//     Reads YAML and trys to extract float64 gauges from top level keys
//       - keys with numeric values
//       - keys with stringy numeric values (eg "123", "123.01")
//       - booleans and stringy bools: converts to 1/0 for true/false
//       - does not support adding dimensions
//
//  Can read YAML data via different 'yamlSourceMethod': read from file ('file'), run a shell
//  pipeline ('shell'), or directly exec a script ('exec').
//
//  Config:
//     metricPrefix     - prefix to add to Metrics, default 'yamlMetrics'.
//     yamlSource       - source of YAML/JSON
//     yamlSourceMethod - method to call source: file, shell, exec (defaut: file)
//     yamlKeyWhitelist - array of regexps to filter keys processed,
//                        useful in particular for facter or similar output
//                        (only used in simple mode)
//
type YamlMetrics struct {
	baseCollector
	metricPrefix     string
	yamlSource       string
	yamlSourceMethod string
	yamlFormat       string
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
	c.yamlSourceMethod = defaultYamlSourceMethod
	c.yamlFormat = defaultYamlFormat
	c.metricPrefix = defaultYamlMetricPrefix
	return c
}

// Configure takes a dictionary of values with which the handler can configure itself
func (c *YamlMetrics) Configure(configMap map[string]interface{}) {
	if v, exists := configMap["yamlSource"]; exists {
		c.yamlSource = v.(string)
	}
	if v, exists := configMap["yamlSourceMethod"]; exists {
		c.yamlSourceMethod = v.(string)
	}
	if v, exists := configMap["yamlFormat"]; exists {
		c.yamlFormat = v.(string)
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
func (c *YamlMetrics) getFulleriteFormatMetrics(yamlData []byte) (metrics []metric.Metric) {
	var m []interface{}
	err := yaml.Unmarshal(yamlData, &m)
	if err != nil {
		c.log.Error("Invalid YAML for fullerite yamlFormat")
	}
	for _, v := range m {
		var metric metric.Metric
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
func (c *YamlMetrics) getSimpleFormatMetrics(yamlData []byte) (metrics []metric.Metric) {
	if len(c.yamlKeyWhitelist) == 0 {
		c.log.Error("Must specify yamlKeyWhitelist for simple format metrics")
		return metrics
	}
	m := make(map[string]interface{})
	err := yaml.Unmarshal(yamlData, &m)
	if err != nil {
		c.log.Error("Could not unmarshal YAML: ", err.Error())
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
	c.log.Debug("GetMetrics: entry")
	if len(yamlData) == 0 {
		c.log.Debug("GetMetrics: yamlData is empty")
		return metrics
	}
	switch format := c.yamlFormat; format {
	case "simple":
		c.log.Debugf("getSimpleFormatMetrics: %s", yamlData)
		return c.getSimpleFormatMetrics(yamlData)
	case "fullerite":
		c.log.Debugf("getFulleriteFormatMetrics: %s", yamlData)
		return c.getFulleriteFormatMetrics(yamlData)
	default:
		c.log.Errorf("%s is not a valid yamlFormat", format)
		return metrics
	}
}

// Collect Compares box IP against leader IP and if true, sends data.
func (c *YamlMetrics) Collect() {
	var y []byte
	var err error
	switch c.yamlSourceMethod {
	case "shell":
		y, err = c.getYamlFromShell(c.yamlSource)
	case "exec":
		y, err = c.getYamlFromExec(c.yamlSource)
	case "file":
		y, err = c.getYamlFromFile(c.yamlSource)
	default:
		c.log.Errorf("Invalid yamlSourceMethod %s", c.yamlSourceMethod)
	}
	if err != nil {
		c.log.Errorf("Could not get YAML data from source %s:%s ", c.yamlSourceMethod, c.yamlSource)
		return
	}
	if metrics := c.GetMetrics(y); len(metrics) > 0 {
		go c.sendMetrics(metrics)
	}
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
