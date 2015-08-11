package main

import (
	"fullerite/config"
	"fullerite/handler"
	"fullerite/metric"
)

func startHandlers(c config.Config) (handlers []handler.Handler) {
	log.Info("Starting handlers...")
	for name, config := range c.Handlers {
		handlers = append(handlers, startHandler(name, c, config))
	}
	return handlers
}

func startHandler(name string, globalConfig config.Config, config map[string]interface{}) handler.Handler {
	log.Debug("Starting handler ", name)
	handler := handler.New(name)

	// apply any global configs
	handler.SetInterval(globalConfig.Interval)
	handler.SetPrefix(globalConfig.Prefix)
	handler.SetDefaultDimensions(globalConfig.DefaultDimensions)

	// now apply the handler level configs
	handler.Configure(config)

	go handler.Run()
	return handler
}

func writeToHandlers(handlers []handler.Handler, metric metric.Metric) {
	for _, handler := range handlers {
		handler.Channel() <- metric
	}
}
