package udp

import (
	"log/slog"
	"net"
	"os"
	"testing"

	"github.com/first-debug/p2p/internal/config"
	"github.com/first-debug/p2p/internal/domain"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer/memory"
	"github.com/google/uuid"
)

func createTestLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

func findFreeUDPPort(t *testing.T) int {
	t.Helper()
	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to resolve UDP addr: %v", err)
	}
	l, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer l.Close()
	return l.LocalAddr().(*net.UDPAddr).Port
}

func TestNewUDPExplorer_InvalidMulticastAddress(t *testing.T) {
	logger := createTestLogger(t)
	port := findFreeUDPPort(t)

	cfg := &config.Config{
		MulticastAddress:       "invalid-address",
		MulticastPort:          port,
		MulticastInterfaceName: "lo",
	}

	peerInfo := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}

	peerStorage := peerstorage.NewMemoryPeerStorage(logger)
	defer peerStorage.Close()

	_, err := NewUDPExplorer(cfg, logger, peerInfo, peerStorage)
	if err == nil {
		t.Error("expected error for invalid multicast address")
	}
}

func TestNewUDPExplorer_Success(t *testing.T) {
	t.Skip("Skipping UDP explorer test - requires specific network configuration")

	logger := createTestLogger(t)
	port := findFreeUDPPort(t)

	cfg := &config.Config{
		MulticastAddress:       "235.5.5.11",
		MulticastPort:          port,
		MulticastInterfaceName: "lo",
	}

	peerInfo := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}

	peerStorage := peerstorage.NewMemoryPeerStorage(logger)

	explorer, err := NewUDPExplorer(cfg, logger, peerInfo, peerStorage)
	if err != nil {
		t.Fatalf("cannot create UDP explorer: %v", err)
	}

	if explorer == nil {
		t.Fatal("NewUDPExplorer returned nil")
	}

	explorer.Stop()
}

func TestUDPExplorer_Emit(t *testing.T) {
	t.Skip("Skipping UDP explorer test - requires specific network configuration")

	logger := createTestLogger(t)
	port := findFreeUDPPort(t)

	cfg := &config.Config{
		MulticastAddress:       "235.5.5.11",
		MulticastPort:          port,
		MulticastInterfaceName: "lo",
	}

	peerInfo := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}

	peerStorage := peerstorage.NewMemoryPeerStorage(logger)

	explorer, err := NewUDPExplorer(cfg, logger, peerInfo, peerStorage)
	if err != nil {
		t.Fatalf("cannot create UDP explorer: %v", err)
	}

	err = explorer.Emit()
	if err != nil {
		t.Errorf("Emit returned error: %v", err)
	}

	explorer.Stop()
}

func TestUDPExplorer_Stop(t *testing.T) {
	t.Skip("Skipping UDP explorer test - requires specific network configuration")

	logger := createTestLogger(t)
	port := findFreeUDPPort(t)

	cfg := &config.Config{
		MulticastAddress:       "235.5.5.11",
		MulticastPort:          port,
		MulticastInterfaceName: "lo",
	}

	peerInfo := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}

	peerStorage := peerstorage.NewMemoryPeerStorage(logger)

	explorer, err := NewUDPExplorer(cfg, logger, peerInfo, peerStorage)
	if err != nil {
		t.Fatalf("cannot create UDP explorer: %v", err)
	}

	explorer.Stop()
}

func TestUDPExplorer_Emit_AfterStop(t *testing.T) {
	t.Skip("Skipping UDP explorer test - requires specific network configuration")

	logger := createTestLogger(t)
	port := findFreeUDPPort(t)

	cfg := &config.Config{
		MulticastAddress:       "235.5.5.11",
		MulticastPort:          port,
		MulticastInterfaceName: "lo",
	}

	peerInfo := domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}

	peerStorage := peerstorage.NewMemoryPeerStorage(logger)

	explorer, err := NewUDPExplorer(cfg, logger, peerInfo, peerStorage)
	if err != nil {
		t.Fatalf("cannot create UDP explorer: %v", err)
	}

	explorer.Stop()

	err = explorer.Emit()
	if err == nil {
		t.Error("expected error when emitting after stop")
	}
}
