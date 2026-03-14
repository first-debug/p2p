package memory

import (
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/storage"
	"github.com/google/uuid"
)

func createTestLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

func TestMemoryPeerStorage_Add(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemoryPeerStorage(logger)

	peer := domain.Peer{
		ID:   uuid.New(),
		Name: "peer1",
		IP:   net.ParseIP("192.168.1.1"),
		Port: 8080,
	}

	if err := s.Add(peer); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	peers, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(peers) != 1 {
		t.Errorf("expected 1 peer, got %d", len(peers))
	}
}

func TestMemoryPeerStorage_Add_Duplicate(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemoryPeerStorage(logger)

	id := uuid.New()
	peer := domain.Peer{
		ID:   id,
		Name: "peer1",
		Port: 8080,
	}

	if err := s.Add(peer); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}

	if err := s.Add(peer); err == nil {
		t.Error("expected error on duplicate Add")
	} else if err != storage.ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestMemoryPeerStorage_GetAll(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemoryPeerStorage(logger)

	peers := []domain.Peer{
		{ID: uuid.New(), Name: "peer1", Port: 8080},
		{ID: uuid.New(), Name: "peer2", Port: 8081},
		{ID: uuid.New(), Name: "peer3", Port: 8082},
	}

	for _, p := range peers {
		if err := s.Add(p); err != nil {
			t.Fatalf("Add returned error: %v", err)
		}
	}

	result, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 peers, got %d", len(result))
	}
}

func TestMemoryPeerStorage_GetAll_Empty(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemoryPeerStorage(logger)

	result, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 peers, got %d", len(result))
	}
}

func TestMemoryPeerStorage_GetByID(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemoryPeerStorage(logger)

	peer := domain.Peer{
		ID:   uuid.New(),
		Name: "test-peer",
		Port: 9000,
	}

	if err := s.Add(peer); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	found, err := s.GetByID(peer.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if found.ID != peer.ID {
		t.Errorf("expected ID %v, got %v", peer.ID, found.ID)
	}
	if found.Name != "test-peer" {
		t.Errorf("expected Name 'test-peer', got %s", found.Name)
	}
}

func TestMemoryPeerStorage_GetByID_NotFound(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemoryPeerStorage(logger)

	_, err := s.GetByID(uuid.New())
	if err == nil {
		t.Error("expected error for non-existent ID")
	}
}

func TestMemoryPeerStorage_RemoveByID(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemoryPeerStorage(logger)

	peer := domain.Peer{
		ID:   uuid.New(),
		Name: "peer1",
		Port: 8080,
	}

	if err := s.Add(peer); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := s.RemoveByID(peer.ID); err != nil {
		t.Fatalf("RemoveByID returned error: %v", err)
	}

	peers, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(peers) != 0 {
		t.Errorf("expected 0 peers after removal, got %d", len(peers))
	}
}

func TestMemoryPeerStorage_RemoveByID_NotFound(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemoryPeerStorage(logger)

	err := s.RemoveByID(uuid.New())
	if err == nil {
		t.Error("expected error for non-existent ID")
	}
}

func TestMemoryPeerStorage_RemoveByName(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemoryPeerStorage(logger)

	peer := domain.Peer{
		ID:   uuid.New(),
		Name: "peer-to-remove",
		Port: 8080,
	}

	if err := s.Add(peer); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := s.RemoveByName("peer-to-remove"); err != nil {
		t.Fatalf("RemoveByName returned error: %v", err)
	}

	peers, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(peers) != 0 {
		t.Errorf("expected 0 peers after removal, got %d", len(peers))
	}
}

func TestMemoryPeerStorage_RemoveByName_NotFound(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemoryPeerStorage(logger)

	err := s.RemoveByName("non-existent")
	if err == nil {
		t.Error("expected error for non-existent name")
	}
}

func TestMemoryPeerStorage_Close(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemoryPeerStorage(logger)
	s.Close()
}

func TestMemoryPeerStorage_ConcurrentAccess(t *testing.T) {
	logger := createTestLogger(t)
	s := NewMemoryPeerStorage(logger)

	errs := make(chan error, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			peer := domain.Peer{
				ID:   uuid.New(),
				Name: "peer",
				Port: 8000 + id,
			}
			errs <- s.Add(peer)
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
