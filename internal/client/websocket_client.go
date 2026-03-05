// Package client предоставляет клиент для подключения к пирам.
package client

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/first-debug/p2p/internal/domain"
	wsserver "github.com/first-debug/p2p/internal/server/websocket"
	"github.com/first-debug/p2p/internal/session"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session-storage"

	"github.com/coder/websocket"
)

// WebSocketClient реализует интерфейс Client для подключения к пирам через WebSocket.
type WebSocketClient struct {
	sessionStorage sessionstorage.SessionStorage
}

// NewWebSocketClient создаёт новый WebSocket клиент.
func NewWebSocketClient(sessionStorage sessionstorage.SessionStorage) *WebSocketClient {
	return &WebSocketClient{
		sessionStorage: sessionStorage,
	}
}

// Connect подключается к указанному пиру и создаёт новую сессию.
func (c *WebSocketClient) Connect(ctx context.Context, peer domain.Peer) (session.Session, error) {
	// Формируем URL для подключения
	url := fmt.Sprintf("ws://%s:%d/ws", peer.IP.String(), peer.Port)

	// Создаём WebSocket подключение
	conn, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{
		HTTPHeader: map[string][]string{
			"PeerID": {string(peer.ID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer: %w", err)
	}

	// Создаём новую сессию
	newSession := wsserver.NewWebSocketSession(conn, &peer, false, time.Now())

	// Сохраняем сессию в хранилище
	if c.sessionStorage != nil {
		if err := c.sessionStorage.Add(newSession); err != nil {
			conn.Close(websocket.StatusInternalError, "failed to add session")
			return nil, fmt.Errorf("failed to add session to storage: %w", err)
		}
	}

	// Запускаем горутину для чтения сообщений
	// Приводим к конкретному типу для доступа к методам Read/Write
	if wsSession, ok := newSession.(*wsserver.WebSocketSession); ok {
		go wsSession.Read(ctx)
		go wsSession.Write(ctx)
	}

	return newSession, nil
}

// Dial создаёт новое подключение к пиру с использованием TCP.
func (c *WebSocketClient) Dial(ctx context.Context, peer domain.Peer) (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", peer.IP.String(), peer.Port)

	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial peer: %w", err)
	}

	return conn, nil
}
