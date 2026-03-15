package config

import "time"

type explorer struct {
	Period    time.Duration `yaml:"period"`
	Multicast *multicast    `yaml:"multicast,omitempty"`
	Broadcast *broadcast    `yaml:"broadcast,omitempty"`
}

type multicast struct {
	Address       string `yaml:"address"`
	Port          int    `yaml:"port"`
	InterfaceName string `yaml:"interface-name"`
}

type broadcast struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}
