package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/first-debug/p2p/internal/domain"
	pb "github.com/first-debug/p2p/internal/proto"
	"github.com/google/uuid"
)

func TestWebSocketSession_GetID(t *testing.T) {
	peer := &domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}

	s := &WebSocketSession{
		readChan:  make(chan *pb.Message, 10),
		writeChan: make(chan *pb.Message, 10),
	}
	s.ID = uuid.New()
	s.Peer = *peer
	s.Incoming = false
	s.LastDial = time.Now()

	id := s.GetID()
	if id != s.ID {
		t.Errorf("GetID: expected %v, got %v", s.ID, id)
	}
}

func TestWebSocketSession_GetLastDial(t *testing.T) {
	peer := &domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}
	now := time.Now()

	s := &WebSocketSession{
		readChan:  make(chan *pb.Message, 10),
		writeChan: make(chan *pb.Message, 10),
	}
	s.ID = uuid.New()
	s.Peer = *peer
	s.Incoming = false
	s.LastDial = now

	lastDial := s.GetLastDial()
	if lastDial != now {
		t.Errorf("GetLastDial: expected %v, got %v", now, lastDial)
	}
}

func TestWebSocketSession_GetReadChannel(t *testing.T) {
	peer := &domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}

	readCh := make(chan *pb.Message, 10)
	s := &WebSocketSession{
		readChan:  readCh,
		writeChan: make(chan *pb.Message, 10),
	}
	s.ID = uuid.New()
	s.Peer = *peer
	s.Incoming = false
	s.LastDial = time.Now()

	ch, err := s.GetReadChannel(context.Background())
	if err != nil {
		t.Fatalf("GetReadChannel returned error: %v", err)
	}
	if ch != readCh {
		t.Error("GetReadChannel: channels don't match")
	}

	sNil := &WebSocketSession{
		readChan:  nil,
		writeChan: make(chan *pb.Message, 10),
	}
	sNil.ID = uuid.New()
	sNil.Peer = *peer
	sNil.Incoming = false
	sNil.LastDial = time.Now()

	_, err = sNil.GetReadChannel(context.Background())
	if err == nil {
		t.Error("expected error for nil read channel")
	}
}

func TestWebSocketSession_GetWriteChannel(t *testing.T) {
	peer := &domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}

	writeCh := make(chan *pb.Message, 10)
	s := &WebSocketSession{
		readChan:  make(chan *pb.Message, 10),
		writeChan: writeCh,
	}
	s.ID = uuid.New()
	s.Peer = *peer
	s.Incoming = false
	s.LastDial = time.Now()

	ch, err := s.GetWriteChannel(context.Background())
	if err != nil {
		t.Fatalf("GetWriteChannel returned error: %v", err)
	}
	if ch != writeCh {
		t.Error("GetWriteChannel: channels don't match")
	}

	sNil := &WebSocketSession{
		readChan:  make(chan *pb.Message, 10),
		writeChan: nil,
	}
	sNil.ID = uuid.New()
	sNil.Peer = *peer
	sNil.Incoming = false
	sNil.LastDial = time.Now()

	_, err = sNil.GetWriteChannel(context.Background())
	if err == nil {
		t.Error("expected error for nil write channel")
	}
}

func TestWebSocketSession_GetPeerID(t *testing.T) {
	peerID := uuid.New()
	peer := &domain.Peer{
		ID:   peerID,
		Port: 8080,
	}

	s := &WebSocketSession{
		readChan:  make(chan *pb.Message, 10),
		writeChan: make(chan *pb.Message, 10),
	}
	s.ID = uuid.New()
	s.Peer = *peer
	s.Incoming = false
	s.LastDial = time.Now()

	id := s.GetPeerID()
	if id != peerID {
		t.Errorf("GetPeerID: expected %v, got %v", peerID, id)
	}
}

func TestWebSocketSession_IsIncoming(t *testing.T) {
	peer := &domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}

	sIncoming := &WebSocketSession{
		readChan:  make(chan *pb.Message, 10),
		writeChan: make(chan *pb.Message, 10),
	}
	sIncoming.ID = uuid.New()
	sIncoming.Peer = *peer
	sIncoming.Incoming = true
	sIncoming.LastDial = time.Now()

	if !sIncoming.IsIncoming() {
		t.Error("IsIncoming: expected true for incoming session")
	}

	sOutgoing := &WebSocketSession{
		readChan:  make(chan *pb.Message, 10),
		writeChan: make(chan *pb.Message, 10),
	}
	sOutgoing.ID = uuid.New()
	sOutgoing.Peer = *peer
	sOutgoing.Incoming = false
	sOutgoing.LastDial = time.Now()

	if sOutgoing.IsIncoming() {
		t.Error("IsIncoming: expected false for outgoing session")
	}
}

func TestWebSocketSession_IsOpen(t *testing.T) {
	peer := &domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}

	sClosed := &WebSocketSession{
		readChan:  make(chan *pb.Message, 10),
		writeChan: make(chan *pb.Message, 10),
	}
	sClosed.ID = uuid.New()
	sClosed.Peer = *peer
	sClosed.Incoming = false
	sClosed.LastDial = time.Now()

	if sClosed.IsOpen() {
		t.Error("IsOpen: expected false for closed session")
	}
}

func TestWebSocketSession_Close(t *testing.T) {
	peer := &domain.Peer{
		ID:   uuid.New(),
		Port: 8080,
	}

	s := &WebSocketSession{
		readChan:  make(chan *pb.Message, 10),
		writeChan: make(chan *pb.Message, 10),
	}
	s.ID = uuid.New()
	s.Peer = *peer
	s.Incoming = false
	s.LastDial = time.Now()

	close(s.readChan)
	close(s.writeChan)

	select {
	case _, ok := <-s.readChan:
		if ok {
			t.Error("readChan should be closed")
		}
	default:
		t.Error("readChan should be closed")
	}
}
