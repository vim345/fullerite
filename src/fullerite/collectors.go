package main

import (
	"fullerite/collector"
	"fullerite/metric"
	"log"
	"strconv"
	"time"
)

func startCollectors(c Config) (collectors []collector.Collector) {
	log.Println("Starting collectors...")

	for name, config := range c.Collectors {
		newCollector := collector.New(name)

		// try and get the config for the interval
		intervalStr, exists := config["interval"]
		var interval int64
		if exists {
			var err error
			interval, err = strconv.ParseInt(intervalStr, 10, 32)
			if err != nil {
				log.Println("Non int value specified for interval in collector", name)
				interval = collector.DefaultCollectionInterval
			}
		} else {
			interval = collector.DefaultCollectionInterval
		}
		newCollector.SetInterval(interval)

		// now let them config themselves
		newCollector.Configure(&config)

		collectors = append(collectors, newCollector)

		go runCollector(&newCollector)
	}
	return collectors
}

func runCollector(inCollector *collector.Collector) {
	for {
		log.Println("Collecting from", inCollector)
		(*inCollector).Collect()
		time.Sleep(time.Duration((*inCollector).Interval()) * time.Second)
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
