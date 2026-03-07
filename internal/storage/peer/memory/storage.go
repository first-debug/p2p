package memory

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/first-debug/p2p/internal/domain"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer"
	"github.com/google/uuid"
)

var logger log.Logger = *log.New(os.Stderr, "[MemoryPeerStorage] ", log.LstdFlags)

type MemoryPeerStorage struct {
	peersMux sync.RWMutex
	peers    map[uuid.UUID]domain.Peer
}

func NewMemoryPeerStorage() peerstorage.PeerStorage {
	return &MemoryPeerStorage{
		peers: make(map[uuid.UUID]domain.Peer),
	}
}

func (s *MemoryPeerStorage) Add(newPeer domain.Peer) error {
	s.peersMux.Lock()
	s.peers[newPeer.ID] = newPeer
	s.peersMux.Unlock()
	logger.Printf("added new Peer = %v", newPeer)
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

func (s *MemoryPeerStorage) GetByID(id uuid.UUID) (domain.Peer, error) {
	s.peersMux.RLock()
	for i, j := range s.peers {
		if id == i {
			return j, nil
		}
	}
	s.peersMux.RUnlock()
	return domain.Peer{}, fmt.Errorf("cannot find Peer with ID = %v", id)
}

func (s *MemoryPeerStorage) RemoveByID(id uuid.UUID) error {
	s.peersMux.Lock()
	for i := range s.peers {
		if id == i {
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
