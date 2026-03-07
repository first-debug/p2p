package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	ws "github.com/first-debug/p2p/internal/client/websocket"
	"github.com/first-debug/p2p/internal/config"
	udpexplorer "github.com/first-debug/p2p/internal/explorer/udp"
	"github.com/first-debug/p2p/internal/manager/cli"
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

	explorer, err := udpexplorer.NewUDPExplorer(cfg, &pb.Peer{
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

	client := ws.NewWebSocketClient(selfInfo, sStorage)

	mgr := cli.NewCliManager(ctx, selfInfo, pStorage, sStorage, client)

	mgrErr := make(chan error)
	go func() {
		mgrErr <- mgr.Run()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-serverErr:
		log.Printf("failed to serve: %v", err)
	case err := <-mgrErr:
		log.Printf("manager exit status: %v", err)
	case sig := <-sigs:
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		s.Stop(ctx)
		log.Printf("terminating: %v", sig)
	}
}
