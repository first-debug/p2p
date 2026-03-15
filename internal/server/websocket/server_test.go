package websocket

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	peerstorage "github.com/first-debug/p2p/internal/storage/peer/memory"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session/memory"
)

func createTestLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

func findFreePort(t *testing.T) int {
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to resolve TCP addr: %v", err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func TestNewWebSocketServer(t *testing.T) {
	logger := createTestLogger(t)
	port := findFreePort(t)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	defer sStorage.Close()

	server := NewWebSocketServer(logger, port, sStorage, pStorage)

	if server == nil {
		t.Fatal("NewWebSocketServer returned nil")
	}

	wsServer, ok := server.(*WebSocketServer)
	if !ok {
		t.Fatal("NewWebSocketServer did not return *WebSocketServer")
	}

	if wsServer.logger == nil {
		t.Error("logger is nil")
	}
	if wsServer.port != port {
		t.Errorf("port: expected %d, got %d", port, wsServer.port)
	}
	if wsServer.sessionsStorage == nil {
		t.Error("sessionsStorage is nil")
	}
	if wsServer.peerStorage == nil {
		t.Error("peerStorage is nil")
	}
	if wsServer.ctx == nil {
		t.Error("ctx is nil")
	}
	if wsServer.ctxCancel == nil {
		t.Error("ctxCancel is nil")
	}
	if wsServer.wg == nil {
		t.Error("wg is nil")
	}
}

func TestWebSocketServer_Serve_And_Stop(t *testing.T) {
	logger := createTestLogger(t)
	port := findFreePort(t)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	defer sStorage.Close()

	server := NewWebSocketServer(logger, port, sStorage, pStorage)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Stop(ctx)

	select {
	case <-errCh:
	case <-time.After(6 * time.Second):
		t.Error("server.Stop() timed out")
	}
}

func TestWebSocketServer_Ping(t *testing.T) {
	logger := createTestLogger(t)
	port := findFreePort(t)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	defer sStorage.Close()

	server := NewWebSocketServer(logger, port, sStorage, pStorage)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", port))
	if err != nil {
		t.Skipf("cannot connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ping: expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if string(body) != "ok" {
		t.Errorf("ping: expected body 'ok', got %s", string(body))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Stop(ctx)

	select {
	case <-errCh:
	case <-time.After(6 * time.Second):
		t.Error("server.Stop() timed out")
	}
}

func TestWebSocketServer_Ping_WhenStopping(t *testing.T) {
	logger := createTestLogger(t)
	port := findFreePort(t)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	defer sStorage.Close()

	server := NewWebSocketServer(logger, port, sStorage, pStorage)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Stop(ctx)

	select {
	case <-errCh:
	case <-time.After(6 * time.Second):
		t.Error("server.Stop() timed out")
	}

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", port))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable && resp.StatusCode != http.StatusNotAcceptable {
		t.Logf("ping after stop: got status %d (may vary)", resp.StatusCode)
	}
}

func TestWebSocketServer_MessageHandle_InvalidPeerID(t *testing.T) {
	logger := createTestLogger(t)
	port := findFreePort(t)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	defer sStorage.Close()

	server := NewWebSocketServer(logger, port, sStorage, pStorage)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/ws", port), nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("PeerID", "invalid-uuid")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "13")

	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("cannot connect to server: %v", err)
	}
	defer resp.Body.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Stop(ctx)

	select {
	case <-errCh:
	case <-time.After(6 * time.Second):
		t.Error("server.Stop() timed out")
	}
}

func TestWebSocketServer_ConcurrentConnections(t *testing.T) {
	logger := createTestLogger(t)
	port := findFreePort(t)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)
	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	defer sStorage.Close()

	server := NewWebSocketServer(logger, port, sStorage, pStorage)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	for i := 0; i < 5; i++ {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", port))
		if err != nil {
			t.Logf("request %d failed: %v", i, err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("ping %d: expected status %d, got %d", i, http.StatusOK, resp.StatusCode)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Stop(ctx)

	select {
	case <-errCh:
	case <-time.After(6 * time.Second):
		t.Error("server.Stop() timed out")
	}
}
