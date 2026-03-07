package websocket

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/first-debug/p2p/internal/domain"
	pb "github.com/first-debug/p2p/internal/proto"
	"github.com/first-debug/p2p/internal/session"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
)

type WebSocketSession struct {
	session.BaseSession

	ctx       context.Context
	ctxCancel context.CancelFunc
	wg        sync.WaitGroup

	connection  *websocket.Conn
	rateLimiter *rate.Limiter
	readChan    chan *pb.Message
	writeChan   chan *pb.Message
}

// NewWebSocketSession принимает указатель на УЖЕ готовое подключение [websocket.Conn] и остальные необходимые данные для создания сессии [WebSocketSession]
func NewWebSocketSession(conn *websocket.Conn, peer *domain.Peer, incoming bool, lastDial time.Time) session.Session {
	logger.Println("Create new session")
	ctx, cancel := context.WithCancel(context.Background())
	ws := &WebSocketSession{
		ctx:         ctx,
		ctxCancel:   cancel,
		connection:  conn,
		rateLimiter: rate.NewLimiter(rate.Every(time.Millisecond*100), 10),
		readChan:    make(chan *pb.Message, 20),
		writeChan:   make(chan *pb.Message, 20),
	}
	ws.ID = uuid.New()
	ws.Peer = *peer
	ws.Incoming = incoming
	ws.LastDial = lastDial

	logger.Printf("Created new incoming session with ID = %v, Peer = {%v, %v, %d}, lastDial = %v",
		ws.ID,
		ws.Peer.ID, ws.Peer.IP, ws.Peer.Port,
		ws.LastDial,
	)

	ws.wg.Go(ws.read)
	ws.wg.Go(ws.write)

	return ws
}

func (s *WebSocketSession) GetID() uuid.UUID {
	return s.ID
}

func (s *WebSocketSession) GetLastDial() time.Time {
	return s.LastDial
}

func (s *WebSocketSession) GetReadChannel(context.Context) (<-chan *pb.Message, error) {
	if s.readChan == nil {
		return nil, errors.New("read channel is nil")
	}
	return s.readChan, nil
}

func (s *WebSocketSession) GetWriteChannel(context.Context) (chan<- *pb.Message, error) {
	if s.writeChan == nil {
		return nil, errors.New("write channel is nil")
	}
	return s.writeChan, nil
}

func (s *WebSocketSession) IsIncoming() bool {
	return s.Incoming
}

func (s *WebSocketSession) IsOpen() bool {
	return s.connection != nil
}

func (s *WebSocketSession) Close(ctx context.Context) {
	s.ctxCancel()
	if s.connection != nil {
		s.connection.Close(websocket.StatusNormalClosure, "manual close")
		s.connection = nil
	}
	s.wg.Wait()
}

func (s *WebSocketSession) read() {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			if s.connection == nil {
				logger.Println("connections closed")
				return
			}

			msg := &pb.Message{}

			// ctx, cancel := context.WithTimeout(s.ctx, time.Second*10)
			// defer cancel()

			// err := s.rateLimiter.Wait(ctx)
			// if err != nil {
			// 	s.closeWithError(err)
			// 	logger.Printf("%v", err)
			// 	return
			// }

			typ, data, err := s.connection.Read(s.ctx)
			if err != nil {
				s.closeWithError(err)
				logger.Printf("%v", err)
				return
			}
			if typ != websocket.MessageBinary {
				errMsg := "unsupported message type"
				s.closeWithError(err)
				logger.Printf("%v", errMsg)
				return
			}

			if err := proto.Unmarshal(data, msg); err != nil {
				logger.Printf("%e", err)
				s.closeWithError(err)
			}
			s.LastDial = time.Now()
			s.readChan <- msg
		}
	}
}

func (s *WebSocketSession) write() {
	for {
		select {
		case <-s.ctx.Done():
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

			ctx, cancel := context.WithTimeout(s.ctx, time.Second*10)
			defer cancel()

			// err := s.rateLimiter.Wait(ctx)
			// if err != nil {
			// 	s.closeWithError(err)
			// 	logger.Printf("%v", err)
			// 	return
			// }

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
			s.LastDial = time.Now()
		}
	}
}

func (s *WebSocketSession) closeWithError(err error) {
	s.ctxCancel()
	if s.connection != nil {
		s.connection.Close(websocket.StatusInternalError, fmt.Sprintf("internal error: %v", err))
		s.connection = nil
	}
	s.wg.Wait()
}
