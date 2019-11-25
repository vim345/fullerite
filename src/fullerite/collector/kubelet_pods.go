package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	l "github.com/Sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"fullerite/config"
	"fullerite/metric"
)

const (
	defaultPort = 10255
)

// KubeletPods collector type.
type KubeletPods struct {
	baseCollector
	timeout       int
	compiledRegex map[string]*Regex
	url           string
}

func init() {
	RegisterCollector("KubeletPods", newKubeletPods)
}

// newKubeletPods creates a new collector for pods returned by kubelet.
func newKubeletPods(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	d := new(KubeletPods)

	d.log = log
	d.channel = channel
	d.interval = initialInterval

	d.name = "KubeletPods"
	d.compiledRegex = make(map[string]*Regex)
	return d
}

// GetURL Returns URL of KubeletPods instance
func (d *KubeletPods) GetURL() string {
	return d.url
}

// Configure takes a dictionary of values with which the handler can configure itself.
func (d *KubeletPods) Configure(configMap map[string]interface{}) {
	if timeout, exists := configMap["kubeletTimeout"]; exists {
		d.timeout = min(config.GetAsInt(timeout, d.interval), d.interval)
	} else {
		d.timeout = d.interval
	}

	var port int
	if kubeletPort, exists := configMap["kubeletPort"]; exists {
		port = config.GetAsInt(kubeletPort, defaultPort)
	} else {
		port = defaultPort
	}
	d.url = fmt.Sprintf("http://localhost:%d/pods", port)

	if generatedDimensions, exists := configMap["generatedDimensions"]; exists {
		for dimension, generator := range generatedDimensions.(map[string]interface{}) {
			for key, regx := range config.GetAsMap(generator) {
				re, err := regexp.Compile(regx)
				if err != nil {
					d.log.Warn("Failed to compile regex: ", regx, err)
				} else {
					d.compiledRegex[dimension] = &Regex{regex: re, tag: key}
				}
			}
		}
	}

	d.configureCommonParams(configMap)
}

// Collect iterates on all the pods and, if possible, collects the
// correspondent statistics.
func (d *KubeletPods) Collect() {
	client := http.Client{
		Timeout: time.Second * time.Duration(d.timeout),
	}

	res, getErr := client.Get(d.url)
	if getErr != nil {
		d.log.Error("Error sending request to kubelet: ", getErr)
		return
	}
	defer res.Body.Close()

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		d.log.Error("Error reading response: ", readErr)
		return
	}

	podList := corev1.PodList{}
	jsonErr := json.Unmarshal(body, &podList)
	if jsonErr != nil {
		d.log.Error("Error parsing response: ", jsonErr)
		return
	}

	metrics := []metric.Metric{}
	for i := range podList.Items {
		metrics = append(metrics, d.getPodInfo(&podList.Items[i])...)
	}
	d.sendMetrics(metrics)
}

// getPodInfo gets pod info for the given pod.
func (d *KubeletPods) getPodInfo(pod *corev1.Pod) []metric.Metric {
	metrics := []metric.Metric{}
	for i := range pod.Spec.Containers {
		metrics = append(metrics, d.getContainerInfo(&pod.Spec.Containers[i])...)
	}
	metric.AddToAll(&metrics, d.extractDimensions(pod))
	return metrics
}

// getContainerInfo gets container info for the given container.
func (d *KubeletPods) getContainerInfo(container *corev1.Container) []metric.Metric {
	metrics := []metric.Metric{}
	if ephemeralStorageLimit, ok := container.Resources.Limits[corev1.ResourceEphemeralStorage]; ok {
		metrics = append(metrics, buildKubeletMetric("KubernetesContainerEphemeralStorageLimit", metric.Gauge, float64(ephemeralStorageLimit.Value())))
	}
	additionalDimensions := map[string]string{
		"container_name": container.Name,
	}
	metric.AddToAll(&metrics, additionalDimensions)
	return metrics
}

func buildKubeletMetric(name string, metricType string, value float64) (m metric.Metric) {
	m = metric.New(name)
	m.MetricType = metricType
	m.Value = value
	return m
}

// Function that extracts additional dimensions from the Kubernetes pod labels
// set up by the user in the configuration file.
func (d KubeletPods) extractDimensions(pod *corev1.Pod) map[string]string {
	ret := map[string]string{}

	for dimension, r := range d.compiledRegex {
		if value, ok := pod.Labels[r.tag]; ok {
			subMatch := r.regex.FindStringSubmatch(value)
			if len(subMatch) > 0 {
				ret[dimension] = strings.Replace(subMatch[len(subMatch)-1], "--", "_", -1)
			}
		}
	}
	d.log.Debug(ret)
	return ret
}

// sendMetrics writes all the metrics received to the collector channel.
func (d KubeletPods) sendMetrics(metrics []metric.Metric) {
	for _, m := range metrics {
		d.Channel() <- m
	}
}
