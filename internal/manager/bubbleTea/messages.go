// Package tuimanager предоставляет TUI менеджер для управления пиром на основе bubbletea.
package tuimanager

import (
	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/session"
)

// === Сообщения для обновления данных ===

// peersLoadedMsg отправляется при загрузке списка пиров из хранилища.
type peersLoadedMsg struct {
	peers []domain.Peer
	err   error
}

// sessionsLoadedMsg отправляется при загрузке списка сессий из хранилища.
type sessionsLoadedMsg struct {
	sessions []session.Session
	err      error
}

// explorerTickMsg отправляется при срабатывании таймера обнаружения пиров.
type explorerTickMsg struct{}

// updateUITickMsg отправляется при срабатывании таймера обновления UI.
type updateUITickMsg struct{}

// === Сообщения для управления формами ===

// connectFormMsg отправляется при отправке формы подключения.
type connectFormMsg struct {
	address string
}

// modalDoneMsg отправляется при завершении модального окна.
type modalDoneMsg struct {
	buttonIndex int
	buttonLabel string
}

// === Сообщения для ошибок и статуса ===

// statusMsg отправляется для обновления текста статуса.
type statusMsg struct {
	text string
}

// errorMsg отправляется при возникновении ошибки.
type errorMsg struct {
	err error
}
