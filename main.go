package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"main/internal/config"
	updexplorer "main/internal/explorer/UPDExplorer"
	tuimanager "main/internal/manager/TuiManager"
	pb "main/internal/proto"
	"main/internal/server/websocket"
	peerstorage "main/internal/storage/peer-storage/memory"
	sessionstorage "main/internal/storage/session-storage/memory"

	"github.com/google/uuid"
)

func main() {
	cfg := config.MustLoad()

	sStorage := sessionstorage.NewMemorySessionStorage()
	s := websocket.NewWebSocketServer(cfg.WebSocketPort, sStorage)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- s.Serve()
	}()

	pStorage := peerstorage.NewMemoryPeerStorage()
	explorer, err := updexplorer.NewUDPExplorer(cfg, &pb.Peer{
		Id:   []byte(uuid.New().String()),
		Port: int32(cfg.WebSocketPort),
	}, pStorage)
	if err != nil {
		log.Printf("%v", err)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.Stop(ctx)

		return
	}

	manager, err := tuimanager.NewTuiManager(explorer, pStorage, sStorage)
	if err != nil {
		log.Printf("%v", err)
	}

	managerErr := make(chan error, 1)
	go func() {
		managerErr <- manager.Run()
	}()

	stop := make(chan any)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				explorer.Emit()
			case <-stop:
				return
			}
		}
	}()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-managerErr:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.Stop(ctx)
		log.Printf("manager stop status: %v", err)
	case err := <-serverErr:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigs:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.Stop(ctx)
		log.Printf("terminating: %v", sig)
	}
}
