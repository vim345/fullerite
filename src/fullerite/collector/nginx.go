package collector

import (
	"fmt"
	"fullerite/config"
	"fullerite/metric"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	l "github.com/Sirupsen/logrus"
)

type NginxStats struct {
	baseCollector
	client   http.Client
	statsURL string
}

const (
	nginxGetTimeout = 10 * time.Second
)

var (
	activeConnectionsRE = regexp.MustCompile(`^Active connections: (?P<conn>\d+)`)
	totalConnectionsRE  = regexp.MustCompile(
		`^\s+(?P<conn>\d+)\s+` +
			`(?P<acc>\d+)\s+(?P<req>\d+)`,
	)
	connectionStatusRE = regexp.MustCompile(
		`^Reading: (?P<reading>\d+) ` +
			`Writing: (?P<writing>\d+) ` +
			`Waiting: (?P<waiting>\d+)`,
	)
)

func init() {
	RegisterCollector("NginxStats", newNginxStats)
}

func newNginxStats(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	m := new(NginxStats)

	m.log = log
	m.channel = channel
	m.interval = initialInterval
	m.name = "NginxStats"
	m.client = http.Client{Timeout: nginxGetTimeout}

	return m
}

func (m *NginxStats) Configure(configMap map[string]interface{}) {
	m.configureCommonParams(configMap)

	c := config.GetAsMap(configMap)

	host := "localhost"
	port := "8080"
	path := "/nginx_status"

	if val, exists := c["reqHost"]; exists {
		host = val
	}
	if val, exists := c["reqPort"]; exists {
		port = val
	}
	if val, exists := c["reqPath"]; exists {
		path = val
	}

	m.statsURL = fmt.Sprintf("http://%s:%s%s", host, port, path)
}

func (m *NginxStats) Collect() {
	for _, metric := range getNginxMetrics(m.client, m.statsURL, m.log) {
		m.Channel() <- metric
	}
}

func queryNginxStats(client http.Client, statsURL string) (string, error) {
	rsp, err := client.Get(statsURL)

	if rsp != nil {
		defer func() {
			io.Copy(ioutil.Discard, rsp.Body)
			rsp.Body.Close()
		}()
	}

	if err != nil {
		return "", err
	}

	if rsp != nil && rsp.StatusCode != 200 {
		err := fmt.Errorf("%s returned %d error code", statsURL, rsp.StatusCode)
		return "", err
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func buildNginxMetric(name string, metricType string, value float64) (m metric.Metric) {
	m = metric.New(name)
	m.MetricType = metricType
	m.Value = value
	return m
}

func getNginxMetrics(client http.Client, statsURL string, log *l.Entry) []metric.Metric {
	contents, err := queryNginxStats(client, statsURL)
	if err != nil {
		log.Error("Could not load stats from nginx: ", err.Error())
		return nil
	}

	metrics := []metric.Metric{}

	for _, line := range strings.Split(contents, "\n") {
		if match := activeConnectionsRE.FindStringSubmatch(line); match != nil {
			conn, _ := strconv.ParseFloat(match[1], 64)
			metrics = append(
				metrics,
				buildNginxMetric("nginx.active_connections", metric.Gauge, conn),
			)
		} else if match := totalConnectionsRE.FindStringSubmatch(line); match != nil {
			conn, _ := strconv.ParseFloat(match[1], 64)
			acc, _ := strconv.ParseFloat(match[2], 64)
			req, _ := strconv.ParseFloat(match[3], 64)
			req_per_conn := req / acc

			metrics = append(
				metrics,
				buildNginxMetric("nginx.conn_accepted", metric.CumulativeCounter, conn),
				buildNginxMetric("nginx.conn_handled", metric.CumulativeCounter, acc),
				buildNginxMetric("nginx.req_handled", metric.CumulativeCounter, req),
				buildNginxMetric("nginx.req_per_conn", metric.Gauge, req_per_conn),
			)
		} else if match := connectionStatusRE.FindStringSubmatch(line); match != nil {
			reading, _ := strconv.ParseFloat(match[1], 64)
			writing, _ := strconv.ParseFloat(match[2], 64)
			waiting, _ := strconv.ParseFloat(match[3], 64)
			metrics = append(
				metrics,
				buildNginxMetric("nginx.act_reads", metric.Gauge, reading),
				buildNginxMetric("nginx.act_writes", metric.Gauge, writing),
				buildNginxMetric("nginx.act_waits", metric.Gauge, waiting),
			)
		}
	}

	return metrics
}
