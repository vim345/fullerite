package util

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// For dependency injection
var (
	ipGetter   = getIps
	httpRegexp = regexp.MustCompile(`http`)
	tcpRegexp  = regexp.MustCompile(`tcp`)
)

// example configuration::
//
// {
//     	"heartbeat_path": "/var/run/nerve/heartbeat",
//		"instance_id": "srv1-devc",
//		"services": {
//	 		"<SERVICE_NAME>.<NAMESPACE>.<otherstuff>": {
//				"host": "<IPADDR>",
//      	    "port": ###,
//      	}
// 		}
//     "services": {
//
// Most imporantly is the port, host and service name. The service name is assumed to be formatted like this::
//
type nerveConfigData struct {
	Services map[string]map[string]interface{}
}

// NerveService is an exported struct containing services' info
type NerveService struct {
	Name      string
	Namespace string
	Host      string
	Port      int
}

// EndPoint defines a struct for endpoints
type EndPoint struct {
	Host string
	Port string
}

// ParseNerveConfig is responsible for taking the JSON string coming in into a list of NerveServices
func ParseNerveConfig(raw *[]byte, namespaceIncluded bool) ([]NerveService, error) {
	services := make(map[string]NerveService)
	results := []NerveService{}
	parsed := new(nerveConfigData)

	err := json.Unmarshal(*raw, parsed)
	if err != nil {
		return results, err
	}

	for rawServiceName, serviceConfig := range parsed.Services {
		host := strings.TrimSpace(serviceConfig["host"].(string))
		service := new(NerveService)
		service.Name = strings.Split(rawServiceName, ".")[0]
		service.Namespace = strings.Split(rawServiceName, ".")[1]
		service.Host = host
		port := extractPort(serviceConfig)

		if port != -1 {
			service.Port = port
			if namespaceIncluded {
				services[service.Name+service.Namespace+":"+strconv.Itoa(port)] = *service
			} else {
				services[service.Name+":"+strconv.Itoa(port)] = *service
			}
		}
	}

	for _, value := range services {
		results = append(results, value)
	}
	return results, nil
}

func extractPort(serviceConfig map[string]interface{}) int {
	checkConfig := make(map[string]interface{})

	if checkArray, ok := serviceConfig["checks"].([]interface{}); ok {
		checkConfig = checkArray[0].(map[string]interface{})
	}

	var uri string

	if uriInterface, ok := checkConfig["uri"]; ok {
		if str, ok := uriInterface.(string); ok {
			uri = str
		}
	}

	if len(uri) == 0 {
		return -1
	}

	uriArray := strings.Split(uri, "/")
	if len(uriArray) > 3 {
		protocol := strings.TrimSpace(uriArray[1])
		port := uriArray[3]
		if !httpRegexp.MatchString(protocol) && !tcpRegexp.MatchString(protocol) {
			return -1
		}
		if portInt, err := strconv.ParseInt(port, 10, 64); err == nil {
			return int(portInt)
		}
	}
	return -1
}

// getIps is responsible for getting all the ips that are associated with this NIC
func getIps() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return []string{}, err
	}

	results := []string{}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return []string{}, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			results = append(results, ip.String())
		}
	}

	return results, nil
}

// CreateMinimalNerveConfig creates a minimal nerve config
func CreateMinimalNerveConfig(config map[string]EndPoint) map[string]map[string]map[string]interface{} {
	minimalNerveConfig := make(map[string]map[string]map[string]interface{})
	serviceConfigs := make(map[string]map[string]interface{})
	for service, endpoint := range config {
		uriEndPoint := fmt.Sprintf("/http/%s/%s/status", service, endpoint.Port)
		serviceConfigs[service] = map[string]interface{}{
			"host": endpoint.Host,
			"port": endpoint.Port,
			"checks": []interface{}{
				map[string]interface{}{
					"uri": uriEndPoint,
				},
			},
		}
	}
	minimalNerveConfig["services"] = serviceConfigs
	return minimalNerveConfig
}
