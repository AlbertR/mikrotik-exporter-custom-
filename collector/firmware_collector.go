package collector

import "github.com/prometheus/client_golang/prometheus"

type firmwareCollector struct {
	props        []string
	descriptions *prometheus.Desc
}

func (c *firmwareCollector) init() {
	c.props = []string{"board-name", "model", "serial-number", "current-firmware", "upgrade-firmware"}


}
