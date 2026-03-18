package cli

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/first-debug/p2p/internal/domain"
	"github.com/google/uuid"
)

func (m *CliManager) catchListPeers() {
	peers, err := m.pStorage.GetAll()
	if err != nil {
		writeError(m.writer, err)
	} else {
		writePeerList(m.writer, peers)
	}
}

func (m *CliManager) catchListSessions() {
	sess, err := m.sStorage.GetAll()
	if err != nil {
		writeError(m.writer, err)
	} else {
		sessLen := len(sess)
		if sessLen == 0 {
			writeWarn(m.writer, "Sessions list is empty.")
			return
		}
		for i, v := range sess {
			fmt.Fprintf(m.writer, "\r%3d) ID: %s\n", i+1, v.GetID())
			fmt.Fprintf(m.writer, "\r%4s%sPeer ID: %s\n", altTab, altTab, v.GetPeerID())
			fmt.Fprintf(m.writer, "\r%4s%sLast dial: %v\n", altTab, altTab, v.GetLastDial().Format(timeFormat))
			if i != sessLen-1 {
				fmt.Fprintln(m.writer)
			}
		}
	}
}

func (m *CliManager) catchConnectCommand(input string) {
	strs := strings.Fields(input)
	if len(strs) != 2 {
		writeWarn(m.writer, "too many arguments for `connect` command.")
		return
	}
	id, err := uuid.Parse(strs[1])
	if err != nil {
		writeError(m.writer, err)
		return
	}
	peer, err := m.pStorage.GetByID(id)
	if err != nil {
		writeError(m.writer, err)
		return
	}
	sess, err := m.client.Connect(context.Background(), &peer)
	if err != nil {
		writeError(m.writer, err)
		return
	}
	wCh, err := sess.GetWriteChannel(m.ctx)
	if err != nil {
		writeError(m.writer, err)
		return
	}
	rCh, err := sess.GetReadChannel(m.ctx)
	if err != nil {
		writeError(m.writer, err)
		return
	}

	m.setMessagingMode(sess.GetPeerID(), rCh, wCh)
}

func (m *CliManager) catchAttachCommand(input string) {
	strs := strings.Fields(input)
	if len(strs) != 2 {
		writeWarn(m.writer, "too many arguments for `attach` command.")
		return
	}
	id, err := uuid.Parse(strs[1])
	if err != nil {
		writeError(m.writer, err)
		return
	}
	sess, err := m.sStorage.GetByID(id)
	if err != nil {
		writeError(m.writer, err)
		return
	}
	wCh, err := sess.GetWriteChannel(m.ctx)
	if err != nil {
		writeError(m.writer, err)
		return
	}
	rCh, err := sess.GetReadChannel(m.ctx)
	if err != nil {
		writeError(m.writer, err)
		return
	}

	m.setMessagingMode(sess.GetPeerID(), rCh, wCh)
}

func (m *CliManager) catchLoadPeersCommand(input string) {
	strs := strings.Fields(input)
	if len(strs) != 2 {
		writeWarn(m.writer, "too many arguments for `load-peers` command.")
		return
	}
	id, err := uuid.Parse(strs[1])
	if err != nil {
		writeError(m.writer, err)
		return
	}
	peer, err := m.pStorage.GetByID(id)
	if err != nil {
		writeError(m.writer, err)
		return
	}
	var peers []domain.Peer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	end := make(chan error)

	wg.Go(func() {
		peers, err = m.client.GetKnownPeers(ctx, &peer)
		end <- err
	})

	select {
	case err, ok := <-end:
		cancel()
		if !ok {
			break
		}
		if err != nil {
			writeError(m.writer, fmt.Errorf("cannot load peers: %s", err.Error()))
			return
		}
	case <-time.Tick(5 * time.Second):
		writeError(m.writer, fmt.Errorf("cannot load peers: %s", err.Error()))
		cancel()
		wg.Wait()
		return
	}
	writePeerList(m.writer, peers)
	for _, p := range peers {
		m.pStorage.Add(p)
	}
}

func (m *CliManager) catchEmitCommand(input string) {
	strs := strings.Fields(input)
	switch len(strs) {
	case 1:
		if err := m.explorer.Emit(); err != nil {
			writeError(m.writer, err)
		}
	case 2:
		if err := m.explorer.TargetEmit(strs[1]); err != nil {
			writeError(m.writer, err)
		}
	default:
		writeWarn(m.writer, "too many arguments for `emit` command.")
	}
}
