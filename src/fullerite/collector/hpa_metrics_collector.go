
package collector

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	autoscalingAnnotation = "autoscaling"
	instanceNameLabelKey = "paasta.yelp.com/instance"
	metricsEndpoints := map[string]string{
		"uwsgi": "status/uwsgi",
		"http": "status"
	}
)

var dimensionSanitizer = strings.NewReplacer(
	".", "_",
	"/", "_")

type HPAMetrics struct {
	baseCollector
	kubeletTimeout       int
	metricsProviderTimeout       int
	podSpecURL           string
	additionalDimensions map[string]string
}

func init() {
	RegisterCollector("HPAMetrics", newHPAMetrics)
}

func newHPAMetrics(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	d := new(HPAMetrics)

	d.log = log
	d.channel = channel
	d.interval = initialInterval

	d.name = "HPAMetrics"
	d.additionalDimensions = make(map[string]string)
	return d
}

// HPAMetrics
func (d *HPAMetrics) Configure(configMap map[string]interface{}) {
	if kubeletTimeout, exists := configMap["kubeletTimeout"]; exists {
		d.kubeletTimeout = min(config.GetAsInt(kubeletTimeout, d.interval), d.interval)
	} else {
		d.kubeletTimeout = d.kubeletTimeout 
	}

	if metricsProviderTimeout, exists := configMap["metricsProviderTimeout"]; exists {
		d.metricsProviderTimeout = min(config.GetAsInt(metricsProviderTimeout, d.interval), d.interval)
	} else {
		d.metricsProviderTimeout = d.metricsProviderTimeout 
	}

	var kubeletPort int
	if kubeletPort, exists := configMap["kubeletPort"]; exists {
		kubeletPort = config.GetAsInt(kubeletPort, defaultPort)
	} else {
		kubeletPort = defaultPort 
	}
	d.podSpecURL = fmt.Sprintf("http://localhost:%d/pods", kubeletPort)

	if additionalDimensions, exists := configMap["additionalDimensions"]; exists {
		d.additionalDimensions = config.GetAsMap(additionalDimensions, d.additionalDimensions)
	}

	d.configureCommonParams(configMap)
}

// Collect iterates on all the pods and, if possible, collects the
// correspondent statistics.
func (d *HPAMetrics) Collect() {
	client := http.Client{
		Timeout: time.Second * time.Duration(d.kubeletTimeout),
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

	for i := range podList.Items {
		go d.CollectMetricsForPod(&podList.Items[i])
	}
}

func (d *HPAMetrics) CollectMetricsForPod(pod *corev1.Pod) {
	metricsName, annotationPresent := pod.GetAnnotations()[autoscalingAnnotation]
	if !annotationPresent || d.allContainersAreReady(pod) {
		return
	}
	podIP := pod.Status.PodIP.IP
	podName := pod.GetName()
	podNamespace := pod.GetNamespace()
	labels := &pod.GetLabels()
	instanceName = labels[instanceNameLabelKey]
	containerPort, _ := d.getContainerPort(pod, instanceName)

	url := fmt.Sprintf("http://%s:%s/%s", podIP, containerPort, metricName, metricsEndpoints[metricsName])
	res, getErr := client.Get(url)
	if getErr != nil {
		d.log.Error("Error sending request to metricsProvider: ", getErr)
		return
	}
	defer res.Body.Close()
	raw, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		d.log.Error("Error reading response: ", readErr)
		return
	}

	var value float64
	switch metricName {
		case "uwsgi": value = parseUWSGIMetrics(raw)
		case "http": value = parseHTTPMetrics(raw)
	}

	labels["kubernetes_namespace"] = podNamespace
	labels["kubernetes_pod_name"] = podName
	for k, v := range d.addDimensions {
		labels[k] = v
	}
	sanitizedDimensions = sanitizeDimensions(labels)

	d.Channel() <- buildHPAMetrics(metricName, sanitizedDimensions, value)
}

func sanitizeDimensions(labels *map[string]string) *map[string]string{
	sanitizedDimensions := make(map[string]string)
	for k, v range labels {
		sanitizedDimensions[dimensionSanitizer.Replace(k)] = dimensionSanitizer.Replace(v) 
	}
	return sanitizedDimensions 
}

func (d *HPAMetrics) allContainersAreReady(pod *corev1.Pod) (bool) {
	for _, status := pod.Status.ContainerStatuses {
		if !status.Ready {
			return false
		}
	}
	return true
}

func (d *HPAMetrics) getContainerPort(pod *corev1.Pod, instanceName string) (int32, bool) {
	// Sanitize instance name. The instance name in label is not sanitized, 
	// but The instance name in container name is sanitized.
	instanceName := strings.ToLower(strings.Replace(instanceName, "_", "--", -1))
	// Remove possible trailing hash added by k8s
	instanceName := instanceName[:min(len(instanceName), 45)]
	for _, container := range pod.Spec.Containers {
		if strings.Contains(container.Name, instanceName) {
			return container.Ports[0].ContainerPort, true
		}
	}
	return 0, false
}

func buildHPAMetric(name string, dimensions map[string]string, value float64) (m metric.Metric) {
	m = metric.New(name)
	m.MetricType = metric.Gauge
	m.Value = value
	m.addDimensions(dimensions)
	return m
}

// ParseUWSGIWorkersStats Counts workers status stats from JSON content and returns metrics
func parseHTTPMetrics(raw []byte) (float64, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal(raw, &result)
	if err != nil {
		return 0, err
	}
	utilization, ok := result["utilization"].(string)
	if !ok {
		return 0, fmt.Errorf("\"utilization\" field not found or not a string")
	}
	return strconv.ParseFloat(utilization, 64) err
}

// ParseUWSGIWorkersStats Counts workers status stats from JSON content and returns metrics
func parseUWSGIMetrics(raw []byte) (float64, error) {
	var utilization float64 = 0 
	result := make(map[string]interface{})
	err := json.Unmarshal(raw, &result)
	if err != nil {
		return utilization, err
	}
	workers, ok := result["workers"].([]interface{})
	if !ok {
		return utilization, fmt.Errorf("\"workers\" field not found or not an array")
	}
	activeWorker := 0
	totalWorker := len(workers)
	for _, worker := range workers {
		workerMap, ok := worker.(map[string]interface{})
		if !ok {
			return utilization, fmt.Errorf("worker record is not a map")
		}
		status, ok := workerMap["status"].(string)
		if !ok {
			return utilization, fmt.Errorf("status not found or not a string")
		}
		if status != "idle" {
			activeWorker++
		}
	}
	utilization = float64(activeWorker)/float64(totalWorker)
	return utilization, err
}