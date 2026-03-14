package session

import (
	"context"
	"testing"
	"time"

	"github.com/first-debug/p2p/internal/domain"
	pb "github.com/first-debug/p2p/internal/proto"
	"github.com/google/uuid"
)

type MockSession struct {
	id       uuid.UUID
	peer     domain.Peer
	incoming bool
	lastDial time.Time
	open     bool
	readCh   <-chan *pb.Message
	writeCh  chan<- *pb.Message
}

func (m *MockSession) GetID() uuid.UUID                              { return m.id }
func (m *MockSession) GetLastDial() time.Time                        { return m.lastDial }
func (m *MockSession) GetReadChannel(context.Context) (<-chan *pb.Message, error) {
	return m.readCh, nil
}
func (m *MockSession) GetWriteChannel(context.Context) (chan<- *pb.Message, error) {
	return m.writeCh, nil
}
func (m *MockSession) GetPeerID() uuid.UUID { return m.peer.ID }
func (m *MockSession) IsIncoming() bool     { return m.incoming }
func (m *MockSession) IsOpen() bool         { return m.open }
func (m *MockSession) Close(context.Context) {
	m.open = false
}

func TestBaseSession_Fields(t *testing.T) {
	id := uuid.New()
	peer := domain.Peer{
		ID:   uuid.New(),
		Name: "test-peer",
		Port: 8080,
	}
	now := time.Now()

	base := BaseSession{
		ID:       id,
		Peer:     peer,
		Incoming: true,
		LastDial: now,
	}

	if base.ID != id {
		t.Errorf("expected ID %v, got %v", id, base.ID)
	}
	if base.Peer.Name != "test-peer" {
		t.Errorf("expected Peer.Name 'test-peer', got %s", base.Peer.Name)
	}
	if !base.Incoming {
		t.Error("expected Incoming to be true")
	}
	if base.LastDial != now {
		t.Errorf("expected LastDial %v, got %v", now, base.LastDial)
	}
}

func TestBaseSession_DefaultValues(t *testing.T) {
	base := BaseSession{}

	if base.ID != uuid.Nil {
		t.Errorf("expected empty ID, got %v", base.ID)
	}
	if base.Peer.ID != uuid.Nil {
		t.Errorf("expected empty Peer.ID, got %v", base.Peer.ID)
	}
	if base.Incoming {
		t.Error("expected Incoming to be false")
	}
	if base.LastDial != (time.Time{}) {
		t.Errorf("expected empty LastDial, got %v", base.LastDial)
	}
}

func TestBaseSession_WaitGroup(t *testing.T) {
	base := BaseSession{}

	done := make(chan struct{})
	go func() {
		base.Wait()
		close(done)
	}()

	base.Add(1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		base.Done()
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("WaitGroup timed out")
	}
}

func TestMockSession_Interface(t *testing.T) {
	id := uuid.New()
	peer := domain.Peer{
		ID:   uuid.New(),
		Port: 9000,
	}
	now := time.Now()
	readCh := make(<-chan *pb.Message)
	writeCh := make(chan<- *pb.Message)

	mock := &MockSession{
		id:       id,
		peer:     peer,
		incoming: false,
		lastDial: now,
		open:     true,
		readCh:   readCh,
		writeCh:  writeCh,
	}

	if mock.GetID() != id {
		t.Errorf("GetID: expected %v, got %v", id, mock.GetID())
	}
	if mock.GetLastDial() != now {
		t.Errorf("GetLastDial: expected %v, got %v", now, mock.GetLastDial())
	}
	if mock.IsIncoming() {
		t.Error("IsIncoming: expected false")
	}
	if !mock.IsOpen() {
		t.Error("IsOpen: expected true")
	}

	ctx := context.Background()
	rCh, err := mock.GetReadChannel(ctx)
	if err != nil {
		t.Errorf("GetReadChannel returned error: %v", err)
	}
	if rCh != readCh {
		t.Error("GetReadChannel: channels don't match")
	}

	wCh, err := mock.GetWriteChannel(ctx)
	if err != nil {
		t.Errorf("GetWriteChannel returned error: %v", err)
	}
	if wCh != writeCh {
		t.Error("GetWriteChannel: channels don't match")
	}

	mock.Close(ctx)
	if mock.IsOpen() {
		t.Error("Close: expected session to be closed")
	}
}
