package main

import (
	"fullerite/handler"
	"fullerite/metric"
	"log"
)

func startHandlers(c Config) (handlers []handler.Handler) {
	log.Println("Starting handlers...")

	defaults := convertToDimensions(&c.DefaultDimensions)

	for name, config := range c.Handlers {
		handler := buildHandler(name)

		// apply any global configs
		handler.SetPrefix(c.Prefix)
		handler.SetDefaultDimensions(&defaults)

		// now apply the handler level configs
		handler.Configure(&config)

		handlers = append(handlers, handler)

		go handler.Run()
	}
	return handlers
}

func convertToDimensions(dimsAsMap *map[string]string) []metric.Dimension {
	defaults := make([]metric.Dimension, 0, len(*dimsAsMap))
	for key, value := range *dimsAsMap {
		dim := metric.Dimension{}
		dim.SetName(key)
		dim.SetValue(value)
		defaults = append(defaults, dim)
	}
	return defaults
}

func buildHandler(name string) handler.Handler {
	log.Println("Building handler", name)
	handler := handler.New(name)
	return handler
}

func writeToHandlers(handlers []handler.Handler, metric metric.Metric) {
	for _, handler := range handlers {
		handler.Channel() <- metric
	}
}
