package collector

import (
	"encoding/json"
	"testing"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"fullerite/metric"
)

func getSUT2() *KubeletPods {
	expectedChan := make(chan metric.Metric)
	var expectedLogger = defaultLog.WithFields(l.Fields{"collector": "fullerite"})

	return newKubeletPods(expectedChan, 10, expectedLogger).(*KubeletPods)
}

func TestKubeletPodsNewKubeletPods(t *testing.T) {
	expectedChan := make(chan metric.Metric)
	var expectedLogger = defaultLog.WithFields(l.Fields{"collector": "fullerite"})

	d := newKubeletPods(expectedChan, 10, expectedLogger).(*KubeletPods)

	assert.Equal(t, d.log, expectedLogger)
	assert.Equal(t, d.channel, expectedChan)
	assert.Equal(t, d.interval, 10)
	assert.Equal(t, d.name, "KubeletPods")
	d.Configure(make(map[string]interface{}))
	assert.Equal(t, d.GetURL(), "http://localhost:10255/pods")
}

func TestKubeletPodsConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})

	d := newKubeletPods(nil, 123, nil).(*KubeletPods)
	d.Configure(config)

	assert.Equal(t, 123, d.Interval())
}

func TestKubeletPodsConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999

	d := newKubeletPods(nil, 123, nil).(*KubeletPods)
	d.Configure(config)

	assert.Equal(t, 9999, d.Interval())
}

func TestKubeletPodsGetPodInfo(t *testing.T) {
	config := make(map[string]interface{})
	dims := []byte(`
	{
		"service_name":  {
			"acme.com/service": ".*"
		},
		"instance_name": {
			"acme.com/instance": ".*"}
	}`)
	var val map[string]interface{}

	err := json.Unmarshal(dims, &val)
	assert.Equal(t, err, nil)
	config["generatedDimensions"] = val

	podJSON := []byte(`
		{
			"metadata": {
				"name": "foo-bar-56dbf584cf-c5pd9",
				"namespace": "somewhere",
				"labels": {
					"acme.com/instance": "bar",
					"acme.com/service": "foo",
					"acme.com/cluster": "baz"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "main",
						"resources": {
							"limits": {
								"cpu": "1100m",
								"ephemeral-storage": "50Gi",
								"memory": "256Mi"
							},
							"requests": {
								"cpu": "100m",
								"ephemeral-storage": "45Gi",
								"memory": "256Mi"
							}
						}
					},
					{
						"name": "aux",
						"resources": {
							"limits": {
								"ephemeral-storage": "33Gi"
							}
						}
					}
				]
			},
			"status": {
				"phase": "Running"
			}
		}`)
	var pod *corev1.Pod
	err = json.Unmarshal(podJSON, &pod)
	assert.Equal(t, err, nil)

	container1Dims := map[string]string{
		"container_name": "main",
		"service_name":   "foo",
		"instance_name":  "bar",
	}
	container2Dims := map[string]string{
		"container_name": "aux",
		"service_name":   "foo",
		"instance_name":  "bar",
	}

	expectedMetrics := []metric.Metric{
		metric.Metric{"KubernetesContainerEphemeralStorageLimit", "gauge", 53687091200, container1Dims},
		metric.Metric{"KubernetesContainerEphemeralStorageLimit", "gauge", 35433480192, container2Dims},
	}

	d := getSUT2()
	d.Configure(config)
	ret := d.getPodInfo(pod)
	assert.Equal(t, ret, expectedMetrics)
}

func TestKubeletPodsGetPodInfoWithoutLimit(t *testing.T) {
	config := make(map[string]interface{})
	dims := []byte(`
	{
		"service_name":  {
			"acme.com/service": ".*"
		},
		"instance_name": {
			"acme.com/instance": ".*"}
	}`)
	var val map[string]interface{}

	err := json.Unmarshal(dims, &val)
	assert.Equal(t, err, nil)
	config["generatedDimensions"] = val

	podJSON := []byte(`
		{
			"metadata": {
				"name": "foo-bar-56dbf584cf-c5pd9",
				"namespace": "somewhere",
				"labels": {
					"acme.com/instance": "bar",
					"acme.com/service": "foo",
					"acme.com/cluster": "baz"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "main"
					}
				]
			},
			"status": {
				"phase": "Running"
			}
		}`)
	var pod *corev1.Pod
	err = json.Unmarshal(podJSON, &pod)
	assert.Equal(t, err, nil)

	expectedMetrics := []metric.Metric{}

	d := getSUT2()
	d.Configure(config)
	ret := d.getPodInfo(pod)
	assert.Equal(t, ret, expectedMetrics)
}
