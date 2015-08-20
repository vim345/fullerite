package collector

import (
	"fullerite/config"
	"fullerite/metric"

	"bufio"
	"encoding/json"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	// DefaultDiamondCollectorPort is the TCP port that diamond
	// collectors write to and we read off of.
	DefaultDiamondCollectorPort = "19191"
)

// Diamond collector type
type Diamond struct {
	BaseCollector
	port     string
	incoming chan []byte
}

// NewDiamond creates a new Diamond collector.
func NewDiamond() *Diamond {
	d := new(Diamond)
	d.log = logrus.WithFields(logrus.Fields{"app": "fullerite", "pkg": "collector", "collector": "Diamond"})
	d.incoming = make(chan []byte)
	d.channel = make(chan metric.Metric)
	d.port = DefaultDiamondCollectorPort
	d.interval = DefaultCollectionInterval
	go d.collectDiamond()
	return d
}

// Configure the collector
func (d *Diamond) Configure(configMap map[string]interface{}) {
	if port, exists := configMap["port"]; exists == true {
		d.port = port.(string)
	}
	if interval, exists := configMap["interval"]; exists == true {
		d.interval = config.GetAsInt(interval, DefaultCollectionInterval)
	}
}

// collectDiamond opens up and reads from the a TCP socket and
// writes what it's read to a local channel. Diamond handler (running in
// separate processes) write to the same port.
//
// When Collect() is called it reads from the local channel converts
// strings to metrics and publishes metrics to handlers.
func (d Diamond) collectDiamond() {
	addr, err := net.ResolveTCPAddr("tcp", ":"+d.port)

	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		d.log.Fatal("Cannot listen on diamond socket", err)
	}
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			d.log.Fatal(err)
		}
		go d.readDiamondMetrics(conn)
	}
}

// readDiamondMetrics reads from the connection
func (d *Diamond) readDiamondMetrics(conn *net.TCPConn) {
	defer conn.Close()
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second)
	reader := bufio.NewReader(conn)
	d.log.Info("Connection started: ", conn.RemoteAddr())
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			d.log.Warn("Error while reading diamond metrics", err)
			break
		}
		d.log.Debug("Read: ", string(line))
		d.incoming <- line
	}
	d.log.Info("Connection closed: ", conn.RemoteAddr())
}

// Collect reads metrics collected from Diamond collectors, converts
// them to fullerite's Metric type and publishes them to handlers.
func (d *Diamond) Collect() {
	for line := range d.incoming {
		var metric metric.Metric
		if err := json.Unmarshal(line, &metric); err != nil {
			d.log.Error("Cannot unmarshal metric line from diamond:", line)
			continue
		}
		metric.AddDimension("diamond", "yes")
		d.Channel() <- metric
	}
}
