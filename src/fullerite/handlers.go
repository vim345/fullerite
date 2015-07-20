package main

import (
	"fullerite/handler"
	"fullerite/metric"
	"log"
)

func startHandlers(c Config) (handlers []handler.Handler) {
	log.Println("Starting handlers...")
	for _, name := range c.Handlers {
		handlers = append(handlers, startHandler(name))
	}
	return handlers
}

func startHandler(name string) handler.Handler {
	log.Println("Starting handler", name)
	handler := handler.New(name)
	readHandlerConfig(handler)
	go handler.Run()
	return handler
}

func readHandlerConfig(handler handler.Handler) {
	// TODO: actually read from configuration file.
	handler.SetInterval(5)
	handler.SetMaxBufferSize(300)
}

func writeToHandlers(handlers []handler.Handler, metric metric.Metric) {
	for _, handler := range handlers {
		handler.Channel() <- metric
	}
}
