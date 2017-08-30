package collector

// Collects metrics produced by chronos. Simply pulls /metrics from the chronos
//  leader and sends all well-formated metrics

import (
	"fmt"
	"fullerite/config"
	"fullerite/dropwizard"
	"fullerite/metric"
	"fullerite/util"
	"net/http"
	"time"

	l "github.com/Sirupsen/logrus"
)

var (
	sendChronosMetrics = (*ChronosStats).sendChronosMetrics
	getChronosMetrics  = (*ChronosStats).getChronosMetrics

	getChronosMetricsURL = func(host string) string { return fmt.Sprintf("http://%s/metrics", host) }
)

const (
	chronosGetTimeout = 10 * time.Second
)

// ChronosStats Collector for chronos leader stats
type ChronosStats struct {
	baseCollector
	client          http.Client
	chronosHost     string
	extraDimensions map[string]string
}

func init() {
	RegisterCollector("ChronosStats", newChronosStats)
}

func newChronosStats(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	m := new(ChronosStats)

	m.log = log
	m.channel = channel
	m.interval = initialInterval
	m.name = "ChronosStats"
	m.client = http.Client{Timeout: chronosGetTimeout}
	m.extraDimensions = make(map[string]string)

	return m
}

// Configure just calls the default configure
func (m *ChronosStats) Configure(configMap map[string]interface{}) {
	m.configureCommonParams(configMap)

	c := config.GetAsMap(configMap)
	if chronosHost, exists := c["chronosHost"]; exists && len(chronosHost) > 0 {
		m.chronosHost = chronosHost
	} else {
		m.log.Error("Chronos host not specified in config")
	}

	if extraDims, exists := configMap["extraDimensions"]; exists {
		dims := config.GetAsMap(extraDims)
		for dim, value := range dims {
			m.extraDimensions[dim] = value
		}
	}
}

// Collect compares the leader against this hosts's hostaname and sends metrics if this is the leader
func (m *ChronosStats) Collect() {
	// Non-chronos-leaders forward requests to the leader, so only the leader's metrics matter
	if leader, err := util.IsLeader(m.chronosHost, "leader", m.client); leader && err == nil {
		go sendChronosMetrics(m)
	} else if err != nil {
		m.log.Error("Error finding leader: ", err)
	} else {
		m.log.Debug("Not the leader, not sending metrics")
	}
}

func (m *ChronosStats) sendChronosMetrics() {
	metrics := getChronosMetrics(m)
	for _, metric := range metrics {
		if !m.ContainsBlacklistedDimension(metric.Dimensions) {
			m.Channel() <- metric
		}
	}
}

func (m *ChronosStats) getChronosMetrics() []metric.Metric {
	url := getChronosMetricsURL(m.chronosHost)

	contents, err := util.GetWrapper(url, m.client)
	if err != nil {
		m.log.Error("Could not load metrics from chronos: ", err.Error())
		return nil
	}

	metrics, err := dropwizard.Parse(contents, "java-1.1", true)

	if err != nil {
		m.log.Error("Unable to decode chronos metrics JSON: ", err)
		return nil
	}

	metric.AddToAll(&metrics, map[string]string{
		"service": "chronos",
	})

	metric.AddToAll(&metrics, m.extraDimensions)

	return metrics
}
