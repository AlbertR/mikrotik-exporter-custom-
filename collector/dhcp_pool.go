package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2/proto"
	"strings"
)

type dhcpPoolCollector struct {
	props        []string
	descriptions *prometheus.Desc
}

func NewDHCPPCollector() routerOSCollector {
	c := &dhcpPoolCollector{}
	c.init()
	return c
}

func (c *dhcpPoolCollector) init() {
	c.props = []string{"name", "next-pool", "ranges"}
	labelNames := []string{"name", "address", "poolname", "nextpool", "ranges"}
	c.descriptions = description("dhcp", "pool_metrics", "dhcp pools metrics", labelNames)
}

func (c *dhcpPoolCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.descriptions
}

func (c *dhcpPoolCollector) collect(ctx *collectorContext) error {
	stat, err := c.fetch(ctx)
	if err != nil {
		return err
	}

	for _, re := range stat {
		c.collectMetrics(ctx, re)
	}
	return nil
}

func (c *dhcpPoolCollector) fetch(ctx *collectorContext) ([]*proto.Sentence, error) {
	reply, err := ctx.client.Run("/ip/pool/print", "=.proplist="+strings.Join(c.props, ","))
	if err != nil {
		log.WithFields(log.Fields{
			"device": ctx.device.Name,
			"error":  err,
		}).Error("error fetch dhcp pool metrics")
		return nil, err
	}

	return reply.Re, nil
}

func (c *dhcpPoolCollector) collectMetrics(ctx *collectorContext, re *proto.Sentence) {

	v := 0.0

	poolname := re.Map["name"]
	nextpool := re.Map["next-pool"]
	ranges := re.Map["ranges"]

	ctx.ch <- prometheus.MustNewConstMetric(c.descriptions, prometheus.GaugeValue, v, ctx.device.Name,
		ctx.device.Address, poolname, nextpool, ranges)
}
