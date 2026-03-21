package websocket

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/coder/websocket"
	"github.com/first-debug/p2p/internal/client"
	"github.com/first-debug/p2p/internal/domain"
	pb "github.com/first-debug/p2p/internal/proto"
	"github.com/first-debug/p2p/internal/session"
	wssession "github.com/first-debug/p2p/internal/session/websocket"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session"
	"google.golang.org/protobuf/proto"
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

func (c *WebSocketClient) GetKnownPeers(ctx context.Context, peer *domain.Peer) ([]domain.Peer, error) {
	url := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", peer.IP.String(), peer.Port),
		Path:   "/peers",
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	req.Header.Add("PeerID", c.selfInfo.ID.String())

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	if res.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("access denied")
	}
	if res == nil {
		c.logger.Error("response is nil")
		return nil, fmt.Errorf("response is nil")
	}
	if res.Body == nil {
		c.logger.Error("response Body is nil")
		return nil, fmt.Errorf("response Body is nil")
	}
	if res.ContentLength <= 0 {
		c.logger.Error("ContentLength is unknown")
		return nil, fmt.Errorf("ContentLength is unknown")
	}

	data := make([]byte, res.ContentLength)

	pbPeers := &pb.KnownPeers{
		Peers: make([]*pb.Peer, 0),
	}

	n, err := res.Body.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		c.logger.Error("cannot read peers list", slog.Any("error", err.Error()))
		return nil, err
	}
	if int64(n) != res.ContentLength {
		errMsg := "the length of the ContentLength and the read information are not equal"
		c.logger.Error(errMsg, slog.Int64("ContentLength", res.ContentLength), slog.Int("read len", n))
		return nil, fmt.Errorf("%s", errMsg)
	}

	err = proto.Unmarshal(data, pbPeers)
	if err != nil {
		c.logger.Error("cannot marshal peers list", slog.Any("error", err.Error()))
		return nil, err
	}

	peers := make([]domain.Peer, 0)
	for _, v := range pbPeers.Peers {
		if peer.ID == c.selfInfo.ID {
			continue
		}
		peer := pb.PbPeerToDomain(v)
		peers = append(peers, peer)
	}

	return peers, nil
}
