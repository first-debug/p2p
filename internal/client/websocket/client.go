package websocket

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/coder/websocket"
	"github.com/first-debug/p2p/internal/client"
	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/session"
	wssession "github.com/first-debug/p2p/internal/session/websocket"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session"
)

type WebSocketClient struct {
	logger   *slog.Logger
	sStorage sessionstorage.SessionStorage
	selfInfo domain.Peer
}

func NewWebSocketClient(log *slog.Logger, peer domain.Peer, sStorage sessionstorage.SessionStorage) client.Client {
	return &WebSocketClient{
		logger:   log.With("module", "WebSocketClient"),
		selfInfo: peer,
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
			"PeerID": {c.selfInfo.ID.String()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	newSession := wssession.NewWebSocketSession(c.logger, conn, peer, false, time.Now())
	if c.sStorage != nil {
		if err := c.sStorage.Add(newSession); err != nil {
			conn.Close(websocket.StatusInternalError, "failed to add session")
			return nil, fmt.Errorf("failed to add session to storage: %w", err)
		}
	}
	return newSession, nil
}
