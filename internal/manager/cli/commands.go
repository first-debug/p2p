package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (m *CliManager) catchListPeers() {
	peers, err := m.pStorage.GetAll()
	if err != nil {
		fmt.Fprintln(m.writer, err.Error())
	} else {
		fmt.Fprint(m.writer, "[")
		for _, v := range peers {
			fmt.Fprintln(m.writer)
			fmt.Fprintln(m.writer, "\t", v.ID)
		}
		fmt.Fprintln(m.writer, "]")
	}
}

func (m *CliManager) catchListSessions() {
	sess, err := m.sStorage.GetAll()
	if err != nil {
		fmt.Fprintln(m.writer, err.Error())
	} else {
		fmt.Fprint(m.writer, "[")
		for _, v := range sess {
			fmt.Fprintln(m.writer)
			fmt.Fprintln(m.writer, "\t", v.GetID())
		}
		fmt.Fprintln(m.writer, "]")
	}
}

func (m *CliManager) catchConnectCommand(input string) {
	strs := strings.Split(input, " ")
	if len(strs) != 2 {
		fmt.Fprintln(m.writer, "too many arguments for `connect` command")
		return
	}
	id, err := uuid.Parse(strs[1])
	if err != nil {
		fmt.Fprintln(m.writer, colorizeText("Error:", colorRed), err.Error())
		return
	}
	peer, err := m.pStorage.GetByID(id)
	if err != nil {
		fmt.Fprintln(m.writer, colorizeText("Error:", colorRed), err.Error())
		return
	}
	sess, err := m.client.Connect(context.Background(), &peer)
	if err != nil {
		fmt.Fprintln(m.writer, colorizeText("Error:", colorRed), err.Error())
		return
	}
	wCh, err := sess.GetWriteChannel(m.ctx)
	if err != nil {
		fmt.Fprintln(m.writer, colorizeText("Error:", colorRed), err.Error())
		return
	}
	rCh, err := sess.GetReadChannel(m.ctx)
	if err != nil {
		fmt.Fprintln(m.writer, colorizeText("Error:", colorRed), err.Error())
		return
	}

	m.setMessagingMode(sess.GetPeerID(), rCh, wCh)
}

func (m *CliManager) catchAttachCommand(input string) {
	strs := strings.Split(input, " ")
	if len(strs) != 2 {
		fmt.Fprintln(m.writer, "too many arguments for `attach` command")
		return
	}
	id, err := uuid.Parse(strs[1])
	if err != nil {
		fmt.Fprintln(m.writer, colorizeText("Error:", colorRed), err.Error())
		return
	}
	sess, err := m.sStorage.GetByID(id)
	if err != nil {
		fmt.Fprintln(m.writer, colorizeText("Error:", colorRed), err.Error())
		return
	}
	wCh, err := sess.GetWriteChannel(m.ctx)
	if err != nil {
		fmt.Fprintln(m.writer, colorizeText("Error:", colorRed), err.Error())
		return
	}
	rCh, err := sess.GetReadChannel(m.ctx)
	if err != nil {
		fmt.Fprintln(m.writer, colorizeText("Error:", colorRed), err.Error())
		return
	}

	m.setMessagingMode(sess.GetPeerID(), rCh, wCh)
}
