package main

import (
	"fullerite/metric"
	"github.com/codegangsta/cli"
	"os"
)

const (
	name    = "fullerite"
	version = "0.0.1"
	desc    = "Diamond compatible metrics collector"
)

func main() {
	app := cli.NewApp()
	app.Name = name
	app.Version = version
	app.Usage = desc
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "/etc/fullerite.conf",
			Usage: "JSON formatted configuration file",
		},
	}
	app.Action = start
	app.Run(os.Args)
}

func start(ctx *cli.Context) {
	c := readConfig(ctx.String("config"))
	collectors := startCollectors(c)
	handlers := startHandlers(c)
	metrics := make(chan metric.Metric)
	readFromCollectors(collectors, metrics)
	for metric := range metrics {
		// Writing to handlers' channels. Sending metrics is
		// handled asynchronously in handlers' Run functions.
		writeToHandlers(handlers, metric)
	}
}
