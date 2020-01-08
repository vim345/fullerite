package collector

import (
	"context"
	"fmt"
	metrics "fullerite/collector/metrics"
	"fullerite/config"
	"fullerite/dropwizard"
	"fullerite/metric"

	l "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
)

const schemaVer = "java-1.1"

// The gRPC Dropwizard Collector collects metrics emitted by java services
// with one of the schemas defined at dropwizard/base_parser.go#L80.
// User needs to specify port and path where the service's metrics endpoint is setup.
type grpcDropwizardCollector struct {
	baseCollector

	endpoints []GrpcEndpoint
	timeout   int
}

func newGrpcDropwizard(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	col := new(grpcDropwizardCollector)

	col.log = log
	col.channel = channel
	col.interval = initialInterval

	col.name = "GrpcDropwizard"
	col.timeout = 3
	return col
}

// GrpcConnector provides common interfaces to connect to a gRPC endpoint.
type GrpcConnector interface {
	getClient() (metrics.MetricsClient, error)
	getName() string
	getPort() string
	getAddr() string
}

// GrpcEndpoint defines a struct for gRPC connections.
type GrpcEndpoint struct {
	// Name is the service name
	Name string
	// Addr is the gRPC service address.
	Addr string
	// Port is the gRPC service port.
	Port string
}

func (e GrpcEndpoint) getName() string {
	return e.Name
}

func (e GrpcEndpoint) getAddr() string {
	return e.Addr
}

func (e GrpcEndpoint) getPort() string {
	return e.Port
}

func (e GrpcEndpoint) getClient() (metrics.MetricsClient, error) {
	var metricsClient metrics.MetricsClient
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", e.getAddr(), e.getPort()), grpc.WithInsecure())

	if err != nil {
		return metricsClient, err
	}

	defer conn.Close()

	return metrics.NewMetricsClient(conn), nil
}

func (g *grpcDropwizardCollector) getMetrics(endpoint GrpcConnector) {
	serviceLog := g.log.WithField("service", endpoint.getName())

	client, err := endpoint.getClient()
	if err != nil {
		return
	}

	res, err := client.Metrics(context.Background(), &metrics.MetricsRequest{})
	if err != nil {
		l.Warningf("Failed to get the results: %s", err)
		return
	}

	metrics, err := dropwizard.Parse([]byte(res.Data), schemaVer, true)

	if err != nil {
		serviceLog.Warn("Failed to parse response into metrics: ", err)
		return
	}

	metric.AddToAll(&metrics, map[string]string{
		"service": endpoint.getName(),
		"port":    endpoint.getPort(),
	})
	serviceLog.Debug("Sending ", len(metrics), " to channel")
	for _, m := range metrics {
		if !g.ContainsBlacklistedDimension(m.Dimensions) {
			g.Channel() <- m
		}
	}
}

func init() {
	RegisterCollector("GrpcDropwizard", newGrpcDropwizard)
}

func (g *grpcDropwizardCollector) Collect() {
	for _, endpoint := range g.endpoints {
		go g.getMetrics(endpoint)
	}
}

func (g *grpcDropwizardCollector) Configure(configMap map[string]interface{}) {
	if val, exists := configMap["endpoints"]; exists {
		val := val.([]interface{})
		g.endpoints = make([]GrpcEndpoint, len(val))
		index := 0
		for _, e := range val {
			endpoint := config.GetAsMap(e)
			g.endpoints[index] = GrpcEndpoint{
				Name: endpoint["service_name"],
				Addr: endpoint["addr"],
				Port: endpoint["port"],
			}
			index++
		}
	}

	if val, exists := configMap["imeout"]; exists {
		g.timeout = config.GetAsInt(val, 2)
	}

	g.configureCommonParams(configMap)
}
