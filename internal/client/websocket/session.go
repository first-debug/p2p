package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/first-debug/p2p/internal/domain"
	pb "github.com/first-debug/p2p/internal/proto"
	"github.com/first-debug/p2p/internal/session"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
)

type WebSocketSession struct {
	session.BaseSession

	connection  *websocket.Conn
	rateLimiter *rate.Limiter
	readChan    chan *pb.Message
	writeChan   chan *pb.Message
}

func NewWebSocketSession(conn *websocket.Conn, peer *domain.Peer, incoming bool, lastDial time.Time) session.Session {
	logger.Println("Create new session")
	ws := &WebSocketSession{
		connection:  conn,
		rateLimiter: rate.NewLimiter(rate.Every(time.Millisecond*100), 10),
		readChan:    make(chan *pb.Message, 20),
		writeChan:   make(chan *pb.Message, 20),
	}
	ws.ID = uuid.New()
	ws.Peer = peer
	ws.Incoming = incoming
	ws.LastDial = lastDial
	go ws.Read(context.Background())
	go ws.Write(context.Background())
	return ws
}

func (s *WebSocketSession) GetID() uuid.UUID {
	return s.ID
}

func (s *WebSocketSession) IsIncoming() bool {
	return s.Incoming
}

func (s *WebSocketSession) GetReadChannel(context.Context) (*chan *pb.Message, error) {
	if s.readChan == nil {
		return nil, errors.New("read channel is nil")
	}
	return &s.readChan, nil
}

func (s *WebSocketSession) GetWriteChannel(context.Context) (*chan *pb.Message, error) {
	if s.writeChan == nil {
		return nil, errors.New("write channel is nil")
	}
	return &s.writeChan, nil
func (s *WebSocketSession) Close(context.Context) {
	if s.connection != nil {
		s.connection.Close(websocket.StatusNormalClosure, "manual close")
		s.connection = nil
	}
}

func (s *WebSocketSession) Read(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if s.connection == nil {
				logger.Println("connections closed")
				return
			}

			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			msg := &pb.Message{}

			err := s.rateLimiter.Wait(ctx)
			if err != nil {
				s.closeWithError(err)
				logger.Printf("%v", err)
				return
			}

			var data []byte

			fmt.Println(s)
			typ, data, err := s.connection.Read(s.ctx)
			if err != nil {
				s.closeWithError(err)
				logger.Printf("%v", err)
				return
			}
			if typ != websocket.MessageBinary {
				errMsg := "unsupported message type"
				logger.Printf("%v", errMsg)
				s.closeWithError(errors.New(errMsg))
				return
			}

			if err := proto.Unmarshal(data, msg); err != nil {
				logger.Printf("%e", err)
				s.closeWithError(err)
			}
			s.readChan <- msg
		}
	}
}

func (s *WebSocketSession) Write(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-s.writeChan:
			if !ok {
				logger.Println("read channel closed")
				return
			}
			if s.connection == nil {
				logger.Println("connections closed")
				return
			}

			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			err := s.rateLimiter.Wait(ctx)
			if err != nil {
				s.closeWithError(err)
				logger.Printf("%v", err)
				return
			}

			data, err := proto.Marshal(msg)
			if err != nil {
				logger.Printf("%e", err)
				s.closeWithError(err)
				return
			}

			err = s.connection.Write(ctx, websocket.MessageBinary, data)
			if err != nil {
				logger.Printf("%e", err)
				s.closeWithError(err)
				return
			}
		}
	}
}


func (s *WebSocketSession) IsOpen() bool {
	return s.connection != nil
}

func (s *WebSocketSession) closeWithError(err error) {
	if s.connection != nil {
		s.connection.Close(websocket.StatusInternalError, fmt.Sprintf("internal error: %v", err))
		s.connection = nil
	}
}
