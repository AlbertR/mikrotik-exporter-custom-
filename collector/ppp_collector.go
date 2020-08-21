package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2/proto"
	"strings"
)

type PPPInterfaceCollector struct {
	props       []string
	description *prometheus.Desc
}

func newPPPInterfacesCollector() routerOSCollector {
	c := &PPPInterfaceCollector{}
	c.init()
	return c
}

func (c *PPPInterfaceCollector) init() {
	c.props = []string{"name", "local-address", "remote-address", "address-list", "dns-server"}
	labelNames := []string{"name", "profilename", "localaddress", "remoteaddress", "addresslist", "dnsserver"}
	c.description = description("ppp_profiles", "metrics", "ppp profiles metrics", labelNames)
}

func (c *PPPInterfaceCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.description
}

func (c *PPPInterfaceCollector) collect(ctx *collectorContext) error {
	stats, err := c.fetch(ctx)
	if err != nil {
		return err
	}

	for _, re := range stats {
		c.collectMetrics(ctx, re)
	}

	return nil
}

func (c *PPPInterfaceCollector) fetch(ctx *collectorContext) ([]*proto.Sentence, error) {
	reply, err := ctx.client.Run("/ppp/profile/print", "=.proplist="+strings.Join(c.props, ","))
	if err != nil {
		log.WithFields(log.Fields{
			"device": ctx.device.Name,
			"error":  err,
		}).Error("error fetching ppp profiles metrics")
		return nil, err
	}
	return reply.Re, nil
}

func (c *PPPInterfaceCollector) collectMetrics(ctx *collectorContext, re *proto.Sentence) {
	v := 0.0

	profilename := re.Map["name"]
	localaddress := re.Map["local-address"]
	remoteaddress := re.Map["remote-address"]
	addresslist := re.Map["address-list"]
	dnsserver := re.Map["dns-server"]

	ctx.ch <- prometheus.MustNewConstMetric(c.description, prometheus.GaugeValue, v, ctx.device.Name,
		profilename, localaddress, remoteaddress, addresslist, dnsserver)
}
