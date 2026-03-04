package peerstorage

import (
	"main/internal/domain"
)

type PeerStorage interface {
	Add(domain.Peer) error
	GetAll() ([]domain.Peer, error)
	// GetPage(start, stop int) ([]domain.PeerInfo, error)
	GetByID([]byte) (domain.Peer, error)
	RemoveByID([]byte) error
	RemoveByName(string) error
}
