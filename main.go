package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	ws "github.com/first-debug/p2p/internal/client/websocket"
	"github.com/first-debug/p2p/internal/config"
	"github.com/first-debug/p2p/internal/domain"
	updexplorer "github.com/first-debug/p2p/internal/explorer/UPDExplorer"
	pb "github.com/first-debug/p2p/internal/proto"
	"github.com/first-debug/p2p/internal/server/websocket"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer-storage/memory"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session-storage/memory"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
)

func main() {
	cfg := config.MustLoad()

	pStorage := peerstorage.NewMemoryPeerStorage()
	sStorage := sessionstorage.NewMemorySessionStorage()

	s := websocket.NewWebSocketServer(cfg.WebSocketPort, sStorage, pStorage)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- s.Serve()
	}()

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

	pStorage.Add(domain.Peer{
		ID:   uuid.Max[:],
		IP:   net.ParseIP("192.168.0.175"),
		Port: 8001,
	})

	client := ws.NewWebSocketClient(sStorage)
	sess, err := client.Connect(context.Background(), &domain.Peer{
		ID:   uuid.Max[:],
		IP:   net.ParseIP("192.168.0.175"),
		Port: 8001,
	})
	if err != nil {
		log.Printf("%v", err)
	} else {
		ch, err := sess.GetWriteChannel(context.Background())
		if err != nil {
			log.Printf("%v", err)
		} else {
			*ch <- &pb.Message{
				SendTime: timestamppb.Now(),
				Message:  "pipapopa",
			}
		}
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-serverErr:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigs:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.Stop(ctx)
		log.Printf("terminating: %v", sig)
	}
}
