package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2/proto"
	"strings"
)

type routeCollector struct {
	props       []string
	description *prometheus.Desc
}

func newRoutesCollector() routerOSCollector {
	c := &routeCollector{}
	c.init()
	return c
}

func (c *routeCollector) init() {
	c.props = []string{"dst-address", "gateway", "distance", "pref-src"}
	labelNames := []string{"name", "dst_address", "gateway", "distance", "pref_src"}
	c.description = description("route", "metrics", "ip route metrics", labelNames)
}

func (c *routeCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.description
}

func (c *routeCollector) collect(ctx *collectorContext) error {
	stats, err := c.fetch(ctx)
	if err != nil {
		return err
	}

	for _, re := range stats {
		c.collectMetrics(ctx, re)
	}

	return nil
}

func (c *routeCollector) fetch(ctx *collectorContext) ([]*proto.Sentence, error) {
	reply, err := ctx.client.Run("/ip/route/print", "=.proplist="+strings.Join(c.props, ","))
	if err != nil {
		log.WithFields(log.Fields{
			"device": ctx.device.Name,
			"error":  err,
		}).Error("error fetching IP route metrics")
		return nil, err
	}

	return reply.Re, nil
}

func (c *routeCollector) collectMetrics(ctx *collectorContext, re *proto.Sentence) {
	dstaddress := re.Map["dst-address"]
	gateway := re.Map["gateway"]
	distance := re.Map["distance"]
	prefsource := re.Map["pref-src"]

	v := 0.0

	ctx.ch <- prometheus.MustNewConstMetric(c.description, prometheus.GaugeValue, v, ctx.device.Name, dstaddress,
		gateway, distance, prefsource)
}
