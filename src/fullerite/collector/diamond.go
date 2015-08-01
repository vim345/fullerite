package collector

import (
	"bufio"
	"encoding/json"
	"fullerite/metric"
	"net"
	"time"
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
	d.incoming = make(chan []byte)
	d.channel = make(chan metric.Metric)
	d.port = DefaultDiamondCollectorPort
	d.interval = DefaultCollectionInterval
	go d.collectDiamond()
	return d
}

// Configure the collector
func (d *Diamond) Configure(config map[string]interface{}) {
	if port, exists := config["port"]; exists == true {
		d.port = port.(string)
	}
	if interval, exists := config["interval"]; exists == true {
		d.interval = int64(interval.(float64))
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
		log.Fatal("Cannot listen on diamond socket", err)
	}
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Fatal(err)
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
	log.Info("Diamond collector connection started: ", conn.RemoteAddr())
	for {
		// TODO: verify that timeout is actually working.
		conn.SetDeadline(time.Now().Add(1e9))
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		log.Debug("Read from Diamond collector: ", string(line))
		d.incoming <- line
	}
	log.Info("Diamond collector connection closed: ", conn.RemoteAddr())
}

// Collect reads metrics collected from Diamond collectors, converts
// them to fullerite's Metric type and publishes them to handlers.
func (d *Diamond) Collect() {
	for line := range d.incoming {
		var metric metric.Metric
		if err := json.Unmarshal(line, &metric); err != nil {
			log.Error("Cannot unmarshal metric line from diamond:", line)
			continue
		}
		metric.AddDimension("diamond", "yes")
		d.Channel() <- metric
	}
}
