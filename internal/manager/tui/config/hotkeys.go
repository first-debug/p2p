// Package config предоставляет конфигурацию для TUI менеджера.
package config

import "github.com/gdamore/tcell/v2"

// HotkeyConfig содержит конфигурацию горячих клавиш.
// Все клавиши вынесены в одно место для удобного изменения.
type HotkeyConfig struct {
	// Основные действия
	Quit          tcell.Key // Выход из приложения
	Refresh       tcell.Key // Обновить списки пиров и сессий
	Connect       tcell.Key // Подключиться к выбранному пиру
	Disconnect    tcell.Key // Закрыть выбранную сессию
	Info          tcell.Key // Показать информацию о выбранном элементе
	Help          tcell.Key // Показать справку
	Emit          tcell.Key // Распространить информацию о себе

	// Навигация
	NextPanel     tcell.Key // Переключиться на следующую панель
	PrevPanel     tcell.Key // Переключиться на предыдущую панель
	SelectUp      tcell.Key // Выбрать элемент выше
	SelectDown    tcell.Key // Выбрать элемент ниже
	SelectEnter   tcell.Key // Выбрать элемент (Enter)
	SelectEscape  tcell.Key // Отменить выбор/закрыть окно

	// Чат
	ChatQuit      tcell.Key // Выйти из чата (сохраняя сессию)
	ChatClose     tcell.Key // Выйти из чата и закрыть сессию
	ChatSend      tcell.Key // Отправить сообщение
}

// DefaultHotkeyConfig возвращает конфигурацию горячих клавиш по умолчанию.
func DefaultHotkeyConfig() *HotkeyConfig {
	return &HotkeyConfig{
		// Основные действия
		Quit:       tcell.KeyCtrlQ,
		Refresh:    tcell.KeyCtrlR,
		Connect:    tcell.KeyCtrlC,
		Disconnect: tcell.KeyCtrlD,
		Info:       tcell.KeyCtrlI,
		Help:       tcell.KeyCtrlH,
		Emit:       tcell.KeyCtrlE,

		// Навигация
		NextPanel:    tcell.KeyTab,
		PrevPanel:    tcell.KeyBacktab,
		SelectUp:     tcell.KeyUp,
		SelectDown:   tcell.KeyDown,
		SelectEnter:  tcell.KeyEnter,
		SelectEscape: tcell.KeyEscape,

		// Чат
		ChatQuit:     tcell.KeyCtrlQ,
		ChatClose:    tcell.KeyCtrlW,
		ChatSend:     tcell.KeyEnter,
	}
}

// HotkeyHelp возвращает строку справки по горячим клавишам.
func (h *HotkeyConfig) HotkeyHelp() string {
	return "Ctrl+Q: Выход | Ctrl+R: Обновить | Ctrl+C: Подключиться | Ctrl+D: Отключить | Ctrl+E: Эмит | Ctrl+I: Инфо | Ctrl+H: Помощь | Tab: Панель | ↑/↓: Выбор | Enter: Выбрать | Esc: Отмена"
}

// HotkeyHelpChat возвращает строку справки по горячим клавишам в режиме чата.
func (h *HotkeyConfig) HotkeyHelpChat() string {
	return "Ctrl+Q: Выйти (сохранить) | Ctrl+W: Выйти и закрыть | Enter: Отправить | Esc: Назад"
}
