package main

import (
	"fullerite/config"
	"fullerite/handler"
	"fullerite/metric"
)

func createHandlers(c config.Config) (handlers []handler.Handler) {
	for name, config := range c.Handlers {
		handlers = append(handlers, createHandler(name, c, config))
	}
	return handlers
}

func createHandler(name string, globalConfig config.Config, instanceConfig map[string]interface{}) handler.Handler {
	handlerInst := handler.New(name)
	if handlerInst == nil {
		return nil
	}

	// apply any global configs
	handlerInst.SetInterval(config.GetAsInt(globalConfig.Interval, handler.DefaultInterval))
	handlerInst.SetPrefix(globalConfig.Prefix)
	handlerInst.SetDefaultDimensions(globalConfig.DefaultDimensions)

	// now apply the handler level configs
	handlerInst.Configure(instanceConfig)

	// now run a listener channel for each collector
	handlerInst.InitListeners(globalConfig)

	return handlerInst
}

func startHandlers(handlers []handler.Handler) {
	log.Info("Starting handlers...")
	for _, handler := range handlers {
		if handler != nil {
			go handler.Run()
		}
	}
}

func writeToHandlers(handlers []handler.Handler, metric metric.Metric) {
	for i := range handlers {
		handlers[i].Channel() <- metric
	}
}
