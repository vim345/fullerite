package main

import (
	"io/ioutil"
	"os"
	"testing"
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
	_, err := readConfig(tmpTestFile)
	if err != nil {
		t.Fail()
	}
}
