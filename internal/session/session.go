package session

import (
	"context"
	"sync"
	"time"

	"github.com/first-debug/p2p/internal/domain"
	pb "github.com/first-debug/p2p/internal/proto"

	"github.com/google/uuid"
)

type BaseSession struct {
	sync.WaitGroup // For no-copy status

	ID       uuid.UUID // Internal ID
	Peer     *domain.Peer
	Incoming bool // Indicate that the instance was created by an external connection
	LastDial time.Time
}

type Session interface {
	GetID() uuid.UUID
	IsIncoming() bool
	GetReadChannel(context.Context) (*chan *pb.Message, error)
	GetWriteChannel(context.Context) (*chan *pb.Message, error)
	Close(context.Context)
	IsOpen() bool
}
