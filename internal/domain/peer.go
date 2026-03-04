package domain

import "net"

type Peer struct {
	ID    []byte
	Name  string
	IP    net.IP
	Port  int
	Files map[string]string
}
