package inmemorysessionstorage

import (
	"context"
	"fmt"
	"sync"

	"main/internal/session"
	sessionstorage "main/internal/storage/session-storage"

	"github.com/google/uuid"
)

type MemorySessionStorage struct {
	sessionsMux sync.RWMutex
	sessions    map[uuid.UUID]session.Session
}

func NewMemorySessionStorage() sessionstorage.SessionStorage {
	return &MemorySessionStorage{
		sessions: make(map[uuid.UUID]session.Session),
	}
}

func (s *MemorySessionStorage) Add(newSession session.Session) error {
	s.sessionsMux.Lock()
	s.sessions[newSession.GetID()] = newSession
	s.sessionsMux.Unlock()
	return nil
}

func (s *MemorySessionStorage) GetAll() ([]session.Session, error) {
	s.sessionsMux.RLock()
	count := len(s.sessions)
	res := make([]session.Session, count)
	count--
	for _, j := range s.sessions {
		res[count] = j
		count--
	}
	s.sessionsMux.RUnlock()
	return res, nil
}

// func (s *MemorySessionStorage) GetPage(start, stop int) ([]session.Session, error) {
// 	return nil, nil
// }

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
	for i, j := range s.sessions {
		if id == i {
			j.Close(ctx)
			delete(s.sessions, i)
			return nil
		}
	}
	s.sessionsMux.Unlock()
	return fmt.Errorf("cannot found Session with ID = %v", id)
}

func (s *MemorySessionStorage) CloseAllByType(ctx context.Context, incoming bool) error {
	s.sessionsMux.Lock()
	for i, j := range s.sessions {
		if incoming == j.IsIncoming() {
			j.Close(ctx)
			delete(s.sessions, i)
			return nil
		}
	}
	s.sessionsMux.Unlock()
	return nil
}
