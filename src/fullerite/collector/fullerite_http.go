package collector

import (
	"fullerite/metric"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	l "github.com/Sirupsen/logrus"
)

const (
	defaultFulleriteHost     = "localhost"
	defaultFulleritePort     = 9090
	defaultFulleritePath     = "metrics"
	defaultFulleriteProtocol = "http"
)

// collects stats from fullerite's http endpoint
type fulleriteHttp struct {
	baseHttpCollector
}

func NewFulleriteHttpCollector(channel chan metric.Metric, initialInterval int, log *l.Entry) *fulleriteHttp {
	inst := new(fulleriteHttp)

	inst.log = log
	inst.channel = channel
	inst.interval = initialInterval

	inst.name = "FulleriteHttp"

	inst.endpoint = fmt.Sprintf("%s://%s:%d/%s",
		defaultFulleriteProtocol,
		defaultFulleriteHost,
		defaultFulleritePort,
		defaultFulleritePath)

	inst.rspHandler = inst.handleResponse
	inst.errHandler = inst.handleError

	return inst
}

func (inst *fulleriteHttp) Configure(configMap map[string]interface{}) {
	if endpoint, exists := configMap["endpoint"]; exists == true {
		inst.endpoint = endpoint.(string)
	}

	inst.configureCommonParams(configMap)
}

func (inst fulleriteHttp) handleError(err error) {
	inst.log.Error("Failed to make GET to ", inst.endpoint, " error is: ", err)
}

// handleResponse assumes the format of the response is a JSON dictionary. It then converts
// them to individual metrics.
func (inst fulleriteHttp) handleResponse(rsp *http.Response) []metric.Metric {
	results := []metric.Metric{}

	txt, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err == nil {
		parsedRsp, parseError := inst.parseResponseText(&txt)
		if parseError != nil {
			inst.log.Error("Failed to parse the request body '", string(txt), "' because of error: ", err)
		} else {
			results = parsedRsp
		}
	} else {
		inst.log.Error("Failed to get the body of the response because of error: ", err)
	}

	return results
}

// parseResponseText takes the raw JSON string and parses that into metrics. The
// format of the JSON string is assumed to be a dictionary and then each key
// creates a metric.
func (inst fulleriteHttp) parseResponseText(raw *[]byte) ([]metric.Metric, error) {
	var parsedMap map[string]float64

	err := json.Unmarshal(*raw, &parsedMap)
	if err != nil {
		return []metric.Metric{}, err
	}

	// now we should parse each of the key/value pairs
	inst.log.Debug("Starting to process the ", len(parsedMap), " keys in the parsed response")
	results := []metric.Metric{}
	for key, value := range parsedMap {
		m := metric.New(key)
		m.Value = value
		m.AddDimension("collector", "fullerite_http")
		results = append(results, m)
	}

	return results, nil
}
