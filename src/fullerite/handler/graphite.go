package handler

import (
	"fmt"
	"fullerite/metric"
	"net"
	"sort"
	"time"

	"github.com/Sirupsen/logrus"
)

// Graphite type
type Graphite struct {
	BaseHandler
	server string
	port   string
}

// NewGraphite returns a new Graphite handler.
func NewGraphite() *Graphite {
	g := new(Graphite)
	g.name = "Graphite"
	g.interval = DefaultInterval
	g.maxBufferSize = DefaultBufferSize
	g.timeout = time.Duration(DefaultTimeoutSec * time.Second)
	g.log = logrus.WithFields(logrus.Fields{"app": "fullerite", "pkg": "handler", "handler": "Graphite"})
	g.channel = make(chan metric.Metric)
	g.emissionTimes = make([]float64, 0)
	return g
}

// Server returns the Graphite server's name or IP
func (g *Graphite) Server() string {
	return g.server
}

// Port returns the Graphite server's port number
func (g *Graphite) Port() string {
	return g.port
}

// Configure accepts the different configuration options for the Graphite handler
func (g *Graphite) Configure(configMap map[string]interface{}) {
	if server, exists := configMap["server"]; exists == true {
		g.server = server.(string)
	} else {
		g.log.Error("There was no server specified for the Graphite Handler, there won't be any emissions")
	}

	if port, exists := configMap["port"]; exists == true {
		g.port = port.(string)
	} else {
		g.log.Error("There was no port specified for the Graphite Handler, there won't be any emissions")
	}
	g.ConfigureCommonParams(&configMap)
}

// Run sends metrics in the channel to the graphite server.
func (g *Graphite) Run() {
	datapoints := make([]string, 0, g.maxBufferSize)

	lastEmission := time.Now()
	lastHandlerMetricsEmission := lastEmission
	for incomingMetric := range g.Channel() {
		datapoint := g.convertToGraphite(incomingMetric)
		g.log.Debug("Graphite datapoint: ", datapoint)
		datapoints = append(datapoints, datapoint)

		emitIntervalPassed := time.Since(lastEmission).Seconds() >= float64(g.interval)
		emitHandlerIntervalPassed := time.Since(lastHandlerMetricsEmission).Seconds() >= float64(g.interval)
		bufferSizeLimitReached := len(datapoints) >= g.maxBufferSize
		doEmit := emitIntervalPassed || bufferSizeLimitReached

		if emitHandlerIntervalPassed {
			lastHandlerMetricsEmission = time.Now()
			m := g.makeEmissionTimeMetric()
			g.resetEmissionTimes()
			m.AddDimension("handler", "Graphite")
			datapoints = append(datapoints, g.convertToGraphite(m))
		}

		if doEmit {
			// emit datapoints
			beforeEmission := time.Now()
			g.emitMetrics(datapoints)
			lastEmission = time.Now()

			emissionTimeInSeconds := lastEmission.Sub(beforeEmission).Seconds()
			g.log.Info("Sending to Graphite took ", emissionTimeInSeconds, " seconds")
			g.emissionTimes = append(g.emissionTimes, emissionTimeInSeconds)

			// reset datapoints
			datapoints = make([]string, 0, g.maxBufferSize)
		}

	}
}

func (g *Graphite) convertToGraphite(incomingMetric metric.Metric) (datapoint string) {
	//orders dimensions so datapoint keeps consistent name
	var keys []string
	dimensions := incomingMetric.GetDimensions(g.DefaultDimensions())
	for k := range dimensions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	datapoint = g.Prefix() + incomingMetric.Name
	for _, key := range keys {
		datapoint = fmt.Sprintf("%s.%s.%s", datapoint, key, dimensions[key])
	}
	datapoint = fmt.Sprintf("%s %f %d\n", datapoint, incomingMetric.Value, time.Now().Unix())
	return datapoint
}

func (g *Graphite) emitMetrics(datapoints []string) {
	g.log.Info("Starting to emit ", len(datapoints), " datapoints")

	if len(datapoints) == 0 {
		g.log.Warn("Skipping send because of an empty payload")
		return
	}

	addr := fmt.Sprintf("%s:%s", g.server, g.port)
	conn, err := net.DialTimeout("tcp", addr, g.timeout)
	if err != nil {
		g.log.Error("Failed to connect ", addr)
	} else {
		for _, datapoint := range datapoints {
			fmt.Fprintf(conn, datapoint)
		}
	}
}
