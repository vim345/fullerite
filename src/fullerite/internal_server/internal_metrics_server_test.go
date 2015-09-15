package internalserver

import (
	"fullerite/handler"

	"encoding/json"
	"testing"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type testHandler struct {
	handler.BaseHandler
	metrics handler.InternalMetrics
}

func (h testHandler) Run()                             {} // noop
func (h testHandler) Configure(map[string]interface{}) {} // noop
func (h testHandler) InternalMetrics() handler.InternalMetrics {
	return h.metrics
}
func (h testHandler) Name() string {
	return "somehandler"
}

func TestBuildResponse(t *testing.T) {
	testLog := l.WithField("testing", "internal_server")

	testMetrics := handler.NewInternalMetrics()
	testMetrics.Counters["somecounter"] = 12.3
	testMetrics.Gauges["somegauge"] = 432.3

	// have to declare this as a pointer b/c some of the base uses pointers
	h := new(testHandler)
	h.metrics = *testMetrics

	testHandlers := []handler.Handler{h}

	srv := internalServer{testLog, &testHandlers}

	rsp := srv.buildResponse()
	assert.NotNil(t, rsp)

	rspFormat := ResponseFormat{}
	err := json.Unmarshal(*rsp, &rspFormat)
	assert.Nil(t, err)

	// in this test ignore the memory stats
	assert.NotNil(t, rspFormat.Memory)
	assert.Equal(t, 1, len(rspFormat.Handlers))

	realHandlerRsp := rspFormat.Handlers["somehandler"]
	assert.Equal(t, 1, len(realHandlerRsp.Counters))
	assert.Equal(t, 12.3, realHandlerRsp.Counters["somecounter"])
	assert.Equal(t, 1, len(realHandlerRsp.Gauges))
	assert.Equal(t, 432.3, realHandlerRsp.Gauges["somegauge"])
}

func TestBuildResponseMemory(t *testing.T) {
	testLog := l.WithField("testing", "internal_server")
	emptyHandlers := []handler.Handler{}

	srv := internalServer{
		log:      testLog,
		handlers: &emptyHandlers,
	}

	rspFormat := new(ResponseFormat)
	rsp := srv.buildResponse()
	err := json.Unmarshal(*rsp, rspFormat)
	assert.Nil(t, err)

	// only care about the memory part
	assert.NotNil(t, rspFormat.Memory)
	assert.NotNil(t, rspFormat.Handlers)

	// only check that there are enough items in the list
	assert.Equal(t, 7, len(rspFormat.Memory.Counters))
	assert.Equal(t, 19, len(rspFormat.Memory.Gauges))
}

func TestRespondToHttp(t *testing.T) {
	// TODO
}
