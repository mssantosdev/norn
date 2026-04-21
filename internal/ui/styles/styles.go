package styles

import (
	"fmt"
	"os"

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
	AppHeader      lipgloss.Style
	Title          lipgloss.Style
	Subtitle       lipgloss.Style
	Label          lipgloss.Style
	Dimmed         lipgloss.Style
	Prompt         lipgloss.Style
	Success        lipgloss.Style
	Error          lipgloss.Style
	Warning        lipgloss.Style
	Box            lipgloss.Style
	HelpKey        lipgloss.Style
	HelpDesc       lipgloss.Style
	TableHeader    lipgloss.Style
	TableCell      lipgloss.Style
	StatusBadge    lipgloss.Style
	SuccessBadge   lipgloss.Style
	WarningBadge   lipgloss.Style
	ErrorBadge     lipgloss.Style
	SelectionBadge lipgloss.Style
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
