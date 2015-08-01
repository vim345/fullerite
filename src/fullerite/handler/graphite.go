package handler

import (
	"fmt"
	"fullerite/metric"
	"net"
	"sort"
	"time"
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
	g.channel = make(chan metric.Metric)
	return g
}

// Configure accepts the different configuration options for the Graphite handler
func (g *Graphite) Configure(config *map[string]string) {
	asmap := *config
	var exists bool
	g.server, exists = asmap["server"]
	if !exists {
		log.Println("There was no server specified for the Graphite Handler, there won't be any emissions")
	}

	g.port, exists = asmap["port"]
	if !exists {
		log.Println("There was no port specified for the Graphite Handler, there won't be any emissions")
	}
}

// Run sends metrics in the channel to the graphite server.
func (g *Graphite) Run() {
	lastEmission := time.Now()
	metrics := make([]string, 0, g.maxBufferSize)
	log.Info("graphite handler started")

	for metric := range g.Channel() {
		log.Println("Sending metric to Graphite:", metric)
		datapoint := g.convertToGraphite(&metric)

		metrics = append(metrics, datapoint)

		//if the datapoints from metric would overflow the buffer, flush it and then add the new datapoints
		if time.Since(lastEmission).Seconds() >= float64(g.interval) || len(metrics) >= g.maxBufferSize {
			g.emitMetrics(metrics)
			lastEmission = time.Now()
			metrics = make([]string, 0, g.maxBufferSize)
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
	log.Info("Starting to emit ", len(datapoints), " datapoints")

	if len(datapoints) == 0 {
		log.Warn("Skipping send because of an empty payload")
		return
	}

	conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%s", g.server, g.port))
	for _, datapoint := range datapoints {
		fmt.Fprintf(conn, datapoint)
		fmt.Println(datapoint)
	}
}
