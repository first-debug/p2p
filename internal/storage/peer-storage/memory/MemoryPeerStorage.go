package inmemorysessionstorage

import (
	"fmt"
	"sync"

	"main/internal/domain"
	peerstorage "main/internal/storage/peer-storage"
)

type MemoryPeerStorage struct {
	peersMux sync.RWMutex
	peers    map[string]domain.Peer
}

func NewMemoryPeerStorage() peerstorage.PeerStorage {
	return &MemoryPeerStorage{
		peers: make(map[string]domain.Peer),
	}
}

func (s *MemoryPeerStorage) Add(newPeer domain.Peer) error {
	s.peersMux.Lock()
	s.peers[string(newPeer.ID)] = newPeer
	s.peersMux.Unlock()
	return nil
}

func (s *MemoryPeerStorage) GetAll() ([]domain.Peer, error) {
	s.peersMux.RLock()
	count := len(s.peers)
	res := make([]domain.Peer, count)
	count--
	for _, j := range s.peers {
		res[count] = j
		count--
	}
	s.peersMux.RUnlock()
	return res, nil
}

func (s *MemoryPeerStorage) GetByID(id []byte) (domain.Peer, error) {
	s.peersMux.RLock()
	for i, j := range s.peers {
		if string(id) == i {
			return j, nil
		}
	}
	s.peersMux.RUnlock()
	return domain.Peer{}, fmt.Errorf("cannot find Peer with ID = %v", id)
}

func (s *MemoryPeerStorage) RemoveByID(id []byte) error {
	s.peersMux.Lock()
	for i := range s.peers {
		if string(id) == i {
			delete(s.peers, i)
			return nil
		}
	}
	s.peersMux.Unlock()
	return fmt.Errorf("cannot found Peer with ID = %v", id)
}

func (s *MemoryPeerStorage) RemoveByName(name string) error {
	s.peersMux.Lock()
	for i, j := range s.peers {
		if name == j.Name {
			delete(s.peers, i)
			return nil
		}
	}
	s.peersMux.Unlock()
	return fmt.Errorf("cannot found Peer with Name = %s", name)
}
