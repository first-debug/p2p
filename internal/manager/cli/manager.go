package cli

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	client "github.com/first-debug/p2p/internal/client"
	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/manager"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session"
	"github.com/google/uuid"
)

var logger log.Logger = *log.New(os.Stderr, "[CliManager] ", log.LstdFlags)

type CliManager struct {
	ctx      context.Context
	selfInfo domain.Peer
	pStorage peerstorage.PeerStorage
	sStorage sessionstorage.SessionStorage
	client   client.Client
}

func NewCliManager(ctx context.Context, peer domain.Peer, p peerstorage.PeerStorage, s sessionstorage.SessionStorage, c client.Client) manager.Manager {
	return &CliManager{
		ctx:      ctx,
		selfInfo: peer,
		pStorage: p,
		sStorage: s,
		client:   c,
	}
}

func (m *CliManager) Run() error {
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	for {
		writer.Flush()
		select {
		case <-m.ctx.Done():
			return nil
		default:
			if scanner.Scan() {
				input := scanner.Text()
				if input == "list peers" {
					peers, err := m.pStorage.GetAll()
					if err != nil {
						fmt.Fprintln(writer, err.Error())
					} else {
						fmt.Fprint(writer, "[")
						for _, v := range peers {
							fmt.Fprintln(writer)
							fmt.Fprintln(writer, "\t", v.ID)
						}
						fmt.Fprintln(writer, "]")
					}
				} else if strings.Contains(input, "connect ") {
					strs := strings.Split(input, " ")
					if len(strs) != 2 {
						fmt.Fprintln(writer, "too many arguments for `connect` command")
						continue
					}
					id, err := uuid.Parse(strs[1])
					if err != nil {
						fmt.Fprintln(writer, err.Error())
						continue
					}
					peer, err := m.pStorage.GetByID(id)
					if err != nil {
						fmt.Fprintln(writer, err.Error())
						continue
					}
					sess, err := m.client.Connect(context.Background(), &peer)
					if err != nil {
						fmt.Fprintln(writer, err.Error())
						continue
					}
					ch, err := sess.GetWriteChannel(m.ctx)
					if err != nil {
						fmt.Fprintln(writer, err.Error())
						continue
					}
					// TODO: add logic for sending messages from the `scanner` to the channel
				} else {
					fmt.Fprintln(writer, "not supported command")
				}
			}
		}
	}
}
