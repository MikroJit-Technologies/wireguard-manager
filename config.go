package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr string     `yaml:"listen_addr"`
	Interface  string     `yaml:"interface"`
	WgConfig   string     `yaml:"wg_config"`
	Server     ServerInfo `yaml:"server"`
	Defaults   Defaults   `yaml:"defaults"`
	Auth       AuthConfig `yaml:"auth"`
}

type ServerInfo struct {
	Endpoint string `yaml:"endpoint"`
	DNS      string `yaml:"dns"`
}

type Defaults struct {
	AllowedIPs string `yaml:"allowed_ips"`
	MTU        int    `yaml:"mtu"`
}

type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func loadConfig(path string) (*Config, error) {
	cfg := &Config{
		ListenAddr: ":8080",
		Interface:  "wg0",
		WgConfig:   "/etc/wireguard/wg0.conf",
		Defaults: Defaults{
			AllowedIPs: "0.0.0.0/0, ::/0",
			MTU:        1420,
		},
	}
	f, err := os.Open(path)
	if err != nil {
		return cfg, nil // use defaults if no config
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
