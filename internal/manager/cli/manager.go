package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	client "github.com/first-debug/p2p/internal/client"
	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/manager"
	pb "github.com/first-debug/p2p/internal/proto"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/chzyer/readline"
)

type inputMode int

const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[0;33m"
	colorBlue   = "\033[0;34m"
	colorNone   = "\033[0m"
)

const (
	menu inputMode = iota
	messaging
)

var (
	menuPromt              string = colorizeText(">> ", colorGreen)
	messagingTemplatePromt string = colorizeText("(%s)>> ", colorRed)
)

type CliManager struct {
	ctx      context.Context
	logger   *slog.Logger
	selfInfo domain.Peer
	pStorage peerstorage.PeerStorage
	sStorage sessionstorage.SessionStorage
	client   client.Client

	rl               *readline.Instance
	currentMode      inputMode
	writer           io.Writer
	currentWriteCh   chan<- *pb.Message
	currentReadCh    <-chan *pb.Message
	sessionCtx       context.Context
	sessionCtxCancel context.CancelFunc
}

func NewCliManager(ctx context.Context, log *slog.Logger, peer domain.Peer, p peerstorage.PeerStorage, s sessionstorage.SessionStorage, c client.Client) manager.Manager {
	return &CliManager{
		ctx:      ctx,
		logger:   log,
		selfInfo: peer,
		pStorage: p,
		sStorage: s,
		client:   c,

		currentMode: menu,
	}
}

func (m *CliManager) Run() error {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          menuPromt,
		HistoryFile:     "/tmp/readline.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold: true,
	})
	if err != nil {
		return err
	}
	m.rl = rl
	defer m.rl.Close()
	m.writer = m.rl.Stdout()

	for {
		select {
		case <-m.ctx.Done():
			return nil
		default:
			input, err := m.rl.Readline()
			if err != nil {
				if err == readline.ErrInterrupt {
					fmt.Fprintln(m.writer, colorizeText("\nUse '/quit' in session or 'exit' to close CLI", colorYellow))
					continue
				}
				return err
			}
			input = strings.TrimSpace(input)

			switch m.currentMode {
			case menu:
				err := m.handelMenu(input)
				if err != nil {
					return nil
				}
			case messaging:
				m.handelMessage(input)
			}
		}
	}
}

func (m *CliManager) handelMenu(input string) (err error) {
	if input == "list peers" {
		m.catchListPeers()
	} else if strings.HasPrefix(input, "connect ") {
		m.catchConnectCommand(input)
	} else if input == "list sessions" {
		m.catchListSessions()
	} else if strings.HasPrefix(input, "attach ") {
		m.catchAttachCommand(input)
	} else if input == "exit" {
		fmt.Fprintln(m.writer, colorizeText("exit", colorGreen))
		return &readline.InterruptError{}
	} else {
		fmt.Fprintln(m.writer, colorizeText("Error:", colorRed), "not supported command")
	}
	return
}

func (m *CliManager) handelMessage(msg string) {
	m.logger.Debug("ждём ввода пользователя")
	if msg == "" {
		return
	}

	if msg == "/quit" {
		m.logger.Debug("пользователь ввёл /quit => выход из сессии")
		fmt.Fprintln(m.writer, colorizeText("detached from the session", colorYellow))
		m.setMenuMode()
		m.sessionCtxCancel()
		return
	}

	sendTime := timestamppb.Now()
	fmt.Fprintf(m.writer, "\033[F\r\033[K")
	fmt.Fprintf(m.writer,
		"%s>> %s\n>> %s%s\n",
		colorGreen, sendTime.AsTime(),
		colorNone,
		msg,
	)

	m.currentWriteCh <- &pb.Message{
		SendTime: sendTime,
		Message:  msg,
	}
}
