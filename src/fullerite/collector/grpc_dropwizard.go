package collector

import (
	"fullerite/config"
	"fullerite/metric"

	l "github.com/Sirupsen/logrus"
)

// The HTTP Dropwizard Collector allows to collect metrics emitted by java/python services
// with one of the schemas defined at dropwizard/base_parser.go#L80.
// User needs to specify port and path where the service'metrics endpoint is setup.
type grpcDropwizardCollector struct {
	baseCollector

	endpoints []ServiceEndpoint
	timeout   int
}

func init() {
	RegisterCollector("GrpcDropwizard", newGRPCDropwizard)
}

func newGRPCDropwizard(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	col := new(grpcDropwizardCollector)

	col.log = log
	col.channel = channel
	col.interval = initialInterval

	col.name = "GrpcDropwizard"
	col.timeout = 3
	return col
}

func (g *grpcDropwizardCollector) Collect() {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) Configure(map[string]interface{}) {
	if val, exists := configMap["endpoints"]; exists {
		val := val.([]interface{})
		h.endpoints = make([]ServiceEndpoint, len(val))
		index := 0
		for _, e := range val {
			endpoint := config.GetAsMap(e)
			h.endpoints[index] = ServiceEndpoint{
				Name: endpoint["service_name"],
				Port: endpoint["port"],
				Path: endpoint["path"],
			}
			index++
		}
	}

	if val, exists := configMap["http_timeout"]; exists {
		h.timeout = config.GetAsInt(val, 2)
	}

	h.configureCommonParams(configMap)
}

func (g *grpcDropwizardCollector) Name() string {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) Channel() chan metric.Metric {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) Interval() int {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) SetInterval(int) {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) CollectorType() string {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) SetCollectorType(string) {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) CanonicalName() string {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) SetCanonicalName(string) {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) Prefix() string {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) SetPrefix(string) {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) Blacklist() []string {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) SetBlacklist([]string) {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) DimensionsBlacklist() map[string]string {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) SetDimensionsBlacklist(map[string]string) {
	panic("not implemented")
}

func (g *grpcDropwizardCollector) ContainsBlacklistedDimension(map[string]string) bool {
	panic("not implemented")
}
