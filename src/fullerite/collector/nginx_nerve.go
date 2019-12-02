package collector

import (
	"fmt"
	"fullerite/config"
	"fullerite/metric"
	"fullerite/util"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	l "github.com/Sirupsen/logrus"
)

type nginxNerveStats struct {
	baseCollector
	client            http.Client
	nerveConfigPath   string
	serviceNameToPath map[string]string
}

var (
	servicePathKeyRE = regexp.MustCompile(`^servicePath\.(.+)$`)
)

func init() {
	RegisterCollector("NginxNerveStats", newNginxNerveStats)
}

func newNginxNerveStats(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	m := new(nginxNerveStats)

	m.log = log
	m.channel = channel
	m.interval = initialInterval
	m.name = "NginxNerveStats"
	m.nerveConfigPath = "/etc/nerve/nerve.conf.json"
	m.client = http.Client{Timeout: nginxGetTimeout}

	return m
}

func (m *nginxNerveStats) Configure(configMap map[string]interface{}) {
	m.configureCommonParams(configMap)
	c := config.GetAsMap(configMap)

	// Convert config keys/values like "servicePath.routing" into a mapping of
	// service name to nginx status path.
	m.serviceNameToPath = make(map[string]string)
	for key, value := range c {
		if match := servicePathKeyRE.FindStringSubmatch(key); match != nil {
			m.serviceNameToPath[match[1]] = value
		}
	}
}

func (m *nginxNerveStats) Collect() {
	rawFileContents, err := ioutil.ReadFile(m.nerveConfigPath)
	if err != nil {
		m.log.Warn("Failed to read the contents of file ", m.nerveConfigPath, " because ", err)
		return
	}
	services, err := util.ParseNerveConfig(&rawFileContents, true)
	if err != nil {
		m.log.Warn("Failed to parse the nerve config at ", m.nerveConfigPath, ": ", err)
		return
	}
	m.log.Debug("Finished parsing Nerve config into ", services)

	for _, service := range services {
		if path, exists := m.serviceNameToPath[service.Name]; exists {
			go m.collectMetricsForService(service, path)
		}
	}
}

func (m *nginxNerveStats) collectMetricsForService(service util.NerveService, path string) {
	serviceLog := m.log.WithField("service", service.Name)
	statsURL := fmt.Sprintf("http://%s:%d%s", service.Host, service.Port, path)

	serviceLog.Debug("Fetching nginx stats from", statsURL)
	metrics := getNginxMetrics(m.client, statsURL, serviceLog)

	metric.AddToAll(&metrics, map[string]string{
		"service_name":      service.Name,
		"service_namespace": service.Namespace,
		"port":              strconv.Itoa(service.Port),
	})

	for _, metric := range metrics {
		m.Channel() <- metric
	}
}
