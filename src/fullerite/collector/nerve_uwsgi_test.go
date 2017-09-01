package collector

import (
	"fullerite/dropwizard"
	"fullerite/metric"
	"fullerite/util"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func getTestNerveConfig() []byte {
	raw := `
	{
	    "heartbeat_path": "/var/run/nerve/heartbeat",
	    "instance_id": "srv1-devc",
	    "services": {
	        "example_service.main.norcal-devc.superregion:norcal-devc.13752.new": {
	            "check_interval": 7,
	            "checks": [
	                {
	                    "fall": 2,
	                    "headers": {},
	                    "host": "127.0.0.1",
	                    "open_timeout": 6,
	                    "port": 6666,
	                    "rise": 1,
	                    "timeout": 6,
	                    "type": "http",
	                    "uri": "/http/example_service.main/13752/status"
	                }
	            ],
	            "host": "10.56.5.21",
	            "port": 13752,
	            "weight": 24,
	            "zk_hosts": [
	                "10.40.5.5:22181",
	                "10.40.5.6:22181",
	                "10.40.1.17:22181"
	            ],
	            "zk_path": "/nerve/superregion:norcal-devc/example_service.main"
	        },
	        "example_service.mesosstage_main.norcal-devc.superregion:norcal-devc.13752.new": {
	            "check_interval": 7,
	            "checks": [
	                {
	                    "fall": 2,
	                    "headers": {},
	                    "host": "127.0.0.1",
	                    "open_timeout": 6,
	                    "port": 6666,
	                    "rise": 1,
	                    "timeout": 6,
	                    "type": "http",
	                    "uri": "/http/example_service.mesosstage_main/13752/status"
	                }
	            ],
	            "host": "10.56.5.21",
	            "port": 22222,
	            "weight": 24,
	            "zk_hosts": [
	                "10.40.5.5:22181",
	                "10.40.5.6:22181",
	                "10.40.1.17:22181"
	            ],
	            "zk_path": "/nerve/superregion:norcal-devc/example_service.mesosstage_main"
	        }
	    }
	}
	`
	return []byte(raw)
}

func getTestUWSGIResponse() string {
	return `{
	"counters": {
		"Acounter":{
			"firstrollup": 134,
			"secondrollup": 89
		}
	},
	"meters": {},
	"timers": {
		"some_timer": {
			"average": 123
		},
		"othertimer": {
			"mean": 345
		}
	},
	"gauges": {
		"some_random_metric": {
			"rollup1": 12
		}
	},
	"histograms": {}
	}
	`
}

func getTestJavaResponse() string {
	return `{
	"counters": {
		"Acounter,dim1=val1,dim2=val2,dim3=val3":{
			"count": 100,
			"firstrollup": 134,
			"secondrollup": 89
		}
	},
	"meters": {},
	"timers": {
		"some_timer": {
			"count": 200,
			"value": 123
		},
		"othertimer,dimX=valX": {
			"mean": 345,
			"m1_rate": 3e-323
		}
	},
	"gauges": {
		"some_random_metric": {
			"rollup1": 12
		}
	},
	"histograms": {}
	}
	`
}

func getTestSchemaUWSGIResponse() string {
	return `{
    "service_dims": {
        "firstdim": "first",
        "seconddim": "second"
    },
	"counters": {
		"Acounter":{
			"firstrollup": 134,
			"secondrollup": 89
		}
	},
	"meters": {},
	"timers": {
		"some_timer": {
			"average": 123
		},
		"othertimer": {
			"mean": 345
		}
	},
	"gauges": {
		"some_random_metric": {
			"rollup1": 12
		}
	},
	"histograms": {}
	}
	`
}

func validateUWSGIResults(t *testing.T, actual []metric.Metric) {
	assert.Equal(t, 5, len(actual))

	for _, m := range actual {
		metricTypeDim, exists := m.GetDimensionValue("type")
		assert.True(t, exists)
		rollup, exists := m.GetDimensionValue("rollup")
		assert.True(t, exists)

		switch m.Name {
		case "some_random_metric":
			assert.Equal(t, "rollup1", rollup)
			assert.Equal(t, "gauge", metricTypeDim)
			assert.Equal(t, 12.0, m.Value)
			assert.Equal(t, metric.Gauge, m.MetricType)
		case "othertimer":
			assert.Equal(t, "mean", rollup)
			assert.Equal(t, "timer", metricTypeDim)
			assert.Equal(t, 345.0, m.Value)
			assert.Equal(t, metric.Gauge, m.MetricType)
		case "some_timer":
			assert.Equal(t, "average", rollup)
			assert.Equal(t, "timer", metricTypeDim)
			assert.Equal(t, 123.0, m.Value)
			assert.Equal(t, metric.Gauge, m.MetricType)
		case "ACounter":
			assert.Equal(t, metric.Counter, m.MetricType)
			val, exists := map[string]float64{
				"firstrollup":  134,
				"secondrollup": 89,
			}[rollup]
			assert.Equal(t, "counter", metricTypeDim)
			assert.True(t, exists)
			assert.Equal(t, val, m.Value)
		}
	}
}

func validateJavaResults(t *testing.T, actual []metric.Metric, serviceName string, servicePort string) {
	assert.Equal(t, 8, len(actual))

	for _, m := range actual {
		metricTypeDim, exists := m.GetDimensionValue("type")
		assert.True(t, exists)
		rollup, exists := m.GetDimensionValue("rollup")
		assert.True(t, exists)

		switch m.Name {
		case "some_random_metric":
			assert.Equal(t, "rollup1", rollup)
			assert.Equal(t, "gauge", metricTypeDim)
			assert.Equal(t, 12.0, m.Value)
			assert.Equal(t, metric.Gauge, m.MetricType)
		case "othertimer":
			val, exists := map[string]float64{
				"mean":    345.0,
				"m1_rate": 3e-323,
			}[rollup]
			assert.Equal(t, "timer", metricTypeDim)
			assert.True(t, exists)
			assert.Equal(t, metric.Gauge, m.MetricType)
			assert.Equal(t, val, m.Value)
			dim, exists := m.GetDimensionValue("dimX")
			assert.True(t, exists)
			assert.Equal(t, "valX", dim)
		case "some_timer":
			val, exists := map[string]float64{
				"value": 123,
				"count": 200,
			}[rollup]
			assert.Equal(t, "timer", metricTypeDim)
			assert.True(t, exists)
			assert.Equal(t, metric.Gauge, m.MetricType)
			assert.Equal(t, val, m.Value)
		case "ACounter":
			assert.Equal(t, metric.Counter, m.MetricType)
			val, exists := map[string]float64{
				"firstrollup":  134,
				"secondrollup": 89,
				"counter":      100,
			}[rollup]
			assert.Equal(t, "counter", metricTypeDim)
			assert.True(t, exists)
			assert.Equal(t, val, m.Value)
			dim, exists := m.GetDimensionValue("dim1")
			assert.True(t, exists)
			assert.Equal(t, "val1", dim)
			dim, exists = m.GetDimensionValue("dim2")
			assert.True(t, exists)
			assert.Equal(t, "val2", dim)
			dim, exists = m.GetDimensionValue("dim3")
			assert.True(t, exists)
			assert.Equal(t, "val3", dim)
		}

		dim, exists := m.GetDimensionValue("service")
		assert.True(t, exists)
		assert.Equal(t, serviceName, dim)

		val, exists := m.GetDimensionValue("port")
		assert.True(t, exists)
		assert.Equal(t, servicePort, val)
	}
}

func validateFullDimensions(t *testing.T, actual []metric.Metric, serviceName, port string) {
	for _, m := range actual {
		assert.Equal(t, 4, len(m.Dimensions))

		val, exists := m.GetDimensionValue("service")
		assert.True(t, exists)
		assert.Equal(t, serviceName, val)

		val, exists = m.GetDimensionValue("port")
		assert.True(t, exists)
		assert.Equal(t, port, val)
	}

}

func validateFullSchemaDimensions(t *testing.T, actual []metric.Metric, serviceName, port string) {
	for _, m := range actual {
		assert.Equal(t, 6, len(m.Dimensions))

		val, exists := m.GetDimensionValue("service")
		assert.True(t, exists)
		assert.Equal(t, serviceName, val)

		val, exists = m.GetDimensionValue("port")
		assert.True(t, exists)
		assert.Equal(t, port, val)

		val, exists = m.GetDimensionValue("firstdim")
		assert.True(t, exists)
		assert.Equal(t, "first", val)

		val, exists = m.GetDimensionValue("seconddim")
		assert.True(t, exists)
		assert.Equal(t, "second", val)
	}

}

func validateEmptyChannel(t *testing.T, c chan metric.Metric) {
	close(c)
	for m := range c {
		t.Fatal("The channel was not empty! got value ", m)
	}
}

func parseURL(url string) (string, string) {
	parts := strings.Split(url, ":")
	ip := strings.Replace(parts[1], "/", "", -1)
	port := parts[2]
	return ip, port
}

func getTestNerveUWSGI() *nerveUWSGICollector {
	return newNerveUWSGI(make(chan metric.Metric), 12, l.WithField("testing", "nerveuwsgi")).(*nerveUWSGICollector)
}

func TestDefaultConfigNerveUWSGI(t *testing.T) {
	inst := getTestNerveUWSGI()
	inst.Configure(make(map[string]interface{}))

	assert.Equal(t, 12, inst.Interval())
	assert.Equal(t, 2, inst.timeout)
	assert.Equal(t, "/etc/nerve/nerve.conf.json", inst.configFilePath)
	assert.Equal(t, "status/metrics", inst.queryPath)
}

func TestConfigNerveUWSGI(t *testing.T) {
	cfg := map[string]interface{}{
		"interval":       345,
		"configFilePath": "/etc/your/moms/house",
		"queryPath":      "littlepiggies",
		"http_timeout":   12,
	}

	inst := getTestNerveUWSGI()
	inst.Configure(cfg)

	assert.Equal(t, 345, inst.Interval())
	assert.Equal(t, "/etc/your/moms/house", inst.configFilePath)
	assert.Equal(t, "littlepiggies", inst.queryPath)
	assert.Equal(t, 12, inst.timeout)
}

func TestErrorQueryEndpointResponse(t *testing.T) {
	//4xx HTTP status code test
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	endpoint := ts.URL + "/status/metrics"
	ts.Close()

	_, _, queryEndpointError := queryEndpoint(endpoint, 10)
	assert.NotNil(t, queryEndpointError)

	//Socket closed test
	tsClosed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	tsClosed.Close()
	closedEndpoint := tsClosed.URL + "/status/metrics"
	_, queryClosedEndpointResponse, queryClosedEndpointError := queryEndpoint(closedEndpoint, 10)
	assert.NotNil(t, queryClosedEndpointError)
	assert.Equal(t, "", queryClosedEndpointResponse)

}

func TestNerveUWSGICollect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, rsp *http.Request) {
		fmt.Fprint(w, getTestUWSGIResponse())
	}))
	defer server.Close()

	// assume format is http://ipaddr:port
	ip, port := parseURL(server.URL)
	minimalNerveConfig := util.CreateMinimalNerveConfig(map[string]util.EndPoint{
		"test_service.things.and.stuff": util.EndPoint{ip, port},
	})

	tmpFile, err := ioutil.TempFile("", "fullerite_testing")
	defer os.Remove(tmpFile.Name())
	assert.Nil(t, err)

	marshalled, err := json.Marshal(minimalNerveConfig)
	assert.Nil(t, err)

	_, err = tmpFile.Write(marshalled)
	assert.Nil(t, err)

	cfg := map[string]interface{}{
		"configFilePath": tmpFile.Name(),
		"queryPath":      "",
	}

	inst := getTestNerveUWSGI()
	inst.Configure(cfg)

	go inst.Collect()

	actual := []metric.Metric{}
	for i := 0; i < 5; i++ {
		actual = append(actual, <-inst.Channel())
	}

	validateUWSGIResults(t, actual)
	validateFullDimensions(t, actual, "test_service", port)
	validateEmptyChannel(t, inst.Channel())
}

func TestNerveUWSGICollectWithSchema(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, rsp *http.Request) {
		w.Header().Set("Metrics-Schema", "uwsgi.1.1")
		fmt.Fprint(w, getTestSchemaUWSGIResponse())
	}))
	defer server.Close()

	// assume format is http://ipaddr:port
	ip, port := parseURL(server.URL)

	minimalNerveConfig := util.CreateMinimalNerveConfig(map[string]util.EndPoint{
		"test_service.things.and.stuff": util.EndPoint{ip, port},
	})

	tmpFile, err := ioutil.TempFile("", "fullerite_testing")
	defer os.Remove(tmpFile.Name())
	assert.Nil(t, err)

	marshalled, err := json.Marshal(minimalNerveConfig)
	assert.Nil(t, err)

	_, err = tmpFile.Write(marshalled)
	assert.Nil(t, err)

	cfg := map[string]interface{}{
		"configFilePath": tmpFile.Name(),
		"queryPath":      "",
	}

	inst := getTestNerveUWSGI()
	inst.Configure(cfg)

	go inst.Collect()

	actual := []metric.Metric{}
	for i := 0; i < 5; i++ {
		actual = append(actual, <-inst.Channel())
	}

	validateUWSGIResults(t, actual)
	validateFullSchemaDimensions(t, actual, "test_service", port)
	validateEmptyChannel(t, inst.Channel())
}

func TestNerveJavaCollectWithSchema(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, rsp *http.Request) {
		w.Header().Set("Metrics-Schema", "java-1.1")
		fmt.Fprint(w, getTestJavaResponse())
	}))
	defer server.Close()

	// assume format is http://ipaddr:port
	ip, port := parseURL(server.URL)

	minimalNerveConfig := util.CreateMinimalNerveConfig(map[string]util.EndPoint{
		"test_service.things.and.stuff": util.EndPoint{ip, port},
	})

	tmpFile, err := ioutil.TempFile("", "fullerite_testing")
	defer os.Remove(tmpFile.Name())
	assert.Nil(t, err)

	marshalled, err := json.Marshal(minimalNerveConfig)
	assert.Nil(t, err)

	_, err = tmpFile.Write(marshalled)
	assert.Nil(t, err)

	cfg := map[string]interface{}{
		"configFilePath": tmpFile.Name(),
		"queryPath":      "",
	}

	inst := getTestNerveUWSGI()
	inst.Configure(cfg)

	go inst.Collect()

	actual := []metric.Metric{}
	for i := 0; i < 8; i++ {
		actual = append(actual, <-inst.Channel())
	}

	validateJavaResults(t, actual, "test_service", port)
	validateEmptyChannel(t, inst.Channel())
}

func TestNerveJavaCollectWithSchemaCumulativeCountersEnabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, rsp *http.Request) {
		w.Header().Set("Metrics-Schema", "java-1.1")
		fmt.Fprint(w, getTestJavaResponse())
	}))
	defer server.Close()

	// assume format is http://ipaddr:port
	ip, port := parseURL(server.URL)

	minimalNerveConfig := util.CreateMinimalNerveConfig(map[string]util.EndPoint{
		"test_service.namespace": util.EndPoint{ip, port},
	})

	tmpFile, err := ioutil.TempFile("", "fullerite_testing")
	defer os.Remove(tmpFile.Name())
	assert.Nil(t, err)

	marshalled, err := json.Marshal(minimalNerveConfig)
	assert.Nil(t, err)

	_, err = tmpFile.Write(marshalled)
	assert.Nil(t, err)

	cfg := map[string]interface{}{
		"configFilePath":    tmpFile.Name(),
		"queryPath":         "",
		"servicesWhitelist": []string{"test_service"},
	}

	inst := getTestNerveUWSGI()
	inst.Configure(cfg)

	go inst.Collect()

	actual := []metric.Metric{}
	flag := true
	for flag == true {
		select {
		case metric := <-inst.Channel():
			actual = append(actual, metric)
		case <-time.After(2 * time.Second):
			flag = false
			break
		}
	}
	assert.Equal(t, 7, len(actual))

	for _, m := range actual {
		switch m.Name {
		case "Acounter.firstrollup":
			assert.Equal(t, 134.0, m.Value)
		case "Acounter.count":
			assert.Equal(t, 100.0, m.Value)
		case "Acounter.secondrollup":
			assert.Equal(t, 89.0, m.Value)
		case "some_timer":
			assert.Equal(t, 123.0, m.Value)
		case "some_timer.count":
			assert.Equal(t, 200.0, m.Value)
			assert.Equal(t, metric.CumulativeCounter, m.MetricType)
		case "othertimer.mean":
			assert.Equal(t, 345.0, m.Value)
		case "some_random_metric.rollup1":
			assert.Equal(t, 12.0, m.Value)
		default:
			t.Fatalf("unknown metric name %s", m.Name)
		}
	}
}

func TestNonConflictingServiceQueries(t *testing.T) {
	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, rsp *http.Request) {
		fmt.Fprint(w, getTestUWSGIResponse())
	}))
	defer goodServer.Close()

	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, rsp *http.Request) {
		return // no response
	}))
	defer badServer.Close()

	goodIP, goodPort := parseURL(goodServer.URL)
	badIP, badPort := parseURL(badServer.URL)
	minimalNerveConfig := util.CreateMinimalNerveConfig(map[string]util.EndPoint{
		"test_service.things.and.stuff":    util.EndPoint{goodIP, goodPort},
		"other_service.does.lots.of.stuff": util.EndPoint{badIP, badPort},
	})

	tmpFile, err := ioutil.TempFile("", "fullerite_testing")
	defer os.Remove(tmpFile.Name())
	assert.Nil(t, err)

	marshalled, err := json.Marshal(minimalNerveConfig)
	assert.Nil(t, err)

	_, err = tmpFile.Write(marshalled)
	assert.Nil(t, err)

	cfg := map[string]interface{}{
		"configFilePath": tmpFile.Name(),
		"queryPath":      "",
	}

	inst := getTestNerveUWSGI()
	inst.Configure(cfg)

	go inst.Collect()

	actual := []metric.Metric{}
	for i := 0; i < 5; i++ {
		actual = append(actual, <-inst.Channel())
	}

	validateUWSGIResults(t, actual)
	validateFullDimensions(t, actual, "test_service", goodPort)
	validateEmptyChannel(t, inst.Channel())
}

func TestUWSGIResponseConversion(t *testing.T) {
	uwsgiRsp := []byte(getTestUWSGIResponse())

	actual, err := dropwizard.Parse(uwsgiRsp, "", false)

	assert.Nil(t, err)
	validateUWSGIResults(t, actual)
	for _, m := range actual {
		assert.Equal(t, 2, len(m.Dimensions))
	}
}

func TestNerveUWSGICollectWithBlacklist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, rsp *http.Request) {
		fmt.Fprint(w, getTestUWSGIResponse())
	}))
	defer server.Close()

	// assume format is http://ipaddr:port
	ip, port := parseURL(server.URL)
	minimalNerveConfig := util.CreateMinimalNerveConfig(map[string]util.EndPoint{
		"test_service.blacklist": util.EndPoint{ip, port},
	})

	tmpFile, err := ioutil.TempFile("", "fullerite_testing")
	defer os.Remove(tmpFile.Name())
	assert.Nil(t, err)

	marshalled, err := json.Marshal(minimalNerveConfig)
	assert.Nil(t, err)

	_, err = tmpFile.Write(marshalled)
	assert.Nil(t, err)

	cfg := map[string]interface{}{
		"configFilePath":       tmpFile.Name(),
		"queryPath":            "",
		"dimensions_blacklist": map[string]string{"rollup": "mean"},
	}

	inst := getTestNerveUWSGI()
	inst.Configure(cfg)

	go inst.Collect()

	actual := []metric.Metric{}
	for i := 0; i < 4; i++ {
		actual = append(actual, <-inst.Channel())
	}

	dropped := metric.Metric{
		Name:       "othertimer",
		MetricType: "gauge",
		Value:      345.0,
		Dimensions: map[string]string{
			"rollup":  "mean",
			"type":    "timer",
			"service": "test_service",
			"port":    port}}
	actual = append(actual, dropped)
	validateUWSGIResults(t, actual)
	validateFullDimensions(t, actual, "test_service", port)
	validateEmptyChannel(t, inst.Channel())
}
