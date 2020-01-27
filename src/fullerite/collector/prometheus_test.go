package collector

import (
	"testing"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"fullerite/metric"
)

func TestPrometheusNewPrometheus(t *testing.T) {
	expectedChan := make(chan metric.Metric)
	var expectedLogger = defaultLog.WithFields(l.Fields{"collector": "fullerite"})

	p := newPrometheus(expectedChan, 10, expectedLogger).(*Prometheus)

	assert.Equal(t, p.log, expectedLogger)
	assert.Equal(t, p.channel, expectedChan)
	assert.Equal(t, p.interval, 10)
	assert.Equal(t, p.name, "Prometheus")
}

func TestPrometheusConfigure(t *testing.T) {
	expectedChan := make(chan metric.Metric)
	var expectedLogger = defaultLog.WithFields(l.Fields{"collector": "fullerite"})

	p := newPrometheus(expectedChan, 10, expectedLogger).(*Prometheus)

	p.Configure(map[string]interface{}{
		"endpoints": []interface{}{
			map[string]interface{}{
				"prefix":         "123/",
				"url":            "https://etcd1.nowhere.com:2379/metrics",
				"serverCaFile":   "/tmp/baz/server-ca.crt",
				"clientCertFile": "/tmp/baz/client-etcd1-nowhere.com.crt",
				"clientKeyFile":  "/tmp/baz/client-etcd1-nowhere.com.key",
				"metrics_whitelist": []string{
					"123",
					"456",
				},
				"metrics_blacklist": []string{
					"78",
				},
				"generated_dimensions": map[string]interface{}{
					"foo": "bar",
				},
			},
		},
	})

	var endpoint *Endpoint
	endpoint = p.endpoints[0]

	assert.Equal(t, endpoint.prefix, "123/")
	assert.Equal(t, endpoint.url, "https://etcd1.nowhere.com:2379/metrics")
	assert.Equal(t, endpoint.serverCaFile, "/tmp/baz/server-ca.crt")
	assert.Equal(t, endpoint.clientCertFile, "/tmp/baz/client-etcd1-nowhere.com.crt")
	assert.Equal(t, endpoint.clientKeyFile, "/tmp/baz/client-etcd1-nowhere.com.key")
	assert.Equal(t, endpoint.timeout, 5)
	assert.Equal(t, *endpoint.metricsWhitelist, map[string]bool{"123": true, "456": true})
	assert.Equal(t, *endpoint.metricsBlacklist, map[string]bool{"78": true})
	assert.Equal(t, endpoint.generatedDimensions, map[string]string{"foo": "bar"})
}
