package styles

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mssantosdev/norn/internal/ui/themes"
	"golang.org/x/term"
)

var (
	BgDark   lipgloss.Color
	BgDarker lipgloss.Color
	BgLight  lipgloss.Color
	Fg       lipgloss.Color
	FgBright lipgloss.Color
	FgMuted  lipgloss.Color
	Blue     lipgloss.Color
	Cyan     lipgloss.Color
	Green    lipgloss.Color
	Orange   lipgloss.Color
	Pink     lipgloss.Color
	Purple   lipgloss.Color
	Red      lipgloss.Color
	Yellow   lipgloss.Color
)

var (
	AppHeader           lipgloss.Style
	Title               lipgloss.Style
	Subtitle            lipgloss.Style
	Label               lipgloss.Style
	Dimmed              lipgloss.Style
	Prompt              lipgloss.Style
	Success             lipgloss.Style
	Error               lipgloss.Style
	Warning             lipgloss.Style
	Box                 lipgloss.Style
	HelpKey             lipgloss.Style
	HelpDesc            lipgloss.Style
	TableHeader         lipgloss.Style
	TableCell           lipgloss.Style
	StatusBadge         lipgloss.Style
	SuccessBadge        lipgloss.Style
	WarningBadge        lipgloss.Style
	ErrorBadge          lipgloss.Style
	SelectionBadge      lipgloss.Style
	SectionBox          lipgloss.Style
	PageHeader          lipgloss.Style
	Divider             lipgloss.Style
	SurfacePanel        lipgloss.Style
	TreeGuide           lipgloss.Style
	TreeBranch          lipgloss.Style
	TreeLeaf            lipgloss.Style
	NotificationSuccess lipgloss.Style
	NotificationError   lipgloss.Style
	NotificationWarning lipgloss.Style
	FormProgress        lipgloss.Style
	FormProgressFill    lipgloss.Style
	EmptyState          lipgloss.Style
	Hint                lipgloss.Style
)

func init() {
	ApplyTheme(themes.Current)
}

func ApplyTheme(theme themes.Theme) {
	BgDark = theme.Background
	BgDarker = theme.Background
	BgLight = theme.Border
	Fg = theme.Foreground
	FgBright = theme.Highlight
	FgMuted = theme.Muted
	Blue = theme.Primary
	Cyan = theme.Secondary
	Green = theme.Success
	Orange = theme.Warning
	Pink = theme.Highlight
	Purple = theme.Secondary
	Red = theme.Error
	Yellow = theme.Warning

	AppHeader = lipgloss.NewStyle().Background(Blue).Foreground(BgDark).Bold(true).Padding(0, 2)
	Title = lipgloss.NewStyle().Foreground(Blue).Bold(true)
	Subtitle = lipgloss.NewStyle().Foreground(FgMuted)
	Label = lipgloss.NewStyle().Foreground(FgBright).Bold(true)
	Dimmed = lipgloss.NewStyle().Foreground(FgMuted)
	Prompt = lipgloss.NewStyle().Foreground(Pink)
	Success = lipgloss.NewStyle().Foreground(Green).Bold(true)
	Error = lipgloss.NewStyle().Foreground(Red).Bold(true)
	Warning = lipgloss.NewStyle().Foreground(Orange).Bold(true)
	Box = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Blue).Padding(1).Background(BgDarker)
	HelpKey = lipgloss.NewStyle().Foreground(Pink).Bold(true)
	HelpDesc = lipgloss.NewStyle().Foreground(FgMuted)
	TableHeader = lipgloss.NewStyle().Foreground(Blue).Bold(true).Underline(true)
	TableCell = lipgloss.NewStyle().Foreground(Fg)
	StatusBadge = lipgloss.NewStyle().Background(Purple).Foreground(BgDark).Bold(true).Padding(0, 1)
	SuccessBadge = lipgloss.NewStyle().Background(Green).Foreground(BgDark).Bold(true).Padding(0, 1)
	WarningBadge = lipgloss.NewStyle().Background(Orange).Foreground(BgDark).Bold(true).Padding(0, 1)
	ErrorBadge = lipgloss.NewStyle().Background(Red).Foreground(BgDark).Bold(true).Padding(0, 1)
	SelectionBadge = lipgloss.NewStyle().Background(Blue).Foreground(BgDark).Bold(true).Padding(0, 1)

	// Facelift v0.0.4 — new styles
	SectionBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Background(theme.Surface).
		Padding(1, 2)

	PageHeader = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(theme.Primary).
		Background(theme.Surface).
		Padding(1, 2).
		Bold(true)

	Divider = lipgloss.NewStyle().
		Foreground(theme.Border)

	SurfacePanel = lipgloss.NewStyle().
		Background(theme.Surface).
		Padding(1, 2)

	TreeGuide = lipgloss.NewStyle().
		Foreground(theme.Border)

	TreeBranch = lipgloss.NewStyle().
		Foreground(theme.Primary)

	TreeLeaf = lipgloss.NewStyle().
		Foreground(theme.Foreground)

	NotificationSuccess = lipgloss.NewStyle().
		Background(theme.Success).
		Foreground(theme.Background).
		Bold(true).
		Padding(0, 2)

	NotificationError = lipgloss.NewStyle().
		Background(theme.Error).
		Foreground(theme.Background).
		Bold(true).
		Padding(0, 2)

	NotificationWarning = lipgloss.NewStyle().
		Background(theme.Warning).
		Foreground(theme.Background).
		Bold(true).
		Padding(0, 2)

	FormProgress = lipgloss.NewStyle().
		Foreground(theme.Muted)

	FormProgressFill = lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)

	EmptyState = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Italic(true)

	Hint = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Italic(true)
}

func TerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width == 0 {
		return 80
	}
	return width
}

func Badge(text string) string {
	return StatusBadge.Render(text)
}

func KV(label, value string) string {
	return fmt.Sprintf("%s %s", Label.Render(label+":"), value)
}

func RenderDivider(width int) string {
	if width <= 0 {
		width = TerminalWidth()
	}
	return Divider.Render(strings.Repeat("─", width))
}

func StatusBadgeFor(status string) string {
	switch strings.ToLower(status) {
	case "active", "enabled", "done", "success", "complete":
		return SuccessBadge.Render(status)
	case "paused", "blocked", "warning", "caution", "review":
		return WarningBadge.Render(status)
	case "error", "failed", "disabled":
		return ErrorBadge.Render(status)
	default:
		return StatusBadge.Render(status)
	}
}

func PageHeaderText(icon, title string) string {
	return PageHeader.Render(fmt.Sprintf("%s %s", icon, title))
}

func SectionTitle(text string) string {
	return Title.Render(text)
}

func EmptyStateText(hint string) string {
	return EmptyState.Render(fmt.Sprintf("🌱 %s", hint))
}
