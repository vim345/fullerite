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
	version = "0.0.2"
	desc    = "Diamond compatible metrics collector"
)

var log = logrus.WithFields(logrus.Fields{"app": "fullerite"})

func init_logrus(ctx *cli.Context) {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC822,
		FullTimestamp:   true,
	})

	if level, err := logrus.ParseLevel(ctx.String("log_level")); err == nil {
		logrus.SetLevel(level)
	} else {
		log.Error(err)
		logrus.SetLevel(logrus.InfoLevel)
	}

	filename := ctx.String("log_file")
	logrus.SetOutput(os.Stderr)
	if filename != "" {
		var f *os.File
		_, err := os.Stat(filename)
		if !os.IsNotExist(err) {
			os.Rename(filename, filename+".prev")
		}
		f, err = os.Create(filename)
		if err != nil {
			log.Error("Cannot create log file ", err)
			log.Warning("Continuing to log to stderr")
		} else {
			logrus.SetOutput(f)
		}
	}
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
		cli.StringFlag{
			Name:  "log_file",
			Value: "",
			Usage: "Log to file",
		},
	}
	app.Action = start
	app.Run(os.Args)
}

func start(ctx *cli.Context) {
	init_logrus(ctx)
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
