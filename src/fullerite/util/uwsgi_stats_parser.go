package util

import (
	"fullerite/metric"

	"encoding/json"
	"fmt"
	"strings"
)

// ParseUWSGIWorkersStats Counts workers status stats from JSON content and returns metrics
func ParseUWSGIWorkersStats(raw []byte) ([]metric.Metric, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal(raw, &result)
	results := []metric.Metric{}
	if err != nil {
		return results, err
	}
	registry := make(map[string]int)
	registry["IdleWorkers"] = 0
	registry["BusyWorkers"] = 0

	workers, ok := result["workers"].([]interface{})
	if !ok {
		return results, fmt.Errorf("\"workers\" field not found or not an array")
	}
	for _, worker := range workers {
		workerMap, ok := worker.(map[string]interface{})
		if !ok {
			return results, fmt.Errorf("worker record is not a map")
		}
		status, ok := workerMap["status"].(string)
		if !ok {
			return results, fmt.Errorf("status not found or not a string")
		}
		// Consider all status starting by sig as just "sig"
		if strings.Index(status, "sig") == 0 {
			status = "sig"
		}
		// Capitalize the metric name (busy => BusyWorkers)
		metricName := strings.Title(status) + "Workers"
		_, exists := registry[metricName]
		if !exists {
			registry[metricName] = 0
		}
		registry[metricName]++
	}
	for key, value := range registry {
		results = append(results, metric.WithValue(key, float64(value)))
	}
	return results, err
}
