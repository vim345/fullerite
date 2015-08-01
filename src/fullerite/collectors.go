package main

import (
	"fullerite/collector"
	"fullerite/metric"
	"time"
)

func startCollectors(c Config) (collectors []collector.Collector) {
	log.Info("Starting collectors...")
	for name, config := range c.Collectors {
		collectors = append(collectors, startCollector(name, config))
	}
	return collectors
}

func startCollector(name string, config map[string]string) collector.Collector {
	log.Debug("Starting collector ", name)
	collector := collector.New(name)
	collector.Configure(config)
	go runCollector(collector)
	return collector
}

func runCollector(collector collector.Collector) {
	log.Info("Running ", collector)
	for {
		collector.Collect()
		time.Sleep(time.Duration(collector.Interval()) * time.Second)
	}
}

func readFromCollectors(collectors []collector.Collector, metrics chan metric.Metric) {
	for _, collector := range collectors {
		go readFromCollector(collector, metrics)
	}
}

func readFromCollector(collector collector.Collector, metrics chan metric.Metric) {
	for metric := range collector.Channel() {
		metrics <- metric
	}
}
