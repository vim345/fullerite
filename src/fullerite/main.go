package main

import (
	"fullerite/metric"

	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const (
	name    = "fullerite"
	version = "0.0.1"
	desc    = "Diamond compatible metrics collector"
)

var log = logrus.WithFields(logrus.Fields{"app": "fullerite"})

func init() {
	// Output to stderr instead of stdout, could also be a file.
	logrus.SetOutput(os.Stderr)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC822,
		FullTimestamp:   true,
	})
}

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
		cli.StringFlag{
			Name:  "log_level, l",
			Value: "info",
			Usage: "Logging level (debug, info, warn, error, fatal, panic)",
		},
	}
	app.Action = start
	app.Run(os.Args)
}

func start(ctx *cli.Context) {
	if level, err := logrus.ParseLevel(ctx.String("log_level")); err == nil {
		logrus.SetLevel(level)
	} else {
		log.Error(err)
		logrus.SetLevel(logrus.InfoLevel)
	}
	log.Info("Starting fullerite...")

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
