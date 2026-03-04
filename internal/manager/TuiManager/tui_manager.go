// Package tuimanager предоставляет TUI (Text User Interface) менеджер для управления пиром.
// Менеджер позволяет:
//   - Обнаруживать пиры в сети
//   - Принимать входящие подключения
//   - Создавать новые подключения к пирам
//   - Управлять активными сессиями
//   - Просматривать информацию о подключенных пирах
//
// Интерфейс построен с использованием библиотеки tview и предоставляет
// интуитивно понятный текстовый интерфейс с цветовым кодированием и горячими клавишами.
package tuimanager

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/first-debug/p2p/internal/domain"
	"github.com/first-debug/p2p/internal/explorer"
	"github.com/first-debug/p2p/internal/session"
	peerstorage "github.com/first-debug/p2p/internal/storage/peer-storage"
	sessionstorage "github.com/first-debug/p2p/internal/storage/session-storage"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// logger - логгер для вывода отладочной информации TUI менеджера.
// Использует стандартный логгер с префиксом [TUIManager] для удобства чтения логов.
var logger = log.New(os.Stderr, "[TUIManager] ", log.LstdFlags)

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

	// application - основное TUI приложение tview
	// Управляет всем интерфейсом, обработкой событий и отрисовкой
	application *tview.Application

	// pages - многостраничный контейнер для переключения между экранами
	// Используется для отображения различных "экранов" приложения
	pages *tview.Pages

	// mainLayout - основной горизонтальный макет приложения
	// Содержит левую панель (пиры) и правую панель (сессии)
	mainLayout *tview.Flex

	// peerList - виджет списка для отображения обнаруженных пиров
	// Позволяет выбирать пиры для подключения
	peerList *tview.List

	// sessionList - виджет списка для отображения активных сессий
	// Показывает текущие подключения и их статус
	sessionList *tview.List

	// statusText - текстовое поле для отображения статуса приложения
	// Показывает текущее состояние, ошибки и уведомления
	statusText *tview.TextView

	// infoText - текстовое поле для отображения подробной информации
	// Используется для показа деталей о выбранном пире или сессии
	infoText *tview.TextView

	// modalForm - модальное окно для ввода данных (например, адреса для подключения)
	// Используется при создании новых подключений
	modalForm *tview.Modal

	// connectForm - форма для подключения к пиру с полем ввода
	// Содержит поле ввода адреса и кнопки управления
	connectForm *tview.Form

	// helpText - текст справки с описанием горячих клавиш
	// Отображается в нижней части экрана
	helpText string

	// ctx - контекст для управления жизненным циклом менеджера
	// Позволяет корректно завершать работу при остановке
	ctx context.Context

	// cancel - функция отмены контекста
	// Вызывается при остановке менеджера
	cancel context.CancelFunc

	// wg - группа ожидания для фоновых задач
	// Используется для ожидания завершения горутин при остановке
	wg sync.WaitGroup

	// updateInterval - интервал обновления UI (в миллисекундах)
	// Определяет, как часто обновляются списки пиров и сессий
	updateInterval time.Duration

	// selectedPeer - выбранный в списке пир
	// Используется для операций с выбранным элементом
	selectedPeer *domain.Peer

	// selectedSession - выбранная в списке сессия
	// Используется для операций с выбранной сессией
	selectedSession session.Session

	// isRunning - флаг работы менеджера
	// Указывает, запущен ли TUI интерфейс
	isRunning bool

	// mu - мьютекс для потокобезопасности
	// Защищает общие данные при доступе из разных горутин
	mu sync.RWMutex
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
//   - error: ошибка при создании (например, при инициализации tview)
func NewTuiManager(
	explorer explorer.Explorer,
	peerStorage peerstorage.PeerStorage,
	sessionStorage sessionstorage.SessionStorage,
) (*TuiManager, error) {
	// Создаем контекст для управления жизненным циклом
	ctx, cancel := context.WithCancel(context.Background())

	// Инициализируем базовую структуру менеджера
	manager := &TuiManager{
		explorer:       explorer,
		peerStorage:    peerStorage,
		sessionStorage: sessionStorage,
		ctx:            ctx,
		cancel:         cancel,
		updateInterval: 500 * time.Millisecond, // Обновление UI каждые 500мс
		isRunning:      false,
		helpText:       "Q: Выход | R: Обновить | C: Подключиться | D: Отключить | I: Инфо | H: Помощь",
	}

	// Инициализируем TUI приложение
	// tview.NewApplication() создает новое приложение с экраном и событиями
	manager.application = tview.NewApplication()

	// Создаем основные UI компоненты
	manager.initUIComponents()

	// Настраиваем макет приложения
	manager.setupLayout()

	// Регистрируем обработчики горячих клавиш
	manager.setupInputCapture()

	// Запускаем фоновые задачи обновления UI
	manager.startBackgroundTasks()

	return manager, nil
}

// initUIComponents инициализирует все UI компоненты.
// Этот метод вызывается один раз при создании менеджера.
func (m *TuiManager) initUIComponents() {
	// ============================================
	// СПИСОК ПИРОВ (левая панель)
	// ============================================
	// tview.List - это виджет списка с возможностью выбора элементов
	// Каждый элемент имеет основной текст (peer name) и вспомогательный (IP:port)
	m.peerList = tview.NewList()
	m.peerList.ShowSecondaryText(true)                                       // Показывать вторичный текст (адрес пира)
	m.peerList.SetHighlightFullLine(true)                                    // Подсвечивать всю строку при выборе
	m.peerList.SetSelectedBackgroundColor(tcell.ColorDarkBlue)               // Цвет выделения
	m.peerList.SetSelectedTextColor(tcell.ColorWhite)                        // Цвет текста выделения
	m.peerList.SetTitle("Обнаруженные пиры [yellow](C) - подключить[white]") // Заголовок списка
	m.peerList.SetBorder(true)                                               // Рисовать рамку вокруг списка
	m.peerList.SetBorderPadding(1, 1, 1, 1)                                  // Отступы внутри рамки

	// Обработчик выбора элемента в списке пиров
	// Вызывается при клике или нажатии Enter на элементе
	// Примечание: SetSelectedFunc принимает функцию с параметром rune (клавиша активации)
	m.peerList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		m.onPeerSelected(index)
	})

	// ============================================
	// СПИСОК СЕССИЙ (правая панель)
	// ============================================
	// Отображает активные подключения с информацией о типе (входящее/исходящее)
	m.sessionList = tview.NewList()
	m.sessionList.ShowSecondaryText(true) // Показывать тип сессии и статус
	m.sessionList.SetHighlightFullLine(true)
	m.sessionList.SetSelectedBackgroundColor(tcell.ColorDarkGreen)
	m.sessionList.SetSelectedTextColor(tcell.ColorWhite)
	m.sessionList.SetTitle("Активные сессии [yellow](D) - отключить[white]")
	m.sessionList.SetBorder(true)
	m.sessionList.SetBorderPadding(1, 1, 1, 1)

	// Обработчик выбора элемента в списке сессий
	m.sessionList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		m.onSessionSelected(index)
	})

	// ============================================
	// ТЕКСТ СТАТУСА (нижняя левая панель)
	// ============================================
	// TextView - виджет для отображения текста
	// Используется для показа текущего статуса приложения
	m.statusText = tview.NewTextView()
	m.statusText.SetDynamicColors(true)                    // Включить поддержку цветов в тексте
	m.statusText.SetText("[green]Статус:[white] Ожидание") // Начальный текст с цветом
	m.statusText.SetTitle("Статус")
	m.statusText.SetBorder(true)
	// Примечание: SetTextAlign отсутствует в tview, используем форматирование текста

	// ============================================
	// ТЕКСТ ИНФОРМАЦИИ (нижняя правая панель)
	// ============================================
	// Отображает подробную информацию о выбранном элементе
	m.infoText = tview.NewTextView()
	m.infoText.SetDynamicColors(true)
	m.infoText.SetText("[yellow]Выберите элемент для просмотра информации[white]")
	m.infoText.SetTitle("Информация")
	m.infoText.SetBorder(true)

	// ============================================
	// МОДАЛЬНОЕ ОКНО
	// ============================================
	// Используется для общих уведомлений и подтверждений
	// Modal - это всплывающее окно поверх основного интерфейса
	m.modalForm = tview.NewModal()
	m.modalForm.SetText("Введите адрес пира (IP:port)")
	m.modalForm.AddButtons([]string{"Подключить", "Отмена"}) // Кнопки действия
	m.modalForm.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		m.onModalDone(buttonIndex, buttonLabel)
	})

	// ============================================
	// ФОРМА ПОДКЛЮЧЕНИЯ
	// ============================================
	// tview.Form - виджет для создания форм с полями ввода
	// Используется для ввода адреса пира при подключении
	m.connectForm = tview.NewForm()
	m.connectForm.AddInputField("IP:port", "", 30, nil, nil)
	// Добавляем кнопки управления формой
	m.connectForm.AddButton("Подключить", func() {
		m.onConnectFormSubmit()
	})
	m.connectForm.AddButton("Отмена", func() {
		m.hideConnectForm()
	})
	// Настраиваем заголовок и рамку
	m.connectForm.SetTitle("Подключение к пиру")
	m.connectForm.SetBorder(true)
	m.connectForm.SetBorderPadding(1, 1, 1, 1)
}

// setupLayout настраивает основную компоновку интерфейса.
// Создает двухколоночный макет с дополнительной нижней панелью.
func (m *TuiManager) setupLayout() {
	// ============================================
	// ОСНОВНОЙ ГОРИЗОНТАЛЬНЫЙ МАКЕТ
	// ============================================
	// tview.Flex - гибкий контейнер, распределяющий пространство между элементами
	// Параметр 0 в AddItem означает "автоматическое распределение"
	// Параметр 1 означает "вес элемента" при распределении пространства
	m.mainLayout = tview.NewFlex()
	m.mainLayout.SetDirection(tview.FlexColumn)
	// Левая колонка - список пиров (занимает 50% ширины)
	m.mainLayout.AddItem(m.peerList, 0, 1, true)
	// Правая колонка - список сессий (занимает 50% ширины)
	m.mainLayout.AddItem(m.sessionList, 0, 1, false)

	// ============================================
	// НИЖНЯЯ ПАНЕЛЬ
	// ============================================
	// Горизонтальный макет для статуса и информации
	bottomLayout := tview.NewFlex()
	bottomLayout.SetDirection(tview.FlexColumn)
	// Статус занимает 30% ширины нижней панели
	bottomLayout.AddItem(m.statusText, 0, 1, false)
	// Информация занимает 70% ширины нижней панели
	bottomLayout.AddItem(m.infoText, 0, 2, false)

	// ============================================
	// ВЕРТИКАЛЬНЫЙ МАКЕТ (основной + нижняя панель)
	// ============================================
	// Объединяем основной макет и нижнюю панель вертикально
	mainVerticalLayout := tview.NewFlex()
	mainVerticalLayout.SetDirection(tview.FlexRow)
	// Основной макет занимает всё доступное пространство кроме 3 строк
	mainVerticalLayout.AddItem(m.mainLayout, 0, 1, true)
	// Нижняя панель фиксированной высоты 3 строки
	mainVerticalLayout.AddItem(bottomLayout, 3, 1, false)

	// ============================================
	// МНОГОСТРАНИЧНЫЙ КОНТЕЙНЕР
	// ============================================
	// tview.Pages позволяет переключаться между разными "экранами"
	// Страница "main" - основной интерфейс
	// Страница "modal" - модальное окно (показывается поверх основного)
	// Страница "connect" - форма подключения
	m.pages = tview.NewPages()
	m.pages.AddPage("main", mainVerticalLayout, true, true)
	m.pages.AddPage("modal", m.modalForm, true, false)     // Скрыта по умолчанию
	m.pages.AddPage("connect", m.connectForm, true, false) // Скрыта по умолчанию

	// Устанавливаем корневой виджет приложения
	// Все события и отрисовка будут идти через этот виджет
	m.application.SetRoot(m.pages, true)
}

// setupInputCapture настраивает обработку глобальных горячих клавиш.
// InputCapture перехватывает все нажатия клавиш до их обработки виджетами.
func (m *TuiManager) setupInputCapture() {
	m.application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// ============================================
		// Q - ВЫХОД ИЗ ПРИЛОЖЕНИЯ
		// ============================================
		// Завершает работу TUI менеджера и освобождает ресурсы
		if event.Key() == tcell.KeyRune && event.Rune() == 'Q' {
			m.Stop(m.ctx)
			return nil // Обрабатываем событие (не передаем дальше)
		}

		// ============================================
		// R - ОБНОВИТЬ СПИСКИ
		// ============================================
		// Принудительно обновляет списки пиров и сессий
		if event.Key() == tcell.KeyRune && event.Rune() == 'R' {
			m.updatePeerList()
			m.updateSessionList()
			m.setStatus("Обновление списков...")
			return nil
		}

		// ============================================
		// C - ПОДКЛЮЧИТЬСЯ К ПИРУ
		// ============================================
		// Открывает форму для ввода адреса пира
		if event.Key() == tcell.KeyRune && event.Rune() == 'C' {
			m.showConnectForm()
			return nil
		}

		// ============================================
		// D - ОТКЛЮЧИТЬСЯ
		// ============================================
		// Закрывает выбранную сессию
		if event.Key() == tcell.KeyRune && event.Rune() == 'D' {
			m.disconnectSelected()
			return nil
		}

		// ============================================
		// I - ПОКАЗАТЬ ИНФОРМАЦИЮ
		// ============================================
		// Обновляет панель информации о выбранном элементе
		if event.Key() == tcell.KeyRune && event.Rune() == 'I' {
			m.showInfo()
			return nil
		}

		// ============================================
		// H - ПОКАЗАТЬ ПОМОЩЬ
		// ============================================
		// Показывает всплывающее окно со списком горячих клавиш
		if event.Key() == tcell.KeyRune && event.Rune() == 'H' {
			m.showHelp()
			return nil
		}

		// ============================================
		// Escape - ЗАКРЫТЬ МОДАЛЬНОЕ ОКНО/ФОРМУ
		// ============================================
		// Закрывает открытые модальные окна при нажатии Escape
		if event.Key() == tcell.KeyEscape {
			m.hideConnectForm()
			m.pages.HidePage("modal")
			return nil
		}

		// ============================================
		// Tab - ПЕРЕКЛЮЧЕНИЕ МЕЖДУ ПАНЕЛЯМИ
		// ============================================
		// Переключает фокус между списком пиров и сессий
		if event.Key() == tcell.KeyTab {
			// Получаем текущий фокус
			if m.application.GetFocus() == m.peerList {
				m.application.SetFocus(m.sessionList)
			} else {
				m.application.SetFocus(m.peerList)
			}
			return nil
		}

		// ============================================
		// Возвращаем событие для дальнейшей обработки
		// ============================================
		// Если клавиша не обработана, передаем её текущему виджету
		return event
	})
}

// startBackgroundTasks запускает фоновые задачи для обновления UI.
// Эти задачи работают в отдельных горутинах и периодически обновляют данные.
func (m *TuiManager) startBackgroundTasks() {
	// ============================================
	// ЗАДАЧА 1: ПЕРИОДИЧЕСКОЕ ОБНОВЛЕНИЕ СПИСКОВ
	// ============================================
	// Запускаем горутину для автоматического обновления списков
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		// Создаем тикер с заданным интервалом
		ticker := time.NewTicker(m.updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// По истечении интервала обновляем списки
				// Обновление выполняется в главной горутине приложения
				// для потокобезопасности с UI
				m.application.QueueUpdateDraw(func() {
					m.updatePeerList()
					m.updateSessionList()
				})
			case <-m.ctx.Done():
				// При получении сигнала отмены завершаем горутину
				return
			}
		}
	}()

	// ============================================
	// ЗАДАЧА 2: ОБНАРУЖЕНИЕ ПИРОВ (если explorer доступен)
	// ============================================
	// Периодически вызываем Emit для поиска новых пиров
	if m.explorer != nil {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()

			discoveryTicker := time.NewTicker(5 * time.Second)
			defer discoveryTicker.Stop()

			for {
				select {
				case <-discoveryTicker.C:
					// Вызываем Emit для отправки/приема multicast сообщений
					if err := m.explorer.Emit(); err != nil {
						logger.Printf("Ошибка обнаружения пиров: %v", err)
					}
				case <-m.ctx.Done():
					return
				}
			}
		}()
	}
}

// updatePeerList обновляет список обнаруженных пиров.
// Получает данные из peerStorage и отображает их в UI.
func (m *TuiManager) updatePeerList() {
	// Получаем все пиры из хранилища
	peers, err := m.peerStorage.GetAll()
	if err != nil {
		logger.Printf("Ошибка получения пиров: %v", err)
		return
	}

	// Очищаем текущий список
	m.peerList.Clear()

	// Если пиров нет, показываем сообщение
	if len(peers) == 0 {
		m.peerList.AddItem("Нет обнаруженных пиров", "", 0, nil)
		m.setStatus("[yellow]Ожидание пиров...[white]")
		return
	}

	// ============================================
	// ДОБАВЛЕНИЕ КАЖДОГО ПИРА В СПИСОК
	// ============================================
	// Для каждого пира создаем элемент списка с именем и адресом
	for i, peer := range peers {
		// Формируем строку адреса IP:port
		address := fmt.Sprintf("%s:%d", peer.IP.String(), peer.Port)

		// Создаем элемент списка
		// mainText - имя пира, secondaryText - адрес
		// 0 - код символа для иконки (можно использовать Unicode символы)
		// Последняя функция - обработчик выбора (nil означает использование SetSelectedFunc)
		m.peerList.AddItem(
			peer.Name, // Основной текст
			address,   // Вторичный текст
			0,         // Иконка
			nil,       // Обработчик (используем глобальный из SetSelectedFunc)
		)

		// Сохраняем выбранного пира
		if m.selectedPeer != nil && string(m.selectedPeer.ID) == string(peer.ID) {
			m.peerList.SetCurrentItem(i)
		}
	}

	// Обновляем статус
	m.setStatus(fmt.Sprintf("[green]Пиров обнаружено: %d[white]", len(peers)))
}

// updateSessionList обновляет список активных сессий.
// Получает данные из sessionStorage и отображает их в UI.
func (m *TuiManager) updateSessionList() {
	// Получаем все сессии из хранилища
	sessions, err := m.sessionStorage.GetAll()
	if err != nil {
		logger.Printf("Ошибка получения сессий: %v", err)
		return
	}

	// Очищаем текущий список
	m.sessionList.Clear()

	// Если сессий нет, показываем сообщение
	if len(sessions) == 0 {
		m.sessionList.AddItem("Нет активных сессий", "", 0, nil)
		return
	}

	// ============================================
	// ДОБАВЛЕНИЕ КАЖДОЙ СЕССИИ В СПИСОК
	// ============================================
	for i, sess := range sessions {
		// ============================================
		// ПОЛУЧЕНИЕ ИНФОРМАЦИИ О ПИРЕ
		// ============================================
		// Используем интерфейс session.Session для получения данных
		// В реальной реализации сессия может содержать ссылку на Peer
		peerName := "Неизвестно"

		// Проверяем, есть ли у сессии метод для получения пира
		// Это временное решение до рефакторинга интерфейса Session
		// Используем type assertion для проверки наличия метода GetPeer()
		if sessionWithPeer, ok := sess.(interface{ GetPeer() *domain.Peer }); ok {
			if peer := sessionWithPeer.GetPeer(); peer != nil {
				peerName = peer.Name
			}
		}
		// Примечание: приведение к *WebSocketSession не будет работать,
		// т.к. реальная сессия находится в пакете websocket и не экспортируется

		// Определяем тип сессии (входящая/исходящая)
		sessionType := "Исходящее"
		if sess.IsIncoming() {
			sessionType = "Входящее"
		}

		// Определяем статус сессии (открыта/закрыта)
		status := "[green]Открыто[white]"
		if !sess.IsOpen() {
			status = "[red]Закрыто[white]"
		}

		// Формируем вторичный текст с типом и статусом
		secondaryText := fmt.Sprintf("%s | %s", sessionType, status)

		// Добавляем элемент в список
		m.sessionList.AddItem(
			peerName,      // Основной текст
			secondaryText, // Вторичный текст
			0,             // Иконка
			nil,           // Обработчик (используем глобальный из SetSelectedFunc)
		)

		// Сохраняем выбранную сессию
		if m.selectedSession != nil && m.selectedSession.GetID() == sess.GetID() {
			m.sessionList.SetCurrentItem(i)
		}
	}
}

// onPeerSelected вызывается при выборе элемента в списке пиров.
// Обновляет внутреннее состояние и панель информации.
func (m *TuiManager) onPeerSelected(index int) {
	// Получаем все пиры для определения выбранного
	peers, err := m.peerStorage.GetAll()
	if err != nil {
		logger.Printf("Ошибка получения пиров: %v", err)
		return
	}

	if index >= 0 && index < len(peers) {
		// Сохраняем выбранного пира
		m.selectedPeer = &peers[index]

		// Обновляем панель информации
		m.showPeerInfo(m.selectedPeer)
	}
}

// onSessionSelected вызывается при выборе элемента в списке сессий.
// Обновляет внутреннее состояние и панель информации.
func (m *TuiManager) onSessionSelected(index int) {
	// Получаем все сессии для определения выбранной
	sessions, err := m.sessionStorage.GetAll()
	if err != nil {
		logger.Printf("Ошибка получения сессий: %v", err)
		return
	}

	if index >= 0 && index < len(sessions) {
		// Сохраняем выбранную сессию
		m.selectedSession = sessions[index]

		// Обновляем панель информации
		m.showSessionInfo(m.selectedSession)
	}
}

// showConnectModal показывает модальное окно для подключения к пиру.
// Позволяет ввести адрес пира в формате IP:port.
// Устарело: используется showConnectForm вместо этой функции.
func (m *TuiManager) showConnectModal() {
	// Показываем модальное окно поверх основного интерфейса
	m.pages.ShowPage("modal")
}

// showConnectForm показывает форму для подключения к пиру.
// Форма содержит поле ввода адреса и кнопки управления.
func (m *TuiManager) showConnectForm() {
	// Сбрасываем поле ввода к пустому значению
	// GetFormItem(0) получает первый элемент формы (поле ввода)
	if input, ok := m.connectForm.GetFormItem(0).(*tview.InputField); ok {
		input.SetText("") // Очищаем поле
	}

	// Показываем страницу формы
	m.pages.ShowPage("connect")

	// Устанавливаем фокус на форму
	// SetFocus переключает фокус ввода на указанный виджет
	m.application.SetFocus(m.connectForm)

	m.setStatus("[yellow]Введите адрес пира...[white]")
}

// hideConnectForm скрывает форму подключения.
// Возвращает фокус на основной интерфейс.
func (m *TuiManager) hideConnectForm() {
	m.pages.HidePage("connect")
	m.application.SetFocus(m.peerList) // Возвращаем фокус на список пиров
	m.setStatus("[green]Статус:[white] Ожидание")
}

// onConnectFormSubmit обрабатывает отправку формы подключения.
// Вызывается при нажатии кнопки "Подключить".
func (m *TuiManager) onConnectFormSubmit() {
	// Получаем введенный адрес из поля ввода
	// GetFormItem(0) - первое поле формы (InputField)
	if input, ok := m.connectForm.GetFormItem(0).(*tview.InputField); ok {
		address := input.GetText()

		// Проверяем, что адрес не пустой
		if address == "" {
			m.setStatus("[red]Введите адрес пира[white]")
			return
		}

		// ============================================
		// ЛОГИКА ПОДКЛЮЧЕНИЯ К ПИРУ
		// ============================================
		// Здесь должна быть логика подключения
		// Для этого нужен метод Connect в менеджере или использование server
		logger.Printf("Попытка подключения к адресу: %s", address)

		// Скрываем форму после отправки
		m.hideConnectForm()

		// Показываем статус подключения
		m.setStatus(fmt.Sprintf("[yellow]Подключение к %s...[white]", address))

		// TODO: Реализовать фактическое подключение
		// 1. Разрешить адрес (IP и порт)
		// 2. Создать WebSocket соединение
		// 3. Создать сессию и добавить в sessionStorage
	}
}

// onModalDone обрабатывает завершение работы модального окна.
// Вызывается при нажатии кнопки в модальном окне.
func (m *TuiManager) onModalDone(buttonIndex int, buttonLabel string) {
	// Скрываем модальное окно
	m.pages.HidePage("modal")

	// Если нажата кнопка "Подключить"
	if buttonLabel == "Подключить" {
		// Получаем введенный адрес из модального окна
		// Примечание: для полноценной реализации нужно добавить поле ввода
		m.setStatus("[yellow]Подключение...[white]")

		// Здесь должна быть логика подключения к пиру
		// Для этого нужен метод Connect в менеджере
	}
}

// disconnectSelected отключает выбранную сессию.
// Закрывает соединение и удаляет сессию из хранилища.
func (m *TuiManager) disconnectSelected() {
	if m.selectedSession == nil {
		m.setStatus("[red]Выберите сессию для отключения[white]")
		return
	}

	// Закрываем сессию через контекст
	m.sessionStorage.CloseByID(m.ctx, m.selectedSession.GetID())

	m.setStatus("[green]Сессия закрыта[white]")
	m.selectedSession = nil

	// Обновляем список сессий
	m.updateSessionList()
}

// showInfo обновляет панель информации.
// Показывает детали о выбранном элементе (пир или сессия).
func (m *TuiManager) showInfo() {
	if m.selectedPeer != nil {
		m.showPeerInfo(m.selectedPeer)
	} else if m.selectedSession != nil {
		m.showSessionInfo(m.selectedSession)
	} else {
		m.infoText.SetText("[yellow]Выберите элемент для просмотра информации[white]")
	}
}

// showPeerInfo отображает подробную информацию о пире.
func (m *TuiManager) showPeerInfo(peer *domain.Peer) {
	// Форматируем информацию в виде текста с цветами
	info := fmt.Sprintf(
		"[green]Имя:[white] %s\n"+
			"[green]ID:[white] %x\n"+
			"[green]IP:[white] %s\n"+
			"[green]Port:[white] %d\n"+
			"[green]Файлов:[white] %d",
		peer.Name,
		peer.ID,
		peer.IP.String(),
		peer.Port,
		len(peer.Files),
	)

	m.infoText.SetText(info)
}

// showSessionInfo отображает подробную информацию о сессии.
func (m *TuiManager) showSessionInfo(sess session.Session) {
	// Получаем ID сессии в строковом формате
	sessionID := sess.GetID().String()

	// ============================================
	// ПОЛУЧЕНИЕ ИНФОРМАЦИИ О ПИРЕ
	// ============================================
	peerName := "Неизвестно"

	// Проверяем, есть ли у сессии метод для получения пира
	// Используем type assertion для проверки наличия метода GetPeer()
	if sessionWithPeer, ok := sess.(interface{ GetPeer() *domain.Peer }); ok {
		if peer := sessionWithPeer.GetPeer(); peer != nil {
			peerName = peer.Name
		}
	}

	// Определяем тип сессии (входящая/исходящая)
	sessionType := "Исходящее"
	if sess.IsIncoming() {
		sessionType = "Входящее"
	}

	// Определяем статус
	status := "Открыто"
	if !sess.IsOpen() {
		status = "Закрыто"
	}

	// Форматируем информацию
	info := fmt.Sprintf(
		"[green]ID сессии:[white] %s\n"+
			"[green]Пир:[white] %s\n"+
			"[green]Тип:[white] %s\n"+
			"[green]Статус:[white] %s",
		sessionID,
		peerName,
		sessionType,
		status,
	)

	m.infoText.SetText(info)
}

// showHelp показывает всплывающее окно с помощью.
func (m *TuiManager) showHelp() {
	// Создаем модальное окно со списком горячих клавиш
	help := tview.NewModal()
	help.SetText(
		"Горячие клавиши:\n\n" +
			"Q - Выход из приложения\n" +
			"R - Обновить списки\n" +
			"C - Подключиться к пиру\n" +
			"D - Отключить сессию\n" +
			"I - Показать информацию\n" +
			"H - Эта справка\n\n" +
			"Навигация:\n" +
			"Tab - Переключение между панелями\n" +
			"Стрелки - Выбор элемента\n" +
			"Escape - Закрыть форму/модальное окно",
	)
	help.AddButtons([]string{"OK"})
	help.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		m.pages.HidePage("help")
	})

	// Добавляем страницу помощи и показываем её
	m.pages.AddPage("help", help, true, false)
	m.pages.ShowPage("help")
}

// setStatus устанавливает текст статуса приложения.
// Использует цветовое форматирование tview.
func (m *TuiManager) setStatus(status string) {
	m.statusText.SetText(fmt.Sprintf("[green]Статус:[white] %s", status))
}

// Run запускает TUI менеджер.
// Блокирует выполнение до завершения приложения.
func (m *TuiManager) Run() error {
	m.mu.Lock()
	m.isRunning = true
	m.mu.Unlock()

	// Запускаем приложение
	// Run() блокирует выполнение до вызова Stop()
	if err := m.application.Run(); err != nil {
		return fmt.Errorf("ошибка запуска приложения: %w", err)
	}

	return nil
}

// Stop останавливает TUI менеджер и освобождает ресурсы.
// Корректно завершает все фоновые задачи.
func (m *TuiManager) Stop(ctx context.Context) {
	m.mu.Lock()
	if !m.isRunning {
		m.mu.Unlock()
		return
	}
	m.isRunning = false
	m.mu.Unlock()

	// Отменяем контекст для сигнализации фоновым задачам
	m.cancel()

	// Ждем завершения фоновых задач
	m.wg.Wait()

	// Останавливаем TUI приложение
	m.application.Stop()

	logger.Println("TUI менеджер остановлен")
}

// WebSocketSession - временная структура для доступа к полям сессии.
// В реальной реализации нужно использовать правильный тип сессии.
// Эта структура дублирует websocket.WebSocketSession для возможности
// приведения типа и получения доступа к полю Peer.
type WebSocketSession struct {
	session.BaseSession
	Peer *domain.Peer
}

// GetPeer возвращает пир, связанный с сессией.
// Этот метод используется для совместимости с интерфейсом.
func (ws *WebSocketSession) GetPeer() *domain.Peer {
	return ws.Peer
}

// Ensure TuiManager implements manager.Manager interface
var _ interface {
	Run() error
	Stop(context.Context)
} = (*TuiManager)(nil)
