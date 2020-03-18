package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestNerveConfig() []byte {
	raw := `
	{
	    "heartbeat_path": "/var/run/nerve/heartbeat",
	    "instance_id": "srv1-devc",
	    "services": {
	        "example_service.main.norcal-devc.superregion:norcal-devc.13752.new": {
	            "check_interval": 7,
	            "checks": [
	                {
	                    "fall": 2,
	                    "headers": {},
	                    "host": "127.0.0.1",
	                    "open_timeout": 6,
	                    "port": 6666,
	                    "rise": 1,
	                    "timeout": 6,
	                    "type": "http",
	                    "uri": "/http/example_service.main/13752/status"
	                }
	            ],
	            "host": "10.56.5.21",
	            "port": 13752,
	            "weight": 24,
	            "zk_hosts": [
	                "10.40.5.5:22181",
	                "10.40.5.6:22181",
	                "10.40.1.17:22181"
	            ],
	            "zk_path": "/nerve/superregion:norcal-devc/example_service.main"
	        },
	        "example_service.mesosstage_main.norcal-devc.superregion:norcal-devc.13752.new": {
	            "check_interval": 7,
	            "checks": [
	                {
	                    "fall": 2,
	                    "headers": {},
	                    "host": "127.0.0.1",
	                    "open_timeout": 6,
	                    "port": 6666,
	                    "rise": 1,
	                    "timeout": 6,
	                    "type": "http",
	                    "uri": "/http/example_service.mesosstage_main/22224/status"
	                }
	            ],
	            "host": "10.56.5.21",
	            "port": 22222,
	            "weight": 24,
	            "zk_hosts": [
	                "10.40.5.5:22181",
	                "10.40.5.6:22181",
	                "10.40.1.17:22181"
	            ],
	            "zk_path": "/nerve/superregion:norcal-devc/example_service.mesosstage_main"
	        },
	        "example_service.another.norcal-devc.superregion:norcal-devc.13752.new": {
	            "check_interval": 7,
	            "checks": [
	                {
	                    "fall": 2,
	                    "headers": {},
	                    "host": "127.0.0.1",
	                    "open_timeout": 6,
	                    "port": 6666,
	                    "rise": 1,
	                    "timeout": 6,
	                    "type": "http",
	                    "uri": "/https/example_service.another/13752/status"
	                }
	            ],
	            "host": "10.56.5.21",
	            "port": 22222,
	            "weight": 24,
	            "zk_hosts": [
	                "10.40.5.5:22181",
	                "10.40.5.6:22181",
	                "10.40.1.17:22181"
	            ],
	            "zk_path": "/nerve/superregion:norcal-devc/example_service.another"
	        },
	        "example_grpc_service.grpc_main.norcal-devc.superregion:norcal-devc.13752.new": {
	            "check_interval": 7,
	            "checks": [
	                {
	                    "fall": 2,
	                    "headers": {},
	                    "host": "127.0.0.1",
	                    "open_timeout": 6,
	                    "port": 12345,
	                    "rise": 1,
	                    "timeout": 6,
	                    "type": "tcp",
	                    "uri": "/tcp/example_grpc_service.another/12345/status"
	                }
	            ],
	            "host": "10.56.5.21",
	            "port": 2222,
	            "weight": 24,
	            "zk_hosts": [
	                "10.40.5.5:22181",
	                "10.40.5.6:22181",
	                "10.40.1.17:22181"
	            ],
	            "zk_path": "/nerve/superregion:norcal-devc/example_grpc_service.another"
	        }
	    }
	}
	`
	return []byte(raw)
}

func noURINerveConfig() []byte {
	return []byte(`
	{
	    "services": {
	        "example_service.main.norcal-devc.superregion:norcal-devc.13752.new": {
	            "check_interval": 7,
	            "checks": [
	                {
	                    "fall": 2,
	                    "headers": {},
	                    "host": "127.0.0.1",
	                    "type": "http"
	                }
	            ],
	            "host": "10.56.5.21",
	            "port": 13752,
	            "weight": 24,
	            "zk_hosts": [
	                "10.40.1.17:22181"
	            ],
	            "zk_path": "/nerve/superregion:norcal-devc/example_service.main"
	        }
             }
        }`)
}

func badURINerveConfig() []byte {
	return []byte(`
	{
	    "services": {
	        "example_service.main.norcal-devc.superregion:norcal-devc.13752.new": {
	            "check_interval": 7,
	            "checks": [
	                {
	                    "fall": 2,
	                    "headers": {},
	                    "host": "127.0.0.1",
	                    "type": "http",
	                    "uri": "/http/example_service.main"
	                }
	            ],
	            "host": "10.56.5.21",
	            "port": 13752,
	            "weight": 24,
	            "zk_hosts": [
	                "10.40.1.17:22181"
	            ],
	            "zk_path": "/nerve/superregion:norcal-devc/example_service.main"
	        }
             }
        }`)
}

func TestNerveConfigParsingTCP(t *testing.T) {
	expected := map[NerveService]bool{
		NerveService{Name: "example_grpc_service", Namespace: "grpc_main", Port: 12345, Host: "10.56.5.21"}: true,
	}

	cfgString := getTestNerveConfig()
	ipGetter = func() ([]string, error) { return []string{"10.56.5.21"}, nil }
	results, err := ParseNerveConfig(&cfgString, true, "tcp")
	assert.Nil(t, err)
	m := make(map[NerveService]bool)
	for _, r := range results {
		m[r] = true
	}
	assert.Equal(t, expected, m)
}

func TestNerveConfigParsing(t *testing.T) {
	expected := map[NerveService]bool{
		NerveService{Name: "example_service", Namespace: "mesosstage_main", Port: 22224, Host: "10.56.5.21"}: true,
		NerveService{Name: "example_service", Namespace: "main", Port: 13752, Host: "10.56.5.21"}:            true,
		NerveService{Name: "example_service", Namespace: "another", Port: 13752, Host: "10.56.5.21"}:         true,
	}

	cfgString := getTestNerveConfig()
	ipGetter = func() ([]string, error) { return []string{"10.56.5.21"}, nil }
	results, err := ParseNerveConfig(&cfgString, true, "http")
	assert.Nil(t, err)
	m := make(map[NerveService]bool)
	for _, r := range results {
		m[r] = true
	}
	assert.Equal(t, expected, m)
}

func TestNerveConfigParsingiNoNamespace(t *testing.T) {
	expected := map[int]bool{
		22224: true, 13752: true,
	}

	cfgString := getTestNerveConfig()
	ipGetter = func() ([]string, error) { return []string{"10.56.5.21"}, nil }
	results, err := ParseNerveConfig(&cfgString, false, "http")
	assert.Nil(t, err)
	m := make(map[int]bool)
	for _, r := range results {
		m[r.Port] = true
	}
	assert.Equal(t, expected, m)
}

func TestHandleBadNerveConfig(t *testing.T) {
	// b/c there is valid json coming in it won't error, just have an empty response
	cfgString := []byte("{}")
	ipGetter = func() ([]string, error) { return []string{"10.56.2.3"}, nil }
	results, err := ParseNerveConfig(&cfgString, true, "http")
	assert.Nil(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
}

func TestHandlePoorlyFormedJson(t *testing.T) {
	cfgString := []byte("notjson")
	ipGetter = func() ([]string, error) { return []string{"10.56.2.3"}, nil }
	results, err := ParseNerveConfig(&cfgString, true, "http")
	assert.NotNil(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
}

func TestNoURI(t *testing.T) {
	cfgString := noURINerveConfig()
	ipGetter = func() ([]string, error) { return []string{"10.56.5.21"}, nil }
	results, _ := ParseNerveConfig(&cfgString, true, "http")
	assert.Equal(t, 0, len(results))
}

func TestBadURI(t *testing.T) {
	cfgString := badURINerveConfig()
	ipGetter = func() ([]string, error) { return []string{"10.56.5.21"}, nil }
	results, _ := ParseNerveConfig(&cfgString, true, "http")
	assert.Equal(t, 0, len(results))
}
