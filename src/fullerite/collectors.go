package main

import (
	"fullerite/collector"
	"fullerite/metric"
	"log"
)

func startCollectors(c Config) (collectors []collector.Collector) {
	log.Println("Starting collectors...")
	for _, name := range c.Collectors {
		collectors = append(collectors, startCollector(name))
	}
	return collectors
}

func startCollector(name string) collector.Collector {
	log.Println("Starting collector", name)
	collector := collector.New(name)
	return collector
}

func readFromCollectors(collectors []collector.Collector) (metrics []metric.Metric) {
	for _, collector := range collectors {
		for _, metric := range readFromCollector(collector) {
			metrics = append(metrics, metric)
		}
	}
	return metrics
}

func readFromCollector(collector collector.Collector) (metrics []metric.Metric) {
	// TODO: read metrics from collectors Channel
	return metrics
}
