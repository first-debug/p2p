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
	pb "github.com/first-debug/p2p/internal/proto"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
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
					wCh, err := sess.GetWriteChannel(m.ctx)
					if err != nil {
						fmt.Fprintln(writer, err.Error())
						continue
					}
					rCh, err := sess.GetReadChannel(m.ctx)
					if err != nil {
						fmt.Fprintln(writer, err.Error())
						continue
					}
					// TODO: add logic for sending messages from the `scanner` to the channel

					go func() {
						for {
							select {
							case <-m.ctx.Done():
								return
							case msg, ok := <-rCh:
								if !ok {
									return
								}
								fmt.Fprintln(writer, "%50s\n", msg.Message)
							}
						}
					}()
					for {
						select {
						case <-m.ctx.Done():
							break
						default:
							if scanner.Scan() {
								msg := scanner.Text()

								wCh <- &pb.Message{
									SendTime: timestamppb.Now(),
									Message:  msg,
								}
							}
						}
					}
				} else if input == "list sessions" {
					sess, err := m.sStorage.GetAll()
					if err != nil {
						fmt.Fprintln(writer, err.Error())
					} else {
						fmt.Fprint(writer, "[")
						for _, v := range sess {
							fmt.Fprintln(writer)
							fmt.Fprintln(writer, "\t", v.GetID())
						}
						fmt.Fprintln(writer, "]")
					}
				} else if strings.Contains(input, "attach ") {
					strs := strings.Split(input, " ")
					if len(strs) != 2 {
						fmt.Fprintln(writer, "too many arguments for `attach` command")
						continue
					}
					id, err := uuid.Parse(strs[1])
					if err != nil {
						fmt.Fprintln(writer, err.Error())
						continue
					}
					sess, err := m.sStorage.GetByID(id)
					if err != nil {
						fmt.Fprintln(writer, err.Error())
						continue
					}
					ch, err := sess.GetReadChannel(m.ctx)
					if err != nil {
						fmt.Fprintln(writer, err.Error())
						continue
					}
					select {
					case <-m.ctx.Done():
						break
					case v, ok := <-ch:
						if !ok {
							fmt.Fprintln(writer, "read channel close")
						}
						fmt.Fprintln(writer, sess.GetID().String(), "(", v.SendTime.String(), "): ", v.Message)
					}
				} else {
					fmt.Fprintln(writer, "not supported command")
				}
			}
		}
	}
}
