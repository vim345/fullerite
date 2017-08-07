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
	"github.com/Sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestYamlMetricsConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	testLog := test_utils.BuildLogger()
	c := NewYamlMetrics(nil, 123, testLog).(*YamlMetrics)
	c.Configure(config)

	assert.Equal(t,
		c.Interval(),
		123,
		"should be the default collection interval",
	)

	assert.Equal(t,
		c.yamlSource,
		"/var/lib/fullerite/yaml_metrics.yaml",
		"should be the default yamlSource",
	)

}

func TestYamlMetricsConfigure(t *testing.T) {
	config := make(map[string]interface{})
	testLog := test_utils.BuildLogger()
	config["interval"] = 9999
	config["yamlSource"] = "/tmp/yaml_metrics.yaml"

	// the channel and logger don't matter
	c := NewYamlMetrics(nil, 12, testLog).(*YamlMetrics)
	c.Configure(config)

	assert.Equal(t,
		c.Interval(),
		9999,
		"should be the defined interval",
	)

	assert.Equal(t,
		"/tmp/yaml_metrics.yaml",
		c.yamlSource,
		"should be the new yaml file",
	)

}

func TestYamlMetricsGetMetricSimple(t *testing.T) {
	testLog := test_utils.BuildLogger()
	y := []byte(heredoc.Doc(`---
		test1: 123
		test2: 456
  `))
	c := NewYamlMetrics(nil, 12, testLog).(*YamlMetrics)
	metrics := c.GetMetrics(y)
	assert.Equal(t,
		float64(123),
		metrics["test1"],
		"test1 value is correct",
	)
	assert.Equal(t,
		float64(456),
		metrics["test2"],
		"test2 value is correct",
	)
}

func TestYamlMetricsGetMetricWithStrings(t *testing.T) {
	testLog := test_utils.BuildLogger()
	y := []byte(heredoc.Doc(`---
		test1: 123
		test2: wibble
  `))
	c := NewYamlMetrics(nil, 12, testLog).(*YamlMetrics)
	metrics := c.GetMetrics(y)
	assert.Equal(t,
		float64(123),
		metrics["test1"],
		"test1 value is correct",
	)
	_, ok := metrics["test2"]
	assert.Equal(t,
		false,
		ok,
		"test2 value does not exist",
	)
}

func TestYamlMetricsGetMetricShouldReturnEmptyAndProduceErrorOnBrokenYaml(t *testing.T) {
	y := []byte(`THIS IS NOT YAML`)
	nullLog, hook := test.NewNullLogger()
	testLog := test_utils.BuildLogger()
	testLog.Logger = nullLog
	c := NewYamlMetrics(nil, 12, testLog).(*YamlMetrics)
	metrics := c.GetMetrics(y)
	assert.Equal(t, 1, len(hook.Entries), "We got one error message")
	assert.Equal(t, 0, len(metrics), "metrics list should be empty")
}

func TestYamlMetricsGetMetricShouldReturnEmptyAndProduceNoErrorOnEmptyYaml(t *testing.T) {
	y := []byte(``)
	nullLog, hook := test.NewNullLogger()
	testLog := test_utils.BuildLogger()
	testLog.Logger = nullLog
	c := NewYamlMetrics(nil, 12, testLog).(*YamlMetrics)
	metrics := c.GetMetrics(y)
	assert.Equal(t, 0, len(hook.Entries), "We did not get error message")
	assert.Equal(t, 0, len(metrics), "metrics list should be empty")
}

func TestYamlMetricsGetMetricWithNestedValues(t *testing.T) {
	testLog := test_utils.BuildLogger()
	y := []byte(heredoc.Doc(`---
		nested:
		- 123
		- 1234
		test1: 123
	`))
	c := NewYamlMetrics(nil, 12, testLog).(*YamlMetrics)
	metrics := c.GetMetrics(y)
	assert.Equal(t,
		float64(123),
		metrics["test1"],
		"test1 value is correct",
	)
	_, ok := metrics["nested"]
	assert.Equal(t,
		false,
		ok,
		"nested value does not exist",
	)
}

func TestYamlMetricsGetMetricWithBooleanValues(t *testing.T) {
	testLog := test_utils.BuildLogger()
	y := []byte(heredoc.Doc(`---
		test_stringy_true: 'true'
		test_stringy_false: 'false'
		test_real_true: true
		test_real_false: false
	`))
	c := NewYamlMetrics(nil, 12, testLog).(*YamlMetrics)
	metrics := c.GetMetrics(y)
	assert.Equal(t, 4, len(metrics), "we should have 4 metrics returned")
	assert.Equal(t, float64(1), metrics["test_stringy_true"], "stringy true is handled as 1")
	assert.Equal(t, float64(0), metrics["test_stringy_false"], "stringy false is handled as 0")
	assert.Equal(t, float64(1), metrics["test_real_true"], "real true is handled as 1")
	assert.Equal(t, float64(0), metrics["test_real_false"], "real false is handled as 0")
}

func TestYamlMetricsGetMetricWithJsonInput(t *testing.T) {
	testLog := test_utils.BuildLogger()
	y := []byte(heredoc.Doc(`{
		"test1": 123,
		"test2": "123",
		"test_stringy_true": "true",
		"test_stringy_false": "false",
		"test_real_true": true,
		"test_real_false": false
	}`))
	c := NewYamlMetrics(nil, 12, testLog).(*YamlMetrics)
	metrics := c.GetMetrics(y)
	assert.Equal(t, 6, len(metrics), "we should have 6 metrics returned")
	assert.Equal(t, float64(123), metrics["test1"], "naked float is returned normally")
	assert.Equal(t, float64(123), metrics["test2"], "string float is returned as float")
	assert.Equal(t, float64(1), metrics["test_stringy_true"], "stringy true is handled as 1")
	assert.Equal(t, float64(0), metrics["test_stringy_false"], "stringy false is handled as 0")
	assert.Equal(t, float64(1), metrics["test_real_true"], "real true is handled as 1")
	assert.Equal(t, float64(0), metrics["test_real_false"], "real false is handled as 0")
}

func TestYamlMetricsGetMetricsWithWhitelist(t *testing.T) {
	config := make(map[string]interface{})
	a := make([]interface{}, 2)
	a[0] = "uptime"
	a[1] = "^sfx_"
	config["yamlKeyWhitelist"] = a
	config["interval"] = 9999
	y := []byte(heredoc.Doc(`---
		uptime: 123
		should_be_sfx_filtered: 123
		sfx_wibble: 666
		sfx_wobble: should_not_happen
	`))
	c := NewYamlMetrics(nil, 12, testLog).(*YamlMetrics)
	c.Configure(config)
	metrics := c.GetMetrics(y)
	assert.Equal(t, 2, len(metrics), "we should have two metrics matching filter: uptime, sfx_wibble")
	assert.Equal(t, float64(123), metrics["uptime"], "uptime is present andi correct")
	assert.Equal(t, float64(666), metrics["sfx_wibble"], "sfx_wibble is present and correct")
	_, ok1 := metrics["sfx_wobble"]
	assert.Equal(t, false, ok1, "sfx_wobble key should still be omitted, as it is not numeric/bool")
	_, ok2 := metrics["should_be_sfx_filtered"]
	assert.Equal(t, false, ok2, "should_be_sfx_filtered key should be filtered out by whitelist")
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
