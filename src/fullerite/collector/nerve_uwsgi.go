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
	"strings"
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

	configFilePath            string
	queryPath                 string
	timeout                   int
	cumCounterEnabledServices []string
	workersStatsEnabled       bool
	workersStatsQueryPath     string
	workersStatsBlacklist     []string
	servicesMetricsWhitelist  []string
	servicesMetricsBlacklist  []string
}

func init() {
	RegisterCollector("NerveUWSGI", newNerveUWSGI)
}

// Default values of configuration fields
func newNerveUWSGI(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	col := new(nerveUWSGICollector)

	col.log = log
	col.channel = channel
	col.interval = initialInterval

	col.name = "NerveUWSGI"
	col.configFilePath = "/etc/nerve/nerve.conf.json"
	col.queryPath = "status/metrics"
	col.workersStatsQueryPath = "status/uwsgi"
	col.timeout = 2

	return col
}

// Rewrites config variables from the global config
func (n *nerveUWSGICollector) Configure(configMap map[string]interface{}) {
	if val, exists := configMap["queryPath"]; exists {
		n.queryPath = val.(string)
	}
	if val, exists := configMap["configFilePath"]; exists {
		n.configFilePath = val.(string)
	}
	if val, exists := configMap["cumCounterEnabledServices"]; exists {
		n.cumCounterEnabledServices = config.GetAsSlice(val)
	}
	if val, exists := configMap["workersStatsBlacklist"]; exists {
		n.workersStatsBlacklist = config.GetAsSlice(val)
	}
	if val, exists := configMap["workersStatsEnabled"]; exists {
		n.workersStatsEnabled = config.GetAsBool(val, false)
	}
	if val, exists := configMap["workersStatsQueryPath"]; exists {
		n.workersStatsQueryPath = val.(string)
	}
	if val, exists := configMap["http_timeout"]; exists {
		n.timeout = config.GetAsInt(val, 2)
	}
	if val, exists := configMap["servicesMetricsWhitelist"]; exists {
		n.servicesMetricsWhitelist = config.GetAsSlice(val)
	}
	if val, exists := configMap["servicesMetricsBlacklist"]; exists {
		if len(n.servicesMetricsWhitelist) > 0 {
			n.log.Error("Only whitelist or blacklist is allowed. Cannot configure with both")
			return
		}
		n.servicesMetricsBlacklist = config.GetAsSlice(val)
	}

	n.configureCommonParams(configMap)
}

// Parses nerve config from HTTP uWSGI stats endpoints
func (n *nerveUWSGICollector) Collect() {
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

	// We need to check multiple conditions here and this is dependent on
	// whether we have a blacklist or whitelist
	for _, service := range services {
		if len(n.servicesMetricsBlacklist) == 0 && len(n.servicesMetricsWhitelist) == 0 {
			go n.queryService(service.Name, service.Port)
		} else if len(n.servicesMetricsWhitelist) > 0 && n.serviceInList(service.Name, n.servicesMetricsWhitelist) {
			go n.queryService(service.Name, service.Port)
		} else {
			if !n.serviceInList(service.Name, n.servicesMetricsBlacklist) {
				go n.queryService(service.Name, service.Port)
			}
		}
	}
}

// Fetches and computes stats from metrics HTTP endpoint,
// calls an additional endpoint if UWSGI is detected
func (n *nerveUWSGICollector) queryService(serviceName string, port int) {
	serviceLog := n.log.WithField("service", serviceName)
	endpoint := fmt.Sprintf("http://localhost:%d/%s", port, n.queryPath)
	serviceLog.Debug("making GET request to ", endpoint)
	rawResponse, schemaVer, err := queryEndpoint(endpoint, n.timeout)
	if err != nil {
		serviceLog.Warn("Failed to query endpoint ", endpoint, ": ", err)
		return
	}
	metrics, err := dropwizard.Parse(
		rawResponse,
		schemaVer,
		n.serviceInList(serviceName, n.cumCounterEnabledServices))
	if err != nil {
		serviceLog.Warn("Failed to parse response into metrics: ", err)
		return
	}
	// If we detect metrics from uwsgi, we try to fetch an additional Workers info
	// If this was a separate collector, there would be no way to figure that out
	// without a costly additional HTTP call.
	// This prevent us from having to maintain a whitelist of services to query
	// or from flooding all non UWSGI services with these requests.
	// We still maintain a blacklist just in case
	if strings.Contains(schemaVer, "uwsgi") && n.workersStatsEnabled && !n.serviceInList(serviceName, n.workersStatsBlacklist) {
		extraDims := dropwizard.ExtractServiceDims(rawResponse)
		serviceLog.Debug("Trying to fetch workers stats")
		uwsgiWorkerStatsEndpoint := fmt.Sprintf("http://localhost:%d/%s", port, n.workersStatsQueryPath)
		uwsgiWorkerStatsMetrics, err := n.tryFetchUWSGIWorkersStats(serviceName, uwsgiWorkerStatsEndpoint)
		if err != nil {
			serviceLog.Info("Could not get additional worker stat metrics")
		} else {
			// Add the metrics to our existing ones so we get the post process for free.
			serviceLog.Debug("Additional workers metrics collected: ", len(uwsgiWorkerStatsMetrics))
			metric.AddToAll(&uwsgiWorkerStatsMetrics, extraDims)
			for _, v := range uwsgiWorkerStatsMetrics {
				metrics = append(metrics, v)
			}
		}
	}

	metric.AddToAll(&metrics, map[string]string{
		"service": serviceName,
		"port":    strconv.Itoa(port),
	})
	serviceLog.Debug("Sending ", len(metrics), " to channel")
	for _, m := range metrics {
		if !n.ContainsBlacklistedDimension(m.Dimensions) {
			n.Channel() <- m
		}
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

// Checks for service name in a list and returns true if the name is present
// and returns False if it is not. This is used by multiple functions.
func (n *nerveUWSGICollector) serviceInList(service string, list []string) bool {
	for _, s := range list {
		if s == service {
			return true
		}
	}
	return false
}

// Fetches and computes status stats from an HTTP endpoint
func (n *nerveUWSGICollector) tryFetchUWSGIWorkersStats(serviceName string, endpoint string) ([]metric.Metric, error) {
	emptyResult := []metric.Metric{}
	serviceLog := n.log.WithField("service", serviceName)
	serviceLog.Debug("making GET request to ", endpoint)
	rawResponse, _, err := queryEndpoint(endpoint, n.timeout)
	if err != nil {
		serviceLog.Info("Failed to query workers stats endpoint ", endpoint, ": ", err)
		return emptyResult, err
	}
	metrics, err := util.ParseUWSGIWorkersStats(rawResponse)
	if err != nil {
		serviceLog.Info("No workers stats retreived: ", err)
		return emptyResult, err
	}
	return metrics, nil
}
