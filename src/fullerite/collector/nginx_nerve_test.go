package collector

import (
	"fullerite/util"
	"net/http"
	"os"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"fullerite/metric"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNginxNerveStatsNewNginxNerveStats(t *testing.T) {
	channel := make(chan metric.Metric)
	log := defaultLog.WithFields(l.Fields{"collector": "NginxNerveStats"})
	stats := newNginxNerveStats(channel, 10, log).(*NginxNerveStats)

	assert.Equal(t, channel, stats.channel)
	assert.Equal(t, 10, stats.interval)
	assert.Equal(t, log, stats.log)
	assert.Equal(t, http.Client{Timeout: getTimeout}, stats.client)
}

func TestNginxNerveStatsConfigureDefaults(t *testing.T) {
	channel := make(chan metric.Metric)
	log := defaultLog.WithFields(l.Fields{"collector": "NginxNerveStats"})
	stats := newNginxNerveStats(channel, 10, log).(*NginxNerveStats)

	configMap := map[string]interface{}{}
	stats.Configure(configMap)
}

func TestNginxNerveStatsConfigure(t *testing.T) {
	channel := make(chan metric.Metric)
	log := defaultLog.WithFields(l.Fields{"collector": "NginxNerveStats"})
	stats := newNginxNerveStats(channel, 10, log).(*NginxNerveStats)

	configMap := map[string]interface{}{
		"servicePath.routing": "/_routing/nginx-status",
		"servicePath.spectre": "/nginx_status",
	}
	stats.Configure(configMap)

	assert.Equal(t, stats.serviceNameToPath, map[string]string{
		"routing": "/_routing/nginx-status",
		"spectre": "/nginx_status",
	})
}

func TestNginxNerveStatsCollect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Active connections: 2")
	}))
	defer ts.Close()
	ip, port := parseURL(ts.URL)
	minimalNerveConfig := util.CreateMinimalNerveConfig(map[string]util.EndPoint{
		"routing.main": util.EndPoint{ip, port},
	})

	tmpFile, err := ioutil.TempFile("", "fullerite_testing")
	defer os.Remove(tmpFile.Name())
	assert.Nil(t, err)

	marshalled, err := json.Marshal(minimalNerveConfig)
	assert.Nil(t, err)

	_, err = tmpFile.Write(marshalled)
	assert.Nil(t, err)

	cfg := map[string]interface{}{
		"servicePath.routing": "/_routing/nginx-status",
	}

	channel := make(chan metric.Metric)
	log := defaultLog.WithFields(l.Fields{"collector": "NginxNerveStats"})
	inst := newNginxNerveStats(channel, 10, log).(*NginxNerveStats)
	inst.nerveConfigPath = tmpFile.Name()
	inst.Configure(cfg)

	inst.Collect()
	metric := <-inst.Channel()
	assert.Equal(t, metric.Value, 2.0)
	assert.Equal(t, metric.Dimensions["service_name"], "routing")
}
