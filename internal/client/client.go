package client

import "context"

type Client interface {
	Connect(context.Context, string) (Client, error)
	Close(context.Context)
	Read(context.Context) ([]byte, error)
	Write(context.Context, []byte) error
	IsOpen() bool
}
