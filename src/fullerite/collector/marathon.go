package collector

// Collects metrics produced by marathon. Simply pulls /metrics from the marathon
//  leader and sends all well-formated metrics

import (
	"encoding/json"
	"fmt"
	"fullerite/config"
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

	getMarathonMetricsURL = func(host string) string { return fmt.Sprintf("http://%s/metrics", host) }
	getMarathonLeaderURL  = func(host string) string { return fmt.Sprintf("http://%s/v2/leader", host) }
)

const (
	marathonGetTimeout = 10 * time.Second
)

// MarathonStats Collector for marathon leader stats
type MarathonStats struct {
	baseCollector
	IP           string
	client       http.Client
	marathonHost string
}

type buildError struct {
	Reason string
}

func (e buildError) Error() string {
	return e.Reason
}

type httpError struct {
	Status int
}

func (e httpError) Error() string {
	return fmt.Sprintf("%s: %s", http.StatusText(e.Status), e.Status)
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

	c := config.GetAsMap(configMap)
	if marathonHost, exists := c["marathonHost"]; exists && len(marathonHost) > 0 {
		m.marathonHost = marathonHost
	} else {
		m.log.Error("Marathon host not specified in config")
	}
}

func (m *MarathonStats) isLeader() bool {
	url := getMarathonLeaderURL(m.marathonHost)

	contents, err := m.marathonGet(url)
	if err != nil {
		m.log.Error("Could not load metrics from marathon: ", err.Error())
		return false
	}

	var leadermap map[string]string

	if decodeErr := json.Unmarshal(contents, &leadermap); decodeErr != nil {
		m.log.Error("Unable to decode leader JSON: ", decodeErr.Error())
		return false
	}

	leader, exists := leadermap["leader"]
	if !exists {
		m.log.Error("Unable to find leader in leader JSON")
		return false
	}

	s := strings.Split(leader, ":")

	h, err := hostname()
	if err != nil {
		m.log.Error("Cannot determine hostname: ", err.Error())
		return false
	}

	return s[0] == h
}

// Collect compares the leader against this hosts's hostaname and sends metrics if this is the leader
func (m *MarathonStats) Collect() {
	// Non-marathon-leaders forward requests to the leader, so only the leader's metrics matter
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

// parse takes a map[string]interface{} and a running slice of metrics and returns
//  an updated slice with the metrics pulled from the map
type parse func(map[string]interface{}, []metric.Metric) ([]metric.Metric, error)

// metricMaker returns a function that will parse the metrics out of a map
// All the metric parsing is nearly identical; the only difference is the type
//  and what the value is named in the json
func metricMaker(valueName string, valueType string) parse {
	return func(value map[string]interface{}, metrics []metric.Metric) ([]metric.Metric, error) {
		for k, v := range value {
			var met metric.Metric

			vmap, ok := v.(map[string]interface{})
			if !ok {
				return metrics, buildError{fmt.Sprintf("%s not in expected format", valueName)}
			}

			v2, exists := vmap[valueName]
			if exists {
				if vfloat, ok := v2.(float64); ok {
					met = metric.WithValue("marathon."+k, vfloat)
					met.MetricType = valueType
					metrics = append(metrics, met)
				}
			}
		}

		return metrics, nil
	}
}

func (m *MarathonStats) unmarshalJSON(b []byte) ([]metric.Metric, error) {
	var f interface{}

	if decodeErr := json.Unmarshal(b, &f); decodeErr != nil {
		return nil, buildError{fmt.Sprintf("Could not convert bytes to JSON: ", decodeErr)}
	}

	u, ok := f.(map[string]interface{})
	if !ok {
		return nil, buildError{"Could not convert JSON to map of strings"}
	}

	metrics := make([]metric.Metric, 0, len(u))

	// Mapping from the name of a metric type returned from Marathon to the function
	//  that will parse it
	jsonToMetricMap := map[string]parse{
		"gauges":   metricMaker("value", metric.Gauge),
		"counters": metricMaker("count", metric.Counter),
		// Meters also include events per second, but we'll ignore those for now
		"meters": metricMaker("count", metric.Counter),
	}

	for k, v := range u {
		if f, exists := jsonToMetricMap[k]; exists {
			// if we have trouble with one metric, keep going
			if vmap, ok := v.(map[string]interface{}); ok {
				var err error

				if metrics, err = f(vmap, metrics); err != nil {
					m.log.Warn("Could not decode ", err)
				}
			} else {
				m.log.Warn("Could not convert %s to proper format: ", k)
			}
		}
	}

	return metrics, nil
}

func (m *MarathonStats) getMarathonMetrics() []metric.Metric {
	url := getMarathonMetricsURL(m.marathonHost)

	contents, err := m.marathonGet(url)
	if err != nil {
		m.log.Error("Could not load metrics from marathon: ", err.Error())
		return nil
	}

	metrics, err := m.unmarshalJSON(contents)
	if err != nil {
		m.log.Error("Unable to decode marathon metrics JSON: ", err)
		return nil
	}

	return metrics
}

func (m *MarathonStats) marathonGet(url string) ([]byte, error) {
	r, err := m.client.Get(url)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, httpError{r.StatusCode}
	}

	contents, _ := ioutil.ReadAll(r.Body)

	return []byte(contents), nil
}
