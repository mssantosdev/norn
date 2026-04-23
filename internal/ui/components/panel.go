package components

import (
	"strings"

	"github.com/mssantosdev/norn/internal/ui/styles"
)

// Panel is a bordered section container with optional title, badge, and footer.
type Panel struct {
	title  string
	width  int
	badge  string
	footer string
	body   []string
}

// NewPanel creates a new panel with the given title.
func NewPanel(title string) Panel {
	return Panel{
		title: title,
		width: styles.TerminalWidth() - 4,
		body:  make([]string, 0),
	}
}

// WithWidth sets a custom width for the panel.
func (p Panel) WithWidth(width int) Panel {
	p.width = width
	return p
}

// WithBadge adds a status badge to the panel title.
func (p Panel) WithBadge(status string) Panel {
	p.badge = status
	return p
}

// WithFooter adds a hint/footer line at the bottom.
func (p Panel) WithFooter(footer string) Panel {
	p.footer = footer
	return p
}

// AddLine adds a content line to the panel body.
func (p Panel) AddLine(line string) Panel {
	p.body = append(p.body, line)
	return p
}

// AddLines adds multiple content lines.
func (p Panel) AddLines(lines []string) Panel {
	p.body = append(p.body, lines...)
	return p
}

// AddKV adds a key-value pair line.
func (p Panel) AddKV(label, value string) Panel {
	p.body = append(p.body, styles.KV(label, value))
	return p
}

// AddEmptyLine adds a blank line for spacing.
func (p Panel) AddEmptyLine() Panel {
	p.body = append(p.body, "")
	return p
}

// AddDivider adds a horizontal divider line.
func (p Panel) AddDivider() Panel {
	p.body = append(p.body, styles.RenderDivider(p.width-4))
	return p
}

// AddSectionTitle adds a section title line.
func (p Panel) AddSectionTitle(title string) Panel {
	p.body = append(p.body, styles.SectionTitle(title))
	return p
}

// Render builds the panel as a styled string.
func (p Panel) Render() string {
	// Build title with optional badge
	titleText := styles.Title.Render(p.title)
	if p.badge != "" {
		titleText = titleText + " " + styles.StatusBadgeFor(p.badge)
	}

	// Build body content
	bodyContent := strings.Join(p.body, "\n")

	// Build footer
	footerContent := ""
	if p.footer != "" {
		footerContent = "\n" + styles.Hint.Render(p.footer)
	}

	// Combine and wrap in section box
	content := titleText
	if bodyContent != "" {
		content += "\n" + bodyContent
	}
	if footerContent != "" {
		content += footerContent
	}

	return styles.SectionBox.Width(p.width).Render(content)
}

// RenderCompact returns a compact panel without outer border (just title + content).
func (p Panel) RenderCompact() string {
	var b strings.Builder

	titleText := styles.Title.Render(p.title)
	if p.badge != "" {
		titleText = titleText + " " + styles.StatusBadgeFor(p.badge)
	}
	b.WriteString(titleText)
	b.WriteString("\n")

	if len(p.body) > 0 {
		b.WriteString(strings.Join(p.body, "\n"))
	}

	if p.footer != "" {
		b.WriteString("\n")
		b.WriteString(styles.Hint.Render(p.footer))
	}

	return b.String()
}
