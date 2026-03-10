package main

import (
	"context"
	"fmt"
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

	fmt.Printf("Using cache directory: %v\n", cfg.CachePath)

	logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	selfInfo = domain.Peer{
		Port: cfg.WebSocketPort,
	}

	if _, err := os.Stat(cfg.IDFile); os.IsNotExist(err) {
		selfInfo.ID = uuid.New()
		if err := os.WriteFile(cfg.IDFile, []byte(selfInfo.ID.String()), 0o600); err != nil {
			panic(err)
		}
	} else {
		idFile, err := os.OpenFile(cfg.IDFile, os.O_RDONLY, 0o600)
		if err != nil {
			panic(err)
		}

		bytes := make([]byte, 36)
		_, err = idFile.Read(bytes)
		if err != nil {
			panic(err)
		}

		id, err := uuid.ParseBytes(bytes)
		if err != nil {
			panic(err)
		}

		selfInfo.ID = id

		idFile.Close()
	}

	fmt.Printf("Self ID: %v\n", selfInfo.ID)

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	pStorage := peerstorage.NewMemoryPeerStorage(logger)
	sStorage := sessionstorage.NewMemorySessionStorage(logger)

	s := websocket.NewWebSocketServer(logger, cfg.WebSocketPort, sStorage, pStorage)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- s.Serve()
	}()

	explorer, err := udpexplorer.NewUDPExplorer(cfg, logger, selfInfo, pStorage)
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
		fmt.Fprint(logFile, "--- ", time.Now(), " ---\n")
	}
}
