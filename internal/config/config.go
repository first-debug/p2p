// Package config provide struct that are representatio of service configuration.
// Also it provide function [MustLoad] which load variable from ENV
package config

import (
	"flag"
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	// PeerName               string `env:"PEER_NAME"`
	// PeerPort               int    `env:"PEER_PORT" env-default:"8001"`
	LogLevel               slog.Level `env:"LOG_LEVEL" env-default:"INFO"`
	WebSocketPort          int        `env:"WEBSOCKET_PORT" env-default:"8001"`
	MulticastAddress       string     `env:"MULTICAST_ADDRESS" env-default:"235.5.5.11"`
	MulticastPort          int        `env:"MULTICAST_PORT" env-default:"8001"`
	MulticastInterfaceName string     `env:"MULTICAST_INTERFACE_NAME" env-default:"wlan0"`
	CachePath              string     `env:"CACHE_PATH"`
	LogFile                string
	IDFile                 string
}

func MustLoad() *Config {
	cfg := &Config{}

	var envPath string
	flag.StringVar(&envPath, "env-file", "", "explicitly specifying the env file to use")
	flag.Parse()

	err := godotenv.Load(envPath)
	if envPath != "" && err != nil {
		panic(err.Error())
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		panic(err.Error())
	}

	if cfg.CachePath == "" {
		cfg.CachePath, err = os.UserCacheDir()
		if err != nil {
			panic(err)
		}
		cfg.CachePath += "/p2p/"
	} else {
		if cfg.CachePath[len(cfg.CachePath)-1] != '/' {
			cfg.CachePath += "/"
		}
	}

	_, stat := os.Stat(cfg.CachePath)
	if os.IsNotExist(stat) {
		if err := os.MkdirAll(cfg.CachePath, 0o755); err != nil {
			panic(err)
		}
	}

	cfg.LogFile = cfg.CachePath + "log.log"
	cfg.IDFile = cfg.CachePath + "id"

	return cfg
}
