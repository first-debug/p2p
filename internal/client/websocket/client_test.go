package websocket

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/first-debug/p2p/internal/domain"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session/memory"
	"github.com/google/uuid"
)

func TestNewWebSocketClient(t *testing.T) {
	logger := createTestLogger(t)
	peer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	defer sStorage.Close()

	client := NewWebSocketClient(logger, peer, sStorage)

	if client == nil {
		t.Fatal("NewWebSocketClient returned nil")
	}

	wsClient, ok := client.(*WebSocketClient)
	if !ok {
		t.Fatal("NewWebSocketClient did not return *WebSocketClient")
	}

	if wsClient.logger == nil {
		t.Error("logger is nil")
	}
	if wsClient.sStorage == nil {
		t.Error("sStorage is nil")
	}
	if wsClient.selfInfo.ID != peer.ID {
		t.Errorf("selfInfo.ID: expected %v, got %v", peer.ID, wsClient.selfInfo.ID)
	}
}

func TestWebSocketClient_Connect_InvalidAddress(t *testing.T) {
	logger := createTestLogger(t)
	selfPeer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	defer sStorage.Close()

	client := NewWebSocketClient(logger, selfPeer, sStorage)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	peer := &domain.Peer{
		ID:   uuid.New(),
		IP:   net.ParseIP("127.0.0.1"),
		Port: 59999,
	}

	_, err := client.Connect(ctx, peer)
	if err == nil {
		t.Error("expected error when connecting to invalid address")
	}
}

func TestWebSocketClient_Connect_NilStorage(t *testing.T) {
	logger := createTestLogger(t)
	selfPeer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}

	client := NewWebSocketClient(logger, selfPeer, nil)

	if client == nil {
		t.Fatal("NewWebSocketClient returned nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	peer := &domain.Peer{
		ID:   uuid.New(),
		IP:   net.ParseIP("127.0.0.1"),
		Port: 59998,
	}

	_, err := client.Connect(ctx, peer)
	if err == nil {
		t.Error("expected error when connecting to invalid address")
	}
}

func TestWebSocketClient_Connect_ContextCancellation(t *testing.T) {
	logger := createTestLogger(t)
	selfPeer := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	defer sStorage.Close()

	client := NewWebSocketClient(logger, selfPeer, sStorage)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	peer := &domain.Peer{
		ID:   uuid.New(),
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8081,
	}

	_, err := client.Connect(ctx, peer)
	if err == nil {
		t.Error("expected error when context is canceled")
	}
}
