package climanager

import (
	"github.com/first-debug/p2p/internal/explorer"
	"github.com/first-debug/p2p/internal/server"
)

type CliManager struct {
	Server   server.Server
	Explorer explorer.Explorer
}
