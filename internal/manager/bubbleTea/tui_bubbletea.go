// Package tuimanager предоставляет TUI менеджер для управления пиром на основе bubbletea.
package tuimanager

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/explorer"
	"github.com/first-debug/p2p/internal/session"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer-storage"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session-storage"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var logger = log.New(os.Stderr, "[TUIManager] ", log.LstdFlags)

// === Стили UI ===

var (
	// Основные цвета
	colorPrimary   = lipgloss.Color("#7D56F4")
	colorSecondary = lipgloss.Color("#585858")
	colorError     = lipgloss.Color("#FF5555")
	colorSuccess   = lipgloss.Color("#50FA7B")
	colorWarning   = lipgloss.Color("#F1FA8C")
	colorInfo      = lipgloss.Color("#8BE9FD")

	// Стили для панелей
	panelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	focusedPanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(colorSuccess).
				Padding(0, 1)

	// Стили для списка
	listStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	listItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(colorPrimary).
				Bold(true)

	// Стили для статуса
	statusStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	// Стили для модальных окон
	modalStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			Background(lipgloss.Color("#1E1E2E"))

	modalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	modalButtonStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	modalButtonSelectedStyle = lipgloss.NewStyle().
					Foreground(colorSuccess).
					Background(colorPrimary).
					Bold(true).
					Padding(0, 2)
)

// === Типы фокуса ===

type focusType int

const (
	focusPeerList focusType = iota
	focusSessionList
	focusConnectForm
	focusModal
)

// === Модель приложения ===

type model struct {
	// Зависимости
	explorer       explorer.Explorer
	peerStorage    peerstorage.PeerStorage
	sessionStorage sessionstorage.SessionStorage

	// Списки
	peerList    list.Model
	sessionList list.Model

	// Текстовое поле для формы подключения
	connectInput textinput.Model

	// Состояние
	focus           focusType
	peers           []domain.Peer
	sessions        []session.Session
	selectedPeer    *domain.Peer
	selectedSession session.Session
	status          string
	errorMsg        string
	width           int
	height          int

	// Модальное окно
	showModal        bool
	modalTitle       string
	modalMessage     string
	modalButtons     []string
	modalButtonIndex int

	// Контекст и управление
	ctx    context.Context
	cancel context.CancelFunc
}

// initModel инициализирует bubbletea модель.
func (m *TuiManager) initModel() model {
	// === Настройка списка пиров ===
	peerDelegate := list.NewDefaultDelegate()
	peerDelegate.Styles.SelectedTitle = selectedItemStyle
	peerDelegate.Styles.NormalTitle = listItemStyle

	peerList := list.New([]list.Item{}, peerDelegate, 0, 0)
	peerList.Title = "Обнаруженные пиры"
	peerList.Styles.Title = modalTitleStyle
	peerList.SetShowHelp(false)
	peerList.SetFilteringEnabled(false)
	peerList.SetShowStatusBar(true)

	// === Настройка списка сессий ===
	sessionDelegate := list.NewDefaultDelegate()
	sessionDelegate.Styles.SelectedTitle = selectedItemStyle
	sessionDelegate.Styles.NormalTitle = listItemStyle

	sessionList := list.New([]list.Item{}, sessionDelegate, 0, 0)
	sessionList.Title = "Активные сессии"
	sessionList.Styles.Title = modalTitleStyle
	sessionList.SetShowHelp(false)
	sessionList.SetFilteringEnabled(false)
	sessionList.SetShowStatusBar(true)

	// === Настройка текстового поля ===
	connectInput := textinput.New()
	connectInput.Placeholder = "IP:port (например, 192.168.1.100:8001)"
	connectInput.CharLimit = 30
	connectInput.Width = 40

	return model{
		explorer:       m.explorer,
		peerStorage:    m.peerStorage,
		sessionStorage: m.sessionStorage,
		peerList:       peerList,
		sessionList:    sessionList,
		connectInput:   connectInput,
		focus:          focusPeerList,
		status:         "Ожидание",
		ctx:            m.ctx,
		cancel:         m.cancel,
	}
}

// === Init ===

func (m model) Init() tea.Cmd {
	return tea.Batch(
		loadPeers(m.peerStorage),
		loadSessions(m.sessionStorage),
		updateUITick(),
		explorerTick(),
	)
}

// === Update ===

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Обработка ошибок
	if errMsg, ok := msg.(errorMsg); ok {
		m.errorMsg = errMsg.err.Error()
		m.status = "Ошибка"
		return m, nil
	}

	// Обработка статуса
	if statusMsg, ok := msg.(statusMsg); ok {
		m.status = statusMsg.text
		return m, nil
	}

	// Обработка загрузки пиров
	if peersMsg, ok := msg.(peersLoadedMsg); ok {
		if peersMsg.err != nil {
			m.errorMsg = peersMsg.err.Error()
		} else {
			m.peers = peersMsg.peers
			m.updatePeerListItems()
		}
		return m, nil
	}

	// Обработка загрузки сессий
	if sessionsMsg, ok := msg.(sessionsLoadedMsg); ok {
		if sessionsMsg.err != nil {
			m.errorMsg = sessionsMsg.err.Error()
		} else {
			m.sessions = sessionsMsg.sessions
			m.updateSessionListItems()
		}
		return m, nil
	}

	// Обработка таймера обновления UI
	if _, ok := msg.(updateUITickMsg); ok {
		cmds = append(cmds, loadPeers(m.peerStorage), loadSessions(m.sessionStorage), updateUITick())
	}

	// Обработка таймера explorer
	if _, ok := msg.(explorerTickMsg); ok {
		if m.explorer != nil {
			go func() {
				if err := m.explorer.Emit(); err != nil {
					logger.Printf("Ошибка обнаружения пиров: %v", err)
				}
			}()
		}
		cmds = append(cmds, explorerTick())
	}

	// Обработка отправки формы подключения
	if connectMsg, ok := msg.(connectFormMsg); ok {
		m.showModal = false
		m.status = fmt.Sprintf("Подключение к %s...", connectMsg.address)
		// Здесь должна быть логика подключения

		// logger.Printf("Попытка подключения к адресу: %s", connectMsg.address)
		return m, nil
	}

	// Обработка размеров экрана
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		m.connectInput.Width = msg.Width/2 - 6
		return m, nil
	}

	// Если показано модальное окно, обрабатываем только его
	if m.showModal {
		return m.updateModal(msg)
	}

	// Если показана форма подключения, обрабатываем только её
	if m.focus == focusConnectForm {
		return m.updateConnectForm(msg)
	}

	// Обработка горячих клавиш
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, keyMap.Quit):
			return m, tea.Quit

		case key.Matches(keyMsg, keyMap.Refresh):
			cmds = append(cmds, loadPeers(m.peerStorage), loadSessions(m.sessionStorage))
			m.status = "Обновление списков..."

		case key.Matches(keyMsg, keyMap.Connect):
			m.showConnectForm()
			return m, nil

		case key.Matches(keyMsg, keyMap.Disconnect):
			return m, m.disconnectSelected()

		case key.Matches(keyMsg, keyMap.Info):
			m.showInfo()
			return m, nil

		case key.Matches(keyMsg, keyMap.Help):
			return m.showHelpModal()

		case key.Matches(keyMsg, keyMap.Tab):
			m.switchFocus()
			return m, nil

		case key.Matches(keyMsg, keyMap.Up):
			if m.focus == focusPeerList {
				m.peerList, _ = m.peerList.Update(msg)
			} else if m.focus == focusSessionList {
				m.sessionList, _ = m.sessionList.Update(msg)
			}

		case key.Matches(keyMsg, keyMap.Down):
			if m.focus == focusPeerList {
				m.peerList, _ = m.peerList.Update(msg)
			} else if m.focus == focusSessionList {
				m.sessionList, _ = m.sessionList.Update(msg)
			}

		case key.Matches(keyMsg, keyMap.Enter):
			if m.focus == focusPeerList && m.peerList.Index() >= 0 && m.peerList.Index() < len(m.peers) {
				m.selectedPeer = &m.peers[m.peerList.Index()]
			} else if m.focus == focusSessionList && m.sessionList.Index() >= 0 && m.sessionList.Index() < len(m.sessions) {
				m.selectedSession = m.sessions[m.sessionList.Index()]
			}
		}
	}

	// Обновляем текущий список в зависимости от фокуса
	if m.focus == focusPeerList {
		m.peerList, _ = m.peerList.Update(msg)
	} else if m.focus == focusSessionList {
		m.sessionList, _ = m.sessionList.Update(msg)
	}

	return m, tea.Batch(cmds...)
}

// updateModal обрабатывает ввод в модальном окне.
func (m model) updateModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, keyMap.Quit) || key.Matches(keyMsg, keyMap.Escape):
			m.showModal = false
			return m, nil

		case key.Matches(keyMsg, keyMap.Enter):
			// Обработка нажатия Enter
			if m.modalButtonIndex == 0 && m.modalTitle == "Помощь" {
				m.showModal = false
			}
			return m, nil

		case key.Matches(keyMsg, keyMap.Left):
			if m.modalButtonIndex > 0 {
				m.modalButtonIndex--
			}

		case key.Matches(keyMsg, keyMap.Right):
			if m.modalButtonIndex < len(m.modalButtons)-1 {
				m.modalButtonIndex++
			}
		}
	}

	return m, nil
}

// updateConnectForm обрабатывает ввод в форме подключения.
func (m model) updateConnectForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, keyMap.Quit):
			return m, tea.Quit

		case key.Matches(keyMsg, keyMap.Escape):
			m.hideConnectForm()
			return m, nil

		case key.Matches(keyMsg, keyMap.Enter):
			address := m.connectInput.Value()
			if address != "" {
				return m, tea.Batch(connectFormCmd(address), func() tea.Msg {
					return statusMsg{text: fmt.Sprintf("Подключение к %s...", address)}
				})
			}
			return m, nil
		}
	}

	// Обновляем текстовое поле
	m.connectInput, _ = m.connectInput.Update(msg)

	return m, tea.Batch(cmds...)
}

// === View ===

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Инициализация..."
	}

	// Вычисляем размеры панелей
	panelWidth := m.width/2 - 4
	panelHeight := m.height - 14

	if panelWidth < 20 {
		panelWidth = 20
	}
	if panelHeight < 5 {
		panelHeight = 5
	}

	m.peerList.SetSize(panelWidth, panelHeight)
	m.sessionList.SetSize(panelWidth, panelHeight)

	var b strings.Builder

	// Верхняя часть с панелями
	peerPanel := m.renderPeerPanel()
	sessionPanel := m.renderSessionPanel()

	// Размещаем панели рядом
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, peerPanel, sessionPanel))
	b.WriteString("\n\n")

	// Информация о выбранном элементе
	b.WriteString(m.renderInfoPanel())
	b.WriteString("\n\n")

	// Статус
	b.WriteString(statusStyle.Render("Статус: " + m.status))
	b.WriteString("\n\n")

	// Помощь
	b.WriteString(m.helpView())

	content := b.String()

	// Модальное окно
	if m.showModal {
		return m.renderModal(content)
	}

	// Форма подключения
	if m.focus == focusConnectForm {
		return m.renderConnectForm(content)
	}

	return content
}

// === Вспомогательные методы ===

// updatePeerListItems обновляет элементы списка пиров.
func (m *model) updatePeerListItems() {
	items := make([]list.Item, 0, len(m.peers))
	for _, peer := range m.peers {
		items = append(items, peerItem{peer: peer})
	}
	m.peerList.SetItems(items)
}

// updateSessionListItems обновляет элементы списка сессий.
func (m *model) updateSessionListItems() {
	items := make([]list.Item, 0, len(m.sessions))
	for _, sess := range m.sessions {
		items = append(items, sessionItem{session: sess})
	}
	m.sessionList.SetItems(items)
}

// renderPeerPanel рендерит панель пиров.
func (m model) renderPeerPanel() string {
	style := panelStyle
	if m.focus == focusPeerList {
		style = focusedPanelStyle
	}
	// Получаем содержимое списка и ограничиваем ширину
	listView := m.peerList.View()
	// Ограничиваем ширину панели
	return style.Render(listView)
}

// renderSessionPanel рендерит панель сессий.
func (m model) renderSessionPanel() string {
	style := panelStyle
	if m.focus == focusSessionList {
		style = focusedPanelStyle
	}
	// Получаем содержимое списка и ограничиваем ширину
	listView := m.sessionList.View()
	// Ограничиваем ширину панели
	return style.Render(listView)
}

// renderInfoPanel рендерит панель информации.
func (m model) renderInfoPanel() string {
	var info string
	panelWidth := m.width/2 - 4

	if panelWidth < 20 {
		panelWidth = 20
	}

	if m.selectedPeer != nil {
		info = fmt.Sprintf(
			"[green]Имя:[white] %s\n"+
				"[green]ID:[white] %x\n"+
				"[green]IP:[white] %s\n"+
				"[green]Port:[white] %d\n"+
				"[green]Файлов:[white] %d",
			m.selectedPeer.Name,
			m.selectedPeer.ID,
			m.selectedPeer.IP.String(),
			m.selectedPeer.Port,
			len(m.selectedPeer.Files),
		)
	} else if m.selectedSession != nil {
		sessionType := "Исходящее"
		if m.selectedSession.IsIncoming() {
			sessionType = "Входящее"
		}
		status := "Открыто"
		if !m.selectedSession.IsOpen() {
			status = "Закрыто"
		}
		peerName := "Неизвестно"
		if sessionWithPeer, ok := m.selectedSession.(interface{ GetPeer() *domain.Peer }); ok {
			if peer := sessionWithPeer.GetPeer(); peer != nil {
				peerName = peer.Name
			}
		}
		info = fmt.Sprintf(
			"[green]ID сессии:[white] %s\n"+
				"[green]Пир:[white] %s\n"+
				"[green]Тип:[white] %s\n"+
				"[green]Статус:[white] %s",
			m.selectedSession.GetID().String(),
			peerName,
			sessionType,
			status,
		)
	} else {
		info = "[yellow]Выберите элемент для просмотра информации[white]"
	}

	infoPanel := panelStyle.Width(panelWidth).Render(info)
	return infoPanel
}

// renderModal рендерит модальное окно поверх основного контента.
func (m model) renderModal(content string) string {
	modalWidth := 50

	// Создаем модальное окно с фиксированной шириной
	modalContent := fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		modalTitleStyle.Render(m.modalTitle),
		m.modalMessage,
		m.renderModalButtons(),
	)

	modal := modalStyle.Width(modalWidth).Render(modalContent)

	// Вычисляем позицию для центрирования
	x := (m.width - modalWidth - 4) / 2
	y := (m.height - 10) / 2

	// Добавляем отступы сверху и слева
	var b strings.Builder
	b.WriteString(content)
	for i := 0; i < y; i++ {
		b.WriteString("\n")
	}
	b.WriteString(strings.Repeat(" ", x))
	b.WriteString(modal)

	return b.String()
}

// renderModalButtons рендерит кнопки модального окна.
func (m model) renderModalButtons() string {
	var buttons []string
	for i, btn := range m.modalButtons {
		if i == m.modalButtonIndex {
			buttons = append(buttons, modalButtonSelectedStyle.Render(btn))
		} else {
			buttons = append(buttons, modalButtonStyle.Render(btn))
		}
	}
	return strings.Join(buttons, "  ")
}

// renderConnectForm рендерит форму подключения.
func (m model) renderConnectForm(content string) string {
	modalWidth := 50

	formContent := fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		modalTitleStyle.Render("Подключение к пиру"),
		m.connectInput.View(),
		"Enter - Подключить | Escape - Отмена",
	)

	form := modalStyle.Width(modalWidth).Render(formContent)

	// Вычисляем позицию для центрирования
	x := (m.width - modalWidth - 4) / 2
	y := (m.height - 10) / 2

	var b strings.Builder
	b.WriteString(content)
	for i := 0; i < y; i++ {
		b.WriteString("\n")
	}
	b.WriteString(strings.Repeat(" ", x))
	b.WriteString(form)

	return b.String()
}

// helpView возвращает строку с подсказками по горячим клавишам.
func (m model) helpView() string {
	return strings.Join([]string{
		"Q: Выход",
		"R: Обновить",
		"C: Подключиться",
		"D: Отключить",
		"I: Инфо",
		"H: Помощь",
		"Tab: Переключение",
		"↑/↓: Навигация",
	}, " | ")
}

// showConnectForm показывает форму подключения.
func (m *model) showConnectForm() {
	m.focus = focusConnectForm
	m.connectInput.SetValue("")
	m.connectInput.Focus()
}

// hideConnectForm скрывает форму подключения.
func (m *model) hideConnectForm() {
	m.focus = focusPeerList
	m.connectInput.Blur()
}

// switchFocus переключает фокус между панелями.
func (m *model) switchFocus() {
	switch m.focus {
	case focusPeerList:
		m.focus = focusSessionList
	case focusSessionList:
		m.focus = focusPeerList
	}
}

// showInfo обновляет информацию о выбранном элементе.
func (m *model) showInfo() {
	// Информация обновляется автоматически в renderInfoPanel
}

// showHelpModal показывает модальное окно помощи.
func (m model) showHelpModal() (tea.Model, tea.Cmd) {
	m.showModal = true
	m.modalTitle = "Помощь"
	m.modalMessage = "Q - Выход\nR - Обновить\nC - Подключиться\nD - Отключить\nI - Инфо\nH - Эта справка\nTab - Переключение между панелями\n↑/↓ - Навигация\nEnter - Выбрать элемент\nEscape - Закрыть форму"
	m.modalButtons = []string{"OK"}
	m.modalButtonIndex = 0
	return m, nil
}

// disconnectSelected отключает выбранную сессию.
func (m *model) disconnectSelected() tea.Cmd {
	if m.selectedSession == nil {
		return func() tea.Msg {
			return statusMsg{text: "Выберите сессию для отключения"}
		}
	}

	m.sessionStorage.CloseByID(m.ctx, m.selectedSession.GetID())
	m.selectedSession = nil

	return tea.Batch(
		loadSessions(m.sessionStorage),
		func() tea.Msg {
			return statusMsg{text: "Сессия закрыта"}
		},
	)
}

// === Команды ===

// loadPeers загружает пиры из хранилища.
func loadPeers(storage peerstorage.PeerStorage) tea.Cmd {
	return func() tea.Msg {
		peers, err := storage.GetAll()
		return peersLoadedMsg{peers: peers, err: err}
	}
}

// loadSessions загружает сессии из хранилища.
func loadSessions(storage sessionstorage.SessionStorage) tea.Cmd {
	return func() tea.Msg {
		sessions, err := storage.GetAll()
		return sessionsLoadedMsg{sessions: sessions, err: err}
	}
}

// updateUITick создает таймер обновления UI.
func updateUITick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return updateUITickMsg{}
	})
}

// explorerTick создает таймер обнаружения пиров.
func explorerTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		return explorerTickMsg{}
	})
}

// connectFormCmd создает команду для подключения.
func connectFormCmd(address string) tea.Cmd {
	return func() tea.Msg {
		return connectFormMsg{address: address}
	}
}

// === Элементы списка ===

// peerItem представляет элемент списка пиров.
type peerItem struct {
	peer domain.Peer
}

func (p peerItem) Title() string       { return p.peer.Name }
func (p peerItem) Description() string { return fmt.Sprintf("%s:%d", p.peer.IP.String(), p.peer.Port) }
func (p peerItem) FilterValue() string { return p.peer.Name }

// sessionItem представляет элемент списка сессий.
type sessionItem struct {
	session session.Session
}

func (s sessionItem) Title() string {
	peerName := "Неизвестно"
	if sessionWithPeer, ok := s.session.(interface{ GetPeer() *domain.Peer }); ok {
		if peer := sessionWithPeer.GetPeer(); peer != nil {
			peerName = peer.Name
		}
	}
	return peerName
}

func (s sessionItem) Description() string {
	sessionType := "Исходящее"
	if s.session.IsIncoming() {
		sessionType = "Входящее"
	}
	status := "Открыто"
	if !s.session.IsOpen() {
		status = "Закрыто"
	}
	return fmt.Sprintf("%s | %s", sessionType, status)
}

func (s sessionItem) FilterValue() string {
	if sessionWithPeer, ok := s.session.(interface{ GetPeer() *domain.Peer }); ok {
		if peer := sessionWithPeer.GetPeer(); peer != nil {
			return peer.Name
		}
	}
	return ""
}

// === Горячие клавиши ===

var keyMap = struct {
	Quit       key.Binding
	Refresh    key.Binding
	Connect    key.Binding
	Disconnect key.Binding
	Info       key.Binding
	Help       key.Binding
	Tab        key.Binding
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Escape     key.Binding
	Left       key.Binding
	Right      key.Binding
}{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("Q", "Выход"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("R", "Обновить"),
	),
	Connect: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("C", "Подключиться"),
	),
	Disconnect: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("D", "Отключить"),
	),
	Info: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("I", "Инфо"),
	),
	Help: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("H", "Помощь"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("Tab", "Переключение"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "Вверх"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "Вниз"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "Выбрать"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "Отмена"),
	),
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "Влево"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "Вправо"),
	),
}
