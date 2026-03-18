package client

import (
	"context"

	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/session"
)

type Client interface {
	Connect(context.Context, *domain.Peer) (session.Session, error)
	GetKnownPeers(context.Context, *domain.Peer) ([]domain.Peer, error)
}
