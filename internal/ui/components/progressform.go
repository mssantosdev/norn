package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mssantosdev/norn/internal/ui/styles"
	"github.com/mssantosdev/norn/internal/ui/themes"
)

// ProgressForm wraps a huh form with a progress indicator.
type ProgressForm struct {
	stepNames []string
	current   int
	total     int
}

// NewProgressForm creates a new progress form with the given step names.
func NewProgressForm(stepNames []string) *ProgressForm {
	return &ProgressForm{
		stepNames: stepNames,
		current:   0,
		total:     len(stepNames),
	}
}

// ProgressBar renders the progress bar as a string.
func (pf *ProgressForm) ProgressBar() string {
	if pf.total <= 1 {
		return ""
	}

	width := 20
	filled := int(float64(width) * float64(pf.current+1) / float64(pf.total))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	stepName := ""
	if pf.current < len(pf.stepNames) {
		stepName = pf.stepNames[pf.current]
	}

	return fmt.Sprintf("\n%s %s %d/%d: %s",
		styles.FormProgressFill.Render(bar[:filled]),
		styles.FormProgress.Render(bar[filled:]),
		pf.current+1,
		pf.total,
		stepName,
	)
}

// Next advances to the next step.
func (pf *ProgressForm) Next() {
	if pf.current < pf.total-1 {
		pf.current++
	}
}

// Prev goes back to the previous step.
func (pf *ProgressForm) Prev() {
	if pf.current > 0 {
		pf.current--
	}
}

// CurrentStep returns the current step name.
func (pf *ProgressForm) CurrentStep() string {
	if pf.current < len(pf.stepNames) {
		return pf.stepNames[pf.current]
	}
	return ""
}

// FormPage represents a single page in a multi-page form.
type FormPage struct {
	Title  string
	Fields []huh.Field
}

// MultiPageForm runs a multi-page form with progress indicator.
func MultiPageForm(pages []FormPage) error {
	for i, page := range pages {
		groups := make([]*huh.Group, 0, len(page.Fields))
		for _, field := range page.Fields {
			groups = append(groups, huh.NewGroup(field))
		}

		// Create progress text
		width := 20
		filled := int(float64(width) * float64(i+1) / float64(len(pages)))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
		_ = fmt.Sprintf("[%s] Step %d of %d: %s", bar, i+1, len(pages), page.Title)

		form := huh.NewForm(groups...)
		if err := form.Run(); err != nil {
			return err
		}
	}
	return nil
}

// ApplyNornTheme applies the current Norn theme colors to a huh form.
func ApplyNornTheme() *huh.Theme {
	t := huh.ThemeBase()
	t.Focused.Base = lipgloss.NewStyle().BorderForeground(themes.Current.Primary)
	t.Focused.Title = lipgloss.NewStyle().Foreground(themes.Current.Primary).Bold(true)
	t.Focused.Description = lipgloss.NewStyle().Foreground(themes.Current.Muted)
	t.Focused.TextInput.Cursor = lipgloss.NewStyle().Foreground(themes.Current.Primary)
	t.Focused.TextInput.Placeholder = lipgloss.NewStyle().Foreground(themes.Current.Muted)
	t.Blurred.Base = lipgloss.NewStyle().BorderForeground(themes.Current.Border)
	t.Blurred.Title = lipgloss.NewStyle().Foreground(themes.Current.Muted)
	return t
}
