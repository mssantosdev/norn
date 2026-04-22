package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mssantosdev/norn/internal/ui/themes"
)

type Spinner struct {
	spinner spinner.Model
	message string
	done    bool
}

func NewSpinner(message string) Spinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(themes.Current.Primary)
	return Spinner{spinner: s, message: message}
}

func (s Spinner) Init() tea.Cmd {
	return s.spinner.Tick
}

func (s Spinner) Update(msg tea.Msg) (Spinner, tea.Cmd) {
	if s.done {
		return s, nil
	}
	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return s, cmd
}

func (s Spinner) View() string {
	if s.done {
		return ""
	}
	return fmt.Sprintf("%s %s", s.spinner.View(), s.message)
}

func (s *Spinner) Finish() {
	s.done = true
}

// RunWithSpinner executes a function while displaying a spinner message.
// The spinner runs in a goroutine and stops when the function completes.
func RunWithSpinner(message string, fn func()) {
	s := NewSpinner(message)
	fmt.Println(s.View())

	done := make(chan bool)
	go func() {
		fn()
		close(done)
	}()

	<-done
	// Clear the spinner line
	fmt.Print("\r\033[K")
}
