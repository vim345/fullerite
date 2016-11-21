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

type nerveUWSGIStatsCollector struct {
	baseCollector

	configFilePath    string
	queryPath         string
	timeout           int
	servicesWhitelist []string
}

func init() {
	RegisterCollector("NerveUWSGIStats", newNerveUWSGIStats)
}

func newNerveUWSGIStats(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	col := new(nerveUWSGIStatsCollector)

	col.log = log
	col.channel = channel
	col.interval = initialInterval

	col.name = "NerveUWSGIStats"
	col.configFilePath = "/etc/nerve/nerve.conf.json"
	col.queryPath = "status/uwsgi"
	col.timeout = 2

	return col
}

func (n *nerveUWSGIStatsCollector) Configure(configMap map[string]interface{}) {
	if val, exists := configMap["queryPath"]; exists {
		n.queryPath = val.(string)
	}
	if val, exists := configMap["configFilePath"]; exists {
		n.configFilePath = val.(string)
	}
	if val, exists := configMap["servicesWhitelist"]; exists {
		n.servicesWhitelist = config.GetAsSlice(val)
	}

	if val, exists := configMap["http_timeout"]; exists {
		n.timeout = config.GetAsInt(val, 2)
	}

	n.configureCommonParams(configMap)
}

func (n *nerveUWSGIStatsCollector) Collect() {
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
		go n.queryService(service.Name, service.Port)
	}
}

func (n *nerveUWSGIStatsCollector) queryService(serviceName string, port int) {
	serviceLog := n.log.WithField("service", serviceName)

	endpoint := fmt.Sprintf("http://localhost:%d/%s", port, n.queryPath)
	serviceLog.Debug("making GET request to ", endpoint)

	rawResponse, err := readJSONFromEndpoint(endpoint, n.timeout)
	if err != nil {
		serviceLog.Warn("Failed to query endpoint ", endpoint, ": ", err)
		return
	}
	metrics, err := parseJSONData(rawResponse, n.serviceInWhitelist(serviceName))
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

func parseJSONData(raw []byte, ccEnabled bool) ([]metric.Metric, error) {
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

// serviceInWhitelist returns true if the service name passed as argument
// is found among the ones whitelisted by the user
func (n *nerveUWSGIStatsCollector) serviceInWhitelist(service string) bool {
	for _, s := range n.servicesWhitelist {
		if s == service {
			return true
		}
	}
	return false
}
