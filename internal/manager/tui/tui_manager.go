// Package tui предоставляет современный TUI (Text User Interface) менеджер для управления пиром.
//
// Архитектура:
//   - config: конфигурация горячих клавиш
//   - styles: стили, темы и оформление компонентов
//   - errors: обработчики ошибок
//   - business: бизнес-логика (операции с пирами, сессиями, чатом)
//   - ui: UI компоненты (макеты, чат, строка состояния)
//
// Интерфейс разделён на части:
//   - Левая панель: список пиров (вверху), информация о пире (внизу)
//   - Правая панель: список сессий (вверху), чат (внизу)
//   - Строка состояния: локальная информация
//   - Подсказка: горячие клавиши
package tui

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/first-debug/p2p/internal/client"
	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/explorer"
	"github.com/first-debug/p2p/internal/manager/tui/business"
	"github.com/first-debug/p2p/internal/manager/tui/config"
	"github.com/first-debug/p2p/internal/manager/tui/errors"
	"github.com/first-debug/p2p/internal/manager/tui/styles"
	"github.com/first-debug/p2p/internal/manager/tui/ui"
	"github.com/first-debug/p2p/internal/session"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer-storage"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session-storage"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// logger - логгер для TUI менеджера.
var logger = errors.Logger

// TuiManager - основной менеджер TUI интерфейса.
type TuiManager struct {
	// Зависимости
	explorer       explorer.Explorer
	peerStorage    peerstorage.PeerStorage
	sessionStorage sessionstorage.SessionStorage
	client         client.Client

	// Конфигурация
	hotkeyConfig *config.HotkeyConfig

	// Стили
	theme  *styles.ColorTheme
	styles *styles.ComponentStyles

	// UI компоненты
	uiComponents *ui.UIComponents
	statusBar    *ui.StatusBar
	chatWindow   *ui.ChatWindow

	// Бизнес-логика
	peerOps      *business.PeerOperations
	sessionOps   *business.SessionOperations
	chatOps      *business.ChatOperations

	// Приложение
	app *tview.Application

	// Контекст и синхронизация
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Состояние
	isRunning      bool
	startTime      time.Time
	leftPanelActive bool

	// Выбранные элементы
	selectedPeer    *domain.Peer
	selectedSession session.Session

	// Данные для отображения
	peers    []domain.Peer
	sessions []session.Session

	// Каналы для обновления UI
	peerUpdateChan    chan []domain.Peer
	sessionUpdateChan chan []session.Session
}

// NewTuiManager создаёт новый TUI менеджер.
func NewTuiManager(
	explorer explorer.Explorer,
	peerStorage peerstorage.PeerStorage,
	sessionStorage sessionstorage.SessionStorage,
	client client.Client,
) (*TuiManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &TuiManager{
		explorer:          explorer,
		peerStorage:       peerStorage,
		sessionStorage:    sessionStorage,
		client:            client,
		hotkeyConfig:      config.DefaultHotkeyConfig(),
		ctx:               ctx,
		cancel:            cancel,
		isRunning:         false,
		startTime:         time.Now(),
		leftPanelActive:   true,
		peerUpdateChan:    make(chan []domain.Peer, 10),
		sessionUpdateChan: make(chan []session.Session, 10),
	}

	// Инициализация стилей
	manager.theme = styles.ModernDarkTheme()
	manager.styles = styles.NewComponentStyles(manager.theme)

	// Инициализация бизнес-логики
	manager.peerOps = business.NewPeerOperations(peerStorage, explorer)
	manager.sessionOps = business.NewSessionOperations(sessionStorage, client)
	manager.chatOps = business.NewChatOperations()

	// Инициализация UI компонентов
	manager.uiComponents = ui.NewUIComponents(manager.styles)
	manager.uiComponents.Init()

	// Инициализация строки состояния
	manager.statusBar = ui.NewStatusBar(manager.uiComponents.StatusBar, manager.styles)

	// Инициализация окна чата
	manager.chatWindow = ui.NewChatWindow(
		manager.uiComponents.ChatView,
		manager.uiComponents.ChatInput,
		manager.chatOps,
		manager.styles,
	)

	// Инициализация приложения
	manager.app = tview.NewApplication()

	// Построение макета
	mainLayout := manager.uiComponents.BuildMainLayout()
	manager.uiComponents.BuildPages(mainLayout)

	// Настройка обработчиков
	manager.setupInputCapture()
	manager.setupListHandlers()

	// Обновление строки состояния
	manager.updateStatusBar()

	return manager, nil
}

// Run запускает TUI менеджер (блокирующий вызов).
func (m *TuiManager) Run() error {
	if m.isRunning {
		return errors.NewErrorf(errors.ErrorTypeUI, "Менеджер уже запущен")
	}

	m.isRunning = true

	// Запуск фоновых задач
	m.startBackgroundTasks()

	// Обработка обновлений UI
	go m.handleUIUpdates()

	// Установка корневого виджета
	m.app.SetRoot(m.uiComponents.Pages, true)

	// Запуск приложения
	if err := m.app.Run(); err != nil {
		return errors.WrapError(err, errors.ErrorTypeUI, "Ошибка запуска приложения")
	}

	return nil
}

// Stop останавливает TUI менеджер.
func (m *TuiManager) Stop(ctx context.Context) error {
	m.cancel()
	m.wg.Wait()
	m.isRunning = false
	return nil
}

// startBackgroundTasks запускает фоновые задачи обновления.
func (m *TuiManager) startBackgroundTasks() {
	// Обновление списков каждые 500мс
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
				m.refreshLists()
			}
		}
	}()

	// Обновление строки состояния каждую секунду
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
				m.app.QueueUpdateDraw(func() {
					m.updateStatusBar()
				})
			}
		}
	}()
}

// handleUIUpdates обрабатывает обновления UI из каналов.
func (m *TuiManager) handleUIUpdates() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case peers := <-m.peerUpdateChan:
			m.app.QueueUpdateDraw(func() {
				m.updatePeerList(peers)
			})
		case sessions := <-m.sessionUpdateChan:
			m.app.QueueUpdateDraw(func() {
				m.updateSessionList(sessions)
			})
		}
	}
}

// refreshLists обновляет списки пиров и сессий.
func (m *TuiManager) refreshLists() {
	// Обновление пиров
	peers, err := m.peerOps.GetPeers()
	if err == nil {
		select {
		case m.peerUpdateChan <- peers:
		default:
		}
	}

	// Обновление сессий
	sessions, err := m.sessionOps.GetSessions()
	if err == nil {
		select {
		case m.sessionUpdateChan <- sessions:
		default:
		}
	}
}

// updatePeerList обновляет список пиров.
func (m *TuiManager) updatePeerList(peers []domain.Peer) {
	m.peers = peers

	var items []string
	for _, peer := range peers {
		items = append(items, fmt.Sprintf("%s (%s:%d)", peer.Name, peer.IP.String(), peer.Port))
	}

	m.uiComponents.SetPeerListItems(items, 0)
	m.statusBar.SetPeerCount(len(peers))
}

// updateSessionList обновляет список сессий.
func (m *TuiManager) updateSessionList(sessions []session.Session) {
	m.sessions = sessions

	var items []string
	for _, sess := range sessions {
		items = append(items, m.sessionOps.FormatSessionListElement(sess))
	}

	m.uiComponents.SetSessionListItems(items, 0)
	m.statusBar.SetSessionCount(len(sessions))
}

// updateStatusBar обновляет строку состояния.
func (m *TuiManager) updateStatusBar() {
	uptime := time.Since(m.startTime)

	info := ui.StatusInfo{
		LocalIP:      ui.GetLocalIP(),
		LocalPort:    8001, // TODO: получить из конфигурации
		PeerID:       "local", // TODO: получить ID локального пира
		PeerCount:    len(m.peers),
		SessionCount: len(m.sessions),
		IsConnected:  len(m.sessions) > 0,
		Uptime:       uptime,
	}

	m.statusBar.Update(info)
}

// setupInputCapture настраивает обработку горячих клавиш.
func (m *TuiManager) setupInputCapture() {
	m.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Обработка горячих клавиш
		switch event.Key() {
		case m.hotkeyConfig.Quit:
			m.app.Stop()
			return nil

		case m.hotkeyConfig.Refresh:
			m.refreshLists()
			return nil

		case m.hotkeyConfig.Emit:
			m.emitPeerInfo()
			return nil

		case m.hotkeyConfig.Info:
			m.showPeerInfo()
			return nil

		case m.hotkeyConfig.Help:
			m.showHelp()
			return nil

		case m.hotkeyConfig.NextPanel:
			m.switchPanel(true)
			return nil

		case m.hotkeyConfig.PrevPanel:
			m.switchPanel(false)
			return nil

		case m.hotkeyConfig.Connect:
			m.connectToSelectedPeer()
			return nil

		case m.hotkeyConfig.Disconnect:
			m.disconnectSelectedSession()
			return nil
		}

		// Обработка символов (для горячих клавиш с Ctrl+буква)
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'Q', 'q':
				if event.Modifiers()&tcell.ModCtrl != 0 {
					m.app.Stop()
					return nil
				}
			case 'R', 'r':
				if event.Modifiers()&tcell.ModCtrl != 0 {
					m.refreshLists()
					return nil
				}
			case 'C', 'c':
				if event.Modifiers()&tcell.ModCtrl != 0 {
					m.connectToSelectedPeer()
					return nil
				}
			case 'D', 'd':
				if event.Modifiers()&tcell.ModCtrl != 0 {
					m.disconnectSelectedSession()
					return nil
				}
			case 'E', 'e':
				if event.Modifiers()&tcell.ModCtrl != 0 {
					m.emitPeerInfo()
					return nil
				}
			case 'I', 'i':
				if event.Modifiers()&tcell.ModCtrl != 0 {
					m.showPeerInfo()
					return nil
				}
			case 'H', 'h':
				if event.Modifiers()&tcell.ModCtrl != 0 {
					m.showHelp()
					return nil
				}
			}
		}

		return event
	})
}

// setupListHandlers настраивает обработчики для списков.
func (m *TuiManager) setupListHandlers() {
	// Обработчик выбора пира
	m.uiComponents.PeerList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(m.peers) {
			m.selectedPeer = &m.peers[index]
			m.uiComponents.SetPeerInfo(m.peerOps.FormatPeerInfo(m.selectedPeer))

			// Проверяем, есть ли уже сессия с этим пиром
			hasSession, sess := m.sessionOps.HasActiveSessionWithPeer(m.selectedPeer)
			if hasSession && sess != nil {
				// Открываем чат с существующей сессией
				m.openChat(sess)
			}
		}
	})

	// Обработчик выбора сессии
	m.uiComponents.SessionList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(m.sessions) {
			m.selectedSession = m.sessions[index]
			m.openChat(m.selectedSession)
		}
	})

	// Обработчик отправки сообщения в чате
	m.uiComponents.ChatInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := m.uiComponents.GetChatInputText()
			if text != "" {
				m.chatWindow.SendMessage(m.ctx, text)
				m.uiComponents.SetInputText("")
			}
		} else if key == tcell.KeyEscape {
			// Выход из чата без закрытия сессии
			m.chatWindow.Close()
			m.app.SetFocus(m.uiComponents.PeerList)
		}
	})
}

// connectToSelectedPeer подключается к выбранному пиру.
func (m *TuiManager) connectToSelectedPeer() {
	if m.selectedPeer == nil {
		m.showNotification("Ошибка", "Выберите пир для подключения")
		return
	}

	// Проверяем, есть ли уже сессия
	hasSession, sess := m.sessionOps.HasActiveSessionWithPeer(m.selectedPeer)
	if hasSession && sess != nil {
		m.openChat(sess)
		return
	}

	// Подключаемся к пиру
	newSession, err := m.sessionOps.ConnectToPeer(m.ctx, *m.selectedPeer)
	if err != nil {
		m.showNotification("Ошибка подключения", errors.HandleError(err, "Не удалось подключиться"))
		return
	}

	// Обновляем список сессий
	m.refreshLists()

	// Открываем чат
	m.openChat(newSession)
}

// openChat открывает чат для сессии.
func (m *TuiManager) openChat(sess session.Session) {
	if sess == nil {
		return
	}

	peerName := "Неизвестно"
	if peer := getPeerFromSession(sess); peer != nil {
		peerName = peer.Name
	}

	m.chatWindow.SetTitle(peerName)
	m.chatWindow.Open(m.ctx, sess)
	m.app.SetFocus(m.uiComponents.ChatInput)
}

// disconnectSelectedSession закрывает выбранную сессию.
func (m *TuiManager) disconnectSelectedSession() {
	if m.selectedSession == nil {
		m.showNotification("Ошибка", "Выберите сессию для отключения")
		return
	}

	err := m.sessionOps.CloseSession(m.ctx, m.selectedSession.GetID())
	if err != nil {
		m.showNotification("Ошибка", errors.HandleError(err, "Не удалось закрыть сессию"))
		return
	}

	// Если чат был открыт с этой сессией, закрываем его
	if m.chatWindow.GetCurrentSession() != nil &&
		m.chatWindow.GetCurrentSession().GetID() == m.selectedSession.GetID() {
		m.chatWindow.Close()
	}

	m.refreshLists()
	m.showNotification("Успех", "Сессия закрыта")
}

// emitPeerInfo распространяет информацию о себе.
func (m *TuiManager) emitPeerInfo() {
	err := m.peerOps.Emit()
	if err != nil {
		m.showNotification("Ошибка", errors.HandleError(err, "Не удалось распространить информацию"))
		return
	}
	m.showNotification("Успех", "Информация распространена")
}

// showPeerInfo показывает информацию о выбранном пире.
func (m *TuiManager) showPeerInfo() {
	if m.selectedPeer != nil {
		m.uiComponents.SetPeerInfo(m.peerOps.FormatPeerInfo(m.selectedPeer))
	}
}

// showHelp показывает справку по горячим клавишам.
func (m *TuiManager) showHelp() {
	helpText := formatHelpText(m.hotkeyConfig)
	m.uiComponents.ShowHelpModal(m.uiComponents.Pages, helpText)
}

// showNotification показывает уведомление.
func (m *TuiManager) showNotification(title, text string) {
	m.uiComponents.ShowModal(m.uiComponents.Pages, title, text, []string{"OK"}, nil)
}

// switchPanel переключает активную панель.
func (m *TuiManager) switchPanel(next bool) {
	if next {
		m.leftPanelActive = !m.leftPanelActive
	}

	if m.leftPanelActive {
		m.app.SetFocus(m.uiComponents.PeerList)
		m.uiComponents.SetPanelActive(true)
	} else {
		m.app.SetFocus(m.uiComponents.SessionList)
		m.uiComponents.SetPanelActive(false)
	}
}

// getPeerFromSession извлекает пир из сессии.
// Поскольку интерфейс Session не предоставляет метода для получения Peer,
// используем рефлексию для получения поля Peer из BaseSession.
func getPeerFromSession(sess session.Session) *domain.Peer {
	if sess == nil {
		return nil
	}

	// Используем рефлексию для получения поля Peer
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

// formatHelpText форматирует текст справки.
func formatHelpText(cfg *config.HotkeyConfig) string {
	return `[yellow]Основные команды:[-]
  Ctrl+Q - Выход из приложения
  Ctrl+R - Обновить списки
  Ctrl+E - Распространить информацию о себе
  Ctrl+C - Подключиться к выбранному пиру
  Ctrl+D - Отключить выбранную сессию
  Ctrl+I - Показать информацию
  Ctrl+H - Эта справка
  Tab - Переключить панель
  ↑/↓ - Навигация по списку
  Enter - Выбрать элемент

[yellow]Чат:[-]
  Ctrl+Q - Выйти из чата (сохранить сессию)
  Ctrl+W - Выйти и закрыть сессию
  Enter - Отправить сообщение
  Escape - Назад

` + cfg.HotkeyHelp()
}
