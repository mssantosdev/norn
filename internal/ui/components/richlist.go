package components

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mssantosdev/norn/internal/ui/styles"
	"github.com/mssantosdev/norn/internal/ui/themes"
)

// RichListItem represents an item in the rich list.
type RichListItem struct {
	ID       string
	Title    string
	Subtitle string
	Status   string
	Detail   string
}

// FilterValue returns the filter value for the item.
func (i RichListItem) FilterValue() string {
	return i.ID + " " + i.Title + " " + i.Subtitle
}

// richListDelegate is the custom delegate for rendering list items.
type richListDelegate struct {
	showStatus bool
	showDetail bool
}

func (d richListDelegate) Height() int {
	height := 1
	if d.showDetail && false { // detail on same line
		height = 1
	}
	return height
}

func (d richListDelegate) Spacing() int { return 1 }

func (d richListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d richListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(RichListItem)
	if !ok {
		return
	}

	// Determine styles based on selection state
	var titleStyle lipgloss.Style
	var subtitleStyle lipgloss.Style

	if index == m.Index() {
		titleStyle = lipgloss.NewStyle().Foreground(themes.Current.Highlight).Bold(true)
		subtitleStyle = lipgloss.NewStyle().Foreground(themes.Current.Highlight)
	} else {
		titleStyle = lipgloss.NewStyle().Foreground(themes.Current.TextPrimary)
		subtitleStyle = lipgloss.NewStyle().Foreground(themes.Current.TextMuted)
	}

	// Build the line
	line := "▸ " + titleStyle.Render(i.Title)

	if i.Subtitle != "" {
		line += " " + subtitleStyle.Render(i.Subtitle)
	}

	if d.showStatus && i.Status != "" {
		line += " " + styles.StatusBadgeFor(i.Status)
	}

	if d.showDetail && i.Detail != "" {
		line += " " + lipgloss.NewStyle().Foreground(themes.Current.Muted).Render(i.Detail)
	}

	fmt.Fprint(w, line)
}

// RichList is a themed, filterable list component.
type RichList struct {
	model      list.Model
	items      []RichListItem
	showStatus bool
	showDetail bool
	quitting   bool
}

// NewRichList creates a new rich list with the given items.
func NewRichList(items []RichListItem) RichList {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	delegate := richListDelegate{}
	width := styles.TerminalWidth() - 4
	if width < 40 {
		width = 40
	}
	height := 15
	if len(items) < height {
		height = len(items)
	}
	if height < 3 {
		height = 3
	}

	m := list.New(listItems, delegate, width, height)
	m.SetShowStatusBar(true)
	m.SetFilteringEnabled(true)
	m.SetShowHelp(false)
	m.SetShowPagination(true)

	// Style the list with theme colors
	m.Styles.Title = lipgloss.NewStyle().Foreground(themes.Current.Primary).Bold(true)
	m.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(themes.Current.Primary)
	m.Styles.FilterCursor = lipgloss.NewStyle().Foreground(themes.Current.Highlight)
	m.Styles.StatusBar = lipgloss.NewStyle().Foreground(themes.Current.Muted)
	m.Styles.StatusBarFilterCount = lipgloss.NewStyle().Foreground(themes.Current.Primary)
	m.Styles.NoItems = lipgloss.NewStyle().Foreground(themes.Current.Muted).Italic(true)
	m.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(themes.Current.Muted)
	m.Styles.HelpStyle = lipgloss.NewStyle().Foreground(themes.Current.Muted)

	return RichList{
		model: m,
		items: items,
	}
}

// WithStatus enables status badge display.
func (rl RichList) WithStatus() RichList {
	rl.showStatus = true
	return rl
}

// WithDetail enables detail display.
func (rl RichList) WithDetail() RichList {
	rl.showDetail = true
	return rl
}

// WithTitle sets the list title.
func (rl RichList) WithTitle(title string) RichList {
	rl.model.Title = title
	return rl
}

// Init implements tea.Model.
func (rl RichList) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (rl RichList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			rl.quitting = true
			return rl, tea.Quit
		}
		if msg.String() == "enter" {
			rl.quitting = true
			return rl, tea.Quit
		}
	}

	var cmd tea.Cmd
	rl.model, cmd = rl.model.Update(msg)
	return rl, cmd
}

// View implements tea.Model.
func (rl RichList) View() string {
	if rl.quitting {
		return ""
	}
	return rl.model.View()
}

// Run runs the interactive rich list and returns the selected item.
func (rl RichList) Run() (*RichListItem, error) {
	p := tea.NewProgram(rl)
	model, err := p.Run()
	if err != nil {
		return nil, err
	}
	rlm := model.(RichList)
	if rlm.model.SelectedItem() == nil {
		return nil, fmt.Errorf("no selection")
	}
	item := rlm.model.SelectedItem().(RichListItem)
	return &item, nil
}

// RenderStatic renders the list as static output (non-interactive).
func (rl RichList) RenderStatic() string {
	var result string
	for _, item := range rl.items {
		line := "▸ " + styles.Title.Render(item.Title)
		if item.Subtitle != "" {
			line += " " + styles.Dimmed.Render(item.Subtitle)
		}
		if rl.showStatus && item.Status != "" {
			line += " " + styles.StatusBadgeFor(item.Status)
		}
		if rl.showDetail && item.Detail != "" {
			line += " " + styles.Dimmed.Render(item.Detail)
		}
		result += line + "\n"
	}
	return result
}
