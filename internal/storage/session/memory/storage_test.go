package memory

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/first-debug/p2p/internal/domain"
	pb "github.com/first-debug/p2p/internal/proto"
	"github.com/first-debug/p2p/internal/session"
	"github.com/first-debug/p2p/internal/storage"
	"github.com/google/uuid"
)

type mockSession struct {
	id       uuid.UUID
	peer     domain.Peer
	incoming bool
	lastDial time.Time
	open     bool
	closed   bool
}

func newMockSession(id uuid.UUID, peer domain.Peer, incoming bool) *mockSession {
	return &mockSession{
		id:       id,
		peer:     peer,
		incoming: incoming,
		lastDial: time.Now(),
		open:     true,
		closed:   false,
	}
}

func (m *mockSession) GetID() uuid.UUID                    { return m.id }
func (m *mockSession) GetLastDial() time.Time              { return m.lastDial }
func (m *mockSession) GetPeerID() uuid.UUID                { return m.peer.ID }
func (m *mockSession) GetReadChannel(context.Context) (<-chan *pb.Message, error) {
	return nil, nil
}
func (m *mockSession) GetWriteChannel(context.Context) (chan<- *pb.Message, error) {
	return nil, nil
}
func (m *mockSession) IsIncoming() bool { return m.incoming }
func (m *mockSession) IsOpen() bool     { return m.open }
func (m *mockSession) Close(context.Context) {
	m.open = false
	m.closed = true
}

func createTestLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

func TestMemorySessionStorage_Add(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemorySessionStorage(logger)
	defer s.Close()

	sess := newMockSession(uuid.New(), domain.Peer{ID: uuid.New(), Port: 8080}, false)

	if err := s.Add(sess); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	sessions, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessions))
	}
}

func TestMemorySessionStorage_Add_Duplicate(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemorySessionStorage(logger)
	defer s.Close()

	id := uuid.New()
	sess := newMockSession(id, domain.Peer{ID: uuid.New(), Port: 8080}, false)

	if err := s.Add(sess); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}

	if err := s.Add(sess); err == nil {
		t.Error("expected error on duplicate Add")
	} else if err != storage.ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestMemorySessionStorage_GetAll(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemorySessionStorage(logger)
	defer s.Close()

	sessions := []session.Session{
		newMockSession(uuid.New(), domain.Peer{ID: uuid.New(), Port: 8080}, false),
		newMockSession(uuid.New(), domain.Peer{ID: uuid.New(), Port: 8081}, true),
		newMockSession(uuid.New(), domain.Peer{ID: uuid.New(), Port: 8082}, false),
	}

	for _, sess := range sessions {
		if err := s.Add(sess); err != nil {
			t.Fatalf("Add returned error: %v", err)
		}
	}

	result, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(result))
	}
}

func TestMemorySessionStorage_GetAll_Empty(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemorySessionStorage(logger)
	defer s.Close()

	result, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(result))
	}
}

func TestMemorySessionStorage_GetByID(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemorySessionStorage(logger)
	defer s.Close()

	id := uuid.New()
	sess := newMockSession(id, domain.Peer{ID: uuid.New(), Port: 9000}, true)

	if err := s.Add(sess); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	found, err := s.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if found.GetID() != id {
		t.Errorf("expected ID %v, got %v", id, found.GetID())
	}
	if !found.IsIncoming() {
		t.Error("expected IsIncoming to be true")
	}
}

func TestMemorySessionStorage_GetByID_NotFound(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemorySessionStorage(logger)
	defer s.Close()

	_, err := s.GetByID(uuid.New())
	if err == nil {
		t.Error("expected error for non-existent ID")
	}
}

func TestMemorySessionStorage_CloseByID(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemorySessionStorage(logger)
	defer s.Close()

	id := uuid.New()
	sess := newMockSession(id, domain.Peer{ID: uuid.New(), Port: 8080}, false)

	if err := s.Add(sess); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	ctx := context.Background()
	if err := s.CloseByID(ctx, id); err != nil {
		t.Fatalf("CloseByID returned error: %v", err)
	}

	if !sess.closed {
		t.Error("expected session to be closed")
	}
	if sess.open {
		t.Error("expected session.IsOpen() to be false")
	}

	sessions, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions after CloseByID, got %d", len(sessions))
	}
}

func TestMemorySessionStorage_CloseByID_NotFound(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemorySessionStorage(logger)
	defer s.Close()

	ctx := context.Background()
	err := s.CloseByID(ctx, uuid.New())
	if err == nil {
		t.Error("expected error for non-existent ID")
	}
}

func TestMemorySessionStorage_CloseAllByType(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemorySessionStorage(logger)
	defer s.Close()

	incomingSess := newMockSession(uuid.New(), domain.Peer{ID: uuid.New(), Port: 8080}, true)
	outgoingSess := newMockSession(uuid.New(), domain.Peer{ID: uuid.New(), Port: 8081}, false)

	if err := s.Add(incomingSess); err != nil {
		t.Fatalf("Add incoming returned error: %v", err)
	}
	if err := s.Add(outgoingSess); err != nil {
		t.Fatalf("Add outgoing returned error: %v", err)
	}

	ctx := context.Background()
	if err := s.CloseAllByType(ctx, true); err != nil {
		t.Fatalf("CloseAllByType returned error: %v", err)
	}

	if !incomingSess.closed {
		t.Error("expected incoming session to be closed")
	}
	if !outgoingSess.open {
		t.Error("expected outgoing session to remain open")
	}
}

func TestMemorySessionStorage_CloseAllByType_NoMatching(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemorySessionStorage(logger)
	defer s.Close()

	ctx := context.Background()
	if err := s.CloseAllByType(ctx, true); err != nil {
		t.Fatalf("CloseAllByType returned error: %v", err)
	}
}

func TestMemorySessionStorage_Close(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemorySessionStorage(logger)
	s.Close()
}

func TestMemorySessionStorage_ConcurrentAccess(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemorySessionStorage(logger)
	defer s.Close()

	errs := make(chan error, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			sess := newMockSession(uuid.New(), domain.Peer{ID: uuid.New(), Port: 8000 + id}, false)
			errs <- s.Add(sess)
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func() {
			_, err := s.GetAll()
			errs <- err
		}()
	}

	for i := 0; i < 10; i++ {
		select {
		case err := <-errs:
			if err != nil && err != storage.ErrAlreadyExists {
				t.Errorf("unexpected error: %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("concurrent test timed out")
		}
	}
}
