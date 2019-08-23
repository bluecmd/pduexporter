package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	g "github.com/soniah/gosnmp"
)

const (
	OidCurrent      = ".1.3.6.1.4.1.17420.1.2.9.1.11.0"
	OidFwVersion    = ".1.3.6.1.4.1.17420.1.2.4.0"
	OidModelNo      = ".1.3.6.1.4.1.17420.1.2.9.1.19.0"
	OidOutletConfig = ".1.3.6.1.4.1.17420.1.2.9.1.14"
	OidOutletStatus = ".1.3.6.1.4.1.17420.1.2.9.1.13.0"
	OidSysLocation  = ".1.3.6.1.2.1.1.6.0"
	OidSysName      = ".1.3.6.1.2.1.1.5.0"
	OidVoltage      = ".1.3.6.1.4.1.17420.1.2.9.1.17.0"
)

var (
	target          = flag.String("host", "127.0.0.1", "ip of pdu")
	community       = flag.String("community", "public", "SNMP community")
	listen          = flag.String("listen", ":9819", "Address to bind to")
	pduSystem       = prometheus.NewDesc("pdu_system", "PDU System tracker", []string{"model", "firmware", "location", "name"}, nil)
	pduOutletStatus = prometheus.NewDesc("pdu_outlet_status", "PDU Outlet boolean status", []string{"name", "index"}, nil)
	pduVoltage      = prometheus.NewDesc("pdu_voltage", "PDU Voltage in Volts", []string{}, nil)
	pduCurrent      = prometheus.NewDesc("pdu_current", "PDU Total Current in Amperes", []string{}, nil)
)

type collector struct{}

func main() {
	flag.Parse()

	g.Default.Target = *target
	g.Default.Community = *community
	g.Default.Version = g.Version1

	err := g.Default.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer g.Default.Conn.Close()

	cmd := &collector{}
	prometheus.MustRegister(cmd)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(*listen, nil); err != nil {
		log.Fatalf("http.ListenAndServe: %v", err)
	}
}

func (_ *collector) Describe(c chan<- *prometheus.Desc) {
	c <- pduSystem
	c <- pduOutletStatus
	c <- pduVoltage
	c <- pduCurrent
}

func (_ *collector) Collect(c chan<- prometheus.Metric) {
	oids := []string{OidFwVersion, OidModelNo, OidSysLocation, OidSysName, OidVoltage, OidOutletStatus, OidCurrent}

	outletMap := []string{}
	g.Default.Walk(OidOutletConfig, func(v g.SnmpPDU) error {
		s := string(v.Value.([]byte))
		name := strings.Split(s, ",")[0]
		outletMap = append(outletMap, name)
		return nil
	})

	fw := "unknown"
	model := "unknown"
	location := "unknown"
	name := "unknown"

	// These PDUs are implemented in the most minimal way possible and
	// have problems doing Get() with more than 3 OIDs.
	for _, o := range oids {
		result, err := g.Default.Get([]string{o})
		if err != nil {
			log.Fatalf("snmp Get() err: %v", err)
		}
		v := result.Variables[0]
		switch o {
		case OidFwVersion:
			fw = string(v.Value.([]byte))
		case OidModelNo:
			model = string(v.Value.([]byte))
		case OidSysLocation:
			location = string(v.Value.([]byte))
		case OidSysName:
			name = string(v.Value.([]byte))
		case OidVoltage:
			voltage := float64(g.ToBigInt(v.Value).Uint64())
			c <- prometheus.MustNewConstMetric(pduVoltage, prometheus.GaugeValue, voltage)
		case OidCurrent:
			// Current is scaled by *10
			current := float64(g.ToBigInt(v.Value).Uint64()) / 10.0
			c <- prometheus.MustNewConstMetric(pduCurrent, prometheus.GaugeValue, current)
		case OidOutletStatus:
			s := string(v.Value.([]byte))
			outlets := strings.Split(s, ",")
			for i, status := range outlets {
				val := 0.0
				if status == "1" {
					val = 1.0
				}
				outletId := strconv.FormatInt(int64(i), 10)
				c <- prometheus.MustNewConstMetric(pduOutletStatus, prometheus.GaugeValue, val, outletMap[i], outletId)
			}
		}
	}

	c <- prometheus.MustNewConstMetric(pduSystem, prometheus.CounterValue, 1.0, model, fw, location, name)
}
