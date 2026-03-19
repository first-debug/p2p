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

	sender            *net.UDPConn
	listener          *net.UDPConn
	marshaledPeerInfo []byte

	peerStorage peerstorage.PeerStorage
}

func NewUDPExplorer(cfg *config.Config, log *slog.Logger, peerInfo domain.Peer, peerStorage peerstorage.PeerStorage) (explorer.Explorer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	e := &UDPExplorer{
		logger:      log.With("module", "UDPExplorer"),
		ctx:         ctx,
		ctxCancel:   cancel,
		peerStorage: peerStorage,
	}

	data, err := proto.Marshal(pb.DomainToPbPeer(&peerInfo))
	if err != nil {
		return e, err
	}

	e.marshaledPeerInfo = data

	e.logger.Info("starting UDP Explorer...")

	if cfg.Explorer.Multicast != nil {
		err = e.setMulticast(cfg)
	}
	if cfg.Explorer.Broadcast != nil {
		if cfg.Explorer.Multicast != nil {
			e.logger.Warn("cannot use 2 exploring method. Now using Multicast UDP")
		} else {
			err = e.setBroadcast(cfg)
		}
	}
	if err != nil {
		return e, err
	}

	e.wg.Go(e.startReceive)
	e.wg.Go(e.checkPeersAvailable)

	return e, err
}

func (e *UDPExplorer) Emit() error {
	_, err := e.sender.Write(e.marshaledPeerInfo)
	if err != nil {
		return err
	}
	return nil
}

func (e *UDPExplorer) TargetEmit(target string) error {
	addr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		e.logger.Debug("cannot resolve UDP Address", "addr", target)
		return err
	}

	sender, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		e.logger.Debug("cannot dial UDP Address", "addr", addr)
		return err
	}
	defer sender.Close()
	sender.SetDeadline(time.Now().Add(300 * time.Millisecond))

	n, err := sender.Write(e.marshaledPeerInfo)
	if err != nil {
		e.logger.Debug("cannot write to UDP Address", "error", err)
		return err
	}
	if n != len(e.marshaledPeerInfo) {
		e.logger.Warn("the length of the data and the written information are not equal", slog.Int("data len", len(e.marshaledPeerInfo)), slog.Int("written len", n))
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	data := make([]byte, 1024)
	for {
		select {
		case <-ticker.C:
			return errors.New("cannot get answer from peer")
		default:
			n, ansAddr, err := sender.ReadFromUDP(data)
			if err != nil {
				e.logger.Debug("cannot read from UDP Address", "error", err)
				return err
			}
			if !addr.IP.Equal(ansAddr.IP) {
				e.logger.Debug("catched request from other peer", "target addr", addr, "answer addr", ansAddr)
				continue
			}

			peer, err := e.parseDomainPeer(n, data)
			if err != nil {
				return err
			}
			// if `peer.IsPublicIP` set to true, `peer.IP` already contains right IP, else `peer.IP` is nil
			if !peer.IsPublicIP {
				peer.IP = addr.IP
			}

			err = e.peerStorage.Add(peer)
			if err != nil && !errors.Is(err, storage.ErrAlreadyExists) {
				return err
			}
			return nil
		}
	}
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

			peer, err := e.parseDomainPeer(n, data)
			if err != nil {
				e.logger.Error("cannot parse Peer from UDP request", slog.String("error", err.Error()))
				continue
			}
			// if `peer.IsPublicIP` set to true, `peer.IP` already contains right IP, else `peer.IP` is nil
			if !peer.IsPublicIP {
				peer.IP = addr.IP
			}

			err = e.peerStorage.Add(peer)
			if err != nil && !errors.Is(err, storage.ErrAlreadyExists) {
				e.logger.Error("cannot add new peer", slog.String("error", err.Error()))
				return
			}

			if n, err := e.listener.WriteTo(e.marshaledPeerInfo, addr); err != nil {
				e.logger.Error("cannot answer to peer", slog.String("error", err.Error()))
			} else {
				e.logger.Debug("answer to peer", slog.Int("bytes", n))
			}
		}
	}
}

func (e *UDPExplorer) setMulticast(cfg *config.Config) error {
	inter, err := getMainInterface()
	if err != nil {
		e.logger.Error("cannot found main interface, please set manualy `interface-name`")
		return err
	}
	e.logger.Info("found interface", "interface-name", inter.Name)

	e.logger.Info("try to use Multicast UDP to explorer other peers")

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", cfg.Explorer.Multicast.Address, cfg.Explorer.Multicast.Port))
	if err != nil {
		return err
	}

	e.listener, err = net.ListenMulticastUDP("udp", inter, addr)
	if err != nil {
		return err
	}

	e.sender, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	return nil
}

func (e *UDPExplorer) setBroadcast(cfg *config.Config) error {
	e.logger.Info("try to use Braodcast UDP to explorer other peers")

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", cfg.Explorer.Broadcast.Port))
	if err != nil {
		return err
	}
	e.listener, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%d", cfg.Explorer.Broadcast.Address, cfg.Explorer.Broadcast.Port))
	if err != nil {
		return err
	}

	e.sender, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	return nil
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

func getMainInterface() (*net.Interface, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	addr := conn.LocalAddr().(*net.UDPAddr)

	if inters, err := net.Interfaces(); len(inters) > 0 {
		if err != nil {
			return nil, err
		}
		for _, i := range inters {
			if i.Name == "lo" {
				continue
			}
			addrs, err := i.Addrs()
			if err != nil {
				continue
			}
			for _, a := range addrs {
				if ipNet, ok := a.(*net.IPNet); ok && ipNet.IP.Equal(addr.IP) {
					return &i, nil
				}
			}
		}
	}
	return nil, errors.New("cannot found interface")
}

func (e *UDPExplorer) parseDomainPeer(n int, data []byte) (domain.Peer, error) {
	var msg pb.Peer
	err := proto.Unmarshal(data[:n], &msg)
	if err != nil {
		return domain.Peer{}, err
	}

	return pb.PbPeerToDomain(&msg), nil
}
