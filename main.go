package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	ws "github.com/first-debug/p2p/internal/client/websocket"
	"github.com/first-debug/p2p/internal/config"
	"github.com/first-debug/p2p/internal/domain"
	udpexplorer "github.com/first-debug/p2p/internal/explorer/udp"
	"github.com/first-debug/p2p/internal/manager/cli"
	"github.com/first-debug/p2p/internal/server/websocket"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer/memory"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session/memory"
	"github.com/google/uuid"
)

var selfInfo domain.Peer

func main() {
	cfg := config.MustLoad()

	fmt.Printf("Using config directory: %v\n", cfg.ConfigDir)

	logFile, err := os.OpenFile(cfg.ConfigDir+"log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	selfInfo = domain.Peer{
		Port: cfg.WebSocket.Port,
	}

	idFile := cfg.ConfigDir + "id"
	if _, err := os.Stat(idFile); os.IsNotExist(err) {
		selfInfo.ID = uuid.New()
		if err := os.WriteFile(idFile, []byte(selfInfo.ID.String()), 0o600); err != nil {
			panic(err)
		}
	} else {
		idFile, err := os.OpenFile(idFile, os.O_RDONLY, 0o600)
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

	s := websocket.NewWebSocketServer(logger, cfg.WebSocket.Port, sStorage, pStorage)

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
		ticker := time.NewTicker(cfg.Explorer.Period)
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

	client := ws.NewWebSocketClient(logger, selfInfo, sStorage)

	mgr := cli.NewCliManager(ctx, logger, selfInfo, pStorage, sStorage, client)

	mgrErr := make(chan error)
	go func() {
		mgrErr <- mgr.Run()
	}()

	defer fmt.Fprint(logFile, "--- ", time.Now(), " ---\n")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-serverErr:
		logger.Error("failed to serve", slog.String("error", err.Error()))
	case err := <-mgrErr:
		if err != nil {
			logger.Error("manager exit with error", slog.String("error", err.Error()))
			break
		}
		logger.Info("close manager")
	case sig := <-sigs:
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		s.Stop(ctx)
		logger.Info("terminating", slog.Any("signal", sig))
	}
}
