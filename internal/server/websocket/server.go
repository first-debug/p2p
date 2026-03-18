package websocket

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/first-debug/p2p/internal/server"
	session "github.com/first-debug/p2p/internal/session/websocket"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session"
	"github.com/google/uuid"

	"github.com/coder/websocket"
)

type WebSocketServer struct {
	logger *slog.Logger

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

func NewWebSocketServer(log *slog.Logger, port int, sessionsStorage sessionstorage.SessionStorage, peerStorage peerstorage.PeerStorage) server.Server {
	ctx, cancel := context.WithCancel(context.Background())
	s := &WebSocketServer{
		logger:          log.With("module", "WebSocketServer"),
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

	s.serveMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if s.isStopping.Load() {
			w.WriteHeader(http.StatusNotAcceptable)
			if _, err := w.Write([]byte("server is stopping")); err != nil {
				s.logger.Error("cannot answer on ping", slog.String("http-error", err.Error()))
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	s.serveMux.HandleFunc("/ws",
		s.stopingMiddleware(s.messageHandle),
	)
	s.isStopping.Store(false)
	return s
}

func (s *WebSocketServer) Serve() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	s.logger.Info("start listening", slog.Any("address", l.Addr()))
	return s.server.Serve(l)
}

func (s *WebSocketServer) Stop(ctx context.Context) {
	s.isStopping.Store(true)
	s.ctxCancel()

	localCtx, ctxCancel := context.WithTimeout(ctx, time.Second*5)
	defer ctxCancel()

	s.sessionsStorage.CloseAllByType(localCtx, true)

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
		s.logger.Error("error during shutdown server", slog.String("error", err.Error()))
	}
}

func (s *WebSocketServer) messageHandle(w http.ResponseWriter, r *http.Request) {
	peerID, err := uuid.Parse(r.Header.Get("PeerID"))
	if err != nil {
		s.logger.Error("cannot extract PeerID from header", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusForbidden)
		return
	}
	peer, err := s.peerStorage.GetByID(peerID)
	if err != nil {
		s.logger.Error("cannot found Peer in local list", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusForbidden)
		return
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		s.logger.Error("cannot connect to peer", slog.String("error", err.Error()), slog.String("PeerID", peer.ID.String()))
		return
	}
	newS := session.NewWebSocketSession(s.logger, c, &peer, true, time.Now())
	if err = s.sessionsStorage.Add(newS); err != nil {
		s.logger.Error("cannot save new Session", slog.String("error", err.Error()))
	}
}
