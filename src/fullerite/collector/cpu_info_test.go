package collector

import (
	"fullerite/metric"
	"test_utils"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCpuInfoCollect(t *testing.T) {
	config := make(map[string]interface{})
	config["procPath"] = ""
	testChannel := make(chan metric.Metric)
	testLogger := test_utils.BuildLogger()

	cpuInfo := NewCpuInfo(testChannel, 100, testLogger)
	cpuInfo.configure(config)

	go cpuInfo.Collect()

	select {
	case m := <-cpuInfo.Channel():
		assert.Equal(t, 2.0, m.Value)
		assert(t, "", m.Dimensions["model"])
		return
	case <-time.After(2 * time.Second):
		t.Fail()
	}
}
