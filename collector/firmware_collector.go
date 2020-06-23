package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2/proto"
	"strings"
)

type firmwareCollector struct {
	props       []string
	description *prometheus.Desc
}

func (c *firmwareCollector) init() {
	c.props = []string{"board-name", "model", "serial-number", "current-firmware", "upgrade-firmware"}

	labelNames := []string{"name", "boardname", "model", "serialnumber", "currentfirmware", "upgradefirmware"}
	c.description = description("firmware", "metrics", "firmware metrics", labelNames)
}

func newFirmwareCollector() routerOSCollector {
	c := &firmwareCollector{}
	c.init()
	return c
}

func (c *firmwareCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.description
}

func (c *firmwareCollector) collect(ctx *collectorContext) error {
	stats, err := c.fetch(ctx)
	if err != nil {
		return err
	}

	for _, re := range stats {
		c.collectMetrics(ctx, re)
	}

	return nil
}

func (c *firmwareCollector) fetch(ctx *collectorContext) ([]*proto.Sentence, error) {
	reply, err := ctx.client.Run("/system/routerboard/print", "=.proplist="+strings.Join(c.props, ","))
	if err != nil {
		log.WithFields(log.Fields{
			"device": ctx.device.Name,
			"error":  err,
		}).Error("error fetching firmware metrics")
		return nil, err
	}

	return reply.Re, nil
}

func (c *firmwareCollector) collectMetrics(ctx *collectorContext, re *proto.Sentence) {
	v := 0.0

	boardname := re.Map["board-name"]
	model := re.Map["model"]
	serialnumber := re.Map["serial-number"]
	currentfirmware := re.Map["current-firmware"]
	upgradefirmware := re.Map["upgrade-firmware"]

	ctx.ch <- prometheus.MustNewConstMetric(c.description, prometheus.GaugeValue, v, ctx.device.Name,
		boardname, model, serialnumber, currentfirmware, upgradefirmware)
}
