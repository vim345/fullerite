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
	hostname = os.Hostname

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

	hostname, err := hostname()

	if err != nil {
		m.log.Error("Cannot determine hostname: ", err.Error())
		return false
	}

	return s[0] == hostname
}

// Collect compares the leader against this hosts's hostaname and sends metrics if this is the leader
func (m *MarathonStats) Collect() {
	if m.isLeader() {
		go sendMarathonMetrics(m)
	}
}

func (m *MarathonStats) sendMarathonMetrics() {
	metrics := getMarathonMetrics(m)

	for _, metric := range metrics {
		m.Channel() <- metric
	}
}

type parse func(map[string]interface{}, []metric.Metric)([]metric.Metric, error)

func metricMaker(valueName string, valueType string) parse {
	return func(value map[string]interface{}, metrics []metric.Metric)([]metric.Metric, error) {
		for k, v := range value {
			var met metric.Metric
			vmap, ok := v.(map[string]interface{})
			if !ok {
				return metrics, buildError{fmt.Sprintf("%s not in expected format", valueName)}
			}
			for k2, v2 := range vmap {
				if k2 == valueName {
					met = metric.WithValue("marathon." + k, v2.(float64))
					met.MetricType = valueType
					metrics = append(metrics, met)
					break
				}
			}
		}

		return metrics, nil
	}
}


func (m *MarathonStats) unmarshalJSON(b []byte) ([]metric.Metric, error) {
	var f interface{}
	decodeErr := json.Unmarshal(b, &f)

	if decodeErr != nil {
		return nil, buildError{fmt.Sprintf("Could not convert bytes to JSON: ", decodeErr)}
	}

	u, ok := f.(map[string]interface{})

	if !ok {
		return nil, buildError{"Could not convert JSON to map of strings"}
	}

	metrics := make([]metric.Metric, 0, len(u))

	jsonToMetricMap := map[string]parse {
		"gauges": metricMaker("value", metric.Gauge),
		"counters": metricMaker("count", metric.Counter),
	}

	for k, v := range u {
		f, exists := jsonToMetricMap[k]
		if exists {
			vmap, ok := v.(map[string]interface{})
			// if we have trouble with one metric, keep going
			if ok {
				var err error
				metrics, err = f(vmap, metrics)

				if err != nil {
					m.log.Warn("Could not decode ", vmap)
				}
			} else {
				m.log.Warn("Could not convert %s to proper format: ", k)
			}
		}
	}

	return metrics, nil
}


func (m *MarathonStats) getMarathonMetrics() []metric.Metric {
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

	metrics, err := m.unmarshalJSON([]byte(contents))

	if err != nil {
		m.log.Error("Unable to decode marathon metrics JSON: ", err)
		return nil
	}

	return metrics
}
