package collector

import (
	"fullerite/config"
	"fullerite/metric"
	"fullerite/util"

	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	l "github.com/Sirupsen/logrus"
)

// Nerve uWSGI collector - fetches counts of idle and busy workers from uWSGI stats servers published over HTTP
// http://uwsgi-docs.readthedocs.io/en/latest/StatsServer.html
// It reads the location of uWSGI stats servers from SmartStack
type nerveUWSGIIdleworkersCollector struct {
	baseCollector

	configFilePath string
	queryPath      string
	timeout        int
}

func init() {
	RegisterCollector("NerveUWSGIIdleworkers", newNerveUWSGIIdleworkers)
}

// Default values of configuration fields
func newNerveUWSGIIdleworkers(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	col := new(nerveUWSGIIdleworkersCollector)

	col.log = log
	col.channel = channel
	col.interval = initialInterval

	col.name = "NerveUWSGIIdleworkers"
	col.configFilePath = "/etc/nerve/nerve.conf.json"
	col.queryPath = "status/uwsgi"
	col.timeout = 2

	return col
}

// Rewrites config variables from the global config
func (n *nerveUWSGIIdleworkersCollector) Configure(configMap map[string]interface{}) {
	if val, exists := configMap["queryPath"]; exists {
		n.queryPath = val.(string)
	}
	if val, exists := configMap["configFilePath"]; exists {
		n.configFilePath = val.(string)
	}

	if val, exists := configMap["http_timeout"]; exists {
		n.timeout = config.GetAsInt(val, 2)
	}

	n.configureCommonParams(configMap)
}

// Parses nerve config from HTTP uWSGI stats endpoints
func (n *nerveUWSGIIdleworkersCollector) Collect() {
	rawFileContents, err := ioutil.ReadFile(n.configFilePath)
	if err != nil {
		n.log.Warn("Failed to read the contents of file ", n.configFilePath, " because ", err)
		return
	}

	services, err := util.ParseNerveConfig(&rawFileContents, false)
	if err != nil {
		n.log.Warn("Failed to parse the nerve config at ", n.configFilePath, ": ", err)
		return
	}
	n.log.Debug("Finished parsing Nerve config into ", services)

	for _, service := range services {
		go n.queryService(service.Name, service.Hostname, service.Port)
	}
}

// Fetches and computes status stats from an HTTP endpoint
func (n *nerveUWSGIIdleworkersCollector) queryService(serviceName string, hostname string, port int) {
	serviceLog := n.log.WithField("service", serviceName)

	endpoint := fmt.Sprintf("http://%s:%d/%s", hostname, port, n.queryPath)
	serviceLog.Debug("making GET request to ", endpoint)

	rawResponse, err := readJSONFromEndpoint(endpoint, n.timeout)
	if err != nil {
		serviceLog.Warn("Failed to query endpoint ", endpoint, ": ", err)
		return
	}
	metrics, err := parseJSONData(rawResponse)
	if err != nil {
		serviceLog.Warn("Failed to parse response into metrics: ", err)
		return
	}

	metric.AddToAll(&metrics, map[string]string{
		"service": serviceName,
		"port":    strconv.Itoa(port),
	})
	serviceLog.Debug("Sending ", len(metrics), " to channel")
	for _, m := range metrics {
		n.Channel() <- m
	}
}

// Counts status stats from JSON content
func parseJSONData(raw []byte) ([]metric.Metric, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal(raw, &result)
	results := []metric.Metric{}
	if err != nil {
		return results, err
	}
	registry := make(map[string]int)
	registry["IdleWorkers"] = 0
	registry["BusyWorkers"] = 0
	registry["SigWorkers"] = 0
	registry["PauseWorkers"] = 0
	registry["CheapWorkers"] = 0
	registry["UnknownStateWorkers"] = 0
	workers, ok := result["workers"].([]interface{})
	if !ok {
		return results, fmt.Errorf("\"workers\" field not found or not an array")
	}
	for _, worker := range workers {
		workerMap, ok := worker.(map[string]interface{})
		if !ok {
			return results, fmt.Errorf("worker record is not a map")
		}
		status, ok := workerMap["status"].(string)
		if !ok {
			return results, fmt.Errorf("status not found or not a string")
		}
		if strings.Index(status, "sig") == 0 {
			status = "sig"
		}
		metricName := strings.Title(status) + "Workers"
		_, exists := registry[metricName]
		if !exists {
			metricName = "UnknownStateWorkers"
		}
		registry[metricName]++
	}
	for key, value := range registry {
		results = append(results, metric.WithValue(key, float64(value)))
	}
	return results, err
}

// Fetches the JSON stats content from HTTP endpoint
func readJSONFromEndpoint(endpoint string, timeout int) ([]byte, error) {
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	rsp, err := client.Get(endpoint)

	if rsp != nil {
		defer func() {
			io.Copy(ioutil.Discard, rsp.Body)
			rsp.Body.Close()
		}()
	}

	if err != nil {
		return []byte{}, err
	}

	if rsp != nil && rsp.StatusCode != 200 {
		err := fmt.Errorf("%s returned %d error code", endpoint, rsp.StatusCode)
		return []byte{}, err
	}

	txt, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return []byte{}, err
	}

	return txt, nil
}
