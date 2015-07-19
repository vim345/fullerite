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
	return handler
}

func writeToHandlers(handlers []handler.Handler, metrics []metric.Metric) {
	for _, handler := range handlers {
		// TODO: create a goroutine for each handler
		writeToHandler(handler)
	}
}

func writeToHandler(handler handler.Handler) {
	// TODO: write to handler's Channel
}
