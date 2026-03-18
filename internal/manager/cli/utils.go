package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/first-debug/p2p/internal/domain"
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
				fmt.Fprintf(m.writer, "\r%s\n", colorizeText("peer disconnected", colorYellow))
				m.sessionCtxCancel()
				m.setMenuMode()
				return
			}

			// termWidth := m.getOutputWidth(os.Stdout)
			fmt.Fprintf(
				m.writer, "\r%s<< %s\n\r<< %s%s\n",
				colorBlue,
				msg.SendTime.AsTime(),
				colorNone, msg.Message,
			)
		}
	}
}

func (m *CliManager) setMessagingMode(peerID uuid.UUID, rCh <-chan *pb.Message, wCh chan<- *pb.Message) {
	m.currentMode = messaging
	m.currentReadCh = rCh
	m.currentWriteCh = wCh
	m.sessionCtx, m.sessionCtxCancel = context.WithCancel(m.ctx)

	m.rl.Config.AutoComplete = messagingCompleter
	m.rl.SetPrompt(fmt.Sprintf(messagingTemplatePromt, strings.Split(peerID.String(), "-")[0]))
	m.rl.Refresh()

	go m.readSessionMessages()
}

func (m *CliManager) setMenuMode() {
	m.currentMode = menu
	m.rl.Config.AutoComplete = menuCompleter
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

func colorizeError(err error) string {
	return fmt.Sprintf("%sError: %s%s", colorRed, colorNone, err.Error())
}

func writeError(writer io.Writer, err error) {
	fmt.Fprintf(writer, "\r%s\n", colorizeError(err))
}

func writeWarn(writer io.Writer, warn string) {
	fmt.Fprintf(writer, "\r%s\n", colorizeText(warn, colorYellow))
}

func writePeerList(writer io.Writer, peers []domain.Peer) {
	peersLen := len(peers)
	if peersLen == 0 {
		writeWarn(writer, "Peers list is empty.")
		return
	}
	for i, v := range peers {
		fmt.Fprintf(writer, "\r%3d) ID: %s\n", i+1, v.ID)
		fmt.Fprintf(writer, "\r%4s%sIP: %s\n", altTab, altTab, v.IP)
		fmt.Fprintf(writer, "\r%4s%sPort: %d\n", altTab, altTab, v.Port)
		if i != peersLen-1 {
			fmt.Fprintln(writer)
		}
	}
}
