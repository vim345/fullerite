package collector

import (
	"fullerite/config"
	"fullerite/dropwizard"
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

const (
	// MetricTypeCounter String for counter metric type
	MetricTypeCounter string = "COUNTER"
	// MetricTypeGauge String for Gauge metric type
	MetricTypeGauge string = "GAUGE"
)

type nerveUWSGICollector struct {
	baseCollector

	configFilePath    string
	queryPath         string
	timeout           int
	servicesWhitelist []string
}

func init() {
	RegisterCollector("NerveUWSGI", newNerveUWSGI)
}

func newNerveUWSGI(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	col := new(nerveUWSGICollector)

	col.log = log
	col.channel = channel
	col.interval = initialInterval

	col.name = "NerveUWSGI"
	col.configFilePath = "/etc/nerve/nerve.conf.json"
	col.queryPath = "status/metrics"
	col.timeout = 2

	return col
}

func (n *nerveUWSGICollector) Configure(configMap map[string]interface{}) {
	if val, exists := configMap["queryPath"]; exists {
		n.queryPath = val.(string)
	}
	if val, exists := configMap["configFilePath"]; exists {
		n.configFilePath = val.(string)
	}
	if val, exists := configMap["servicesWhitelist"]; exists {
		n.servicesWhitelist = config.GetAsSlice(val)
	}

	n.configureCommonParams(configMap)
}

func (n *nerveUWSGICollector) Collect() {
	rawFileContents, err := ioutil.ReadFile(n.configFilePath)
	if err != nil {
		n.log.Warn("Failed to read the contents of file ", n.configFilePath, " because ", err)
		return
	}

	servicePortMap, err := util.ParseNerveConfig(&rawFileContents)
	if err != nil {
		n.log.Warn("Failed to parse the nerve config at ", n.configFilePath, ": ", err)
		return
	}
	n.log.Debug("Finished parsing Nerve config into ", servicePortMap)

	for port, service := range servicePortMap {
		go n.queryService(service.Name, port)
	}
}

func (n *nerveUWSGICollector) queryService(serviceName string, port int) {
	serviceLog := n.log.WithField("service", serviceName)

	endpoint := fmt.Sprintf("http://localhost:%d/%s", port, n.queryPath)
	serviceLog.Debug("making GET request to ", endpoint)

	rawResponse, schemaVer, err := queryEndpoint(endpoint, n.timeout)
	if err != nil {
		serviceLog.Warn("Failed to query endpoint ", endpoint, ": ", err)
		return
	}
	metrics, err := dropwizard.Parse(rawResponse, schemaVer, n.serviceInWhitelist(serviceName))
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

func queryEndpoint(endpoint string, timeout int) ([]byte, string, error) {
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
		return []byte{}, "", err
	}

	if rsp != nil && rsp.StatusCode != 200 {
		err := fmt.Errorf("%s returned %d error code", endpoint, rsp.StatusCode)
		return []byte{}, "", err
	}

	schemaVer := rsp.Header.Get("Metrics-Schema")
	if schemaVer == "" {
		schemaVer = "default"
	}

	txt, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return []byte{}, "", err
	}

	return txt, schemaVer, nil
}

// serviceInWhitelist returns true if the service name passed as argument
// is found among the ones whitelisted by the user
func (n *nerveUWSGICollector) serviceInWhitelist(service string) bool {
	for _, s := range n.servicesWhitelist {
		if s == service {
			return true
		}
	}
	return false
}
