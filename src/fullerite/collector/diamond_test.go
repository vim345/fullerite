package collector_test

import (
	"fullerite/collector"
	"fullerite/metric"

	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiamondConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	d := collector.NewDiamond()
	d.Configure(config)

	assert.Equal(t,
		d.Interval(),
		collector.DefaultCollectionInterval,
		"should be the default collection interval",
	)
}

func TestDiamondConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999
	config["port"] = "0"
	d := collector.NewDiamond()
	d.Configure(config)

	assert := assert.New(t)
	assert.Equal(d.Interval(), 9999, "should be the defined interval")
	assert.Equal(d.Port(), "0", "should be the defined port")
}

func TestDiamondCollect(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	config := make(map[string]interface{})
	config["port"] = "0"

	d := collector.NewDiamond()
	d.Configure(config)

	// start collecting Diamond metrics
	go d.Collect()

	conn, err := connectToDiamondCollector(d)
	require.Nil(t, err, "should connect")
	require.NotNil(t, conn, "should connect")

	emitTestMetric(conn)

	select {
	case m := <-d.Channel():
		assert.Equal(t, m.Name, "test")
	case <-time.After(1 * time.Second):
		t.Fail()
	}
}



func connectToDiamondCollector(d *collector.Diamond) (net.Conn, error) {
	// emit a Diamond metric
	var (
		conn net.Conn
		err  error
	)
	for retry := 0; retry < 3; retry++ {
		if conn, err = net.DialTimeout("tcp", "localhost:"+d.Port(), 2*time.Second); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return conn, err
}

func emitTestMetric(conn net.Conn) {
	m := metric.New("test")
	b, _ := json.Marshal(m)
	fmt.Fprintf(conn, string(b)+"\n")
	fmt.Fprintf(conn, string(b)+"\n")
}
