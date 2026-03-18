package domain

import (
	"net"

	"github.com/google/uuid"
)

type Peer struct {
	ID         uuid.UUID
	Name       string
	IP         net.IP
	IsPublicIP bool
	Port       int
	Files      map[string]string
}
