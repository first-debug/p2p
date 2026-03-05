package business

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	"github.com/first-debug/p2p/internal/client"
	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/manager/tui/errors"
	"github.com/first-debug/p2p/internal/session"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session-storage"

	"github.com/google/uuid"
)

// SessionOperations содержит операции для работы с сессиями.
type SessionOperations struct {
	storage sessionstorage.SessionStorage
	client  client.Client
}

// NewSessionOperations создаёт новые операции для работы с сессиями.
func NewSessionOperations(storage sessionstorage.SessionStorage, client client.Client) *SessionOperations {
	return &SessionOperations{
		storage: storage,
		client:  client,
	}
}

// GetSessions возвращает список всех активных сессий.
func (op *SessionOperations) GetSessions() ([]session.Session, error) {
	sessions, err := op.storage.GetAll()
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeStorage, "Не удалось получить список сессий")
	}
	return sessions, nil
}

// GetSessionByID возвращает сессию по ID.
func (op *SessionOperations) GetSessionByID(id uuid.UUID) (session.Session, error) {
	sess, err := op.storage.GetByID(id)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeSession, "Не удалось получить сессию по ID")
	}
	return sess, nil
}

// ConnectToPeer подключается к пиру и создаёт новую сессию.
func (op *SessionOperations) ConnectToPeer(ctx context.Context, peer domain.Peer) (session.Session, error) {
	if op.client == nil {
		return nil, errors.NewErrorf(errors.ErrorTypeSession, "Client не инициализирован")
	}

	sess, err := op.client.Connect(ctx, peer)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeNetwork, "Не удалось подключиться к пиру")
	}

	return sess, nil
}

// CloseSession закрывает сессию по ID.
func (op *SessionOperations) CloseSession(ctx context.Context, id uuid.UUID) error {
	err := op.storage.CloseByID(ctx, id)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeSession, "Не удалось закрыть сессию")
	}
	return nil
}

// CloseAllSessionsByType закрывает все сессии указанного типа.
func (op *SessionOperations) CloseAllSessionsByType(ctx context.Context, incoming bool) error {
	err := op.storage.CloseAllByType(ctx, incoming)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeSession, "Не удалось закрыть сессии по типу")
	}
	return nil
}

// FormatSessionInfo форматирует информацию о сессии для отображения.
func (op *SessionOperations) FormatSessionInfo(sess session.Session) string {
	if sess == nil {
		return "Сессия не выбрана"
	}

	status := "Закрыта"
	if sess.IsOpen() {
		status = "Открыта"
	}

	direction := "Исходящее"
	if sess.IsIncoming() {
		direction = "Входящее"
	}

	result := ""
	result += "ID: " + sess.GetID().String() + "\n"
	result += "Направление: " + direction + "\n"
	result += "Статус: " + status + "\n"

	if peer := getPeerFromSession(sess); peer != nil {
		result += "\nПир:\n"
		result += "  Имя: " + peer.Name + "\n"
		result += "  IP: " + peer.IP.String() + "\n"
		result += "  Порт: " + fmt.Sprintf("%d", peer.Port) + "\n"
	}

	return result
}

// FormatSessionListElement форматирует сессию для отображения в списке.
func (op *SessionOperations) FormatSessionListElement(sess session.Session) string {
	if sess == nil {
		return ""
	}

	direction := "→"
	if sess.IsIncoming() {
		direction = "←"
	}

	status := "●"
	if !sess.IsOpen() {
		status = "○"
	}

	peerInfo := "Неизвестно"
	if peer := getPeerFromSession(sess); peer != nil {
		peerInfo = peer.Name
	}

	return direction + " " + peerInfo + " " + status
}

// getPeerFromSession извлекает пир из сессии.
// Поскольку интерфейс Session не предоставляет метода для получения Peer,
// используем рефлексию для получения поля Peer из BaseSession.
func getPeerFromSession(sess session.Session) *domain.Peer {
	if sess == nil {
		return nil
	}

	// Используем рефлексию для получения поля Peer
	// Это работает, потому что WebSocketSession встраивает BaseSession с полем Peer
	v := reflect.ValueOf(sess)

	// Если это указатель, разыменовываем
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Ищем поле Peer (может быть в BaseSession)
	peerField := v.FieldByName("Peer")
	if peerField.IsValid() && !peerField.IsZero() {
		if peer, ok := peerField.Interface().(*domain.Peer); ok {
			return peer
		}
	}

	// Возвращаем nil, если не удалось получить Peer
	return nil
}

// HasActiveSessionWithPeer проверяет, есть ли активная сессия с указанным пиром.
func (op *SessionOperations) HasActiveSessionWithPeer(peer *domain.Peer) (bool, session.Session) {
	sessions, err := op.GetSessions()
	if err != nil {
		return false, nil
	}

	for _, sess := range sessions {
		if !sess.IsOpen() {
			continue
		}

		if sessPeer := getPeerFromSession(sess); sessPeer != nil {
			if bytes.Equal(sessPeer.ID, peer.ID) {
				return true, sess
			}
		}
	}

	return false, nil
}

// RefreshSessions обновляет список сессий.
func (op *SessionOperations) RefreshSessions() ([]session.Session, error) {
	return op.GetSessions()
}
