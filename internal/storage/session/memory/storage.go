package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/first-debug/p2p/internal/session"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session"

	"github.com/google/uuid"
)

type MemorySessionStorage struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	wg        sync.WaitGroup

	sessionsMux sync.RWMutex
	sessions    map[uuid.UUID]session.Session
}

func NewMemorySessionStorage() sessionstorage.SessionStorage {
	ctx, cancel := context.WithCancel(context.Background())
	storage := &MemorySessionStorage{
		ctx:       ctx,
		ctxCancel: cancel,
		sessions:  make(map[uuid.UUID]session.Session),
	}

	return storage
}

func (s *MemorySessionStorage) Add(newSession session.Session) error {
	s.sessionsMux.Lock()
	defer s.sessionsMux.Unlock()

	if _, exist := s.sessions[newSession.GetID()]; exist {
		return storage.ErrAlreadyExists
	}
	s.sessions[newSession.GetID()] = newSession
	return nil
}

func (s *MemorySessionStorage) GetAll() ([]session.Session, error) {
	s.sessionsMux.RLock()
	defer s.sessionsMux.RUnlock()

	count := len(s.sessions)
	res := make([]session.Session, count)
	count--
	for _, j := range s.sessions {
		res[count] = j
		count--
	}
	return res, nil
}

func (s *MemorySessionStorage) GetByID(id uuid.UUID) (session.Session, error) {
	s.sessionsMux.RLock()
	for i, j := range s.sessions {
		if id == i {
			return j, nil
		}
	}
	s.sessionsMux.RUnlock()
	return nil, fmt.Errorf("cannot find Session with ID = %v", id)
}

func (s *MemorySessionStorage) CloseByID(ctx context.Context, id uuid.UUID) error {
	s.sessionsMux.Lock()
	defer s.sessionsMux.Unlock()

	for i, j := range s.sessions {
		if id == i {
			j.Close(ctx)
			delete(s.sessions, i)
			return nil
		}
	}
	return fmt.Errorf("cannot found Session with ID = %v", id)
}

func (s *MemorySessionStorage) CloseAllByType(ctx context.Context, incoming bool) error {
	s.sessionsMux.Lock()
	defer s.sessionsMux.Unlock()

	for i, j := range s.sessions {
		if incoming == j.IsIncoming() {
			j.Close(ctx)
			delete(s.sessions, i)
			return nil
		}
	}
	return nil
}

func (s *MemorySessionStorage) Close() {
	s.ctxCancel()
	s.wg.Wait()
}
