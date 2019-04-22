package collector

import (
	"fmt"
	"net/http"
	"testing"

	"fullerite/metric"
	"net/http/httptest"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNginxStatsNewNginxStats(t *testing.T) {
	channel := make(chan metric.Metric)
	log := defaultLog.WithFields(l.Fields{"collector": "Nginx"})
	stats := newNginxStats(channel, 10, log).(*nginxStats)

	assert.Equal(t, channel, stats.channel)
	assert.Equal(t, 10, stats.interval)
	assert.Equal(t, log, stats.log)
	assert.Equal(t, http.Client{Timeout: getTimeout}, stats.client)
}

func TestNginxStatsConfigureDefaults(t *testing.T) {
	channel := make(chan metric.Metric)
	log := defaultLog.WithFields(l.Fields{"collector": "Nginx"})
	stats := newNginxStats(channel, 10, log).(*nginxStats)

	configMap := map[string]interface{}{}
	stats.Configure(configMap)
	assert.Equal(t, stats.statsURL, "http://localhost:8080/nginx_status")
}

func TestNginxStatsConfigureCustomStatsLocation(t *testing.T) {
	channel := make(chan metric.Metric)
	log := defaultLog.WithFields(l.Fields{"collector": "Nginx"})
	stats := newNginxStats(channel, 10, log).(*nginxStats)

	configMap := map[string]interface{}{
		"reqHost": "yelp.com",
		"reqPort": "1234",
		"reqPath": "/my-cool-status",
	}
	stats.Configure(configMap)
	assert.Equal(t, stats.statsURL, "http://yelp.com:1234/my-cool-status")
}

func TestNginxStatsQueryNginxStatsSuccess(t *testing.T) {
	channel := make(chan metric.Metric)
	log := defaultLog.WithFields(l.Fields{"collector": "Nginx"})
	stats := newNginxStats(channel, 10, log).(*nginxStats)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "some response here")
	}))
	defer ts.Close()

	stats.statsURL = ts.URL
	contents, err := queryNginxStats(stats.client, stats.statsURL)
	assert.Equal(t, err, nil)
	assert.Equal(t, contents, "some response here\n")
}

func TestNginxStatsQueryNginxStatsFailure(t *testing.T) {
	channel := make(chan metric.Metric)
	log := defaultLog.WithFields(l.Fields{"collector": "Nginx"})
	stats := newNginxStats(channel, 10, log).(*nginxStats)

	stats.statsURL = "invalid-url"
	contents, err := queryNginxStats(stats.client, stats.statsURL)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, contents, "")
}

func TestBuildNginxMetric(t *testing.T) {
	expected := metric.Metric{"nginx.test", "gauge", 0.1, map[string]string{}}
	actual := buildNginxMetric("nginx.test", metric.Gauge, 0.1)
	assert.Equal(t, expected, actual)
}

func TestGetNginxMetrics(t *testing.T) {
	channel := make(chan metric.Metric)
	log := defaultLog.WithFields(l.Fields{"collector": "Nginx"})
	stats := newNginxStats(channel, 10, log).(*nginxStats)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Response copied verbatim from an nginx status page from the routing service.
		fmt.Fprintln(w, "Active connections: 2")
		fmt.Fprintln(w, "server accepts handled requests")
		fmt.Fprintln(w, " 82130 82130 84211")
		fmt.Fprintln(w, "Reading: 0 Writing: 1 Waiting: 1")
	}))
	defer ts.Close()

	stats.statsURL = ts.URL
	metrics := getNginxMetrics(stats.client, stats.statsURL, stats.log)
	assert.Equal(t, metrics, []metric.Metric{
		buildNginxMetric("nginx.active_connections", metric.Gauge, 2),
		buildNginxMetric("nginx.conn_accepted", metric.CumulativeCounter, 82130),
		buildNginxMetric("nginx.conn_handled", metric.CumulativeCounter, 82130),
		buildNginxMetric("nginx.req_handled", metric.CumulativeCounter, 84211),
		buildNginxMetric("nginx.req_per_conn", metric.Gauge, 84211.0/82130.0),
		buildNginxMetric("nginx.act_reads", metric.Gauge, 0),
		buildNginxMetric("nginx.act_writes", metric.Gauge, 1),
		buildNginxMetric("nginx.act_waits", metric.Gauge, 1),
	})
}
