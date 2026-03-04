package climanager

import (
	"main/internal/explorer"
	"main/internal/server"
)

type CliManager struct {
	Server   server.Server
	Explorer explorer.Explorer
}
