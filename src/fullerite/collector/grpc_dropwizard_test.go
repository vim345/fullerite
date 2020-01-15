package collector

import (
	"context"
	"errors"
	"fullerite/metric"
	"testing"

	metrics "fullerite/collector/metrics"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const (
	PORT        = "46459"
	ADDR        = "127.0.0.1"
	ServiceName = "ranking_platform_query_engine"
	INTERVAL    = 5
	TIMEOUT     = 5
)

var grpcEndpoint = GrpcEndpoint{
	Name: ServiceName,
	Addr: ADDR,
	Port: PORT,
}

type MockedGrpcClinet struct{}

func (m MockedGrpcClinet) Metrics(
	ctx context.Context, in *metrics.MetricsRequest, opts ...grpc.CallOption) (*metrics.MetricsResponse, error) {
	return &metrics.MetricsResponse{
		Data: `{
  "version": "4.0.0",
  "service_dims": {
    "git_sha": "aabbcc",
    "deploy_group": "canary"
  },
  "gauges": {
    "jvm.attribute.uptime": {
      "value": 252892259
    }
  }
}`,
	}, nil
}

type BadMockedGrpcClinet struct{}

func (m BadMockedGrpcClinet) Metrics(
	ctx context.Context, in *metrics.MetricsRequest, opts ...grpc.CallOption) (*metrics.MetricsResponse, error) {
	var response *metrics.MetricsResponse
	return response, errors.New("I am a fake error")
}

type MockedGrpcEndpoint struct {
	GrpcEndpoint
}

func (e MockedGrpcEndpoint) getClient() (metrics.MetricsClient, error) {
	return &MockedGrpcClinet{}, nil
}

type BadMockedGrpcEndpoint struct {
	GrpcEndpoint
}

func (e BadMockedGrpcEndpoint) getClient() (metrics.MetricsClient, error) {
	return &BadMockedGrpcClinet{}, nil
}

func getTestGrpcDropwizard() *grpcDropwizardCollector {
	return newGrpcDropwizard(make(chan metric.Metric), 12, l.WithField("test", "grpcDropwizard")).(*grpcDropwizardCollector)
}

func TestConfigGrpcDropwizard(t *testing.T) {
	service := map[string]string{}
	service["service_name"] = ServiceName
	service["addr"] = ADDR
	service["port"] = PORT

	endpoints := make([]interface{}, 1)
	endpoints[0] = service

	cfg := map[string]interface{}{
		"interval":  INTERVAL,
		"endpoints": endpoints,
		"timeout":   TIMEOUT,
	}

	inst := getTestGrpcDropwizard()

	inst.Configure(cfg)

	assert.Equal(t, INTERVAL, inst.Interval())
	assert.Equal(t, ServiceName, inst.endpoints[0].Name)
	assert.Equal(t, PORT, inst.endpoints[0].Port)
	assert.Equal(t, TIMEOUT, inst.timeout)
}

func TestGetMetrics(t *testing.T) {
	mockedGrpcEndpoint := &MockedGrpcEndpoint{grpcEndpoint}

	inst := getTestGrpcDropwizard()

	go inst.getMetrics(mockedGrpcEndpoint)

	for data := range inst.Channel() {
		assert.Equal(t, data.Dimensions["git_sha"], "aabbcc")
		assert.Equal(t, data.Dimensions["deploy_group"], "canary")
		assert.Equal(t, "gauge", data.MetricType)
		assert.Equal(t, 6, len(data.Dimensions))
		assert.Equal(t, "jvm.attribute.uptime", data.Name)
		close(inst.Channel())
	}
}

func TestGetMetricsWithErrors(t *testing.T) {
	mockedGrpcEndpoint := &BadMockedGrpcEndpoint{grpcEndpoint}

	inst := getTestGrpcDropwizard()

	// If there's an error, it means no message is published to the metrics channel. That's why
	// code can normally finish and there's no need to close the channel separately similar to
	// how it's done when a valid response is returned from the server and published into
	// the metrics channel.
	inst.getMetrics(mockedGrpcEndpoint)
}
