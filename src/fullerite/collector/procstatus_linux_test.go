// +build linux

package collector

import (
	"fullerite/metric"

	"testing"
	"time"

	l "github.com/Sirupsen/logrus"
)

func TestProcStatusCollect(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999

	dims := make(map[string][2]string)
	dims["module"] = [2]string{"cmdline", ".*"}

	config["generatedDimensions"] = dims

	channel := make(chan metric.Metric)

	testLog = l.WithFields(l.Fields{"testing": "procstatus_linux"})
	ps := NewProcStatus(channel, 12, testLog)
	ps.Configure(config)

	go ps.Collect()

	select {
	case <-ps.Channel():
		return
	case <-time.After(2 * time.Second):
		t.Fail()
	}
}
