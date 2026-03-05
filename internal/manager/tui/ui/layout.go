// Package ui предоставляет UI компоненты для TUI менеджера.
package ui

import (
	"github.com/first-debug/p2p/internal/manager/tui/styles"
	"github.com/rivo/tview"
)

// UIComponents содержит все UI компоненты.
type UIComponents struct {
	// Основные контейнеры
	Pages      *tview.Pages
	MainLayout *tview.Flex

	// Левая панель
	PeerList       *tview.List
	PeerInfoText   *tview.TextView
	LeftPanel      *tview.Flex

	// Правая панель
	SessionList    *tview.List
	ChatView       *tview.TextView
	ChatInput      *tview.InputField
	RightPanel     *tview.Flex

	// Строка состояния
	StatusBar      *tview.TextView

	// Модальные окна
	ModalForm      *tview.Modal
	HelpModal      *tview.TextView

	// Стили
	Styles         *styles.ComponentStyles
}

// NewUIComponents создаёт новые UI компоненты.
func NewUIComponents(styles *styles.ComponentStyles) *UIComponents {
	return &UIComponents{
		Styles: styles,
	}
}

// Init initializes все UI компоненты.
func (c *UIComponents) Init() {
	// Инициализация списков
	c.PeerList = tview.NewList()
	c.PeerList.SetBorder(true)
	c.PeerList.SetTitle(" ПИРЫ ")
	c.PeerList.SetTitleAlign(tview.AlignLeft)
	c.Styles.ApplyListStyles(c.PeerList)

	c.SessionList = tview.NewList()
	c.SessionList.SetBorder(true)
	c.SessionList.SetTitle(" СЕССИИ ")
	c.SessionList.SetTitleAlign(tview.AlignLeft)
	c.Styles.ApplyListStyles(c.SessionList)

	// Инициализация текстовых полей
	c.PeerInfoText = tview.NewTextView()
	c.PeerInfoText.SetBorder(true)
	c.PeerInfoText.SetTitle(" ИНФОРМАЦИЯ О ПИРЕ ")
	c.PeerInfoText.SetTitleAlign(tview.AlignLeft)
	c.PeerInfoText.SetScrollable(true)
	c.Styles.ApplyTextViewStyles(c.PeerInfoText, true)

	c.ChatView = tview.NewTextView()
	c.ChatView.SetBorder(true)
	c.ChatView.SetTitle(" ЧАТ ")
	c.ChatView.SetTitleAlign(tview.AlignLeft)
	c.ChatView.SetScrollable(true)
	c.Styles.ApplyChatStyles(c.ChatView)

	c.ChatInput = tview.NewInputField()
	c.ChatInput.SetLabel("> ")
	c.ChatInput.SetFieldWidth(0)
	c.Styles.ApplyInputFieldStyles(c.ChatInput)

	// Строка состояния
	c.StatusBar = tview.NewTextView()
	c.StatusBar.SetBackgroundColor(c.Styles.GetTheme().StatusBarBg)
	c.StatusBar.SetTextColor(c.Styles.GetTheme().StatusBarFg)

	// Модальные окна
	c.ModalForm = tview.NewModal()
	c.Styles.ApplyModalStyles(c.ModalForm)

	c.HelpModal = tview.NewTextView()
	c.HelpModal.SetBorder(true)
	c.HelpModal.SetTitle(" ГОРЯЧИЕ КЛАВИШИ ")
	c.HelpModal.SetDynamicColors(true)
	c.Styles.ApplyTextViewStyles(c.HelpModal, true)
}

// BuildMainLayout строит основной макет приложения.
func (c *UIComponents) BuildMainLayout() *tview.Flex {
	// Левая панель: список пиров + информация о пире
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	leftPanel.AddItem(c.PeerList, 0, 1, true)
	leftPanel.AddItem(c.PeerInfoText, 0, 1, false)
	c.LeftPanel = leftPanel

	// Правая панель: список сессий + чат + ввод
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	rightPanel.AddItem(c.SessionList, 0, 1, false)
	rightPanel.AddItem(c.ChatView, 0, 2, false)
	rightPanel.AddItem(c.ChatInput, 1, 1, false)
	c.RightPanel = rightPanel

	// Основной горизонтальный макет
	mainLayout := tview.NewFlex().SetDirection(tview.FlexColumn)
	mainLayout.AddItem(leftPanel, 0, 1, true)
	mainLayout.AddItem(rightPanel, 0, 1, false)
	c.MainLayout = mainLayout

	return mainLayout
}

// BuildPages создаёт контейнер страниц.
func (c *UIComponents) BuildPages(mainLayout *tview.Flex) *tview.Pages {
	pages := tview.NewPages()
	pages.AddPage("main", mainLayout, true, true)
	c.Pages = pages
	return pages
}

// SetPeerListItems устанавливает элементы списка пиров.
func (c *UIComponents) SetPeerListItems(items []string, selected int) {
	c.PeerList.Clear()
	for _, item := range items {
		c.PeerList.AddItem(item, "", 0, nil)
	}
	if selected >= 0 && selected < len(items) {
		c.PeerList.SetCurrentItem(selected)
	}
}

// SetSessionListItems устанавливает элементы списка сессий.
func (c *UIComponents) SetSessionListItems(items []string, selected int) {
	c.SessionList.Clear()
	for _, item := range items {
		c.SessionList.AddItem(item, "", 0, nil)
	}
	if selected >= 0 && selected < len(items) {
		c.SessionList.SetCurrentItem(selected)
	}
}

// SetPeerInfo устанавливает информацию о пире.
func (c *UIComponents) SetPeerInfo(text string) {
	c.PeerInfoText.SetText(text)
}

// AppendChatMessage добавляет сообщение в чат.
func (c *UIComponents) AppendChatMessage(formatted string) {
	currentText := c.ChatView.GetText(false)
	c.ChatView.SetText(currentText + "\n" + formatted)
	c.ChatView.ScrollToEnd()
}

// ClearChat очищает чат.
func (c *UIComponents) ClearChat() {
	c.ChatView.Clear()
}

// SetChatInputText устанавливает текст поля ввода чата.
func (c *UIComponents) SetChatInputText(text string) {
	c.ChatInput.SetText(text)
}

// GetChatInputText возвращает текст поля ввода чата.
func (c *UIComponents) GetChatInputText() string {
	return c.ChatInput.GetText()
}

// SetStatusBarText устанавливает текст строки состояния.
func (c *UIComponents) SetStatusBarText(text string) {
	c.StatusBar.SetText(" " + text)
}

// ShowModal показывает модальное окно.
func (c *UIComponents) ShowModal(pages *tview.Pages, title, text string, buttons []string, handler func(int, string)) {
	modal := tview.NewModal()
	modal.SetText(text)
	modal.SetTitle(title)
	modal.SetBackgroundColor(c.Styles.GetTheme().StatusBarBg)
	modal.SetTextColor(c.Styles.GetTheme().StatusBarFg)
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if handler != nil {
			handler(buttonIndex, buttonLabel)
		}
		pages.RemovePage("modal")
	})
	pages.AddPage("modal", modal, true, true)
}

// ShowHelpModal показывает справку по горячим клавишам.
func (c *UIComponents) ShowHelpModal(pages *tview.Pages, helpText string) {
	c.HelpModal.SetText(helpText)
	
	// Центрированное модальное окно
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(c.HelpModal, 20, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)
	
	pages.AddPage("help", flex, true, true)
}

// SetFocus устанавливает фокус на указанный виджет.
func (c *UIComponents) SetFocus(app *tview.Application, widget tview.Primitive) {
	app.SetFocus(widget)
}

// GetFocused возвращает текущий сфокусированный виджет.
func (c *UIComponents) GetFocused(app *tview.Application) tview.Primitive {
	return app.GetFocus()
}

// SetPanelActive устанавливает активность панели (подсветка границы).
func (c *UIComponents) SetPanelActive(leftActive bool) {
	if leftActive {
		c.PeerList.SetBorderColor(c.Styles.GetTheme().BorderActive)
		c.SessionList.SetBorderColor(c.Styles.GetTheme().Border)
	} else {
		c.PeerList.SetBorderColor(c.Styles.GetTheme().Border)
		c.SessionList.SetBorderColor(c.Styles.GetTheme().BorderActive)
	}
}

// HideChatInput скрывает поле ввода чата (переключает фокус).
func (c *UIComponents) HideChatInput() {
	// InputField не имеет SetVisible, поэтому просто не используем его
}

// ShowChatInput показывает поле ввода чата.
func (c *UIComponents) ShowChatInput() {
	// InputField не имеет SetVisible
}

// IsChatVisible возвращает true, если чат видим.
func (c *UIComponents) IsChatVisible() bool {
	return true
}

// SetInputText устанавливает текст поля ввода чата.
func (c *UIComponents) SetInputText(text string) {
	c.ChatInput.SetText(text)
}

// GetInputText возвращает текст поля ввода чата.
func (c *UIComponents) GetInputText() string {
	return c.ChatInput.GetText()
}
