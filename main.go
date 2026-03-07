package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/first-debug/p2p/internal/config"
	updexplorer "github.com/first-debug/p2p/internal/explorer/udp"
	pb "github.com/first-debug/p2p/internal/proto"
	"github.com/first-debug/p2p/internal/server/websocket"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer/memory"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session/memory"

	"github.com/google/uuid"
)

func main() {
	cfg := config.MustLoad()

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	pStorage := peerstorage.NewMemoryPeerStorage()
	sStorage := sessionstorage.NewMemorySessionStorage()

	s := websocket.NewWebSocketServer(cfg.WebSocketPort, sStorage, pStorage)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- s.Serve()
	}()

	explorer, err := updexplorer.NewUDPExplorer(cfg, &pb.Peer{
		ID:   pb.ToPbUUID(uuid.New()),
		Port: int32(cfg.WebSocketPort),
	}, pStorage)
	if err != nil {
		log.Printf("%v", err)
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		s.Stop(ctx)

		return
	}

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				explorer.Emit()
			case <-ctx.Done():
				return
			}
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-serverErr:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigs:
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		s.Stop(ctx)
		log.Printf("terminating: %v", sig)
	}
}
