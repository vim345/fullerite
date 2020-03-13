package collector

import (
	"fullerite/metric"

	grpcMetrics "fullerite/collector/metrics"
	"testing"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var sampleOne = grpcMetrics.Sample{
	Name:        "S1",
	LabelNames:  []string{"grpc_type", "grpc_method"},
	LabelValues: []string{"BIDI_STREAMING", "ServerReflectionInfo"},
	Value:       800.0,
}

var sampleTwo = grpcMetrics.Sample{
	Name:        "S2",
	LabelNames:  []string{"grpc_type", "grpc_method"},
	LabelValues: []string{"BIDI_STREAMING", "ServerReflectionInfo"},
	Value:       900.0,
}

var sampleThree = grpcMetrics.Sample{
	Name:        "S3",
	LabelNames:  []string{"grpc_type", "grpc_method"},
	LabelValues: []string{"BIDI_STREAMING", "ServerReflectionInfo"},
	Value:       1000.0,
}

var sampleFour = grpcMetrics.Sample{
	Name:        "S4",
	LabelNames:  []string{"grpc_type", "grpc_method"},
	LabelValues: []string{"BIDI_STREAMING", "ServerReflectionInfo"},
	Value:       400.0,
}

var sampleFive = grpcMetrics.Sample{
	Name:        "S5",
	LabelValues: []string{"grpc_type", "grpc_method"},
	LabelNames:  []string{"BIDI_STREAMING", "ServerReflectionInfo"},
	Value:       500.0,
}

var metricOne = grpcMetrics.MetricFamilySamples{
	Name:    "grpc_server_handled_total",
	Type:    grpcMetrics.SampleType_COUNTER,
	Help:    "Total number of RPCs completed on the server.",
	Samples: []*grpcMetrics.Sample{&sampleOne, &sampleTwo, &sampleThree},
}

var metricTwo = grpcMetrics.MetricFamilySamples{
	Name:    "grpc_server_started_total",
	Type:    grpcMetrics.SampleType_GAUGE,
	Help:    "Total number of RPCs started on the server.",
	Samples: []*grpcMetrics.Sample{&sampleFour, &sampleFive},
}

type mockedGRPCGetter struct{}

// Get retrieves content from the metrics gRPC endpoint.
func (m *mockedGRPCGetter) Get() ([]*grpcMetrics.MetricFamilySamples, error) {
	return []*grpcMetrics.MetricFamilySamples{&metricOne, &metricTwo}, nil
}

func init() {
	l.SetLevel(l.DebugLevel)
}

func getTestNerveGRPC() *nerveGRPCCollector {
	return newNerveGRPC(make(chan metric.Metric), 12, l.WithField("testing", "nervegrpc")).(*nerveGRPCCollector)
}

func TestDefaultConfigNerveGRPC(t *testing.T) {
	inst := getTestNerveGRPC()
	inst.Configure(make(map[string]interface{}))

	assert.Equal(t, 12, inst.Interval())
	assert.Equal(t, 2, inst.timeout)
	assert.Equal(t, "/etc/nerve/nerve.conf.json", inst.configFilePath)
}

func TestConfigNerveGRPC(t *testing.T) {
	cfg := map[string]interface{}{
		"interval":          345,
		"servicesWhitelist": []string{"test_service"},
		"configFilePath":    "/etc/your/moms/house",
		"http_timeout":      12,
	}

	inst := getTestNerveGRPC()
	inst.Configure(cfg)

	assert.Equal(t, 345, inst.Interval())
	assert.Equal(t, "/etc/your/moms/house", inst.configFilePath)
	assert.Equal(t, []string{"test_service"}, inst.servicesWhitelist)
	assert.Equal(t, 12, inst.timeout)
}

func TestQueryService(t *testing.T) {
	inst := getTestNerveGRPC()
	go inst.queryService("grpc_service", 8888, &mockedGRPCGetter{})

	actual := []metric.Metric{}
	lenLines := 5
	for i := 0; i < lenLines; i++ {
		actual = append(actual, <-inst.Channel())
	}

	validateGRPCResults(t, actual, lenLines)
}

func validateGRPCResults(t *testing.T, actual []metric.Metric, length int) {
	assert.Equal(t, length, len(actual))

	for _, m := range actual {
		switch m.Name {
		case "grpc_service_S1":
			grpcTypeDim, exists := m.GetDimensionValue("grpc_type")
			assert.True(t, exists)
			assert.Equal(t, "BIDI_STREAMING", grpcTypeDim)
			grpcMethodDim, exists := m.GetDimensionValue("grpc_method")
			assert.True(t, exists)
			assert.Equal(t, "ServerReflectionInfo", grpcMethodDim)
			assert.Equal(t, 800.0, m.Value)
			assert.Equal(t, metric.Counter, m.MetricType)
		case "grpc_service_S2":
			assert.Equal(t, 900.0, m.Value)
			assert.Equal(t, metric.Counter, m.MetricType)
		case "grpc_service_S3":
			assert.Equal(t, 1000.0, m.Value)
			assert.Equal(t, metric.Counter, m.MetricType)
		case "grpc_service_S4":
			assert.Equal(t, 400.0, m.Value)
			assert.Equal(t, metric.Gauge, m.MetricType)
		case "grpc_service_S5":
			assert.Equal(t, 500.0, m.Value)
			assert.Equal(t, metric.Gauge, m.MetricType)
		default:
			t.Fatal("Unexpected metric name: " + m.Name)
		}

	}
}
