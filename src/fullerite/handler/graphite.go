package handler

import (
	"fullerite/metric"
	"time"
	"fmt"
	"net"
)

// Graphite type
type Graphite struct {
	BaseHandler
	server string
	port string
}

// NewGraphite returns a new Graphite handler.
func NewGraphite() *Graphite {
	g := new(Graphite)
	g.name = "Graphite"
	g.maxBufferSize = DefaultBufferSize
	g.channel = make(chan metric.Metric)
	return g
}

// Configure : accepts the different configuration options for the graphite handler
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


	for metric := range g.channel {
		log.Println("Sending metric to Graphite:", metric)
		datapoints := g.convertToGraphite(&metric)

		//if the datapoints from metric would overflow the buffer, flush it and then add the new datapoints
		if time.Since(lastEmission).Seconds() >= float64(g.interval) || len(metrics) + len(*datapoints) > g.maxBufferSize {
			g.emitMetrics(metrics)
			lastEmission = time.Now()
			metrics = make([]string, 0, g.maxBufferSize)
		}
		metrics = append(metrics, *datapoints...)
	}

}

func (g *Graphite) convertToGraphite(metric *metric.Metric) *[]string{
	outname := g.Prefix() + (*metric).Name
	dimensions := metric.GetDimensions(g.DefaultDimensions())
	datapoints := make([]string, 0, len(dimensions) + 1)
	// for key in dimensions, generate a new metric data point, add to a list, return
	//what timestamp to use?
	datapoints = append(datapoints, fmt.Sprintf("%s %f %s\n", outname, metric.Value, time.Now())) //find out what time to use

	for key, value := range dimensions {
		//create a list of datapoints for this metric, then append that list the a global list
		datapoints = append(datapoints, fmt.Sprintf("%s.%s %f %s\n", outname, key, value, time.Now()))
	}

	return &datapoints
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
	}
}


















