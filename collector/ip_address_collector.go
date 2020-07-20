package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2/proto"
	"strings"
)

type IPAddressCollector struct {
	props       []string
	description *prometheus.Desc
}

func newIPAddressCollector() routerOSCollector {
	c := &IPAddressCollector{}
	c.init()
	return c
}

func (c *IPAddressCollector) init() {
	c.props = []string{"address", "interface", "netmask", "network"}
	labelNames := []string{"name", "address", "interface", "mask", "network"}
	c.description = description("ipaddress", "metrics", "ip addresses metrics", labelNames)
}

func (c *IPAddressCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.description
}

func (c *IPAddressCollector) collect(ctx *collectorContext) error {
	stats, err := c.fetch(ctx)
	if err != nil {
		return err
	}

	for _, re := range stats {
		c.collectMetrics(ctx, re)
	}

	return nil
}

func (c *IPAddressCollector) fetch(ctx *collectorContext) ([]*proto.Sentence, error) {
	reply, err := ctx.client.Run("/ip/address/print", ".=proplist="+strings.Join(c.props, ","))
	if err != nil {
		log.WithFields(log.Fields{
			"device": ctx.device.Name,
			"error":  err,
		}).Error("error fetching ip address metrics")
		return nil, err
	}

	return reply.Re, nil
}

func (c *IPAddressCollector) collectMetrics(ctx *collectorContext, re *proto.Sentence) {
	address := re.Map["address"]
	ipinterface := re.Map["interface"]
	mask := re.Map["netmask"]
	network := re.Map["network"]

	v := 0.0

	ctx.ch <- prometheus.MustNewConstMetric(c.description, prometheus.GaugeValue, v, ctx.device.Name, address,
		ipinterface, mask, network)
}
