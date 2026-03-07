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
	defer s.peersMux.Unlock()

	s.peers[newPeer.ID] = newPeer
	logger.Printf("added new Peer = %v", newPeer)
	return nil
}

func (s *MemoryPeerStorage) GetAll() ([]domain.Peer, error) {
	s.peersMux.RLock()
	defer s.peersMux.RUnlock()

	count := len(s.peers)
	res := make([]domain.Peer, count)
	count--
	for _, j := range s.peers {
		res[count] = j
		count--
	}
	return res, nil
}

func (s *MemoryPeerStorage) GetByID(id uuid.UUID) (domain.Peer, error) {
	s.peersMux.RLock()
	defer s.peersMux.RUnlock()

	for i, j := range s.peers {
		if id == i {
			return j, nil
		}
	}
	return domain.Peer{}, fmt.Errorf("cannot find Peer with ID = %v", id)
}

func (s *MemoryPeerStorage) RemoveByID(id uuid.UUID) error {
	s.peersMux.Lock()
	defer s.peersMux.Unlock()

	for i := range s.peers {
		if id == i {
			delete(s.peers, i)
			return nil
		}
	}
	return fmt.Errorf("cannot found Peer with ID = %v", id)
}

func (s *MemoryPeerStorage) RemoveByName(name string) error {
	s.peersMux.Lock()
	defer s.peersMux.Unlock()

	for i, j := range s.peers {
		if name == j.Name {
			delete(s.peers, i)
			return nil
		}
	}
	return fmt.Errorf("cannot found Peer with Name = %s", name)
}
