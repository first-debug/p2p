// Package config provide struct that are representatio of service configuration.
// Also it provide function [MustLoad] which load variable from ENV
package config

import (
	"flag"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// PeerName               string `env:"PEER_NAME"`
	// PeerPort               int    `env:"PEER_PORT" env-default:"8001"`
	LogLevel   slog.Level `yaml:"log-level"`
	WebSocket  webSocket  `yaml:"websocket"`
	Explorer   explorer   `yaml:"explorer"`
	IP         *net.IP    `yaml:"ip,omitempty"`
	IsPublicIP bool       `yaml:"is-public-ip"`
	ConfigDir  string     `yaml:"-"`
	HistoryDir string     `yaml:"-"`
}

func MustLoad() *Config {
	cfg := &Config{}

	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	cfg.ConfigDir = strings.TrimRight(strings.TrimRight(configDir, "/"), "\\") + "/p2p/"
	cfg.HistoryDir = cfg.ConfigDir + "history/"

	// make a config and a history directories
	if err := makeDirIfNotExists(cfg.HistoryDir); err != nil {
		panic(err)
	}

	var ConfigFile string
	flag.StringVar(&ConfigFile, "config", cfg.ConfigDir+"config.yaml", "explicitly specifying the env file to use")
	flag.Parse()

	if err := parseConfigFile(ConfigFile, cfg); err != nil {
		panic(err)
	}

	return cfg
}

func makeDirIfNotExists(configDir string) (err error) {
	_, err = os.Stat(configDir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return err
		}
	}

	return nil
}

func parseConfigFile(ConfigFile string, config *Config) error {
	var file *os.File
	file, err := os.Open(ConfigFile)
	if err == nil {
		defer file.Close()

		return yaml.NewDecoder(file).Decode(config)
	}

	if os.IsNotExist(err) {
		config.WebSocket = webSocket{Port: 8001}
		config.Explorer = explorer{
			Period: 20 * time.Second,
			Multicast: &multicast{
				Address: "235.5.5.11",
				Port:    8001,
			},
		}

		if err = createConfigFile(ConfigFile, config); err != nil {
			return err
		}
	}
	return nil
}

func createConfigFile(ConfigFile string, config *Config) error {
	file, err := os.Create(ConfigFile)
	if err != nil {
		return err
	}

	if err = yaml.NewEncoder(file).Encode(config); err != nil {
		return err
	}

	return nil
}
