package collector

import (
	"fullerite/metric"
	"testing"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func getTestHTTPDropwizard() *httpDropwizardCollector {
	return newHTTPDropwizard(make(chan metric.Metric), 12, l.WithField("test", "httpDropwizard")).(*httpDropwizardCollector)
}

func TestDefaultConfigHttpDropwizard(t *testing.T) {
	c := getTestHTTPDropwizard()
	c.Configure(make(map[string]interface{}))

	assert.Equal(t, 12, c.Interval())
	assert.Equal(t, 3, c.timeout)
	assert.Nil(t, c.endpoints)
}

func TestConfigHttpDropwizard(t *testing.T) {
	service := map[string]string{}
	service["service_name"] = "test_name"
	service["port"] = "3400"
	service["path"] = "path0/path1"
	endpoints := make([]interface{}, 1)
	endpoints[0] = service
	cfg := map[string]interface{}{
		"interval":     5,
		"http_timeout": 10,
		"endpoints":    endpoints,
	}

	inst := getTestHTTPDropwizard()
	inst.Configure(cfg)

	assert.Equal(t, 5, inst.Interval())
	assert.Equal(t, 10, inst.timeout)
	assert.Equal(t, "test_name", inst.endpoints[0].Name)
	assert.Equal(t, "3400", inst.endpoints[0].Port)
	assert.Equal(t, "path0/path1", inst.endpoints[0].Path)
}
