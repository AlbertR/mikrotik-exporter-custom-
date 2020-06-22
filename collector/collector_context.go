package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	routeros "gopkg.in/routeros.v2"
	"mikrotik_exporter/config"
)

type collectorContext struct {
	ch     chan<- prometheus.Metric
	device *config.Device
	client *routeros.Client
}
