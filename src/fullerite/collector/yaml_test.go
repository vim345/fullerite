package collector

import (
	"fmt"
	"fullerite/metric"
	"fullerite/test_utils"
	"io/ioutil"
	"os"

	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	l "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestYamlMetricsConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	testLog := test_utils.BuildLogger()
	c := NewYamlMetrics(nil, 123, testLog).(*YamlMetrics)
	c.Configure(config)
	assert.Equal(t, c.Interval(), 123, "should be the default collection interval")
	assert.Equal(t, c.yamlSource, "/var/lib/fullerite/yaml_metrics.yaml", "should be the default yamlSource")
}

func TestYamlMetricsConfigure(t *testing.T) {
	config := make(map[string]interface{})
	testLog := test_utils.BuildLogger()
	config["interval"] = 9999
	config["yamlSource"] = "/tmp/yaml_metrics.yaml"
	c := NewYamlMetrics(nil, 12, testLog).(*YamlMetrics)
	c.Configure(config)
	assert.Equal(t, c.Interval(), 9999, "should be the defined interval")
	assert.Equal(t, "/tmp/yaml_metrics.yaml", c.yamlSource, "should be the new yaml file")
}

func getMetricsTestHarness(y []byte, log *l.Entry, config map[string]interface{}) []metric.Metric {
	if log == nil {
		log = test_utils.BuildLogger()
	}
	if len(config) == 0 {
		config = make(map[string]interface{})
		config["metricPrefix"] = ""
	}
	c := NewYamlMetrics(nil, 12, log).(*YamlMetrics)
	c.Configure(config)
	return c.GetMetrics(y)
}

func TestYamlMetricsGetMetricSimple(t *testing.T) {
	y := []byte(heredoc.Doc(`---
		test1: 123
		test2: 456
  `))
	should := []metric.Metric{
		{Name: "test1", Value: 123},
		{Name: "test2", Value: 456},
	}
	metrics := getMetricsTestHarness(y, nil, nil)
	compareShouldAndGot(t, should, metrics)
}

func TestYamlMetricsGetMetricWithStrings(t *testing.T) {
	y := []byte(heredoc.Doc(`---
		test1: 123
		test2: wibble
  `))
	should := []metric.Metric{
		{Name: "test1", Value: 123},
	}
	metrics := getMetricsTestHarness(y, nil, nil)
	compareShouldAndGot(t, should, metrics)
}

func TestYamlMetricsGetMetricShouldReturnEmptyAndProduceErrorOnBrokenYaml(t *testing.T) {
	y := []byte(`THIS IS NOT YAML`)
	nullLog, hook := test.NewNullLogger()
	testLog := test_utils.BuildLogger()
	testLog.Logger = nullLog
	metrics := getMetricsTestHarness(y, testLog, nil)
	assert.Equal(t, 1, len(hook.Entries), "We got one error message")
	assert.Equal(t, 0, len(metrics), "metrics list should be empty")
}

func TestYamlMetricsGetMetricShouldReturnEmptyAndProduceNoErrorOnEmptyYaml(t *testing.T) {
	y := []byte(``)
	nullLog, hook := test.NewNullLogger()
	testLog := test_utils.BuildLogger()
	testLog.Logger = nullLog
	metrics := getMetricsTestHarness(y, testLog, nil)
	assert.Equal(t, 0, len(hook.Entries), "We did not get error message")
	assert.Equal(t, 0, len(metrics), "metrics list should be empty")
}

func TestYamlMetricsGetMetricWithNestedValues(t *testing.T) {
	y := []byte(heredoc.Doc(`---
		nested:
		- 123
		- 1234
		test1: 123
	`))
	should := []metric.Metric{
		{Name: "test1", Value: 123},
	}
	metrics := getMetricsTestHarness(y, nil, nil)
	compareShouldAndGot(t, should, metrics)
}

func TestYamlMetricsGetMetricWithFulleriteFormatNoMetrics(t *testing.T) {
	y := []byte(heredoc.Doc(`---
		format: fullerite
		metrics:
		test3: 456
	`))
	metrics := getMetricsTestHarness(y, nil, nil)
	assert.Equal(t, 0, len(metrics), "no metrics are returned")
}

func TestYamlMetricsGetMetricWithFulleriteFormat(t *testing.T) {
	y := []byte(heredoc.Doc(`---
		format: fullerite
		test3: 456
		metrics:
		  - name: test1
		    value: 123
		    type: gauge
		    dimensions:
		      dim1: dim1_value
		      dim2: dim2_value
		  - name: test2
		    value: 789
		    type: gauge
		    dimensions:
		      dim1: dim1_value
		      dim2: dim2_value
	`))
	should := []metric.Metric{
		{Name: "test1", Value: 123, MetricType: "gauge"},
		{Name: "test2", Value: 789, MetricType: "gauge"},
	}
	metrics := getMetricsTestHarness(y, nil, nil)
	assert.Equal(t, 2, len(metrics), "two metrics are returned")
	for i, v := range metrics {
		s := should[i]
		assert.Equal(t, s.Name, v.Name, fmt.Sprintf("%d: %s name is correct", i, s.Name))
		assert.Equal(t, s.Value, v.Value, fmt.Sprintf("%d: %s value is correct", i, s.Name))
		assert.Equal(t, s.MetricType, v.MetricType, fmt.Sprintf("%d: %s type is correct", i, s.MetricType))
		assert.Equal(t, "dim1_value", v.Dimensions["dim1"], "dim1 is correct")
		assert.Equal(t, "dim2_value", v.Dimensions["dim2"], "dim1 is correct")
	}
}

func TestYamlMetricsGetMetricWithFulleriteFormatBogusMetrics(t *testing.T) {
	y := []byte(heredoc.Doc(`---
		format: fullerite
		metrics:
		  - name: true_is_not_valid
		    value: true
		    type: gauge
		  - name: string_number_is_not_valid
		    value: "123"
		    type: gauge
		  - name: string_is_not_valid
		    value: "wibble"
		    type: gauge
		test3: 456
	`))
	nullLog, hook := test.NewNullLogger()
	testLog := test_utils.BuildLogger()
	testLog.Logger = nullLog
	metrics := getMetricsTestHarness(y, testLog, nil)
	assert.Equal(t, 0, len(metrics), "no metrics are returned")
	assert.Equal(t, 3, len(hook.Entries), "We got error messages for each bogus case")
}

func TestYamlMetricsGetMetricWithFulleriteFormatNoDimensions(t *testing.T) {
	y := []byte(heredoc.Doc(`---
		format: fullerite
		metrics:
		  - name: test1
		    value: 123
		    type: gauge
		  - name: test2
		    value: 789
		    type: gauge
	`))
	should := []metric.Metric{
		{Name: "test1", Value: 123},
		{Name: "test2", Value: 789},
	}
	metrics := getMetricsTestHarness(y, nil, nil)
	compareShouldAndGot(t, should, metrics)
}

func compareShouldAndGot(t *testing.T, should []metric.Metric, got []metric.Metric) {
	expectedNum := len(should)
	count := 0
	assert.Equal(t, len(should), len(got), fmt.Sprintf("we should have %d metrics", len(should)))
	for _, v1 := range should {
		for _, v2 := range got {
			if v1.Name == v2.Name {
				assert.Equal(t, v2.Value, v1.Value, fmt.Sprintf("%s value is correct", v2.Name))
				count += 1
			}
		}
	}
	assert.Equal(t, expectedNum, count, "We got the expected number Name matches")
}

func TestYamlMetricsGetMetricWithBooleanValues(t *testing.T) {
	y := []byte(heredoc.Doc(`---
		test_stringy_true: 'true'
		test_stringy_false: 'false'
		test_real_true: true
		test_real_false: false
	`))
	should := []metric.Metric{
		{Name: "test_stringy_true", Value: 1},
		{Name: "test_stringy_false", Value: 0},
		{Name: "test_real_true", Value: 1},
		{Name: "test_real_false", Value: 0},
	}
	metrics := getMetricsTestHarness(y, nil, nil)
	compareShouldAndGot(t, should, metrics)
}

func TestYamlMetricsGetMetricWithJsonInput(t *testing.T) {
	y := []byte(heredoc.Doc(`{
		"test_json_real_value": 123,
		"test_json_string_value": "123",
		"test_stringy_true": "true",
		"test_stringy_false": "false",
		"test_real_true": true,
		"test_real_false": false
	}`))
	should := []metric.Metric{
		{Name: "test_json_real_value", Value: 123},
		{Name: "test_json_string_value", Value: 123},
		{Name: "test_stringy_true", Value: 1},
		{Name: "test_stringy_false", Value: 0},
		{Name: "test_real_true", Value: 1},
		{Name: "test_real_false", Value: 0},
	}
	metrics := getMetricsTestHarness(y, nil, nil)
	compareShouldAndGot(t, should, metrics)
}

func TestYamlMetricsGetMetricsWithWhitelist(t *testing.T) {
	config := make(map[string]interface{})
	a := make([]interface{}, 2)
	a[0] = "uptime"
	a[1] = "^sfx_"
	config["yamlKeyWhitelist"] = a
	config["interval"] = 9999
	config["metricPrefix"] = ""
	y := []byte(heredoc.Doc(`---
		uptime: 123
		should_be_sfx_filtered: 123
		sfx_wibble: 666
		sfx_wobble: should_not_happen
	`))
	should := []metric.Metric{
		{Name: "uptime", Value: 123},
		{Name: "sfx_wibble", Value: 666},
	}
	metrics := getMetricsTestHarness(y, nil, config)
	compareShouldAndGot(t, should, metrics)
}

func TestYamlMetricsCollectOnceDefaultPrefix(t *testing.T) {
	testLogger := test_utils.BuildLogger()
	config := make(map[string]interface{})
	yamlFile := "/tmp/yaml_metrics.yaml"
	defer os.Remove(yamlFile)
	config["yamlSource"] = yamlFile
	y := []byte(heredoc.Doc(`---
		test1: 123
	`))
	err := ioutil.WriteFile(yamlFile, y, 0644)
	if err != nil {
		t.Fatal("Could not write YAML file")
	}
	testChannel := make(chan metric.Metric)
	c := NewYamlMetrics(testChannel, 123, testLogger).(*YamlMetrics)
	c.Configure(config)
	go c.Collect()
	select {
	case m := <-c.Channel():
		assert.Equal(t, "YamlMetrics.test1", m.Name)
		assert.Equal(t, float64(123), m.Value)
		return
	case <-time.After(4 * time.Second):
		t.Fail()
	}
}

func TestYamlMetricsCollectOnceNoPrefix(t *testing.T) {
	testLogger := test_utils.BuildLogger()
	config := make(map[string]interface{})
	yamlFile := "/tmp/yaml_metrics.yaml"
	defer os.Remove(yamlFile)
	config["yamlSource"] = yamlFile
	config["metricPrefix"] = ""
	y := []byte(heredoc.Doc(`---
		test1: 123
	`))
	err := ioutil.WriteFile(yamlFile, y, 0644)
	if err != nil {
		t.Fatal("Could not write YAML file")
	}
	testChannel := make(chan metric.Metric)
	c := NewYamlMetrics(testChannel, 123, testLogger).(*YamlMetrics)
	c.Configure(config)
	go c.Collect()
	select {
	case m := <-c.Channel():
		assert.Equal(t, "test1", m.Name)
		assert.Equal(t, float64(123), m.Value)
		return
	case <-time.After(4 * time.Second):
		t.Fail()
	}
}

func TestYamlMetricsCollectOnceNewPrefix(t *testing.T) {
	testLogger := test_utils.BuildLogger()
	config := make(map[string]interface{})
	yamlFile := "/tmp/yaml_metrics.yaml"
	defer os.Remove(yamlFile)
	config["yamlSource"] = yamlFile
	config["metricPrefix"] = "wibble"
	y := []byte(heredoc.Doc(`---
		test1: 123
	`))
	err := ioutil.WriteFile(yamlFile, y, 0644)
	if err != nil {
		t.Fatal("Could not write YAML file")
	}
	testChannel := make(chan metric.Metric)
	c := NewYamlMetrics(testChannel, 123, testLogger).(*YamlMetrics)
	c.Configure(config)
	go c.Collect()
	select {
	case m := <-c.Channel():
		assert.Equal(t, "wibble.test1", m.Name)
		assert.Equal(t, float64(123), m.Value)
		return
	case <-time.After(4 * time.Second):
		t.Fail()
	}
}

func TestYamlMetricsCollectNoShellExec(t *testing.T) {
	testLogger := test_utils.BuildLogger()
	config := make(map[string]interface{})
	execFile := "/tmp/yaml_metrics_exec.yaml"
	defer os.Remove(execFile)
	config["yamlSource"] = fmt.Sprintf("exec:%s", execFile)
	y := []byte(heredoc.Doc(`#!/bin/sh
		echo "{test1: 123, test2: 456}"
	`))
	err := ioutil.WriteFile(execFile, y, 0700)
	if err != nil {
		t.Fatal("Could not write exec file")
	}
	testChannel := make(chan metric.Metric)
	c := NewYamlMetrics(testChannel, 123, testLogger).(*YamlMetrics)
	c.Configure(config)
	go c.Collect()
	select {
	case m := <-c.Channel():
		assert.Equal(t, "YamlMetrics.test1", m.Name)
		assert.Equal(t, float64(123), m.Value)
		return
	case <-time.After(4 * time.Second):
		t.Fail()
	}
}

func TestYamlMetricsCollectShellExec(t *testing.T) {
	testLogger := test_utils.BuildLogger()
	config := make(map[string]interface{})
	shellCommand := `echo 123 | sed 's/\(.*\)/{testshell: \1}/'`
	config["yamlSource"] = fmt.Sprintf("shell:%s", shellCommand)
	testChannel := make(chan metric.Metric)
	c := NewYamlMetrics(testChannel, 123, testLogger).(*YamlMetrics)
	c.Configure(config)
	go c.Collect()
	select {
	case m := <-c.Channel():
		assert.Equal(t, "YamlMetrics.testshell", m.Name)
		assert.Equal(t, float64(123), m.Value)
		return
	case <-time.After(4 * time.Second):
		t.Fail()
	}
}
