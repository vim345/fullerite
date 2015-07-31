package main

import (
	"fullerite/handler"
	"fullerite/metric"
	"log"
)

func startHandlers(c Config) (handlers []handler.Handler) {
	log.Println("Starting handlers...")

	for name, config := range c.Handlers {
		handler := handler.New(name)

		// apply any global configs
		handler.SetInterval(c.Interval)
		handler.SetPrefix(c.Prefix)
		handler.SetDefaultDimensions(&c.DefaultDimensions)

		// now apply the handler level configs
		handler.Configure(&config)

		handlers = append(handlers, handler)

		go handler.Run()
	}
	return handlers
}

func writeToHandlers(handlers []handler.Handler, metric metric.Metric) {
	for _, handler := range handlers {
		handler.Channel() <- metric
	}
}
