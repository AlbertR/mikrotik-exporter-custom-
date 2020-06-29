package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"

	"mikrotik_exporter/collector"
	"mikrotik_exporter/config"
)

var (
	address    = flag.String("address", "", "address of the device to monitor")
	configFile = flag.String("config-file", "", "config file to load")
	device     = flag.String("device", "", "single device to monitor")
	insecure   = flag.Bool("insecure", false, "skips verification of server certificate when using TLS ("+
		"not recommended)")
	logFormat   = flag.String("log-format", "json", "logformat text or json (default json)")
	logLevel    = flag.String("log-level", "debug", "log level")
	metricsPath = flag.String("path", "/metrics", "path to answer requests on")
	password    = flag.String("password", "", "password for authentication for single device")
	deviceport  = flag.String("deviceport", "8728", "port for single device")
	port        = flag.String("port", ":9436", "port number to listen on")
	user        = flag.String("user", "", "user for authentication with single device")
	ver         = flag.Bool("version", false, "find the version of binary")

	withBgp = flag.Bool("with-bgp", false, "retrieves BGP routing information")

	withDHCP     = flag.Bool("with-dhcp", false, "retrives DHCP server metrics")
	withDHCPL    = flag.Bool("with-dhcpl", false, "retrives DHCP server lease metrics")
	withFirmware = flag.Bool("with-firmware", false, "retrives Firmware metrics")

	cfg *config.Config

	appVersion = "DEVELOPMENT"
	shortSha   = "0xFACEBEAD"
)

func init() {
	prometheus.MustRegister(version.NewCollector("mikrotik_exporter"))
}

func main() {
	flag.Parse()
	if *ver {
		fmt.Printf("\nVersion:   %s\nShort SHA: %s\n\n", appVersion, shortSha)
		os.Exit(0)
	}

	configureLog()

	c, err := loadConfig()
	if err != nil {
		log.Errorf("Could not load config: %v", err)
		os.Exit(3)
	}
	cfg = c

	startServer()
}

func configureLog() {

	ll, err := log.ParseLevel(*logLevel)

	if err != nil {
		panic(err)
	}

	log.SetLevel(ll)

	if *logFormat == "text" {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

func loadConfig() (*config.Config, error) {

	if *configFile != "" {
		return loadConfigFromFile()
	}

	return loadConfigFromFlags()
}

func loadConfigFromFile() (*config.Config, error) {

	b, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}

	return config.Load(bytes.NewReader(b))
}

func loadConfigFromFlags() (*config.Config, error) {

	if *device == "" || *address == "" || *user == "" || *password == "" {
		return nil, fmt.Errorf("missing required param for single device configuration")
	}

	return &config.Config{
		Devices: []config.Device{
			config.Device{
				Name:     *device,
				Address:  *address,
				User:     *user,
				Password: *password,
				Port:     *deviceport,
			},
		},
	}, nil
}

func startServer() {
	h, err := createMetricsHandler()

	if err != nil {
		log.Fatal(err)
	}
	http.Handle(*metricsPath, h)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>Mikrotik Exporter</title></head>
			<body>
			<h1>Mikrotik Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Info("Listening on ", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}

func createMetricsHandler() (http.Handler, error) {
	opts := collectorOptions()
	nc, err := collector.NewCollector(cfg, opts...)
	if err != nil {
		return nil, err
	}

	registry := prometheus.NewRegistry()
	err = registry.Register(nc)
	if err != nil {
		return nil, err
	}

	return promhttp.HandlerFor(registry,
		promhttp.HandlerOpts{
			ErrorLog:      log.New(),
			ErrorHandling: promhttp.ContinueOnError,
		}), nil
}

func collectorOptions() []collector.Option {
	opts := []collector.Option{}

	if *withBgp || cfg.Features.BGP {
		opts = append(opts, collector.WithBGP())
	}

	if *withDHCP || cfg.Features.DHCP {
		opts = append(opts, collector.WithDHCP())
	}

	if *withDHCPL || cfg.Features.DHCPL {
		opts = append(opts, collector.WithDHCPL())
	}

	if *withFirmware || cfg.Features.Firmware {
		opts = append(opts, collector.WithFirmware())
	}

	return opts
}
