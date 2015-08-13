package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testConfiguration = `{
    "prefix": "test.",
    "interval": 10,
    "defaultDimensions": {
        "application": "fullerite",
        "host": "dev33-devc"
    },

    "diamond_collectors_path": "src/diamond/collectors",
    "diamond_collectors": {
        "CPUCollector": {"enabled": true, "interval": 10},
        "PingCollector": {"enabled": true, "target_google": "google.com", "interval": 10, "bin": "/bin/ping"}
    },

    "collectors": {
        "Test": {
            "metricName": "TestMetric",
            "interval": 10
        },
        "Diamond":{
            "port": "19191",
            "interval": 10
        }
    },

    "handlers": {
        "Graphite": {
            "server": "10.40.11.51",
            "port": "2003",
            "timeout": 2
        },
        "SignalFx": {
            "authToken": "secret_token",
            "endpoint": "https://ingest.signalfx.com/v2/datapoint",
            "interval": 10,
            "timeout": 2
        }
    }
}
`
var tmpTestFile string

func TestMain(m *testing.M) {
	if f, err := ioutil.TempFile("/tmp", "fullerite"); err == nil {
		f.WriteString(testConfiguration)
		tmpTestFile = f.Name()
		f.Close()
		defer os.Remove(tmpTestFile)
	}
	m.Run()
}

func TestParseExampleConfig(t *testing.T) {
	_, err := ReadConfig(tmpTestFile)
	if err != nil {
		t.Fail()
	}
}

func TestGetInt(t *testing.T) {
	val, err := GetAsInt("10")
	assert.Equal(t, val, 10)
	assert.Nil(t, err)

	val, err = GetAsInt("notanint")
	assert.Nil(t, val)
	assert.NotNil(t, err)

	val, err = GetAsInt(12)
	assert.Equal(t, val, 12)
	assert.Nil(t, err)

	val, err = GetAsInt(12.123)
	assert.Equal(t, val, 12)
	assert.Nil(t, err)

	var asint int
	asint, err = GetAsInt(12)
	assert.Equal(t, asint, 12)
	assert.Nil(t, err)
}

func TestGetFloat(t *testing.T) {
	val, err := GetAsFloat("10")
	assert.Equal(t, val, 10)
	assert.Nil(t, err)

	val, err = GetAsFloat("10.21")
	assert.Equal(t, val, 10.21)
	assert.Nil(t, err)

	val, err = GetAsFloat("notanint")
	assert.Nil(t, val)
	assert.NotNil(t, err)

	val, err = GetAsFloat(12.123)
	assert.Equal(t, val, 12)
	assert.Nil(t, err)
}
