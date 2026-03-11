package memory

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/first-debug/p2p/internal/session"
	"github.com/first-debug/p2p/internal/storage"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session"

	"github.com/google/uuid"
)

type MemorySessionStorage struct {
	logger *slog.Logger

	ctx       context.Context
	ctxCancel context.CancelFunc
	wg        sync.WaitGroup

	sessionsMux sync.RWMutex
	sessions    map[uuid.UUID]session.Session
}

func NewMemorySessionStorage(log *slog.Logger) sessionstorage.SessionStorage {
	ctx, cancel := context.WithCancel(context.Background())
	storage := &MemorySessionStorage{
		logger:    log.With("module", "MemorySessionStorage"),
		ctx:       ctx,
		ctxCancel: cancel,
		sessions:  make(map[uuid.UUID]session.Session),
	}

	storage.wg.Go(storage.checkSessionsAvailable)

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
	defer s.sessionsMux.RUnlock()

	for i, j := range s.sessions {
		if id == i {
			return j, nil
		}
	}
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

func (s *MemorySessionStorage) checkSessionsAvailable() {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			sessions, err := s.GetAll()
			if err != nil {
				s.logger.Error(err.Error())
			}
			s.sessionsMux.Lock()
			for _, i := range sessions {
				if !i.IsOpen() {
					delete(s.sessions, i.GetID())
				}
			}
			s.sessionsMux.Unlock()
		}
	}
}
