package config_test

import (
	"fullerite/config"

	"io/ioutil"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testBadConfiguration = `{
    "prefix": "test.",
    malformed JSON File {123!!!!
}
`

var testGoodConfiguration = `{
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
var (
	tmpTestGoodFile, tmpTestBadFile string
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.ErrorLevel)
	if f, err := ioutil.TempFile("/tmp", "fullerite"); err == nil {
		f.WriteString(testGoodConfiguration)
		tmpTestGoodFile = f.Name()
		f.Close()
		defer os.Remove(tmpTestGoodFile)
	}
	if f, err := ioutil.TempFile("/tmp", "fullerite"); err == nil {
		f.WriteString(testBadConfiguration)
		tmpTestBadFile = f.Name()
		f.Close()
		defer os.Remove(tmpTestBadFile)
	}
	os.Exit(m.Run())
}

func TestGetInt(t *testing.T) {
	assert := assert.New(t)

	val := config.GetAsInt("10", 123)
	assert.Equal(val, 10)

	val = config.GetAsInt("notanint", 123)
	assert.Equal(val, 123)

	val = config.GetAsInt(12.123, 123)
	assert.Equal(val, 12)

	val = config.GetAsInt(12, 123)
	assert.Equal(val, 12)
}

func TestGetFloat(t *testing.T) {
	assert := assert.New(t)

	val := config.GetAsFloat("10", 123)
	assert.Equal(val, 10.0)

	val = config.GetAsFloat("10.21", 123)
	assert.Equal(val, 10.21)

	val = config.GetAsFloat("notanint", 123)
	assert.Equal(val, 123.0)

	val = config.GetAsFloat(12.123, 123)
	assert.Equal(val, 12.123)
}

func TestParseGoodConfig(t *testing.T) {
	_, err := config.ReadConfig(tmpTestGoodFile)
	assert.Nil(t, err, "should succeed")
}

func TestParseBadConfig(t *testing.T) {
	_, err := config.ReadConfig(tmpTestBadFile)
	assert.NotNil(t, err, "should fail")
}
