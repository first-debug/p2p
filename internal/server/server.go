// Package server contains intarface which define schema of exchange between peers
package server

import (
	"context"
)

type Server interface {
	Start(context.Context)
	Stop(context.Context)
	GetSessions() []Session
}

type Session interface {
	Read(context.Context) ([]byte, error)
	Write(context.Context, []byte) error
	Close(context.Context)
	IsOpen() bool
}
