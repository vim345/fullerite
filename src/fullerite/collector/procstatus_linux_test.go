// +build linux

package collector

import (
	"fullerite/metric"

	"testing"
	"time"
)

func TestProcStatusCollect(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999

	channel := make(chan metric.Metric)

	ps := NewProcStatus(channel, 12, nil)
	ps.Configure(config)

	go ps.Collect()

	select {
	case <-ps.Channel():
		return
	case <-time.After(2 * time.Second):
		t.Fail()
	}
}
