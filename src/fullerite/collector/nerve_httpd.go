package collector

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"fullerite/config"
	"fullerite/metric"
	"fullerite/util"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	l "github.com/Sirupsen/logrus"
)

var (
	getNerveHTTPDMetrics = (*NerveHTTPD).getMetrics
	knownApacheMetrics   = map[string]string{
		"ReqPerSec":        metric.Gauge,
		"BytesPerSec":      metric.Gauge,
		"BytesPerReq":      metric.Gauge,
		"BusyWorkers":      metric.Gauge,
		"Total Accesses":   metric.CumulativeCounter,
		"IdleWorkers":      metric.Gauge,
		"StartingWorkers":  metric.Gauge,
		"ReadingWorkers":   metric.Gauge,
		"WritingWorkers":   metric.Gauge,
		"KeepaliveWorkers": metric.Gauge,
		"DnsWorkers":       metric.Gauge,
		"ClosingWorkers":   metric.Gauge,
		"LoggingWorkers":   metric.Gauge,
		"FinishingWorkers": metric.Gauge,
		"CleanupWorkers":   metric.Gauge,
		"StandbyWorkers":   metric.Gauge,
		"CPULoad":          metric.Gauge,
	}
	metricRegexp = regexp.MustCompile(`^([A-Za-z ]+):\s+(.+)$`)
)

// NerveHTTPD discovers Apache servers via Nerve config
// and reports metric for them
type NerveHTTPD struct {
	baseCollector

	configFilePath    string
	queryPath         string
	timeout           int
	statusTTL         time.Duration
	servicesWhitelist []string
}

type nerveHTTPDResponse struct {
	data   []byte
	err    error
	status int
}

func init() {
	RegisterCollector("NerveHTTPD", newNerveHTTPD)
}

func newNerveHTTPD(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	c := new(NerveHTTPD)
	c.channel = channel
	c.interval = initialInterval
	c.log = log

	c.name = "NerveHTTPD"
	c.configFilePath = "/etc/nerve/nerve.conf.json"
	c.queryPath = "server-status?auto"
	c.timeout = 2
	c.statusTTL = time.Duration(60) * time.Minute
	return c
}

// Configure the collector
func (c *NerveHTTPD) Configure(configMap map[string]interface{}) {
	if val, exists := configMap["queryPath"]; exists {
		c.queryPath = val.(string)
	}

	if val, exists := configMap["configFilePath"]; exists {
		c.configFilePath = val.(string)
	}

	if val, exists := configMap["status_ttl"]; exists {
		tmpStatusTTL := config.GetAsInt(val, 3600)
		c.statusTTL = time.Duration(tmpStatusTTL) * time.Second
	}

	if val, exists := configMap["servicesWhitelist"]; exists {
		c.servicesWhitelist = config.GetAsSlice(val)
	}

	c.configureCommonParams(configMap)
}

// Collect the metrics
func (c *NerveHTTPD) Collect() {
	rawFileContents, err := ioutil.ReadFile(c.configFilePath)
	if err != nil {
		c.log.Warn("Failed to read the contents of file ", c.configFilePath, " because ", err)
		return
	}
	services, err := util.ParseNerveConfig(&rawFileContents, true)
	if err != nil {
		c.log.Warn("Failed to parse the nerve config at ", c.configFilePath, ": ", err)
		return
	}
	c.log.Debug("Finished parsing Nerve config into ", services)

	for _, service := range services {
		if c.serviceInWhitelist(service) {
			go c.emitHTTPDMetric(service)
		}
	}
}

func (c *NerveHTTPD) serviceInWhitelist(service util.NerveService) bool {
	for _, s := range c.servicesWhitelist {
		if s == service.Name+"."+service.Namespace {
			return true
		}
	}
	return false
}

func (c *NerveHTTPD) emitHTTPDMetric(service util.NerveService) {
	metrics := getNerveHTTPDMetrics(c, service)
	for _, metric := range metrics {
		c.Channel() <- metric
	}
	c.Channel() <- metric.Sentinel()
}

func (c *NerveHTTPD) getMetrics(service util.NerveService) []metric.Metric {
	results := []metric.Metric{}
	serviceLog := c.log.WithField("service", service.Name)

	endpoint := fmt.Sprintf("http://%s:%d/%s", service.Host, service.Port, c.queryPath)
	serviceLog.Debug("making GET request to ", endpoint)

	httpResponse := fetchApacheMetrics(endpoint, service.Port)

	if httpResponse.status != 200 {
		serviceLog.Warn("Failed to query endpoint ", endpoint, ": ", httpResponse.err)
		return results
	}
	apacheMetrics := extractApacheMetrics(httpResponse.data)
	metric.AddToAll(&apacheMetrics, map[string]string{
		"service_name":      service.Name,
		"service_namespace": service.Namespace,
		"port":              strconv.Itoa(service.Port),
	})
	return apacheMetrics
}

func extractApacheMetrics(data []byte) []metric.Metric {
	results := []metric.Metric{}
	reader := bytes.NewReader(data)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		metricLine := scanner.Text()
		resultMatch := metricRegexp.FindStringSubmatch(metricLine)
		if len(resultMatch) > 0 {
			k := resultMatch[1]
			v := resultMatch[2]
			if k == "IdleWorkers" {
				continue
			}

			if k == "Scoreboard" {
				scoreBoardMetrics := extractScoreBoardMetrics(k, v)
				results = append(results, scoreBoardMetrics...)
			}

			metric, err := buildApacheMetric(k, v)
			if err == nil {
				results = append(results, metric)
			}
		}

	}
	return results
}

func buildApacheMetric(key, value string) (metric.Metric, error) {
	var tmpMetric metric.Metric
	if metricType, ok := knownApacheMetrics[key]; ok {
		whiteRegexp := regexp.MustCompile(`\s+`)
		metricName := whiteRegexp.ReplaceAllString(key, "")
		metricValue, err := strconv.ParseFloat(value, 64)

		if err != nil {
			return tmpMetric, err
		}
		m := metric.WithValue(metricName, metricValue)
		m.MetricType = metricType
		return m, nil
	}
	return tmpMetric, errors.New("invalid metric")
}

func extractScoreBoardMetrics(key, value string) []metric.Metric {
	results := []metric.Metric{}
	charCounter := func(str string, pattern string) float64 {
		return float64(strings.Count(str, pattern))
	}
	metricWithValueAndType := func(str string, value float64) metric.Metric {
		m := metric.WithValue(str, value)
		m.MetricType = knownApacheMetrics[str]
		return m
	}
	results = append(results, metricWithValueAndType("IdleWorkers", charCounter(value, "_")))
	results = append(results, metricWithValueAndType("StartingWorkers", charCounter(value, "S")))
	results = append(results, metricWithValueAndType("ReadingWorkers", charCounter(value, "R")))
	results = append(results, metricWithValueAndType("WritingWorkers", charCounter(value, "W")))
	results = append(results, metricWithValueAndType("KeepaliveWorkers", charCounter(value, "K")))
	results = append(results, metricWithValueAndType("DnsWorkers", charCounter(value, "D")))
	results = append(results, metricWithValueAndType("ClosingWorkers", charCounter(value, "C")))
	results = append(results, metricWithValueAndType("LoggingWorkers", charCounter(value, "L")))
	results = append(results, metricWithValueAndType("FinishingWorkers", charCounter(value, "G")))
	results = append(results, metricWithValueAndType("CleanupWorkers", charCounter(value, "I")))
	results = append(results, metricWithValueAndType("StandbyWorkers", charCounter(value, "_")))
	return results
}

func fetchApacheMetrics(endpoint string, timeout int) *nerveHTTPDResponse {
	response := new(nerveHTTPDResponse)
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	rsp, err := client.Get(endpoint)
	response.err = err
	if rsp != nil {
		response.status = rsp.StatusCode
	}

	if err != nil {
		return response
	}

	txt, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		response.err = err
		return response
	}
	response.data = txt
	return response
}
