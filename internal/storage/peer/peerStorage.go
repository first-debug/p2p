package peerstorage

import (
	"github.com/first-debug/p2p/internal/domain"
	"github.com/google/uuid"
)

type PeerStorage interface {
	Add(domain.Peer) error
	GetAll() ([]domain.Peer, error)
	// GetPage(start, stop int) ([]domain.PeerInfo, error)
	GetByID(uuid.UUID) (domain.Peer, error)
	RemoveByID(uuid.UUID) error
	RemoveByName(string) error
}
