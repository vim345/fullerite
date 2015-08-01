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
	g.maxBufferSize = DefaultBufferSize
	g.log = logrus.WithFields(logrus.Fields{"app": "fullerite", "pkg": "handler", "handler": "Graphite"})
	g.channel = make(chan metric.Metric)
	return g
}

// Configure accepts the different configuration options for the Graphite handler
func (g *Graphite) Configure(config map[string]interface{}) {
	if server, exists := config["server"]; exists == true {
		g.server = server.(string)
	} else {
		g.log.Error("There was no server specified for the Graphite Handler, there won't be any emissions")
	}
	if port, exists := config["port"]; exists == true {
		g.port = port.(string)
	} else {
		g.log.Error("There was no port specified for the Graphite Handler, there won't be any emissions")
	}
}

// Run sends metrics in the channel to the graphite server.
func (g *Graphite) Run() {
	datapoints := make([]string, 0, g.maxBufferSize)

	lastEmission := time.Now()
	for incomingMetric := range g.Channel() {
		datapoint := g.convertToGraphite(&incomingMetric)
		g.log.Debug("Graphite datapoint: ", datapoint)
		datapoints = append(datapoints, datapoint)
		if time.Since(lastEmission).Seconds() >= float64(g.interval) || len(datapoints) >= g.maxBufferSize {
			g.emitMetrics(datapoints)
			lastEmission = time.Now()
			datapoints = make([]string, 0, g.maxBufferSize)
		}
	}
}

func (g *Graphite) convertToGraphite(metric *metric.Metric) (datapoint string) {
	datapoint = g.Prefix() + (*metric).Name
	dimensions := metric.GetDimensions(g.DefaultDimensions())

	//orders keys so datapoint keeps consistent name
	var keys []string
	for k := range dimensions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		//create a list of datapoints for this metric, then append that list the a global list
		datapoint = fmt.Sprintf("%s.%s.%s", datapoint, key, dimensions[key])
	}
	datapoint = fmt.Sprintf("%s %f %d\n", datapoint, metric.Value, time.Now().Unix())
	return datapoint
}

func (g *Graphite) emitMetrics(datapoints []string) {
	g.log.Info("Starting to emit ", len(datapoints), " datapoints")

	if len(datapoints) == 0 {
		g.log.Warn("Skipping send because of an empty payload")
		return
	}

	conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%s", g.server, g.port))
	for _, datapoint := range datapoints {
		fmt.Fprintf(conn, datapoint)
	}
}
