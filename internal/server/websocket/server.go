package websocket

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/first-debug/p2p/internal/server"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session"
	"github.com/google/uuid"

	"github.com/coder/websocket"
)

var logger log.Logger = *log.New(os.Stderr, "[WebSocketServer] ", log.LstdFlags)

type WebSocketServer struct {
	server *http.Server
	port   int

	serveMux http.ServeMux

	sessionsStorage sessionstorage.SessionStorage
	peerStorage     peerstorage.PeerStorage

	ctx        context.Context
	ctxCancel  context.CancelFunc
	wg         *sync.WaitGroup
	isStopping atomic.Bool
}

func NewWebSocketServer(port int, sessionsStorage sessionstorage.SessionStorage, peerStorage peerstorage.PeerStorage) server.Server {
	ctx, cancel := context.WithCancel(context.Background())
	s := &WebSocketServer{
		port:            port,
		sessionsStorage: sessionsStorage,
		peerStorage:     peerStorage,
		ctx:             ctx,
		ctxCancel:       cancel,
		wg:              &sync.WaitGroup{},
	}
	s.server = &http.Server{
		Handler:      &s.serveMux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	s.serveMux.HandleFunc("/ws", s.messageHandle)
	s.isStopping.Store(false)
	return s
}

func (s *WebSocketServer) Serve() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	logger.Printf("listening on ws://%v", l.Addr())
	return s.server.Serve(l)
}

func (s *WebSocketServer) Stop(ctx context.Context) {
	s.isStopping.Store(true)
	s.ctxCancel()

	localCtx, ctxCancel := context.WithTimeout(ctx, time.Second*5)
	defer ctxCancel()

	// s.sessionsStorage.CloseAllByType(true)

	end := make(chan any)
	go func() {
		s.wg.Wait()
		end <- struct{}{}
	}()

	select {
	case <-end:
	case <-localCtx.Done():
	}
	if err := s.server.Shutdown(localCtx); err != nil {
		logger.Fatalf("%e", err)
	}
}

func (s *WebSocketServer) messageHandle(w http.ResponseWriter, r *http.Request) {
	if s.isStopping.Load() {
		w.WriteHeader(http.StatusNotAcceptable)
		if _, err := w.Write([]byte("server is stopping")); err != nil {
			logger.Fatalf("cannot emit about server status: %e", err)
		}
		return
	}
	logger.Println("start handler")
	peerID := r.Header.Get("PeerID")
	peer, err := s.peerStorage.GetByID([]byte(peerID))
	if err != nil {
		logger.Printf("%v", err)
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		logger.Printf("%v", err)
		return
	}
	newS := NewWebSocketSession(c, &peer, true, time.Now())
	s.sessionsStorage.Add(newS)
}
