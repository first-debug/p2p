// Package config provide struct that are representatio of service configuration.
// Also it provide function [MustLoad] which load variable from ENV
package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	// PeerName               string `env:"PEER_NAME"`
	// PeerPort               int    `env:"PEER_PORT" env-default:"8001"`
	WebSocketPort          int    `env:"WEBSOCKET_PORT" env-default:"8001"`
	MulticastAddress       string `env:"MULTICAST_ADDRESS" env-default:"235.5.5.11"`
	MulticastPort          int    `env:"MULTICAST_PORT" env-default:"8001"`
	MulticastInterfaceName string `env:"MULTICAST_INTERFACE_NAME" env-default:"wlan0"`
	CachePath              string `env:"CACHE_PATH"`
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

	var cacheDir string

	if cfg.CachePath == "" {
		cacheDir, err = os.UserCacheDir()
		if err != nil {
			panic(err)
		}
		cacheDir += "/p2p/"
	} else {
		fmt.Println(2)
		cacheDir = cfg.CachePath
		if cacheDir[len(cacheDir)-1] != '/' {
			cacheDir += "/"
		}
	}

	_, stat := os.Stat(cacheDir)
	if os.IsNotExist(stat) {
		if err := os.MkdirAll(cacheDir, 0o755); err != nil {
			panic(err)
		}
	}

	cfg.LogFile = cacheDir + "log.log"
	cfg.IDFile = cacheDir + "id"

	return cfg
}
