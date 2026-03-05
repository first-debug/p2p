// Package business предоставляет бизнес-логику для TUI менеджера.
package business

import (
	"context"
	"fmt"

	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/explorer"
	"github.com/first-debug/p2p/internal/manager/tui/errors"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer-storage"
)

// PeerOperations содержит операции для работы с пирами.
type PeerOperations struct {
	storage peerstorage.PeerStorage
	explorer explorer.Explorer
}

// NewPeerOperations создаёт новые операции для работы с пирами.
func NewPeerOperations(storage peerstorage.PeerStorage, explorer explorer.Explorer) *PeerOperations {
	return &PeerOperations{
		storage:  storage,
		explorer: explorer,
	}
}

// GetPeers возвращает список всех пиров из хранилища.
func (op *PeerOperations) GetPeers() ([]domain.Peer, error) {
	peers, err := op.storage.GetAll()
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeStorage, "Не удалось получить список пиров")
	}
	return peers, nil
}

// GetPeerByID возвращает пир по ID.
func (op *PeerOperations) GetPeerByID(id []byte) (*domain.Peer, error) {
	peer, err := op.storage.GetByID(id)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeStorage, "Не удалось получить пир по ID")
	}
	return &peer, nil
}

// Emit распространяет информацию о себе в сеть.
func (op *PeerOperations) Emit() error {
	if op.explorer == nil {
		return errors.NewErrorf(errors.ErrorTypeNetwork, "Explorer не инициализирован")
	}
	err := op.explorer.Emit()
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeNetwork, "Не удалось распространить информацию о себе")
	}
	return nil
}

// RefreshPeers обновляет список пиров из хранилища.
func (op *PeerOperations) RefreshPeers(ctx context.Context) ([]domain.Peer, error) {
	// Сначала пробуем обновить через explorer, если он есть
	if op.explorer != nil {
		// Emit может быть долгим, поэтому делаем это в горутине если нужно
		// Но для простоты вызываем напрямую
		_ = op.explorer.Emit()
	}

	// Получаем обновлённый список
	return op.GetPeers()
}

// FormatPeerInfo форматирует информацию о пире для отображения.
func (op *PeerOperations) FormatPeerInfo(peer *domain.Peer) string {
	if peer == nil {
		return "Пир не выбран"
	}

	result := ""
	result += "Имя: " + peer.Name + "\n"
	result += "ID: " + formatPeerID(peer.ID) + "\n"
	result += "IP: " + peer.IP.String() + "\n"
	result += "Порт: " + formatInt(peer.Port) + "\n"

	if len(peer.Files) > 0 {
		result += "\nФайлы:\n"
		for name, path := range peer.Files {
			result += "  • " + name + ": " + path + "\n"
		}
	}

	return result
}

// formatPeerID форматирует ID пир для отображения.
func formatPeerID(id []byte) string {
	if len(id) == 0 {
		return "N/A"
	}
	if len(id) > 8 {
		return fmt.Sprintf("%x...", id[:8])
	}
	return fmt.Sprintf("%x", id)
}

// formatInt форматирует число для отображения.
func formatInt(n int) string {
	return fmt.Sprintf("%d", n)
}

