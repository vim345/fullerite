package collector

import (
	"encoding/json"
	"fmt"
	"fullerite/metric"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	l "github.com/Sirupsen/logrus"
)

var (
	sendMarathonMetrics = (*MarathonStats).sendMarathonMetrics
	getMarathonMetrics  = (*MarathonStats).getMarathonMetrics

	getMarathonMetricsURL = func(ip string) string { return fmt.Sprintf("http://%s/metrics", ip) }
	getMarathonLeaderURL  = func(ip string) string { return fmt.Sprintf("http://%s/v2/leader", ip) }
)

const (
	marathonGetTimeout = 10 * time.Second
)

// MarathonStats Collector for marathon leader stats
type MarathonStats struct {
	baseCollector
	IP     string
	client http.Client
}

type buildError struct {
	Reason string
}

func (e buildError) Error() string {
	return e.Reason
}

func init() {
	RegisterCollector("MarathonStats", newMarathonStats)
}

func newMarathonStats(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	m := new(MarathonStats)

	m.log = log
	m.channel = channel
	m.interval = initialInterval
	m.name = "MarathonStats"
	m.client = http.Client{Timeout: marathonGetTimeout}

	if ip, err := externalIP(); err != nil {
		m.log.Error("Cannot determine IP: ", err.Error())
	} else {
		m.IP = ip
	}

	return m
}

// Configure just calls the default configure
func (m *MarathonStats) Configure(configMap map[string]interface{}) {
	m.configureCommonParams(configMap)
}

func (m *MarathonStats) isLeader() bool {
	url := getMarathonLeaderURL(m.IP)
	r, err := m.client.Get(url)

	if err != nil {
		m.log.Error("Could not load leader from marathon", err.Error())
		return false
	}

	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		m.log.Error("Not 200 response code from marathon: ", r.Status)
		return false
	}

	contents, _ := ioutil.ReadAll(r.Body)

	var leadermap map[string]string
	decodeErr := json.Unmarshal([]byte(contents), &leadermap)

	if decodeErr != nil {
		m.log.Error("Unable to decode leader JSON: ", decodeErr.Error())
		return false
	}

	leader, exists := leadermap["leader"]

	if !exists {
		m.log.Error("Unable to find leader in leader JSON")
		return false
	}

	s := strings.Split(leader, ":")

	hostname, err := os.Hostname()

	if err != nil {
		m.log.Error("Cannot determine hostname: ", err.Error())
	}

	return s[0] == hostname
}

// Collect compares the leader against this hosts's hostaname and sends metrics if this is the leader
func (m *MarathonStats) Collect() {
	if m.isLeader() {
		go sendMarathonMetrics(m)
	}
}

func toSendMetric(name string) bool {
	if name == "counters" {
		return true
	} else if name == "gauges" {
		return true
	} else {
		return false
	}
}

func (m *MarathonStats) sendMarathonMetrics() {
	metrics := m.getMarathonMetrics()

	for metricType, metricValues := range metrics {
		if toSendMetric(metricType) {
			for k, v := range metricValues {
				s, err := buildMarathonMetric(metricType, k, v)
				if err != nil {
					m.log.Error("Error building Marathon Metric: ", err)
				}
				m.Channel() <- s
			}
		}
	}
}

func (m *MarathonStats) getMarathonMetrics() map[string]map[string]interface{} {
	url := getMarathonMetricsURL(m.IP)
	r, err := m.client.Get(url)

	if err != nil {
		m.log.Error("Could not load metrics from marathon", err.Error())
		return nil
	}

	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		m.log.Error("Not 200 response code from marathon", r.Status)
		return nil
	}

	contents, _ := ioutil.ReadAll(r.Body)
	var types map[string]string

	decodeErr := json.Unmarshal([]byte(contents), &types)

	if decodeErr != nil {
		m.log.Error("Unable to decode marathon metrics JSON: ", decodeErr.Error())
		return nil
	}

	var allMetrics map[string]map[string]interface{}
	var metric map[string]interface{}

	for k, v := range types {
		if k != "version" {
			decodeErr := json.Unmarshal([]byte(v), &metric)

			if decodeErr != nil {
				m.log.Error("Unable to decode marathon metrics JSON: ", decodeErr.Error())
				return nil
			}

			allMetrics[k] = metric
		}
	}

	return allMetrics
}

func buildMarathonMetric(metricType string, k string, v interface{}) (metric.Metric, error) {
	var m metric.Metric
	var err error
	switch metricType {
	case "gauges":
		m, err = buildMarathonMetricWithValueName(k, v.(string), "value")
	case "counters":
		m, err = buildMarathonMetricWithValueName(k, v.(string), "count")
	default:
		err = buildError{fmt.Sprintf("%s is not a supported metric type", metricType)}
	}

	return m, err
}

func buildMarathonMetricWithValueName(k string, v string, valueName string) (metric.Metric, error) {
	var valueMap map[string]float64
	var m metric.Metric

	decodeErr := json.Unmarshal([]byte(v), &valueMap)

	if decodeErr != nil {
		return m, decodeErr
	}

	m = metric.WithValue(k, valueMap[valueName])

	return m, nil
}
