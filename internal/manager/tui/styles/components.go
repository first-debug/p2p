package styles

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ComponentStyles содержит стили для UI компонентов.
type ComponentStyles struct {
	theme *ColorTheme
}

// NewComponentStyles создаёт новые стили компонентов с заданной темой.
func NewComponentStyles(theme *ColorTheme) *ComponentStyles {
	return &ComponentStyles{theme: theme}
}

// ApplyListStyles применяет стили к списку.
func (s *ComponentStyles) ApplyListStyles(list *tview.List) {
	list.SetMainTextColor(s.theme.ListText)
	list.SetSelectedTextColor(s.theme.ListSelectedFg)
	list.SetSelectedBackgroundColor(s.theme.ListSelectedBg)
	list.SetHighlightFullLine(true)
}

// ApplyTextViewStyles применяет стили к TextView.
func (s *ComponentStyles) ApplyTextViewStyles(text *tview.TextView, bordered bool) {
	text.SetTextColor(s.theme.Foreground)
	if bordered {
		text.SetBorder(true)
		text.SetBorderColor(s.theme.Border)
		text.SetTitleColor(s.theme.PanelTitle)
	}
}

// ApplyChatStyles применяет стили к чату.
func (s *ComponentStyles) ApplyChatStyles(chat *tview.TextView) {
	chat.SetDynamicColors(true)
	chat.SetWordWrap(true)
	chat.SetBorder(true)
	chat.SetBorderColor(s.theme.Border)
	chat.SetTitleColor(s.theme.PanelTitle)
}

// ApplyInputFieldStyles применяет стили к полю ввода.
func (s *ComponentStyles) ApplyInputFieldStyles(input *tview.InputField) {
	input.SetFieldBackgroundColor(s.theme.StatusBarBg)
	input.SetFieldTextColor(s.theme.StatusBarFg)
	input.SetLabelColor(s.theme.PanelTitle)
}

// ApplyFormStyles применяет стили к форме.
func (s *ComponentStyles) ApplyFormStyles(form *tview.Form) {
	form.SetButtonBackgroundColor(s.theme.ListSelectedBg)
	form.SetButtonTextColor(s.theme.ListSelectedFg)
	form.SetFieldBackgroundColor(s.theme.StatusBarBg)
	form.SetFieldTextColor(s.theme.StatusBarFg)
	form.SetLabelColor(s.theme.PanelTitle)
}

// ApplyModalStyles применяет стили к модальному окну.
func (s *ComponentStyles) ApplyModalStyles(modal *tview.Modal) {
	modal.SetBackgroundColor(s.theme.StatusBarBg)
	modal.SetTextColor(s.theme.StatusBarFg)
}

// GetTheme возвращает текущую тему.
func (s *ComponentStyles) GetTheme() *ColorTheme {
	return s.theme
}

// FormatMessage форматирует сообщение для чата с цветами.
func (s *ComponentStyles) FormatMessage(text string, incoming bool, timestamp string) string {
	color := s.theme.ChatOutgoing
	if incoming {
		color = s.theme.ChatIncoming
	}
	// Формат: [timestamp] сообщение с цветом
	return tview.Escape(timestamp) + " " +
		"[" + color.Name() + "::b]" + tview.Escape(text) + "[-::]"
}

// FormatSystemMessage форматирует системное сообщение.
func (s *ComponentStyles) FormatSystemMessage(text string) string {
	return "[" + s.theme.ChatSystem.Name() + "::i]*** " + tview.Escape(text) + " [-::]"
}

// GetPanelTitleStyle возвращает стиль для заголовка панели.
func (s *ComponentStyles) GetPanelTitleStyle() tcell.Style {
	return tcell.StyleDefault.Foreground(s.theme.PanelTitle)
}

// GetBorderStyle возвращает стиль для границы.
func (s *ComponentStyles) GetBorderStyle(active bool) tcell.Style {
	if active {
		return tcell.StyleDefault.Foreground(s.theme.BorderActive)
	}
	return tcell.StyleDefault.Foreground(s.theme.Border)
}
