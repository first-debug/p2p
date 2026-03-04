// Package server contains intarface which define schema of exchange between peers
package server

import (
	"context"
)

type Server interface {
	Serve() error
	Stop(context.Context)
}
