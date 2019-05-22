package collector

import (
	"fullerite/config"
	"fullerite/metric"
	"fullerite/util"

	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	l "github.com/Sirupsen/logrus"
)

// Nerve uWSGI worker stats collector
// fetches counts of idle and busy workers from uWSGI stats servers published over HTTP
// http://uwsgi-docs.readthedocs.io/en/latest/StatsServer.html
// It reads the location of uWSGI stats servers from SmartStack
type uWSGINerveWorkerStatsCollector struct {
	baseCollector

	configFilePath    string
	queryPath         string
	timeout           int
	servicesWhitelist []string
}

func init() {
	RegisterCollector("UWSGINerveWorkerStats", newUWSGINerveWorkerStats)
}

// Default values of configuration fields
func newUWSGINerveWorkerStats(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	col := new(uWSGINerveWorkerStatsCollector)

	col.log = log
	col.channel = channel
	col.interval = initialInterval

	col.name = "UWSGINerveWorkerStats"
	col.configFilePath = "/etc/nerve/nerve.conf.json"
	col.queryPath = "status/uwsgi"
	col.timeout = 2

	return col
}

// Rewrites config variables from the global config
func (n *uWSGINerveWorkerStatsCollector) Configure(configMap map[string]interface{}) {
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

// Parses nerve config from HTTP uWSGI stats endpoints
func (n *uWSGINerveWorkerStatsCollector) Collect() {
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
		if n.serviceInWhitelist(service) {
			go n.queryService(service.Name, service.Port)
		}
	}
}

// Fetches and computes status stats from an HTTP endpoint
func (n *uWSGINerveWorkerStatsCollector) queryService(serviceName string, port int) {
	serviceLog := n.log.WithField("service", serviceName)

	endpoint := fmt.Sprintf("http://localhost:%d/%s", port, n.queryPath)
	serviceLog.Debug("making GET request to ", endpoint)

	rawResponse, err := readJSONFromEndpoint(endpoint, n.timeout)
	if err != nil {
		serviceLog.Warn("Failed to query endpoint ", endpoint, ": ", err)
		return
	}
	metrics, err := util.ParseUWSGIWorkersStats(rawResponse)
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

// serviceInWhitelist returns true if the service name passed as argument
// is found among the ones whitelisted by the user
func (n *uWSGINerveWorkerStatsCollector) serviceInWhitelist(service util.NerveService) bool {
	for _, s := range n.servicesWhitelist {
		if s == service.Name {
			return true
		}
	}
	return false
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
