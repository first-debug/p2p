// Package styles предоставляет стили и темы для TUI менеджера.
package styles

import "github.com/gdamore/tcell/v2"

// ColorTheme содержит цветовую тему для TUI.
type ColorTheme struct {
	// Основные цвета
	Background       tcell.Color
	Foreground       tcell.Color
	Border           tcell.Color
	BorderActive     tcell.Color
	BorderInactive   tcell.Color

	// Цвета для списков
	ListText         tcell.Color
	ListSelectedBg   tcell.Color
	ListSelectedFg   tcell.Color
	ListSeparator    tcell.Color

	// Цвета для статусов
	StatusInfo       tcell.Color
	StatusWarning    tcell.Color
	StatusError      tcell.Color
	StatusSuccess    tcell.Color

	// Цвета для чата
	ChatIncoming     tcell.Color
	ChatOutgoing     tcell.Color
	ChatTimestamp    tcell.Color
	ChatSystem       tcell.Color

	// Цвета для панелей
	PanelHeader      tcell.Color
	PanelTitle       tcell.Color
	PanelSubtitle    tcell.Color

	// Цвета для строки состояния
	StatusBarBg      tcell.Color
	StatusBarFg      tcell.Color
	StatusBarHighlight tcell.Color
}

// ModernDarkTheme возвращает современную тёмную тему.
func ModernDarkTheme() *ColorTheme {
	return &ColorTheme{
		// Основные цвета
		Background:       tcell.ColorDefault,
		Foreground:       tcell.ColorWhite,
		Border:           tcell.ColorGray,      // Тёмно-серый
		BorderActive:     tcell.ColorGreen,     // Зелёный акцент
		BorderInactive:   tcell.ColorGray,      // Тёмно-серый

		// Цвета для списков
		ListText:         tcell.ColorWhite,
		ListSelectedBg:   tcell.ColorGreen,     // Зелёный
		ListSelectedFg:   tcell.ColorBlack,
		ListSeparator:    tcell.ColorGray,

		// Цвета для статусов
		StatusInfo:       tcell.ColorDodgerBlue,    // Синий
		StatusWarning:    tcell.ColorYellow,   // Жёлтый
		StatusError:      tcell.ColorRed,     // Красный
		StatusSuccess:    tcell.ColorGreen,     // Зелёный

		// Цвета для чата
		ChatIncoming:     tcell.ColorLightGreen,   // Светло-зелёный
		ChatOutgoing:     tcell.ColorLightBlue,    // Светло-синий
		ChatTimestamp:    tcell.ColorDimGray,   // Серый
		ChatSystem:       tcell.ColorOrange,   // Оранжевый

		// Цвета для панелей
		PanelHeader:      tcell.ColorGray,
		PanelTitle:       tcell.ColorGreen,
		PanelSubtitle:    tcell.ColorDimGray,

		// Цвета для строки состояния
		StatusBarBg:      tcell.ColorDarkGray,
		StatusBarFg:      tcell.ColorWhite,
		StatusBarHighlight: tcell.ColorGreen,
	}
}

// ModernLightTheme возвращает современную светлую тему.
func ModernLightTheme() *ColorTheme {
	return &ColorTheme{
		// Основные цвета
		Background:       tcell.ColorWhite,
		Foreground:       tcell.ColorBlack,
		Border:           tcell.ColorSilver,
		BorderActive:     tcell.ColorGreen,
		BorderInactive:   tcell.ColorSilver,

		// Цвета для списков
		ListText:         tcell.ColorBlack,
		ListSelectedBg:   tcell.ColorGreen,
		ListSelectedFg:   tcell.ColorWhite,
		ListSeparator:    tcell.ColorSilver,

		// Цвета для статусов
		StatusInfo:       tcell.ColorBlue,
		StatusWarning:    tcell.ColorYellow,
		StatusError:      tcell.ColorRed,
		StatusSuccess:    tcell.ColorGreen,

		// Цвета для чата
		ChatIncoming:     tcell.ColorGreen,
		ChatOutgoing:     tcell.ColorBlue,
		ChatTimestamp:    tcell.ColorDimGray,
		ChatSystem:       tcell.ColorOrange,

		// Цвета для панелей
		PanelHeader:      tcell.ColorSilver,
		PanelTitle:       tcell.ColorGreen,
		PanelSubtitle:    tcell.ColorDimGray,

		// Цвета для строки состояния
		StatusBarBg:      tcell.ColorSilver,
		StatusBarFg:      tcell.ColorBlack,
		StatusBarHighlight: tcell.ColorGreen,
	}
}

// GetStyle возвращает стиль с заданными цветами.
func GetStyle(fg, bg tcell.Color) tcell.Style {
	return tcell.StyleDefault.Foreground(fg).Background(bg)
}

// GetBorderStyle возвращает стиль для границы.
func GetBorderStyle(color tcell.Color) tcell.Style {
	return tcell.StyleDefault.Foreground(color)
}
