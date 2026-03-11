package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	pb "github.com/first-debug/p2p/internal/proto"
	"github.com/google/uuid"
	"golang.org/x/term"
)

func (m *CliManager) readSessionMessages() {
	for {
		select {
		case <-m.sessionCtx.Done():
			return
		case msg, ok := <-m.currentReadCh:
			if !ok {
				m.logger.Debug("канал сообщений закрыт")
				fmt.Fprintln(m.writer, colorizeText("peer disconnected", colorYellow))
				m.sessionCtxCancel()
				m.setMenuMode()
				return
			}

			termWidth := m.getOutputWidth(os.Stdout)
			fmt.Fprintf(
				m.writer, "%s%*s%s\n%*s\n",
				colorBlue,
				termWidth, msg.SendTime.AsTime(),
				colorNone,
				termWidth, msg.Message,
			)
		}
	}
}

func (m *CliManager) setMessagingMode(peerID uuid.UUID, rCh <-chan *pb.Message, wCh chan<- *pb.Message) {
	m.currentMode = messaging
	m.currentReadCh = rCh
	m.currentWriteCh = wCh
	m.sessionCtx, m.sessionCtxCancel = context.WithCancel(m.ctx)

	m.rl.SetPrompt(fmt.Sprintf(messagingTemplatePromt, strings.Split(peerID.String(), "-")[0]))
	m.rl.Refresh()

	go m.readSessionMessages()
}

func (m *CliManager) setMenuMode() {
	m.currentMode = menu
	m.rl.SetPrompt(menuPromt)
	m.rl.Refresh()
}

func (m *CliManager) getOutputWidth(target *os.File) int {
	fd := int(target.Fd())
	if term.IsTerminal(fd) {
		w, _, err := term.GetSize(fd)
		if err != nil {
			m.logger.Error("term.GetSize", slog.String("error", err.Error()))
			return -1
		}
		return w
	}
	return -1
}

func colorizeText(text, color string) string {
	return fmt.Sprintf("%s%s%s", color, text, colorNone)
}
