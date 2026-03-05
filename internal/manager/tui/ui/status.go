package ui

import (
	"fmt"
	"net"
	"time"

	"github.com/first-debug/p2p/internal/manager/tui/styles"
	"github.com/rivo/tview"
)

// StatusInfo содержит информацию для отображения в строке состояния.
type StatusInfo struct {
	LocalIP        string
	LocalPort      int
	PeerID         string
	PeerCount      int
	SessionCount   int
	IsConnected    bool
	Uptime         time.Duration
}

// StatusBar предоставляет функциональность строки состояния.
type StatusBar struct {
	view   *tview.TextView
	styles *styles.ComponentStyles
	info   StatusInfo
}

// NewStatusBar создаёт новую строку состояния.
func NewStatusBar(view *tview.TextView, styles *styles.ComponentStyles) *StatusBar {
	return &StatusBar{
		view:   view,
		styles: styles,
	}
}

// Update обновляет информацию в строке состояния.
func (sb *StatusBar) Update(info StatusInfo) {
	sb.info = info
	sb.view.SetText(sb.formatStatus(info))
}

// formatStatus форматирует информацию для строки состояния.
func (sb *StatusBar) formatStatus(info StatusInfo) string {
	theme := sb.styles.GetTheme()
	
	// Форматируем uptime
	uptimeStr := formatDuration(info.Uptime)
	
	// Статус подключения
	connectionStatus := "○"
	connectionColor := theme.StatusError
	if info.IsConnected {
		connectionStatus = "●"
		connectionColor = theme.StatusSuccess
	}
	
	// Форматируем строку состояния
	return fmt.Sprintf(
		"[%s]%s[-:-:-] IP: %s | Порт: %d | ID: %s | Пиры: %d | Сессии: %d | Время работы: %s",
		connectionColor.String(),
		connectionStatus,
		info.LocalIP,
		info.LocalPort,
		info.PeerID,
		info.PeerCount,
		info.SessionCount,
		uptimeStr,
	)
}

// SetPeerCount устанавливает количество пиров.
func (sb *StatusBar) SetPeerCount(count int) {
	sb.info.PeerCount = count
	sb.Update(sb.info)
}

// SetSessionCount устанавливает количество сессий.
func (sb *StatusBar) SetSessionCount(count int) {
	sb.info.SessionCount = count
	sb.Update(sb.info)
}

// SetConnected устанавливает статус подключения.
func (sb *StatusBar) SetConnected(connected bool) {
	sb.info.IsConnected = connected
	sb.Update(sb.info)
}

// GetInfo возвращает текущую информацию.
func (sb *StatusBar) GetInfo() StatusInfo {
	return sb.info
}

// formatDuration форматирует длительность в человекочитаемый формат.
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dч %dм %dс", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dм %dс", minutes, seconds)
	}
	return fmt.Sprintf("%dс", seconds)
}

// GetLocalIP возвращает локальный IP адрес.
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	
	return "127.0.0.1"
}

// GetLocalIPs возвращает все локальные IP адреса.
func GetLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return []string{"127.0.0.1"}
	}
	
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}
	
	if len(ips) == 0 {
		return []string{"127.0.0.1"}
	}
	
	return ips
}
