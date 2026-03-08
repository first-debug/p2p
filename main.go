package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/first-debug/p2p/internal/config"
	"github.com/first-debug/p2p/internal/domain"
	udpexplorer "github.com/first-debug/p2p/internal/explorer/udp"
	"github.com/first-debug/p2p/internal/server/websocket"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer/memory"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session/memory"
	"github.com/google/uuid"
)

var selfInfo domain.Peer

func main() {
	cfg := config.MustLoad()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	selfInfo = domain.Peer{
		ID:   uuid.New(),
		Port: cfg.WebSocketPort,
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)

	s := websocket.NewWebSocketServer(logger, cfg.WebSocketPort, sStorage, pStorage)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- s.Serve()
	}()

	explorer, err := udpexplorer.NewUDPExplorer(cfg, logger, &pb.Peer{
		ID:   pb.ToPbUUID(selfInfo.ID),
		Port: int32(cfg.WebSocketPort),
	}, pStorage)
	if err != nil {
		logger.Error("cannot start Explorer", slog.String("error", err.Error()))
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
		logger.Error("failed to serve", slog.String("error", err.Error()))
	case sig := <-sigs:
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		s.Stop(ctx)
		logger.Info("terminating", slog.Any("signal", sig))
	}
}
