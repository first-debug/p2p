package cli

import (
	"bufio"
	"context"
	"log"
	"os"

	client "github.com/first-debug/p2p/internal/client"
	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/manager"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session"
)

var logger log.Logger = *log.New(os.Stderr, "[CliManager] ", log.LstdFlags)

type CliManager struct {
	ctx      context.Context
	selfInfo domain.Peer
	pStorage peerstorage.PeerStorage
	sStorage sessionstorage.SessionStorage
	client   client.Client
}

func NewCliManager(ctx context.Context, peer domain.Peer, p peerstorage.PeerStorage, s sessionstorage.SessionStorage, c client.Client) manager.Manager {
	return &CliManager{
		ctx:      ctx,
		selfInfo: peer,
		pStorage: p,
		sStorage: s,
		client:   c,
	}
}

func (m *CliManager) Run() error {
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	for {
		writer.Flush()
		select {
		case <-m.ctx.Done():
			return nil
		default:
			if scanner.Scan() {
			}
		}
	}
}
