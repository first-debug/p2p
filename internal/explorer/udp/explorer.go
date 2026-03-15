package udp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/first-debug/p2p/internal/config"
	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/explorer"
	pb "github.com/first-debug/p2p/internal/proto"
	"github.com/first-debug/p2p/internal/storage"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer"

	"google.golang.org/protobuf/proto"
)

type UDPExplorer struct {
	logger *slog.Logger

	ctx       context.Context
	ctxCancel context.CancelFunc
	wg        sync.WaitGroup

	sender   *net.UDPConn
	listener *net.UDPConn
	peerInfo *pb.Peer

	peerStorage peerstorage.PeerStorage
}

func NewUDPExplorer(cfg *config.Config, log *slog.Logger, peerInfo domain.Peer, peerStorage peerstorage.PeerStorage) (explorer.Explorer, error) {
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

	ctx, cancel := context.WithCancel(context.Background())

	e := &UDPExplorer{
		logger:      log.With("module", "UDPExplorer"),
		ctx:         ctx,
		ctxCancel:   cancel,
		listener:    listener,
		sender:      sender,
		peerInfo:    pb.DomainToPbPeer(&peerInfo),
		peerStorage: peerStorage,
	}

	e.logger.Info("UDP Explorer started")

	e.wg.Go(e.startReceive)
	e.wg.Go(e.checkPeersAvailable)

	return e, err
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

func (e *UDPExplorer) startReceive() {
	e.logger.Info("starting recive information from other peers")
	data := make([]byte, 1024)
	for {
		select {
		case <-e.ctx.Done():
			return
		default:
			n, addr, err := e.listener.ReadFromUDP(data)
			if err != nil {
				e.logger.Error("cannot read", slog.String("error", err.Error()))
				continue
			}

			// compare local address with addres of incoming request
			ipParts := strings.Split(e.sender.LocalAddr().String(), ":")
			if len(ipParts) != 2 {
				e.logger.Error("cannot parse local IPv4", slog.String("address", e.sender.LocalAddr().String()))
			}
			if addr.IP.String() == ipParts[0] {
				continue
			}

			var msg pb.Peer
			err = proto.Unmarshal(data[:n], &msg)
			if err != nil {
				e.logger.Error("cannot unmarshal UDP request", slog.String("error", err.Error()))
				continue
			}
			peer := pb.PbPeerToDomain(&msg)
			peer.IP = addr.IP
			err = e.peerStorage.Add(peer)
			if err != nil && !errors.Is(err, storage.ErrAlreadyExists) {
				e.logger.Error("Cannot add new peer", slog.String("error", err.Error()))
			}
		}
	}
}

func (e *UDPExplorer) checkPeersAvailable() {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			peers, err := e.peerStorage.GetAll()
			if err != nil {
				e.logger.Error("cannot get list of peers", slog.String("error", err.Error()))
			}
			for _, i := range peers {
				res, err := http.Get(fmt.Sprintf("http://%v:%v/ping", i.IP, i.Port))
				if err == nil && res.StatusCode == http.StatusOK {
					continue
				}
				e.peerStorage.RemoveByID(i.ID)
			}
		}
	}
}
