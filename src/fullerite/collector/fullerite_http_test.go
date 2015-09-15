package collector

import (
	"fullerite/metric"

	"bytes"
	"io"
	"net/http"
	"testing"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func getTestInstance() *fulleriteHttp {
	testChannel := make(chan metric.Metric)
	testLog = l.WithFields(l.Fields{"testing": "fullerite_http"})

	inst := NewFulleriteHttpCollector(testChannel, 12, testLog)
	return inst
}

type noopCloser struct {
	io.Reader
}

func (n noopCloser) Close() error { return nil }

func buildTestHttpResponse(body string) *http.Response {
	rsp := new(http.Response)

	rsp.Body = noopCloser{bytes.NewBufferString(body)}

	return rsp
}

func TestMakeNewFulleriteHttp(t *testing.T) {
	inst := getTestInstance()
	assert.NotNil(t, inst)

	assert.Equal(t, "http://localhost:9090/metrics", inst.endpoint)
	assert.NotNil(t, inst.errHandler)
	assert.NotNil(t, inst.rspHandler)
}

func TestConfigureFulleriteHttp(t *testing.T) {
	config := make(map[string]interface{})
	config["endpoint"] = "http://somewhere:234/marp"
	config["interval"] = 123

	inst := getTestInstance()
	inst.Configure(config)

	assert.Equal(t, "http://somewhere:234/marp", inst.endpoint)
	assert.Equal(t, 123, inst.Interval())
}

func TestHandleBadInputFulleriteHttp(t *testing.T) {
	rsp := buildTestHttpResponse("teststring")

	inst := getTestInstance()
	results := inst.rspHandler(rsp)

	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
}

func TestHandleNotJson(t *testing.T) {
	txt := []byte("not json")
	inst := getTestInstance()
	metrics, err := inst.parseResponseText(&txt)

	assert.NotNil(t, metrics)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(metrics))
}

func TestHandlePopulatedResponseFulleriteHttp(t *testing.T) {
	testString := `
	{
    	"somemetric": 42,
    	"anothermetric": 56.4
	}
	`
	asBytes := []byte(testString)

	inst := getTestInstance()
	metrics, err := inst.parseResponseText(&asBytes)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(metrics))

	for _, m := range metrics {
		assert.Equal(t, 1, len(m.Dimensions))
		dimVal, exists := m.GetDimensionValue("collector", map[string]string{})
		assert.True(t, exists)
		assert.Equal(t, "fullerite_http", dimVal)
		if m.Name == "somemetric" {
			assert.Equal(t, 42.0, m.Value)
		} else if m.Name == "anothermetric" {
			assert.Equal(t, 56.4, m.Value)
		} else {
			t.FailNow()
		}
	}
}
