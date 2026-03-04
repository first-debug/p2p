// Package config provide struct that are representatio of service configuration.
// Also it provide function [MustLoad] which load variable from ENV
package config

import (
	"flag"

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
}

func MustLoad() *Config {
	cfg := &Config{}

	var envPath string
	flag.StringVar(&envPath, "env-file", "", "explicitly specifying the env file to use")
	flag.Parse()

	if err := godotenv.Load(envPath); envPath != "" && err != nil {
		panic(err.Error())
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		panic(err.Error())
	}

	return cfg
}
