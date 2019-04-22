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

	actual := newProcNetUDPStats(c, i, l).(*procNetUDPStats)

	assert.Equal(t, "ProcNetUDPStats", actual.Name())
	assert.Equal(t, c, actual.Channel())
	assert.Equal(t, i, actual.Interval())
	assert.Equal(t, l, actual.log)
}

func TestProcNetUDPStatsConfigureMissingRemote(t *testing.T) {
	l := defaultLog.WithFields(l.Fields{"collector": "ProcNetUDPStats"})

	fakeCollector := newProcNetUDPStats(nil, 0, l).(*procNetUDPStats)
	fakeCollector.Configure(map[string]interface{}{
		"localAddressWhitelist": "7F000001:613",
	})

	assert.NotNil(t, fakeCollector.localAddressWhitelist)
	assert.Nil(t, fakeCollector.remoteAddressWhitelist)
}

func TestProcNetUDPStatsConfigureMissingLocal(t *testing.T) {
	l := defaultLog.WithFields(l.Fields{"collector": "ProcNetUDPStats"})

	fakeCollector := newProcNetUDPStats(nil, 0, l).(*procNetUDPStats)
	fakeCollector.Configure(map[string]interface{}{
		"remoteAddressWhitelist": "7F000001:613",
	})

	assert.Nil(t, fakeCollector.localAddressWhitelist)
	assert.NotNil(t, fakeCollector.remoteAddressWhitelist)
}

func TestProcNetUDPStatsConfigure(t *testing.T) {
	l := defaultLog.WithFields(l.Fields{"collector": "ProcNetUDPStats"})

	fakeCollector := newProcNetUDPStats(nil, 0, l).(*procNetUDPStats)
	fakeCollector.Configure(map[string]interface{}{
		"localAddressWhitelist":  "7F000001:613",
		"remoteAddressWhitelist": "7F000001:613",
	})

	assert.NotNil(t, fakeCollector.localAddressWhitelist)
	assert.NotNil(t, fakeCollector.remoteAddressWhitelist)
}

func TestParse(t *testing.T) {
	out := `sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
 3152: FEFFFEA9:4ED6 00000000:0000 07 00000000:00000000 00:00000000 00000000 65534        0 3841266873 2 ffff88021734b480 0
15747: FEFFFEA9:8009 FEFFFEA9:1FBD 01 00000000:00000000 00:00000000 00000000  4404        0 1989081677 2 ffff8806859712c0 100`
	fakeCollector := &procNetUDPStats{}

	lines := fakeCollector.parseProcNetUDPLines(out)
	assert.Equal(t, 2, len(lines))

	assert.Equal(t, "FEFFFEA9:4ED6", lines[0].localAddress)
	assert.Equal(t, "00000000:0000", lines[0].remoteAddress)
	assert.Equal(t, "0", lines[0].drops)
	assert.Equal(t, "FEFFFEA9:8009", lines[1].localAddress)
	assert.Equal(t, "FEFFFEA9:1FBD", lines[1].remoteAddress)
	assert.Equal(t, "100", lines[1].drops)
}
