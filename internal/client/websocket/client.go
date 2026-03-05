package client

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	wsserver "github.com/first-debug/p2p/internal/server/websocket"

	"github.com/coder/websocket"
	"github.com/first-debug/p2p/internal/client"
	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/session"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session-storage"
)

var logger log.Logger = *log.New(os.Stderr, "[WebSocketClient] ", log.LstdFlags)

type WebSocketClient struct {
	sStorage sessionstorage.SessionStorage
}

func NewWebSocketClient(sStorage sessionstorage.SessionStorage) client.Client {
	return &WebSocketClient{
		sStorage: sStorage,
	}
}

func (c *WebSocketClient) Connect(ctx context.Context, peer *domain.Peer) (session.Session, error) {
	url := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", peer.IP.String(), peer.Port),
		Path:   "/ws",
	}
	conn, _, err := websocket.Dial(ctx, url.String(), &websocket.DialOptions{
		HTTPHeader: map[string][]string{
			"PeerID": {string(peer.ID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	newSession := wsserver.NewWebSocketSession(conn, peer, false, time.Now())
	if c.sStorage != nil {
		if err := c.sStorage.Add(newSession); err != nil {
			conn.Close(websocket.StatusInternalError, "failed to add session")
			return nil, fmt.Errorf("failed to add session to storage: %w", err)
		}
	}
	return newSession, nil
}
