package components

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mssantosdev/norn/internal/ui/styles"
)

// NotificationType represents the type of notification.
type NotificationType int

const (
	SuccessNotification NotificationType = iota
	ErrorNotification
	WarningNotification
	InfoNotification
)

// Notification represents a toast notification.
type Notification struct {
	msg       string
	notifType NotificationType
	duration  time.Duration
	timer     time.Time
	visible   bool
}

// NewNotification creates a new notification.
func NewNotification(msg string, notifType NotificationType, duration time.Duration) Notification {
	return Notification{
		msg:       msg,
		notifType: notifType,
		duration:  duration,
		visible:   true,
	}
}

// ShowSuccess creates a success notification.
func ShowSuccess(msg string) Notification {
	return NewNotification(msg, SuccessNotification, 3*time.Second)
}

// ShowError creates an error notification.
func ShowError(msg string) Notification {
	return NewNotification(msg, ErrorNotification, 5*time.Second)
}

// ShowWarning creates a warning notification.
func ShowWarning(msg string) Notification {
	return NewNotification(msg, WarningNotification, 4*time.Second)
}

// ShowInfo creates an info notification.
func ShowInfo(msg string) Notification {
	return NewNotification(msg, InfoNotification, 3*time.Second)
}

// Init implements tea.Model.
func (n Notification) Init() tea.Cmd {
	n.timer = time.Now()
	return tea.Tick(n.duration, func(time.Time) tea.Msg {
		return dismissMsg{}
	})
}

// Update implements tea.Model.
func (n Notification) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case dismissMsg:
		n.visible = false
		return n, tea.Quit
	}
	return n, nil
}

// View implements tea.Model.
func (n Notification) View() string {
	if !n.visible {
		return ""
	}

	var prefix string
	var style string

	switch n.notifType {
	case SuccessNotification:
		prefix = "✓"
		style = styles.NotificationSuccess.Render(prefix + " " + n.msg)
	case ErrorNotification:
		prefix = "✗"
		style = styles.NotificationError.Render(prefix + " " + n.msg)
	case WarningNotification:
		prefix = "⚠"
		style = styles.NotificationWarning.Render(prefix + " " + n.msg)
	default:
		prefix = "ℹ"
		style = styles.StatusBadge.Render(prefix + " " + n.msg)
	}

	// Center the notification
	width := styles.TerminalWidth()
	lines := strings.Split(style, "\n")
	var result strings.Builder
	for _, line := range lines {
		padding := (width - len(line)) / 2
		if padding < 0 {
			padding = 0
		}
		result.WriteString(strings.Repeat(" ", padding))
		result.WriteString(line)
		result.WriteString("\n")
	}
	return result.String()
}

type dismissMsg struct{}

// RunNotification displays a notification and returns immediately (non-blocking feel).
func RunNotification(n Notification) {
	// For now, just print the notification inline
	// Full bubble tea notification would require running a separate program
	fmt.Println(n.View())
}
