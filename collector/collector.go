package collector

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2"
	"io"
	"mikrotik_exporter/config"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	namespace      = "mikrotik"
	apiPort        = "8728"
	apiPortTLS     = "8729"
	dnsPort        = 53
	DefaultTimeout = 5 * time.Second
)

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"mikrotik_exporter: duration of a collector scrape",
		[]string{"device"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"mikrotik_exporter: whether a collector succeeded",
		[]string{"device"},
		nil,
	)
)

type collector struct {
	devices     []config.Device
	collectors  []routerOSCollector
	timeout     time.Duration
	enableTLS   bool
	insecureTLS bool
}

// WithBGP enables BGP routing metrics
func WithBGP() Option {
	return func(c *collector) {
		c.collectors = append(c.collectors, newBGPCollector())
	}
}

// WithDHCP enables DHCP server metrics
func WithDHCP() Option {
	return func(c *collector) {
		c.collectors = append(c.collectors, newDHCPCollector())
	}
}

//WithDHCP enables DHCP server lease metrics
func WithDHCPL() Option {
	return func(c *collector) {
		c.collectors = append(c.collectors, newDHCPLCollector())
	}
}

// WithFirmware enables firmware metrics
func WithFirmware() Option {
	return func(c *collector) {
		c.collectors = append(c.collectors, newFirmwareCollector())
	}
}

// WithWLANIF enables wlan interfaces matrics
func WithWLANIF() Option {
	return func(c *collector) {
		c.collectors = append(c.collectors, newWLANIFCollector())
	}
}

// WithWLANSTA enables wlan station metrics
func WithWLANSTA() Option {
	return func(c *collector) {
		c.collectors = append(c.collectors, newWLANSTACollector())
	}
}

// Option applies option to collector
type Option func(*collector)

func NewCollector(cfg *config.Config, opts ...Option) (prometheus.Collector, error) {

	log.WithFields(log.Fields{
		"numDevices": len(cfg.Devices),
	}).Info("setting up collector for devices")

	c := &collector{
		devices: cfg.Devices,
		timeout: DefaultTimeout,
		collectors: []routerOSCollector{
			newInterfaceCollector(),
			newResourceCollector(),
		},
	}

	for _, o := range opts {
		o(c)
	}

	return c, nil
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc

	for _, co := range c.collectors {
		co.describe(ch)
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {

	wg := sync.WaitGroup{}

	var realDevices []config.Device

	for _, dev := range c.devices {
		if (config.SrvRecord{}) != dev.Srv {
			log.WithFields(log.Fields{
				"SRV": dev.Srv.Record,
			}).Info("SRV configuration detected")
			conf, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
			dnsServer := net.JoinHostPort(conf.Servers[0], strconv.Itoa(dnsPort))
			if (config.DnsServer{}) != dev.Srv.Dns {
				dnsServer = net.JoinHostPort(dev.Srv.Dns.Address, strconv.Itoa(dev.Srv.Dns.Port))
				log.WithFields(log.Fields{
					"DnsServer": dnsServer,
				}).Info("Custom DNS config detected")
			}
			dnsMsg := new(dns.Msg)
			dnsCli := new(dns.Client)

			dnsMsg.RecursionDesired = true
			dnsMsg.SetQuestion(dns.Fqdn(dev.Srv.Record), dns.TypeSRV)
			r, _, err := dnsCli.Exchange(dnsMsg, dnsServer)

			if err != nil {
				os.Exit(1)
			}

			for _, k := range r.Answer {
				if s, ok := k.(*dns.SRV); ok {
					d := config.Device{}
					d.Name = strings.TrimRight(s.Target, ".")
					d.Address = strings.TrimRight(s.Target, ".")
					d.User = dev.User
					d.Password = dev.Password
					c.getIdentity(&d)
					realDevices = append(realDevices, d)
				}
			}
		} else {
			realDevices = append(realDevices, dev)
		}
	}

	wg.Add(len(realDevices))
	for _, dev := range realDevices {
		go func(d config.Device) {
			c.collectForDevice(d, ch)
			wg.Done()
		}(dev)
	}

	wg.Wait()
}

func (c *collector) getIdentity(d *config.Device) error {
	cl, err := c.connect(d)

	if err != nil {
		log.WithFields(log.Fields{
			"device": d.Name,
			"error":  err,
		}).Error("error dialing device")
		return err
	}
	defer cl.Close()
	reply, err := cl.Run("/system/identity/print")
	if err != nil {
		log.WithFields(log.Fields{
			"device": d.Name,
			"error":  err,
		}).Error("error fetching ethernet interfaces")
		return err
	}
	for _, id := range reply.Re {
		d.Name = id.Map["name"]
	}

	return nil
}

func (c *collector) collectForDevice(d config.Device, ch chan<- prometheus.Metric) {
	begin := time.Now()

	err := c.connectAndCollect(&d, ch)

	duration := time.Since(begin)
	var success float64
	if err != nil {
		log.Errorf("ERROR: %s collector failed after %fs: %s", d.Name, duration.Seconds(), err)
		success = 0
	} else {
		log.Debugf("OK: %s collector succeeded after %fs.", d.Name, duration.Seconds())
		success = 1
	}

	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), d.Name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, d.Name)
}

func (c *collector) connectAndCollect(d *config.Device, ch chan<- prometheus.Metric) error {
	cl, err := c.connect(d)
	if err != nil {
		log.WithFields(log.Fields{
			"device": d.Name,
			"error":  err,
		}).Error("error dialing device")
		return err
	}
	defer cl.Close()

	for _, co := range c.collectors {
		ctx := &collectorContext{ch, d, cl}
		err = co.collect(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *collector) connect(d *config.Device) (*routeros.Client, error) {
	var conn net.Conn
	var err error

	log.WithField("device", d.Name).Debug("trying to Dial")
	if !c.enableTLS {
		if (d.Port) == "" {
			d.Port = apiPort
		}
		conn, err = net.DialTimeout("tcp", d.Address+":"+d.Port, c.timeout)
		if err != nil {
			return nil, err
		}
	} else {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: c.insecureTLS,
		}
		if (d.Port) == "" {
			d.Port = apiPortTLS
		}
		conn, err = tls.DialWithDialer(&net.Dialer{
			Timeout: c.timeout,
		},
			"tcp", d.Address+":"+d.Port, tlsCfg)
		if err != nil {
			return nil, err
		}
	}
	log.WithField("device", d.Name).Debug("done dialing")
	client, err := routeros.NewClient(conn)
	if err != nil {
		return nil, err
	}
	log.WithField("device", d.Name).Debug("got client")

	log.WithField("device", d.Name).Debug("trying to login")
	r, err := client.Run("/login", "=name="+d.User, "=password="+d.Password)
	if err != nil {
		return nil, err
	}
	ret, ok := r.Done.Map["ret"]
	if !ok {
		// Login method post-6.43 one stage, cleartext and no challenge
		if r.Done != nil {
			return client, nil
		}
		return nil, errors.New("RouterOS: /login: no ret (challenge) received")
	}
	// Login method pre-6.43 two stages, challenge
	b, err := hex.DecodeString(ret)
	if err != nil {
		return nil, fmt.Errorf("RouterOS: /login: invalid ret (challenge) hex string received: %s", err)
	}

	r, err = client.Run("/login", "=name="+d.User, "=response="+challengeResponse(b, d.Password))
	if err != nil {
		return nil, err
	}
	log.WithField("device", d.Name).Debug("done with login")

	return client, nil
}

func challengeResponse(cha []byte, password string) string {
	h := md5.New()
	h.Write([]byte{0})
	_, _ = io.WriteString(h, password)
	h.Write(cha)
	return fmt.Sprintf("00%x", h.Sum(nil))
}
