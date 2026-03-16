package cli

import (
	"context"
	"errors"
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

	"github.com/chzyer/readline"
	"github.com/mattn/go-colorable"
	"golang.org/x/term"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type inputMode int

const (
	menu inputMode = iota
	messaging
)

const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[0;33m"
	colorBlue   = "\033[0;34m"
	colorNone   = "\033[0m"
)

const (
	altTab     = "  "
	timeFormat = "15:04:05, 2 Jan 2006"
)

var (
	menuPromt              string = fmt.Sprintf("\r%s", colorizeText(">> ", colorGreen))
	messagingTemplatePromt string = fmt.Sprintf("\r%s", colorizeText("(%s)>> ", colorRed))
	historyFileName        string = "history.tmp"
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
	historyDir       string
}

func NewCliManager(ctx context.Context, log *slog.Logger, peer domain.Peer, p peerstorage.PeerStorage, s sessionstorage.SessionStorage, c client.Client, historyDir string) manager.Manager {
	return &CliManager{
		ctx:      ctx,
		logger:   log,
		selfInfo: peer,
		pStorage: p,
		sStorage: s,
		client:   c,

		currentMode: menu,
		historyDir:  historyDir,
	}
}

func (m *CliManager) Run() error {
	oldState, err := term.MakeRaw(0)
	if err != nil {
		return err
	}
	defer term.Restore(0, oldState)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          menuPromt,
		HistoryFile:     fmt.Sprintf("%s/%s", m.historyDir, historyFileName),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		Stdout:          colorable.NewColorableStdout(),

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
					switch m.currentMode {
					case menu:
						writeWarn(m.writer, "Use 'exit' to close CLI.")
					case messaging:
						writeWarn(m.writer, "Use '/quit' to detach feom session.")
					}
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
		fmt.Fprintf(m.writer, "\r%s\n", colorizeText("exit", colorGreen))
		return &readline.InterruptError{}
	} else {
		writeError(m.writer, errors.New("not supported command"))
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
		writeWarn(m.writer, "detached from the session.")
		m.setMenuMode()
		m.sessionCtxCancel()
		return
	}

	sendTime := timestamppb.Now()
	fmt.Fprintf(m.writer, "\033[F\r\033[K")
	fmt.Fprintf(m.writer,
		"%s>> %s\n\r>> %s%s\n",
		colorGreen, sendTime.AsTime(),
		colorNone,
		msg,
	)

	m.currentWriteCh <- &pb.Message{
		SendTime: sendTime,
		Message:  msg,
	}
}
