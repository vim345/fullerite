package collector

import (
	"bufio"
	"encoding/json"
	"fullerite/metric"
	"log"
	"net"
	"time"
)

const (
	// DiamondCollectorPort is the TCP port that diamond
	// collectors write to and we read off of.
	DiamondCollectorPort = "19191"
)

// Diamond collector type
type Diamond struct {
	interval int
	channel  chan metric.Metric
	incoming chan []byte
}

// NewDiamond creates a new Diamond collector.
func NewDiamond() *Diamond {
	d := new(Diamond)
	d.incoming = make(chan []byte)
	d.channel = make(chan metric.Metric)
	go d.collectDiamond()
	return d
}

// collectDiamond opens up and reads from the a TCP socket and
// writes what it's read to a local channel. Diamond handler (running in
// separate processes) write to the same port.
//
// When Collect() is called it reads from the local channel converts
// strings to metrics and publishes metrics to handlers.
func (d Diamond) collectDiamond() {
	// TODO: we need to make sure that this goroutine is always up
	// and running.

	addr, err := net.ResolveTCPAddr("tcp", ":"+DiamondCollectorPort)
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
func (d Diamond) readDiamondMetrics(conn *net.TCPConn) {
	defer conn.Close()
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second)
	reader := bufio.NewReader(conn)
	log.Println("Diamond collector is starting a new reader...")
	for {
		// TODO: verify that timeout is actually working.
		conn.SetDeadline(time.Now().Add(1e9))
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		log.Println("Read from Diamond collector: " + string(line))
		d.incoming <- line
	}
}

// Collect reads metrics collected from Diamond collectors, converts
// them to fullerite's Metric type and publishes them to handlers.
func (d Diamond) Collect() {
	for line := range d.incoming {
		var metric metric.Metric
		if err := json.Unmarshal(line, &metric); err != nil {
			log.Println("Cannot unmarshal metric line from diamond: " + string(line))
			continue
		}
		metric.AddDimension("diamond", "yes")
		d.Channel() <- metric
	}
}

// Name of the collector.
func (d Diamond) Name() string {
	return "Diamond"
}

// Interval returns the collect rate of the collector.
func (d Diamond) Interval() int {
	return d.interval
}

// Channel returns the internal metrics channel. fullerite reads from
// this channel to pass metrics to the handlers.
func (d Diamond) Channel() chan metric.Metric {
	return d.channel
}

// String returns the collector name in printable format.
func (d Diamond) String() string {
	return d.Name() + "Collector"
}

// SetInterval sets the collect rate of the collector.
func (d *Diamond) SetInterval(interval int) {
	d.interval = interval
}
