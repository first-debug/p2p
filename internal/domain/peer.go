package domain

import (
	"net"

	"github.com/google/uuid"
)

type Peer struct {
	ID    uuid.UUID
	Name  string
	IP    net.IP
	Port  int
	Files map[string]string
}
