package collector

import (
	"fullerite/metric"
	"testing"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewProcNetUDPStats(t *testing.T) {
	c := make(chan metric.Metric)
	i := 10
	l := defaultLog.WithFields(l.Fields{"collector": "Mesos"})

	actual := newProcNetUDPStats(c, i, l).(*ProcNetUDPStats)

	assert.Equal(t, "ProcNetUDPStats", actual.Name())
	assert.Equal(t, c, actual.Channel())
	assert.Equal(t, i, actual.Interval())
	assert.Equal(t, l, actual.log)
}

func TestProcNetUDPStatsConfigureMissingRemote(t *testing.T) {
	l := defaultLog.WithFields(l.Fields{"collector": "ProcNetUDPStats"})

	fake_collector := newProcNetUDPStats(nil, 0, l).(*ProcNetUDPStats)
	fake_collector.Configure(map[string]interface{}{
		"localAddressWhitelist": "7F000001:613",
	})

	assert.NotNil(t, fake_collector.localAddressWhitelist)
	assert.Nil(t, fake_collector.remoteAddressWhitelist)
}

func TestProcNetUDPStatsConfigureMissingLocal(t *testing.T) {
	l := defaultLog.WithFields(l.Fields{"collector": "ProcNetUDPStats"})

	fake_collector := newProcNetUDPStats(nil, 0, l).(*ProcNetUDPStats)
	fake_collector.Configure(map[string]interface{}{
		"remoteAddressWhitelist": "7F000001:613",
	})

	assert.Nil(t, fake_collector.localAddressWhitelist)
	assert.NotNil(t, fake_collector.remoteAddressWhitelist)
}

func TestProcNetUDPStatsConfigure(t *testing.T) {
	l := defaultLog.WithFields(l.Fields{"collector": "ProcNetUDPStats"})

	fake_collector := newProcNetUDPStats(nil, 0, l).(*ProcNetUDPStats)
	fake_collector.Configure(map[string]interface{}{
		"localAddressWhitelist":  "7F000001:613",
		"remoteAddressWhitelist": "7F000001:613",
	})

	assert.NotNil(t, fake_collector.localAddressWhitelist)
	assert.NotNil(t, fake_collector.remoteAddressWhitelist)
}

func TestParse(t *testing.T) {
	out := `sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
 3152: FEFFFEA9:4ED6 00000000:0000 07 00000000:00000000 00:00000000 00000000 65534        0 3841266873 2 ffff88021734b480 0
15747: FEFFFEA9:8009 FEFFFEA9:1FBD 01 00000000:00000000 00:00000000 00000000  4404        0 1989081677 2 ffff8806859712c0 0`
	fake_collector := &ProcNetUDPStats{}

	lines := fake_collector.parseProcNetUDPLines(out)
	assert.Equal(t, 2, len(lines))
}
