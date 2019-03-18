package collector

import (
	"fullerite/metric"

	l "github.com/Sirupsen/logrus"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

const (
	KUBECONFIG="/nail/home/sagarp/.kube/config"
)


type K8sStats struct {
	baseCollector
	client *kubernetes.Clientset
}

func init() {
	RegisterCollector("K8sStats", newK8sStats)
}


func handleError(err error) {
    if err  != nil {
        panic(err)
    }
}


// newK8sStats Simple constructor to set properties for the embedded baseCollector.
func newK8sStats(channel chan metric.Metric, intialInterval int, log *l.Entry) Collector {
	m := new(K8sStats)

	m.log = log
	m.channel = channel
	m.interval = intialInterval
	m.name = "K8sStats"
	config, err := clientcmd.BuildConfigFromFlags("", KUBECONFIG)
    handleError(err)

    // creates the clientset
    clientset, err1 := kubernetes.NewForConfig(config)
	handleError(err1)

	m.client = clientset

	return m
}

// Configure Override *baseCollector.Configure().
func (m *K8sStats) Configure(configMap map[string]interface{}) {
	m.configureCommonParams(configMap)
}

// Collect Compares box IP against leader IP and if true, sends data.
func (m *K8sStats) Collect() {
	go m.sendK8sMetrics()
}

// sendMetrics Send to baseCollector channel.
func (m *K8sStats) sendK8sMetrics() {
	for k, v := range getK8sMetrics(m) {
		s := buildK8sMetric(k, float64(v))
		m.Channel() <- s
	}
}


func getK8sMetrics(m *K8sStats) map[string]int64 {
	// Query k8s API to get the metrics necessary
	output, err := m.client.CoreV1().Nodes().List(metav1.ListOptions{})
	handleError(err)

	totalCpus := resource.Quantity{}
	totalMem := resource.Quantity{}
	totalStorage := resource.Quantity{}
	freeCpus := resource.Quantity{}
	freeMem := resource.Quantity{}
	freeStorage := resource.Quantity{}

	for _, item := range output.Items {
		freeCpus.Add(item.Status.Allocatable["cpu"])
		freeMem.Add(item.Status.Allocatable["memory"])
		freeStorage.Add(item.Status.Allocatable["ephemeral-storage"])

		totalCpus.Add(item.Status.Capacity["cpu"])
		totalMem.Add(item.Status.Capacity["memory"])
		totalStorage.Add(item.Status.Capacity["ephemeral-storage"])

	}

	metrics := make(map[string]int64, 0)
	metrics["free_cpus"] = freeCpus.Value()
	metrics["free_cpus"] = freeMem.Value()
	metrics["free_disk"] = freeStorage.Value()
	metrics["total_cpus"] = totalCpus.Value()
	metrics["total_cpus"] = totalMem.Value()
	metrics["total_disk"] = totalStorage.Value()

	return metrics
}

// buildMetric Build a fullerite metric.
func buildK8sMetric(k string, v float64) metric.Metric {
	m := metric.New("k8s." + k)
	m.Value = v
	m.type = "guage"
	return m
}
