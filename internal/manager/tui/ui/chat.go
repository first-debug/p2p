package ui

import (
	"context"
	"sync"

	"github.com/first-debug/p2p/internal/manager/tui/business"
	"github.com/first-debug/p2p/internal/manager/tui/styles"
	"github.com/first-debug/p2p/internal/session"
	"github.com/rivo/tview"
)

// ChatWindow представляет окно чата с сессией.
type ChatWindow struct {
	mu           sync.RWMutex
	chatView     *tview.TextView
	chatInput    *tview.InputField
	chatOps      *business.ChatOperations
	styles       *styles.ComponentStyles
	currentSession session.Session
	readerCancel   context.CancelFunc
	messageChan    <-chan business.ChatMessage
	isActive     bool
}

// NewChatWindow создаёт новое окно чата.
func NewChatWindow(
	chatView *tview.TextView,
	chatInput *tview.InputField,
	chatOps *business.ChatOperations,
	styles *styles.ComponentStyles,
) *ChatWindow {
	return &ChatWindow{
		chatView:  chatView,
		chatInput: chatInput,
		chatOps:   chatOps,
		styles:    styles,
		isActive:  false,
	}
}

// Open открывает чат для указанной сессии.
func (cw *ChatWindow) Open(ctx context.Context, sess session.Session) error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	// Если уже открыт чат с этой сессией, ничего не делаем
	if cw.isActive && cw.currentSession != nil {
		if cw.currentSession.GetID() == sess.GetID() {
			return nil
		}
		// Закрываем предыдущий чат
		cw.closeInternal()
	}

	cw.currentSession = sess
	cw.isActive = true

	// Очищаем чат и добавляем системное сообщение
	cw.chatView.SetText("[green]Чат открыт[-]")

	// Запускаем чтение сообщений
	messageChan, cancel, err := cw.chatOps.StartChatReader(ctx, sess)
	if err != nil {
		return err
	}

	cw.readerCancel = cancel
	cw.messageChan = messageChan

	// Запускаем горутину для обработки входящих сообщений
	go cw.handleMessages(ctx)

	// Показываем поле ввода (устанавливаем фокус)
	// tview InputField не имеет SetVisible, поэтому просто устанавливаем фокус

	return nil
}

// handleMessages обрабатывает входящие сообщения в фоновом режиме.
func (cw *ChatWindow) handleMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-cw.messageChan:
			if !ok {
				// Канал закрыт
				cw.mu.Lock()
				cw.chatView.SetText(cw.chatView.GetText(false) + "\n" + cw.chatOps.FormatMessage(
					cw.chatOps.CreateSystemMessage("Сессия закрыта"),
				))
				cw.mu.Unlock()
				return
			}
			cw.mu.Lock()
			cw.chatView.SetText(cw.chatView.GetText(false) + "\n" + cw.chatOps.FormatMessage(msg))
			cw.mu.Unlock()
		}
	}
}

// SendMessage отправляет сообщение в чат.
func (cw *ChatWindow) SendMessage(ctx context.Context, text string) error {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	if !cw.isActive || cw.currentSession == nil {
		return nil
	}

	if text == "" {
		return nil
	}

	// Отправляем сообщение
	err := cw.chatOps.SendMessage(ctx, cw.currentSession, text)
	if err != nil {
		return err
	}

	// Добавляем исходящее сообщение в чат
	cw.chatView.SetText(cw.chatView.GetText(false) + "\n" + cw.chatOps.FormatMessage(
		cw.chatOps.CreateOutgoingMessage(text),
	))

	return nil
}

// Close закрывает чат, не закрывая сессию.
func (cw *ChatWindow) Close() {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.closeInternal()
}

// CloseWithSession закрывает чат и сессию.
func (cw *ChatWindow) CloseWithSession(ctx context.Context) error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if cw.currentSession != nil {
		cw.currentSession.Close(ctx)
	}

	cw.closeInternal()
	return nil
}

// closeInternal внутренняя функция закрытия чата.
func (cw *ChatWindow) closeInternal() {
	if cw.readerCancel != nil {
		cw.readerCancel()
	}

	cw.isActive = false
	cw.currentSession = nil
	cw.messageChan = nil
	cw.readerCancel = nil

	// Скрываем поле ввода (переключаем фокус)
	// tview InputField не имеет SetVisible, поэтому просто очищаем

	// Добавляем системное сообщение
	cw.chatView.SetText(cw.chatView.GetText(false) + "\n" + cw.chatOps.FormatMessage(
		cw.chatOps.CreateSystemMessage("Чат закрыт"),
	))
}

// IsActive возвращает true, если чат активен.
func (cw *ChatWindow) IsActive() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.isActive
}

// GetCurrentSession возвращает текущую сессию.
func (cw *ChatWindow) GetCurrentSession() session.Session {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.currentSession
}

// SetTitle устанавливает заголовок окна чата.
func (cw *ChatWindow) SetTitle(title string) {
	cw.chatView.SetTitle(" ЧАТ: " + title + " ")
}

// Clear очищает чат.
func (cw *ChatWindow) Clear() {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.chatView.Clear()
}

// SetInputText устанавливает текст поля ввода.
func (cw *ChatWindow) SetInputText(text string) {
	cw.chatInput.SetText(text)
}

// GetInputText возвращает текст поля ввода.
func (cw *ChatWindow) GetInputText() string {
	return cw.chatInput.GetText()
}

// ClearInput очищает поле ввода.
func (cw *ChatWindow) ClearInput() {
	cw.chatInput.SetText("")
}
