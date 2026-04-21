package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mssantosdev/norn/internal/ui/logger"
	"github.com/mssantosdev/norn/internal/ui/styles"
)

type HelpTopic struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Version     string        `json:"version"`
	Usage       string        `json:"usage"`
	Commands    []CommandHelp `json:"commands,omitempty"`
}

type CommandHelp struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Usage       string     `json:"usage"`
	Flags       []FlagHelp `json:"flags,omitempty"`
	Examples    []string   `json:"examples,omitempty"`
}

type FlagHelp struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
}

func renderHelpText(topic HelpTopic) string {
	var b strings.Builder
	width := styles.TerminalWidth()

	// Header
	header := styles.AppHeader.Render(topic.Name)
	if topic.Version != "" {
		header = styles.AppHeader.Render(fmt.Sprintf("%s %s", topic.Name, topic.Version))
	}
	b.WriteString(header)
	b.WriteString("\n")

	// Description
	if topic.Description != "" {
		b.WriteString(styles.Subtitle.Render(topic.Description))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Usage
	if topic.Usage != "" {
		b.WriteString(styles.Label.Render("Usage:"))
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(topic.Usage)
		b.WriteString("\n\n")
	}

	// Commands
	if len(topic.Commands) > 0 {
		b.WriteString(styles.Label.Render("Commands:"))
		b.WriteString("\n")
		for _, cmd := range topic.Commands {
			key := fmt.Sprintf("  %-12s", cmd.Name)
			desc := wrapText(cmd.Description, width-16)
			lines := strings.Split(desc, "\n")
			for i, line := range lines {
				if i == 0 {
					b.WriteString(styles.HelpKey.Render(key))
					b.WriteString(styles.HelpDesc.Render(line))
				} else {
					b.WriteString(strings.Repeat(" ", 16))
					b.WriteString(styles.HelpDesc.Render(line))
				}
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

	// Examples
	for _, cmd := range topic.Commands {
		if len(cmd.Examples) > 0 {
			b.WriteString(styles.Label.Render(fmt.Sprintf("Examples for %s:", cmd.Name)))
			b.WriteString("\n")
			for _, ex := range cmd.Examples {
				b.WriteString("  ")
				b.WriteString(styles.Dimmed.Render(ex))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}

	// Footer
	b.WriteString(styles.Dimmed.Render("Use --format=json for machine-readable output"))
	b.WriteString("\n")

	return b.String()
}

func renderHelpJSON(topic HelpTopic) string {
	data, err := json.MarshalIndent(topic, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err)
	}
	return string(data)
}

func wrapText(text string, width int) string {
	if width <= 0 {
		width = 80
	}
	if len(text) <= width {
		return text
	}
	var b strings.Builder
	words := strings.Fields(text)
	currentLen := 0
	for i, word := range words {
		if i > 0 && currentLen+1+len(word) > width {
			b.WriteString("\n")
			currentLen = 0
		} else if i > 0 {
			b.WriteString(" ")
			currentLen++
		}
		b.WriteString(word)
		currentLen += len(word)
	}
	return b.String()
}

func showHelp(topic HelpTopic, args []string) bool {
	hasHelp := false
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			hasHelp = true
			break
		}
	}
	if !hasHelp {
		return false
	}
	format := "text"
	for _, a := range args {
		if a == "--format=json" {
			format = "json"
		}
	}
	if format == "json" {
		logger.Print(renderHelpJSON(topic))
	} else {
		logger.Print(renderHelpText(topic))
	}
	return true
}

func rootHelp() HelpTopic {
	return HelpTopic{
		Name:        "norn",
		Description: "Weave-aware multi-agent harness",
		Version:     "0.0.1",
		Usage:       "norn <command> [flags]",
		Commands: []CommandHelp{
			{
				Name:        "init",
				Description: "Bootstrap a new Norn project or workspace",
				Usage:       "norn init [flags]",
				Flags: []FlagHelp{
					{Name: "--no-interactive", Description: "Run in non-interactive mode"},
					{Name: "--enable-opencode", Description: "Enable OpenCode integration"},
					{Name: "--mode=folder|branch", Description: "Planning mode (default: folder)"},
					{Name: "--name=<name>", Description: "Project name"},
				},
				Examples: []string{
					"norn init",
					"norn init --no-interactive --name=my-project --enable-opencode",
				},
			},
			{
				Name:        "status",
				Description: "Show current workspace status",
				Usage:       "norn status",
			},
			{
				Name:        "detect",
				Description: "Detect languages, tools, and frameworks",
				Usage:       "norn detect",
			},
			{
				Name:        "fates",
				Description: "Manage agent fates",
				Usage:       "norn fates <list|show <name>>",
			},
			{
				Name:        "patterns",
				Description: "Manage pattern documents",
				Usage:       "norn patterns <list|add|show|remove>",
			},
			{
				Name:        "skills",
				Description: "Manage skill documents",
				Usage:       "norn skills <list|add|show|remove>",
			},
			{
				Name:        "tools",
				Description: "Manage tool permission definitions",
				Usage:       "norn tools <list|add|show|edit|remove>",
			},
			{
				Name:        "weaves",
				Description: "Manage weave planning artifacts",
				Usage:       "norn weaves <list|add|show|remove>",
				Flags: []FlagHelp{
					{Name: "--surface=shared|local|both", Description: "Planning surface for writes"},
				},
			},
			{
				Name:        "threads",
				Description: "Manage thread planning artifacts",
				Usage:       "norn threads <list <weave-id>|add <weave-id> <title> <summary>|show <weave-id> <thread-id>|remove <weave-id> <thread-id>>",
				Flags: []FlagHelp{
					{Name: "--surface=shared|local|both", Description: "Planning surface for writes"},
				},
			},
			{
				Name:        "warps",
				Description: "Manage runtime warp lanes",
				Usage:       "norn warps <list|add|assign|assignment|show|remove>",
				Flags: []FlagHelp{
					{Name: "--view=runtime", Description: "Show runtime ownership view"},
				},
			},
			{
				Name:        "runes",
				Description: "Manage configuration",
				Usage:       "norn runes <show|resolve|edit>",
				Flags: []FlagHelp{
					{Name: "--scope=global|workspace|local", Description: "Config scope"},
					{Name: "--format=table|yaml", Description: "Output format"},
					{Name: "--set path=value", Description: "Set config value"},
					{Name: "--unset path", Description: "Unset config value"},
				},
			},
			{
				Name:        "chat",
				Description: "Validate and manage OpenCode integration",
				Usage:       "norn chat <validate>",
			},
		},
	}
}

func chatHelp() HelpTopic {
	return HelpTopic{
		Name:        "norn chat",
		Description: "Validate and manage OpenCode integration",
		Usage:       "norn chat <command>",
		Commands: []CommandHelp{
			{
				Name:        "validate",
				Description: "Check if opencode binary is available on PATH",
				Usage:       "norn chat validate",
			},
			{
				Name:        "status",
				Description: "Show OpenCode integration status",
				Usage:       "norn chat status",
			},
			{
				Name:        "export",
				Description: "Export OpenCode agents and config",
				Usage:       "norn chat export [--output=<dir>]",
				Flags: []FlagHelp{
					{Name: "--output=<dir>", Description: "Output directory for config export (default: current directory)"},
				},
			},
			{
				Name:        "assist",
				Description: "Get AI assistance for planning artifacts",
				Usage:       "norn chat assist [--prompt=<prompt>]",
				Flags: []FlagHelp{
					{Name: "--prompt=<text>", Description: "What to ask the AI to generate"},
				},
				Examples: []string{
					"norn chat assist --prompt=\"Generate starter patterns for a Go API\"",
				},
			},
			{
				Name:        "preview",
				Description: "Preview AI-generated artifacts without saving",
				Usage:       "norn chat preview --prompt=<prompt>",
				Flags: []FlagHelp{
					{Name: "--prompt=<text>", Description: "What to ask the AI to generate"},
				},
				Examples: []string{
					"norn chat preview --prompt=\"Generate deployment skills\"",
				},
			},
		},
	}
}
