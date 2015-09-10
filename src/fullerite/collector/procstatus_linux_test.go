// +build linux

package collector

import (
	"testing"
	"time"
)

func TestProcStatusCollect(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999

	ps := NewProcStatus()
	ps.Configure(config)

	go ps.Collect()

	select {
	case <-ps.Channel():
		return
	case <-time.After(2 * time.Second):
		t.Fail()
	}
}
