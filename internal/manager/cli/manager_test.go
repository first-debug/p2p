package cli

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/session"
	pb "github.com/first-debug/p2p/internal/proto"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer/memory"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session/memory"
	"github.com/google/uuid"
)

func createTestLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

type mockClient struct {
	connectFunc func(context.Context, *domain.Peer) (session.Session, error)
}

func (m *mockClient) Connect(ctx context.Context, peer *domain.Peer) (session.Session, error) {
	if m.connectFunc != nil {
		return m.connectFunc(ctx, peer)
	}
	return nil, nil
}

type mockSession struct {
	id       uuid.UUID
	peer     domain.Peer
	incoming bool
	lastDial time.Time
	open     bool
	readCh   <-chan *pb.Message
	writeCh  chan<- *pb.Message
}

func (m *mockSession) GetID() uuid.UUID                    { return m.id }
func (m *mockSession) GetLastDial() time.Time              { return m.lastDial }
func (m *mockSession) GetPeerID() uuid.UUID                { return m.peer.ID }
func (m *mockSession) GetReadChannel(context.Context) (<-chan *pb.Message, error) {
	return m.readCh, nil
}
func (m *mockSession) GetWriteChannel(context.Context) (chan<- *pb.Message, error) {
	return m.writeCh, nil
}
func (m *mockSession) IsIncoming() bool { return m.incoming }
func (m *mockSession) IsOpen() bool     { return m.open }
func (m *mockSession) Close(context.Context) {
	m.open = false
}

func TestNewCliManager(t *testing.T) {
	logger := createTestLogger(t)
	ctx := context.Background()
	peer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	defer sStorage.Close()
	client := &mockClient{}

	manager := NewCliManager(ctx, logger, peer, pStorage, sStorage, client)

	if manager == nil {
		t.Fatal("NewCliManager returned nil")
	}

	mgr, ok := manager.(*CliManager)
	if !ok {
		t.Fatal("NewCliManager did not return *CliManager")
	}

	if mgr.ctx == nil {
		t.Error("ctx is nil")
	}
	if mgr.logger == nil {
		t.Error("logger is nil")
	}
	if mgr.selfInfo.ID != peer.ID {
		t.Errorf("selfInfo.ID: expected %v, got %v", peer.ID, mgr.selfInfo.ID)
	}
	if mgr.pStorage == nil {
		t.Error("pStorage is nil")
	}
	if mgr.sStorage == nil {
		t.Error("sStorage is nil")
	}
	if mgr.client == nil {
		t.Error("client is nil")
	}
}

func TestCliManager_Run_ContextCancellation(t *testing.T) {
	logger := createTestLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	peer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	defer sStorage.Close()
	client := &mockClient{}

	manager := NewCliManager(ctx, logger, peer, pStorage, sStorage, client)

	cancel()

	done := make(chan error, 1)
	go func() {
		done <- manager.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Error("Run did not exit after context cancellation")
	}
}

func TestCliManager_getOutputWidth_NonTerminal(t *testing.T) {
	logger := createTestLogger(t)
	ctx := context.Background()
	peer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	defer sStorage.Close()
	client := &mockClient{}

	manager := NewCliManager(ctx, logger, peer, pStorage, sStorage, client)
	mgr := manager.(*CliManager)

	width := mgr.getOutputWidth(os.Stdout)
	if width < -1 {
		t.Errorf("unexpected width: %d", width)
	}
}

func TestCliManager_ListPeers_Empty(t *testing.T) {
	logger := createTestLogger(t)
	ctx := context.Background()
	peer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	defer sStorage.Close()
	client := &mockClient{}

	manager := NewCliManager(ctx, logger, peer, pStorage, sStorage, client)
	mgr := manager.(*CliManager)

	peers, err := mgr.pStorage.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(peers) != 0 {
		t.Errorf("expected 0 peers, got %d", len(peers))
	}
}

func TestCliManager_ListPeers_WithPeers(t *testing.T) {
	logger := createTestLogger(t)
	ctx := context.Background()
	peer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	defer sStorage.Close()
	client := &mockClient{}

	testPeers := []domain.Peer{
		{ID: uuid.New(), Name: "peer1", Port: 8081},
		{ID: uuid.New(), Name: "peer2", Port: 8082},
	}
	for _, p := range testPeers {
		if err := pStorage.Add(p); err != nil {
			t.Fatalf("Add returned error: %v", err)
		}
	}

	manager := NewCliManager(ctx, logger, peer, pStorage, sStorage, client)
	mgr := manager.(*CliManager)

	peers, err := mgr.pStorage.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(peers) != 2 {
		t.Errorf("expected 2 peers, got %d", len(peers))
	}
}

func TestCliManager_ListSessions_Empty(t *testing.T) {
	logger := createTestLogger(t)
	ctx := context.Background()
	peer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	defer sStorage.Close()
	client := &mockClient{}

	manager := NewCliManager(ctx, logger, peer, pStorage, sStorage, client)
	mgr := manager.(*CliManager)

	sessions, err := mgr.sStorage.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestCliManager_GetPeerByID_NotFound(t *testing.T) {
	logger := createTestLogger(t)
	ctx := context.Background()
	peer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	defer sStorage.Close()
	client := &mockClient{}

	manager := NewCliManager(ctx, logger, peer, pStorage, sStorage, client)
	mgr := manager.(*CliManager)

	_, err := mgr.pStorage.GetByID(uuid.New())
	if err == nil {
		t.Error("expected error for non-existent peer ID")
	}
}

func TestCliManager_GetSessionByID_NotFound(t *testing.T) {
	logger := createTestLogger(t)
	ctx := context.Background()
	peer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	defer sStorage.Close()
	client := &mockClient{}

	manager := NewCliManager(ctx, logger, peer, pStorage, sStorage, client)
	mgr := manager.(*CliManager)

	_, err := mgr.sStorage.GetByID(uuid.New())
	if err == nil {
		t.Error("expected error for non-existent session ID")
	}
}

func TestMockSession_Interface(t *testing.T) {
	id := uuid.New()
	peer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	now := time.Now()
	readCh := make(<-chan *pb.Message)
	writeCh := make(chan<- *pb.Message)

	mock := &mockSession{
		id:       id,
		peer:     peer,
		incoming: true,
		lastDial: now,
		open:     true,
		readCh:   readCh,
		writeCh:  writeCh,
	}

	if mock.GetID() != id {
		t.Errorf("GetID: expected %v, got %v", id, mock.GetID())
	}
	if mock.GetLastDial() != now {
		t.Errorf("GetLastDial: expected %v, got %v", now, mock.GetLastDial())
	}
	if !mock.IsIncoming() {
		t.Error("IsIncoming: expected true")
	}
	if !mock.IsOpen() {
		t.Error("IsOpen: expected true")
	}

	ctx := context.Background()
	rCh, err := mock.GetReadChannel(ctx)
	if err != nil {
		t.Errorf("GetReadChannel returned error: %v", err)
	}
	if rCh != readCh {
		t.Error("GetReadChannel: channels don't match")
	}

	wCh, err := mock.GetWriteChannel(ctx)
	if err != nil {
		t.Errorf("GetWriteChannel returned error: %v", err)
	}
	if wCh != writeCh {
		t.Error("GetWriteChannel: channels don't match")
	}

	mock.Close(ctx)
	if mock.IsOpen() {
		t.Error("Close: expected session to be closed")
	}
}
