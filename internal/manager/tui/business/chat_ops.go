package business

import (
	"context"
	"sync"
	"time"

	"github.com/first-debug/p2p/internal/manager/tui/errors"
	"github.com/first-debug/p2p/internal/session"
	pb "github.com/first-debug/p2p/internal/proto"
)

// ChatMessage представляет сообщение в чате.
type ChatMessage struct {
	Text      string
	Incoming  bool
	Timestamp time.Time
	IsSystem  bool
}

// ChatOperations содержит операции для работы с чатом.
type ChatOperations struct {
	mu sync.RWMutex
}

// NewChatOperations создаёт новые операции для работы с чатом.
func NewChatOperations() *ChatOperations {
	return &ChatOperations{}
}

// StartChatReader запускает чтение сообщений из сессии в фоновом режиме.
// Возвращает канал для получения сообщений и функцию для остановки.
func (op *ChatOperations) StartChatReader(ctx context.Context, sess session.Session) (<-chan ChatMessage, context.CancelFunc, error) {
	if sess == nil {
		return nil, nil, errors.NewErrorf(errors.ErrorTypeSession, "Сессия не инициализирована")
	}

	readChan, err := sess.GetReadChannel(ctx)
	if err != nil {
		return nil, nil, errors.WrapError(err, errors.ErrorTypeSession, "Не удалось получить канал чтения")
	}

	messageChan := make(chan ChatMessage, 100)
	childCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer close(messageChan)
		for {
			select {
			case <-childCtx.Done():
				return
			case msg, ok := <-readChan:
				if !ok {
					// Канал закрыт - сессия завершена
					messageChan <- ChatMessage{
						Text:      "Сессия закрыта",
						IsSystem:  true,
						Timestamp: time.Now(),
					}
					return
				}
				if msg != nil {
					messageChan <- ChatMessage{
						Text:      msg.Message,
						Incoming:  true,
						Timestamp: time.Now(),
					}
				}
			}
		}
	}()

	return messageChan, cancel, nil
}

// SendMessage отправляет сообщение через сессию.
func (op *ChatOperations) SendMessage(ctx context.Context, sess session.Session, text string) error {
	if sess == nil {
		return errors.NewErrorf(errors.ErrorTypeSession, "Сессия не инициализирована")
	}

	writeChan, err := sess.GetWriteChannel(ctx)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeSession, "Не удалось получить канал записи")
	}

	msg := &pb.Message{
		Message: text,
	}

	select {
	case writeChan <- msg:
		return nil
	case <-ctx.Done():
		return errors.NewErrorf(errors.ErrorTypeSession, "Контекст отменён во время отправки")
	default:
		// Канал может быть полон или закрыт
		return errors.NewErrorf(errors.ErrorTypeSession, "Не удалось отправить сообщение")
	}
}

// FormatMessage форматирует сообщение для отображения в чате.
func (op *ChatOperations) FormatMessage(msg ChatMessage) string {
	timeStr := msg.Timestamp.Format("15:04:05")
	
	if msg.IsSystem {
		return "[#d09a66::i]*** " + msg.Text + "[-::]"
	}

	if msg.Incoming {
		return "[" + timeStr + " #98c379::b]" + msg.Text + "[-::]"
	}
	return "[" + timeStr + " #61afef::b]" + msg.Text + "[-::]"
}

// CreateSystemMessage создаёт системное сообщение.
func (op *ChatOperations) CreateSystemMessage(text string) ChatMessage {
	return ChatMessage{
		Text:      text,
		IsSystem:  true,
		Timestamp: time.Now(),
	}
}

// CreateOutgoingMessage создаёт исходящее сообщение.
func (op *ChatOperations) CreateOutgoingMessage(text string) ChatMessage {
	return ChatMessage{
		Text:      text,
		Incoming:  false,
		Timestamp: time.Now(),
	}
}

// CloseChat закрывает чат, останавливая чтение сообщений.
func (op *ChatOperations) CloseChat(cancel context.CancelFunc) {
	if cancel != nil {
		cancel()
	}
}
