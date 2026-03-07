package udp

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/first-debug/p2p/internal/config"
	"github.com/first-debug/p2p/internal/explorer"
	pb "github.com/first-debug/p2p/internal/proto"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer"

	"google.golang.org/protobuf/proto"
)

var logger log.Logger = *log.New(os.Stderr, "[UDPExplorer] ", log.LstdFlags)

type UDPExplorer struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	wg        sync.WaitGroup

	sender   *net.UDPConn
	listener *net.UDPConn
	peerInfo *pb.Peer

	peerStorage peerstorage.PeerStorage
}

func NewUDPExplorer(cfg *config.Config, peerInfo *pb.Peer, peerStorage peerstorage.PeerStorage) (explorer.Explorer, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", cfg.MulticastAddress, cfg.MulticastPort))
	if err != nil {
		return nil, err
	}

	inter, err := net.InterfaceByName(cfg.MulticastInterfaceName)
	if err != nil {
		if inters, err := net.Interfaces(); len(inters) > 0 {
			for _, i := range inters {
				if i.Name == "lo" {
					continue
				} else {
					inter = &i
				}
			}
		} else {
			return nil, err
		}
	}

	listener, err := net.ListenMulticastUDP("udp", inter, addr)
	if err != nil {
		return nil, err
	}

	sender, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	logger.Println("UDP Explorer started")

	ctx, cancel := context.WithCancel(context.Background())

	e := &UDPExplorer{
		ctx:         ctx,
		ctxCancel:   cancel,
		listener:    listener,
		sender:      sender,
		peerInfo:    peerInfo,
		peerStorage: peerStorage,
	}
	e.wg.Go(e.startReceive)

	return e, err
}

func (e *UDPExplorer) startReceive() {
	logger.Println("starting recive information from other peers")
	e.wg.Go(func() {
		data := make([]byte, 1024)
		for {
			select {
			case <-e.stop:
				return
			default:
				n, addr, err := e.listener.ReadFromUDP(data)
				ipParts := strings.Split(e.sender.LocalAddr().String(), ":")
				if len(ipParts) != 2 {
					logger.Printf("cannot parse local IPv4 from string '%v'", e.sender.LocalAddr().String())
				}
				if addr.IP.String() == ipParts[0] {
					continue
				}

				if err != nil {
					logger.Printf("Read error: %v", err)
					continue
				}
				var msg pb.Peer
				err = proto.Unmarshal(data[:n], &msg)
				if err != nil {
					logger.Printf("cannot unmarshal UDP request: %v", err)
					continue
				}
				peer := pb.PbPeerToDomain(&msg)
				peer.IP = addr.IP
				err = e.peerStorage.Add(peer)
				if err != nil {
					logger.Printf("Cannot add new peer: %v", err)
				}
			}
		}
	})
}

func (e *UDPExplorer) Emit() error {
	data, err := proto.Marshal(e.peerInfo)
	if err != nil {
		return err
	}
	_, err = e.sender.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (e *UDPExplorer) Stop() {
	e.ctxCancel()
	e.wg.Wait()
	e.listener.Close()
	e.sender.Close()
}
