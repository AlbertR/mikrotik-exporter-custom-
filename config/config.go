package config

import (
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Config represent the configuration for the exporter
type Config struct {
	Devices  []Device `yaml:"devices"`
	Features struct {
		BGP      bool `yaml:"bgp,omitempty"`
		DHCP     bool `yaml:"dhcp,omitempty"`
		DHCPL    bool `yaml:"dhcpl,omitempty"`
		Firmware bool `yaml:"firmware,omitempty"`
		WLANIF   bool `yaml:"wlanif,omitempty"`
	} `yaml:"features,omitempty"`
}

// Device represent a target device
type Device struct {
	Name     string    `yaml:"name"`
	Address  string    `yaml:"address,omitempty"`
	Srv      SrvRecord `yaml:"srv,omitempty"`
	User     string    `yaml:"user"`
	Password string    `yaml:"password"`
	Port     string    `yaml:"port"`
}

type SrvRecord struct {
	Record string    `yaml:"record"`
	Dns    DnsServer `yaml:"dns,omitempty"`
}

type DnsServer struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

// Load reads YAML from reader and unmarshals in Config
func Load(r io.Reader) (*Config, error) {

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	c := &Config{}

	err = yaml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
