// Package tuimanager предоставляет TUI (Text User Interface) менеджер для управления пиром.
// Менеджер позволяет:
//   - Обнаруживать пиры в сети
//   - Принимать входящие подключения
//   - Создавать новые подключения к пирам
//   - Управлять активными сессиями
//   - Просматривать информацию о подключенных пирах
//
// Интерфейс построен с использованием библиотеки bubbletea и предоставляет
// интуитивно понятный текстовый интерфейс с цветовым кодированием и горячими клавишами.
package tuimanager

import (
	"context"
	"fmt"

	"github.com/first-debug/p2p/internal/explorer"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer-storage"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session-storage"

	tea "github.com/charmbracelet/bubbletea"
)

// TuiManager - основной структ TUI менеджера.
// Реализует интерфейс manager.Manager и предоставляет полнофункциональный
// текстовый интерфейс для управления пиром.
type TuiManager struct {
	// explorer - компонент для обнаружения пиров в сети (multicast/UDP)
	explorer explorer.Explorer

	// peerStorage - хранилище информации о пирах
	peerStorage peerstorage.PeerStorage

	// sessionStorage - хранилище активных сессий
	sessionStorage sessionstorage.SessionStorage

	// ctx - контекст для управления жизненным циклом менеджера
	ctx context.Context

	// cancel - функция отмены контекста
	cancel context.CancelFunc

	// isRunning - флаг работы менеджера
	isRunning bool

	// program - bubbletea программа
	program *tea.Program
}

// NewTuiManager создает новый экземпляр TUI менеджера.
//
// Параметры:
//   - explorer: компонент обнаружения пиров
//   - peerStorage: хранилище информации о пирах
//   - sessionStorage: хранилище сессий
//
// Возвращает:
//   - *TuiManager: настроенный экземпляр менеджера
//   - error: ошибка при создании
func NewTuiManager(
	explorer explorer.Explorer,
	peerStorage peerstorage.PeerStorage,
	sessionStorage sessionstorage.SessionStorage,
) (*TuiManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &TuiManager{
		explorer:       explorer,
		peerStorage:    peerStorage,
		sessionStorage: sessionStorage,
		ctx:            ctx,
		cancel:         cancel,
		isRunning:      false,
	}

	return manager, nil
}

// Run запускает TUI менеджер.
// Блокирует выполнение до завершения приложения.
func (m *TuiManager) Run() error {
	m.isRunning = true

	// Инициализируем модель
	initialModel := m.initModel()

	// Создаем и запускаем bubbletea программу
	// tea.WithAltScreen использует альтернативный экран терминала
	m.program = tea.NewProgram(initialModel, tea.WithAltScreen())

	// Запускаем программу (блокирующий вызов)
	if _, err := m.program.Run(); err != nil {
		return fmt.Errorf("ошибка запуска приложения: %w", err)
	}

	return nil
}

// Stop останавливает TUI менеджер и освобождает ресурсы.
// Корректно завершает все фоновые задачи.
func (m *TuiManager) Stop(ctx context.Context) {
	if !m.isRunning {
		return
	}
	m.isRunning = false

	// Отменяем контекст для сигнализации фоновым задачам
	m.cancel()

	// Останавливаем bubbletea программу
	if m.program != nil {
		m.program.Quit()
	}

	logger.Println("TUI менеджер остановлен")
}

// Ensure TuiManager implements manager.Manager interface
var _ interface {
	Run() error
	Stop(context.Context)
} = (*TuiManager)(nil)
