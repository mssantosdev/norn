package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mssantosdev/norn/internal/detect"
	"github.com/mssantosdev/norn/internal/export"
	"github.com/mssantosdev/norn/internal/fates"
	"github.com/mssantosdev/norn/internal/loom"
	"github.com/mssantosdev/norn/internal/norn"
	"github.com/mssantosdev/norn/internal/opencode"
	"github.com/mssantosdev/norn/internal/patterns"
	"github.com/mssantosdev/norn/internal/skills"
	"github.com/mssantosdev/norn/internal/threads"
	toolstore "github.com/mssantosdev/norn/internal/tools"
	"github.com/mssantosdev/norn/internal/ui/logger"
	"github.com/mssantosdev/norn/internal/ui/styles"
	"github.com/mssantosdev/norn/internal/weaves"
)

var errUsage = errors.New("usage: norn <init|status|detect|fates|patterns|skills|tools|weaves|threads|warps|runes|chat|export|completion>")

func Run(args []string) error {
	if len(args) == 0 {
		logger.Print(renderHelpText(rootHelp()))
		return nil
	}

	// Handle top-level --help when no command is given or first arg is --help
	if args[0] == "--help" || args[0] == "-h" {
		showHelp(rootHelp(), args)
		return nil
	}

	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "status":
		return runStatus()
	case "detect":
		return runDetect()
	case "fates":
		return runFates(args[1:])
	case "patterns":
		return runDocCollection("patterns", args[1:])
	case "skills":
		return runDocCollection("skills", args[1:])
	case "tools":
		return runTools(args[1:])
	case "weaves":
		return runWeaves(args[1:])
	case "threads":
		return runThreads(args[1:])
	case "warps":
		return runWarps(args[1:])
	case "runes":
		return runRunes(args[1:])
	case "export":
		return runExport(args[1:])
	case "chat":
		if showHelp(chatHelp(), args[1:]) {
			return nil
		}
		return runChat(args[1:])
	case "completion":
		return runCompletion(args[1:])
	default:
		return errUsage
	}
}

func runInit(args []string) error {
	if showHelp(initHelp(), args) {
		return nil
	}
	opts, err := parseInitArgs(args)
	if err != nil {
		return err
	}
	if !opts.NonInteractive {
		if err := runInitForm(&opts); err != nil {
			return err
		}
	}
	detected, err := detect.Scan(".")
	if err != nil {
		return err
	}
	if len(opts.Languages) == 0 {
		opts.Languages = detected.Languages
	}
	if len(opts.Tools) == 0 {
		opts.Tools = detected.Tools
	}
	if len(opts.Frameworks) == 0 {
		opts.Frameworks = detected.Frameworks
	}
	if opts.Name == "" {
		abs, _ := filepath.Abs(".")
		opts.Name = filepath.Base(abs)
	}
	if opts.Theme == "" {
		opts.Theme = "tokyonight"
	}
	if opts.Skeleton == "" {
		opts.Skeleton = "standard"
	}
	if opts.OpenCodeModel == "" {
		opts.OpenCodeModel = "github-copilot/gpt-5.4-mini"
	}
	if opts.OpenCodeAgent == "" {
		opts.OpenCodeAgent = "build"
	}
	if opts.PlanningPath == "" {
		opts.PlanningPath = ".norn"
	}

	workspace := norn.Workspace{
		Root: ".",
		Runes: norn.RuneFile{
			Name:        opts.Name,
			Mode:        detectWorkspaceMode("."),
			Preferences: norn.PreferencesConfig{Language: "en", Verbosity: "normal"},
			UI:          norn.UIConfig{Theme: opts.Theme},
			Planning:    norn.PlanningConfig{Path: opts.PlanningPath},
			OpenCode:    norn.OpenCodeConfig{Enabled: opts.EnableOpenCode, Provider: "github-copilot", Model: opts.OpenCodeModel, Agent: opts.OpenCodeAgent, ResponseLanguage: "en", DraftingMode: "ask"},
			Tooling:     norn.ToolingConfig{Languages: opts.Languages, Tools: opts.Tools, Frameworks: opts.Frameworks},
			Hydra:       norn.HydraConfig{Enabled: detectWorkspaceMode(".") == norn.WorkspaceModeWorkspace},
		},
	}
	planningPath, err := loom.Ensure(workspace.Root, opts)
	if err != nil {
		return err
	}
	workspace.Runes.Planning.Path = planningPath
	if err := ensureWorkspacePaths(workspace); err != nil {
		return err
	}
	if err := norn.Save(workspace); err != nil {
		return err
	}
	if err := fates.Bootstrap(norn.FatesRoot(workspace), workspace.Runes.OpenCode.Model); err != nil {
		return err
	}
	if err := bootstrapCommands(workspace, detected); err != nil {
		return err
	}
	if err := fates.ExportOpenCode(norn.FatesRoot(workspace), norn.ToolsRoot(workspace), norn.OpenCodeAgentsRoot(workspace)); err != nil {
		return err
	}
	if workspace.Runes.OpenCode.Enabled && strings.TrimSpace(opts.OpenCodePrompt) != "" {
		assist, err := opencode.AssistInit(workspace.Runes.OpenCode, opts.OpenCodePrompt)
		if err != nil {
			logger.Warn("opencode assistance failed", "error", err)
		} else {
			for _, doc := range assist.Patterns {
				_ = patterns.Save(filepath.Join(norn.SharedPlanningRoot(workspace), "patterns"), doc)
			}
			for _, doc := range assist.Skills {
				_ = skills.Save(filepath.Join(norn.SharedPlanningRoot(workspace), "skills"), doc)
			}
			for _, doc := range assist.Weaves {
				_ = patterns.Save(filepath.Join(norn.SharedPlanningRoot(workspace), "weaves"), doc)
			}
		}
	}
	logger.Info("workspace initialized", "name", workspace.Runes.Name, "planning", workspace.Runes.Planning.Path)
	return nil
}

func parseInitArgs(args []string) (norn.InitOptions, error) {
	opts := norn.InitOptions{}
	for _, arg := range args {
		switch {
		case arg == "--no-interactive":
			opts.NonInteractive = true
		case arg == "--enable-opencode":
			opts.EnableOpenCode = true
		case strings.HasPrefix(arg, "--name="):
			opts.Name = strings.TrimPrefix(arg, "--name=")
		case strings.HasPrefix(arg, "--path="):
			opts.PlanningPath = strings.TrimPrefix(arg, "--path=")
		case strings.HasPrefix(arg, "--theme="):
			opts.Theme = strings.TrimPrefix(arg, "--theme=")
		case strings.HasPrefix(arg, "--languages="):
			opts.Languages = splitCSV(strings.TrimPrefix(arg, "--languages="))
		case strings.HasPrefix(arg, "--tools="):
			opts.Tools = splitCSV(strings.TrimPrefix(arg, "--tools="))
		case strings.HasPrefix(arg, "--frameworks="):
			opts.Frameworks = splitCSV(strings.TrimPrefix(arg, "--frameworks="))
		case strings.HasPrefix(arg, "--model="):
			opts.OpenCodeModel = strings.TrimPrefix(arg, "--model=")
		case strings.HasPrefix(arg, "--agent="):
			opts.OpenCodeAgent = strings.TrimPrefix(arg, "--agent=")
		case strings.HasPrefix(arg, "--skeleton="):
			opts.Skeleton = strings.TrimPrefix(arg, "--skeleton=")
		case strings.HasPrefix(arg, "--prompt="):
			opts.OpenCodePrompt = strings.TrimPrefix(arg, "--prompt=")
		default:
			return opts, fmt.Errorf("unknown init argument: %s", arg)
		}
	}
	return opts, nil
}

func runInitForm(opts *norn.InitOptions) error {
	projectName := opts.Name
	skeleton := opts.Skeleton
	openCodeEnabled := opts.EnableOpenCode
	openCodePrompt := opts.OpenCodePrompt
	theme := opts.Theme
	if theme == "" {
		theme = "tokyonight"
	}
	if skeleton == "" {
		skeleton = "standard"
	}

	getAIHelp := false

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("What should this project be called? This name appears in status and exports.").
				Placeholder("my-awesome-project").
				Value(&projectName),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable OpenCode integration?").
				Description("Export Norn's fates and skills as OpenCode agents. This lets you chat with the skald for planning help, and ensures your team's AI tools follow project conventions.").
				Value(&openCodeEnabled),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Get AI help with initial setup?").
				Description("The skald agent will suggest initial patterns and documentation based on your project description.").
				Value(&getAIHelp),
		).WithHideFunc(func() bool { return !openCodeEnabled }),
		huh.NewGroup(
			huh.NewText().
				Title("Describe your project for the skald").
				Description("Help the skald understand your project. Be specific about the tech stack and goals.\n\nExamples:\n• A Go CLI tool for managing Docker containers with Cobra\n• A React + TypeScript frontend with real-time WebSocket updates\n• A Rust microservice that processes payments with Stripe integration").
				Placeholder("A [language] [type] that [does what]").
				Value(&openCodePrompt),
		).WithHideFunc(func() bool { return !openCodeEnabled || !getAIHelp }),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Project scaffold").
				Description("What to create in the project directory.").
				Options(
					huh.NewOption("Standard — README, constitution, and default folders", "standard"),
					huh.NewOption("Empty — minimal structure, add artifacts manually", "empty"),
				).Value(&skeleton),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Theme").
				Description("Color theme for all Norn CLI output.").
				Options(themeOptions()...).
				Value(&theme),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}

	var previewFiles string
	if skeleton == "standard" {
		previewFiles = "\n\nFiles to create:\n[NEW] .norn/README.md\n[NEW] .norn/constitution.md\n[NEW] .norn/weaves/\n[NEW] .norn/patterns/\n[NEW] .norn/skills/\n[NEW] .norn/fates/keeper.yaml\n[NEW] .norn/fates/weaver.yaml\n[NEW] .norn/fates/judge.yaml\n[NEW] .norn/fates/fates.yaml\n[NEW] .norn/fates/skald.yaml"
	} else {
		previewFiles = "\n\nFiles to create:\n[NEW] .norn/"
	}
	preview := fmt.Sprintf("Project: %s\nScaffold: %s\nOpenCode: %t%s\nTheme: %s", projectName, skeleton, openCodeEnabled, previewFiles, theme)
	confirmed := true
	confirm := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title("Preview").Description(preview),
			huh.NewConfirm().Title("Create workspace with these files?").Value(&confirmed),
		),
	)
	if err := confirm.Run(); err != nil {
		return err
	}
	if !confirmed {
		return fmt.Errorf("init cancelled")
	}

	opts.Name = projectName
	opts.Skeleton = skeleton
	opts.EnableOpenCode = openCodeEnabled
	if openCodeEnabled && getAIHelp {
		opts.OpenCodePrompt = openCodePrompt
	}
	opts.Theme = theme
	return nil
}

func runStatus() error {
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	logger.Print(styles.Title.Render("Norn Status"))
	logger.Print(styles.KV("root", w.Root))
	logger.Print(styles.KV("workspace mode", string(w.Runes.Mode)))
	logger.Print(styles.KV("planning path", w.Runes.Planning.Path))
	logger.Print(styles.KV("hydra", fmt.Sprintf("%t", w.Runes.Hydra.Enabled)))
	logger.Print(styles.KV("opencode", fmt.Sprintf("%t", w.Runes.OpenCode.Enabled)))
	logger.Print(styles.KV("languages", strings.Join(w.Runes.Tooling.Languages, ", ")))
	logger.Print(styles.KV("tools", strings.Join(w.Runes.Tooling.Tools, ", ")))
	logger.Print(styles.KV("frameworks", strings.Join(w.Runes.Tooling.Frameworks, ", ")))
	logger.Print(styles.KV("fates", fmt.Sprintf("%d", norn.CountFiles(norn.FatesRoot(w), ".yaml"))))
	logger.Print(styles.KV("tools", fmt.Sprintf("%d", norn.CountFiles(norn.ToolsRoot(w), ".yaml"))))
	logger.Print(styles.KV("weaves", fmt.Sprintf("%d", countWeaves(w))))
	logger.Print(styles.KV("threads", fmt.Sprintf("%d", countThreads(w))))
	logger.Print(styles.KV("patterns", fmt.Sprintf("%d", norn.CountFiles(filepath.Join(norn.SharedPlanningRoot(w), "patterns"), ".md"))))
	logger.Print(styles.KV("skills", fmt.Sprintf("%d", norn.CountFiles(filepath.Join(norn.SharedPlanningRoot(w), "skills"), ".md"))))
	return nil
}

func runDetect() error {
	detected, err := detect.Scan(".")
	if err != nil {
		return err
	}
	logger.Print(styles.Title.Render("Detection"))
	logger.Print(styles.KV("languages", strings.Join(detected.Languages, ", ")))
	logger.Print(styles.KV("tools", strings.Join(detected.Tools, ", ")))
	logger.Print(styles.KV("frameworks", strings.Join(detected.Frameworks, ", ")))
	if len(detected.Locations) > 0 {
		logger.Print(styles.KV("locations", strings.Join(detected.Locations, ", ")))
	}
	return nil
}

func runFates(args []string) error {
	if showHelp(fatesHelp(), args) {
		return nil
	}
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	if len(args) == 0 || args[0] == "list" {
		items, err := fates.List(norn.FatesRoot(w))
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render("Fates"))
		for _, item := range items {
			logger.Print(styles.KV(item.Name, item.Description))
		}
		return nil
	}
	if args[0] == "show" {
		fateName := ""
		if len(args) >= 2 {
			fateName = args[1]
		} else {
			items, err := fates.List(norn.FatesRoot(w))
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no fates available")
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.Name, Title: item.Description})
			}
			selected, err := promptArtifactSelection("Select a fate", artifactItems)
			if err != nil {
				return err
			}
			fateName = selected
		}
		item, err := fates.Load(norn.FatesRoot(w), fateName)
		if err != nil {
			return err
		}
		rendered, err := fates.RenderOpenCode(item, norn.ToolsRoot(w))
		if err != nil {
			return err
		}
		logger.Print(rendered)
		return nil
	}
	if len(args) == 1 && args[0] == "add" {
		return runFatesAddInteractive(w)
	}
	if len(args) >= 2 && args[0] == "add" {
		if len(args) < 3 {
			return fmt.Errorf("usage: norn fates add <name> <description>")
		}
		return runFatesAddNonInteractive(w, args[1], strings.Join(args[2:], " "))
	}
	if args[0] == "edit" {
		fateName := ""
		if len(args) >= 2 && !strings.HasPrefix(args[1], "--set=") {
			fateName = args[1]
		} else {
			items, err := fates.List(norn.FatesRoot(w))
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no fates available")
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.Name, Title: item.Description})
			}
			selected, err := promptArtifactSelection("Select a fate to edit", artifactItems)
			if err != nil {
				return err
			}
			fateName = selected
		}
		sets := []string{}
		startIdx := 2
		if len(args) >= 2 && strings.HasPrefix(args[1], "--set=") {
			startIdx = 1
		}
		for _, arg := range args[startIdx:] {
			if strings.HasPrefix(arg, "--set=") {
				sets = append(sets, strings.TrimPrefix(arg, "--set="))
			}
		}
		if len(sets) == 0 {
			return runFatesEditInteractive(w, fateName)
		}
		return runFatesEditNonInteractive(w, fateName, sets)
	}
	if args[0] == "remove" {
		fateName := ""
		if len(args) >= 2 {
			fateName = args[1]
		} else {
			items, err := fates.List(norn.FatesRoot(w))
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no fates available")
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.Name, Title: item.Description})
			}
			selected, err := promptArtifactSelection("Select a fate to remove", artifactItems)
			if err != nil {
				return err
			}
			fateName = selected
		}
		if err := fates.Delete(norn.FatesRoot(w), fateName); err != nil {
			return err
		}
		return fates.ExportOpenCode(norn.FatesRoot(w), norn.ToolsRoot(w), norn.OpenCodeAgentsRoot(w))
	}
	return fmt.Errorf("usage: norn fates <list|show|add|edit|remove>")
}

func runFatesAddInteractive(w norn.Workspace) error {
	name := ""
	description := ""
	model := "github-copilot/gpt-5.4-mini"
	temperature := "0.2"
	body := ""
	allowEdit := false
	form := huh.NewForm(
		huh.NewGroup(huh.NewInput().Title("Name").Description("Unique identifier (lowercase, no spaces) — e.g., 'keeper', 'custom-agent'").Placeholder("my-agent").Value(&name)),
		huh.NewGroup(huh.NewInput().Title("Description").Description("One-line summary of this fate's responsibilities").Placeholder("Reviews API contracts and validates schemas").Value(&description)),
		huh.NewGroup(huh.NewInput().Title("Model").Description("AI model to use — e.g., github-copilot/gpt-5.4-mini").Placeholder("github-copilot/gpt-5.4-mini").Value(&model)),
		huh.NewGroup(huh.NewInput().Title("Temperature").Description("Creativity: 0.0 strict, 0.2 balanced, 0.5 creative").Placeholder("0.2").Value(&temperature)),
		huh.NewGroup(huh.NewText().Title("Body").Description("System prompt defining this fate's behavior, constraints, and responsibilities. This is exported to OpenCode as the agent definition.").Placeholder("You are the reviewer fate. Check all API changes for backward compatibility.").Value(&body)),
		huh.NewGroup(huh.NewConfirm().Title("Allow edit?").Description("Whether this fate can modify files directly. Keep false for review-only fates.").Value(&allowEdit)),
	)
	if err := form.Run(); err != nil {
		return err
	}
	item := norn.FateSource{
		Name:        name,
		Description: description,
		Model:       model,
		Temperature: temperature,
		Body:        body,
		AllowEdit:   allowEdit,
	}
	preview := fmt.Sprintf("Name: %s\nDescription: %s\nModel: %s\nTemperature: %s\nAllowEdit: %t\n\n%s",
		item.Name, item.Description, item.Model, item.Temperature, item.AllowEdit, item.Body)
	confirmed := true
	confirm := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title("Preview").Description(preview),
			huh.NewConfirm().Title("Create fate?").Value(&confirmed),
		),
	)
	if err := confirm.Run(); err != nil {
		return err
	}
	if !confirmed {
		return fmt.Errorf("add cancelled")
	}
	if err := fates.Save(norn.FatesRoot(w), item); err != nil {
		return err
	}
	return fates.ExportOpenCode(norn.FatesRoot(w), norn.ToolsRoot(w), norn.OpenCodeAgentsRoot(w))
}

func runFatesAddNonInteractive(w norn.Workspace, name, description string) error {
	item := norn.FateSource{
		Name:        name,
		Description: description,
		Model:       "github-copilot/gpt-5.4-mini",
		Temperature: "0.2",
		Body:        fmt.Sprintf("You are the %s fate. ", name),
	}
	if err := fates.Save(norn.FatesRoot(w), item); err != nil {
		return err
	}
	logger.Info("fate created", "name", name)
	return fates.ExportOpenCode(norn.FatesRoot(w), norn.ToolsRoot(w), norn.OpenCodeAgentsRoot(w))
}

func runFatesEditInteractive(w norn.Workspace, name string) error {
	item, err := fates.Load(norn.FatesRoot(w), name)
	if err != nil {
		return err
	}
	description := item.Description
	model := item.Model
	temperature := item.Temperature
	body := item.Body
	allowEdit := item.AllowEdit
	form := huh.NewForm(
		huh.NewGroup(huh.NewInput().Title("Description").Description("One-line summary of this fate's responsibilities").Placeholder("Reviews API contracts and validates schemas").Value(&description)),
		huh.NewGroup(huh.NewInput().Title("Model").Description("AI model to use — e.g., github-copilot/gpt-5.4-mini").Placeholder("github-copilot/gpt-5.4-mini").Value(&model)),
		huh.NewGroup(huh.NewInput().Title("Temperature").Description("Creativity: 0.0 strict, 0.2 balanced, 0.5 creative").Placeholder("0.2").Value(&temperature)),
		huh.NewGroup(huh.NewText().Title("Body").Description("System prompt defining this fate's behavior, constraints, and responsibilities. This is exported to OpenCode as the agent definition.").Placeholder("You are the reviewer fate. Check all API changes for backward compatibility.").Value(&body)),
		huh.NewGroup(huh.NewConfirm().Title("Allow edit?").Description("Whether this fate can modify files directly. Keep false for review-only fates.").Value(&allowEdit)),
	)
	if err := form.Run(); err != nil {
		return err
	}
	updated := norn.FateSource{
		Name:        name,
		Description: description,
		Model:       model,
		Temperature: temperature,
		Body:        body,
		AllowEdit:   allowEdit,
	}
	if err := fates.Save(norn.FatesRoot(w), updated); err != nil {
		return err
	}
	return fates.ExportOpenCode(norn.FatesRoot(w), norn.ToolsRoot(w), norn.OpenCodeAgentsRoot(w))
}

func runFatesEditNonInteractive(w norn.Workspace, name string, sets []string) error {
	item, err := fates.Load(norn.FatesRoot(w), name)
	if err != nil {
		return err
	}
	for _, set := range sets {
		parts := strings.SplitN(set, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --set %q; expected field=value", set)
		}
		field := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch field {
		case "description":
			item.Description = value
		case "model":
			item.Model = value
		case "temperature":
			item.Temperature = value
		case "body":
			item.Body = value
		case "allow_edit":
			item.AllowEdit = value == "true"
		default:
			return fmt.Errorf("unknown field %q; expected description, model, temperature, body, or allow_edit", field)
		}
	}
	if err := fates.Save(norn.FatesRoot(w), item); err != nil {
		return err
	}
	logger.Info("fate updated", "name", name)
	return fates.ExportOpenCode(norn.FatesRoot(w), norn.ToolsRoot(w), norn.OpenCodeAgentsRoot(w))
}

func runDocCollection(kind string, args []string) error {
	if showHelp(docCollectionHelp(kind), args) {
		return nil
	}
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	root := filepath.Join(norn.SharedPlanningRoot(w), kind)
	if len(args) == 0 || args[0] == "list" {
		items, err := listDocs(kind, root)
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render(strings.Title(kind)))
		for _, item := range items {
			logger.Print(styles.KV(item.ID, item.Title))
		}
		return nil
	}
	if len(args) >= 3 && args[0] == "add" {
		doc := norn.Document{ID: slug(args[1]), Title: args[1], Summary: strings.Join(args[2:], " "), Body: strings.Join(args[2:], " ")}
		return saveDoc(kind, root, doc)
	}
	if args[0] == "show" {
		docID := ""
		if len(args) >= 2 {
			docID = args[1]
		} else {
			items, err := listDocs(kind, root)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no %s available", kind)
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Summary})
			}
			selected, err := promptArtifactSelection(fmt.Sprintf("Select a %s", kind), artifactItems)
			if err != nil {
				return err
			}
			docID = selected
		}
		doc, err := loadDoc(kind, root, docID)
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render(doc.Title))
		logger.Print(doc.Body)
		return nil
	}
	if args[0] == "edit" {
		docID := ""
		if len(args) >= 2 && !strings.HasPrefix(args[1], "--set=") {
			docID = args[1]
		} else {
			items, err := listDocs(kind, root)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no %s available", kind)
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Summary})
			}
			selected, err := promptArtifactSelection(fmt.Sprintf("Select a %s to edit", kind), artifactItems)
			if err != nil {
				return err
			}
			docID = selected
		}
		sets := []string{}
		startIdx := 2
		if len(args) >= 2 && strings.HasPrefix(args[1], "--set=") {
			startIdx = 1
		}
		for _, arg := range args[startIdx:] {
			if strings.HasPrefix(arg, "--set=") {
				sets = append(sets, strings.TrimPrefix(arg, "--set="))
			}
		}
		if len(sets) == 0 {
			return runDocEditInteractive(kind, root, docID)
		}
		return runDocEditNonInteractive(kind, root, docID, sets)
	}
	if args[0] == "remove" {
		docID := ""
		if len(args) >= 2 {
			docID = args[1]
		} else {
			items, err := listDocs(kind, root)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no %s available", kind)
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Summary})
			}
			selected, err := promptArtifactSelection(fmt.Sprintf("Select a %s to remove", kind), artifactItems)
			if err != nil {
				return err
			}
			docID = selected
		}
		return deleteDoc(kind, root, docID)
	}
	return fmt.Errorf("usage: norn %s <list|add|show|edit|remove>", kind)
}

func runDocEditInteractive(kind, root, id string) error {
	doc, err := loadDoc(kind, root, id)
	if err != nil {
		return err
	}
	title := doc.Title
	summary := doc.Summary
	body := doc.Body
	form := huh.NewForm(
		huh.NewGroup(huh.NewInput().Title("Title").Description("Human-readable name for this artifact").Placeholder("API Authentication Pattern").Value(&title)),
		huh.NewGroup(huh.NewInput().Title("Summary").Description("One-line description of what this artifact covers").Placeholder("How to authenticate API requests using JWT tokens").Value(&summary)),
		huh.NewGroup(huh.NewText().Title("Body").Description("Full content. Use Markdown formatting. This is what fates and humans read for context.").Placeholder("## Overview\n\nDescribe the pattern, convention, or skill here.\n\n## Examples\n\nProvide concrete examples.").Value(&body)),
	)
	if err := form.Run(); err != nil {
		return err
	}
	updated := norn.Document{ID: id, Title: title, Summary: summary, Body: body}
	return saveDoc(kind, root, updated)
}

func runDocEditNonInteractive(kind, root, id string, sets []string) error {
	doc, err := loadDoc(kind, root, id)
	if err != nil {
		return err
	}
	for _, set := range sets {
		parts := strings.SplitN(set, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --set %q; expected field=value", set)
		}
		field := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch field {
		case "title":
			doc.Title = value
		case "summary":
			doc.Summary = value
		case "body":
			doc.Body = value
		default:
			return fmt.Errorf("unknown field %q; expected title, summary, or body", field)
		}
	}
	if err := saveDoc(kind, root, doc); err != nil {
		return err
	}
	logger.Info(kind+" updated", "id", id)
	return nil
}

func runTools(args []string) error {
	if showHelp(toolsHelp(), args) {
		return nil
	}
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	root := norn.ToolsRoot(w)
	if len(args) == 0 || args[0] == "list" {
		items, err := toolstore.List(root)
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render("Tools"))
		for _, item := range items {
			logger.Print(styles.KV(item.ID, fmt.Sprintf("%s [%s]", item.Command, strings.Join(item.Roles, ", "))))
		}
		return nil
	}
	if len(args) >= 4 && args[0] == "add" {
		item := norn.ManagedTool{ID: slug(args[1]), Title: args[1], Category: args[2], Command: strings.Join(args[3:], " "), Roles: []string{"weaver"}}
		if err := toolstore.Save(root, item); err != nil {
			return err
		}
		return fates.ExportOpenCode(norn.FatesRoot(w), root, norn.OpenCodeAgentsRoot(w))
	}
	if args[0] == "show" {
		toolID := ""
		if len(args) >= 2 {
			toolID = args[1]
		} else {
			items, err := toolstore.List(root)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no tools available")
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Description})
			}
			selected, err := promptArtifactSelection("Select a tool", artifactItems)
			if err != nil {
				return err
			}
			toolID = selected
		}
		item, err := toolstore.Load(root, toolID)
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render(item.Title))
		logger.Print(styles.KV("ID", item.ID))
		logger.Print(styles.KV("Category", item.Category))
		logger.Print(styles.KV("Command", item.Command))
		logger.Print(styles.KV("Pattern", item.Pattern))
		logger.Print(styles.KV("Risk", item.Risk))
		logger.Print(styles.KV("Roles", strings.Join(item.Roles, ", ")))
		if item.Description != "" {
			logger.Print(styles.KV("Description", item.Description))
		}
		return nil
	}
	if args[0] == "edit" {
		toolID := ""
		if len(args) >= 2 && !strings.HasPrefix(args[1], "--set=") {
			toolID = args[1]
		} else {
			items, err := toolstore.List(root)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no tools available")
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Description})
			}
			selected, err := promptArtifactSelection("Select a tool to edit", artifactItems)
			if err != nil {
				return err
			}
			toolID = selected
		}
		sets := []string{}
		startIdx := 2
		if len(args) >= 2 && strings.HasPrefix(args[1], "--set=") {
			startIdx = 1
		}
		for _, arg := range args[startIdx:] {
			if strings.HasPrefix(arg, "--set=") {
				sets = append(sets, strings.TrimPrefix(arg, "--set="))
			}
		}
		if len(sets) == 0 {
			return runToolsEditInteractive(w, root, toolID)
		}
		return runToolsEditNonInteractive(w, root, toolID, sets)
	}
	if args[0] == "remove" {
		toolID := ""
		if len(args) >= 2 {
			toolID = args[1]
		} else {
			items, err := toolstore.List(root)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no tools available")
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Description})
			}
			selected, err := promptArtifactSelection("Select a tool to remove", artifactItems)
			if err != nil {
				return err
			}
			toolID = selected
		}
		if err := toolstore.Delete(root, toolID); err != nil {
			return err
		}
		return fates.ExportOpenCode(norn.FatesRoot(w), root, norn.OpenCodeAgentsRoot(w))
	}
	return fmt.Errorf("usage: norn tools <list|add|show|edit|remove>")
}

func runToolsEditInteractive(w norn.Workspace, root, id string) error {
	item, err := toolstore.Load(root, id)
	if err != nil {
		return err
	}
	title := item.Title
	description := item.Description
	category := item.Category
	command := item.Command
	pattern := item.Pattern
	risk := item.Risk
	rolesStr := strings.Join(item.Roles, ", ")
	form := huh.NewForm(
		huh.NewGroup(huh.NewInput().Title("Title").Description("Human-readable name — e.g., 'Run Tests'").Placeholder("Run Tests").Value(&title)),
		huh.NewGroup(huh.NewInput().Title("Description").Description("What this tool does and when to use it").Placeholder("Runs the full test suite with coverage reporting").Value(&description)),
		huh.NewGroup(huh.NewInput().Title("Category").Description("Tool type: build, test, lint, deploy, setup").Placeholder("test").Value(&category)),
		huh.NewGroup(huh.NewInput().Title("Command").Description("Exact shell command to run. Use !command syntax for OpenCode: !go test ./...").Placeholder("go test ./...").Value(&command)),
		huh.NewGroup(huh.NewInput().Title("Pattern").Description("Glob pattern to match files this tool operates on. Leave empty to derive from command.").Placeholder("*.go").Value(&pattern)),
		huh.NewGroup(
			huh.NewSelect[string]().Title("Risk").Description("Safety level: low (read-only), medium (can modify), high (destructive)").Options(
				huh.NewOption("Low — read-only", "low"),
				huh.NewOption("Medium — can modify files", "medium"),
				huh.NewOption("High — destructive operations", "high"),
			).Value(&risk),
		),
		huh.NewGroup(huh.NewInput().Title("Roles").Description("Which fates can invoke this tool — comma-separated").Placeholder("weaver, fates").Value(&rolesStr)),
	)
	if err := form.Run(); err != nil {
		return err
	}
	roles := splitCSV(rolesStr)
	if pattern == "" {
		pattern = command + "*"
	}
	updated := norn.ManagedTool{
		ID:          id,
		Title:       title,
		Description: description,
		Category:    category,
		Command:     command,
		Pattern:     pattern,
		Risk:        risk,
		Roles:       roles,
	}
	if err := toolstore.Save(root, updated); err != nil {
		return err
	}
	return fates.ExportOpenCode(norn.FatesRoot(w), root, norn.OpenCodeAgentsRoot(w))
}

func runToolsEditNonInteractive(w norn.Workspace, root, id string, sets []string) error {
	item, err := toolstore.Load(root, id)
	if err != nil {
		return err
	}
	for _, set := range sets {
		parts := strings.SplitN(set, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --set %q; expected field=value", set)
		}
		field := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch field {
		case "title":
			item.Title = value
		case "description":
			item.Description = value
		case "category":
			item.Category = value
		case "command":
			item.Command = value
		case "pattern":
			item.Pattern = value
		case "risk":
			item.Risk = value
		case "roles":
			item.Roles = splitCSV(value)
		default:
			return fmt.Errorf("unknown field %q; expected title, description, category, command, pattern, risk, or roles", field)
		}
	}
	if item.Pattern == "" {
		item.Pattern = item.Command + "*"
	}
	if err := toolstore.Save(root, item); err != nil {
		return err
	}
	logger.Info("command updated", "id", id)
	return fates.ExportOpenCode(norn.FatesRoot(w), root, norn.OpenCodeAgentsRoot(w))
}

func runWeaves(args []string) error {
	if showHelp(weavesHelp(), args) {
		return nil
	}
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	root := norn.SharedPlanningRoot(w)
	if len(args) == 0 || args[0] == "list" {
		items, err := weaves.ListMerged(root, norn.OverlayPlanningRoot(w))
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render("Weaves"))
		for _, item := range items {
			logger.Print(styles.KV(item.ID, item.Title))
		}
		return nil
	}
	if len(args) >= 3 && args[0] == "add" {
		if len(args) < 3 {
			return fmt.Errorf("usage: norn weaves add <title> <summary>")
		}
		title := args[1]
		summary := strings.Join(args[2:], " ")
		doc := norn.Document{ID: slug(title), Title: title, Summary: summary, Body: weaves.DefaultBody(title, summary)}
		return weaves.SaveToSurface(norn.SharedPlanningRoot(w), doc)
	}
	if len(args) == 1 && args[0] == "add" {
		doc, err := promptWeaveCreation(w)
		if err != nil {
			return err
		}
		return weaves.SaveToSurface(norn.SharedPlanningRoot(w), doc)
	}
	if args[0] == "show" {
		weaveID := ""
		if len(args) >= 2 {
			weaveID = args[1]
			// Try partial match resolution
			items, err := weaves.ListMerged(root, norn.OverlayPlanningRoot(w))
			if err != nil {
				return err
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Summary})
			}
			resolved, resolveErr := resolveArtifactID(weaveID, artifactItems)
			if resolveErr == nil {
				weaveID = resolved
			}
			// If resolution fails, try direct load (might be exact match not in list)
		} else {
			items, err := weaves.ListMerged(root, norn.OverlayPlanningRoot(w))
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no weaves available; create a weave first")
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Summary})
			}
			selected, err := promptArtifactSelection("Select a weave", artifactItems)
			if err != nil {
				return err
			}
			weaveID = selected
		}
		doc, err := weaves.LoadMerged(root, norn.OverlayPlanningRoot(w), weaveID)
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render(doc.Title))
		logger.Print(doc.Body)
		return nil
	}
	if args[0] == "remove" {
		weaveID := ""
		if len(args) >= 2 {
			weaveID = args[1]
			// Try partial match resolution
			items, err := weaves.ListMerged(root, norn.OverlayPlanningRoot(w))
			if err != nil {
				return err
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Summary})
			}
			resolved, resolveErr := resolveArtifactID(weaveID, artifactItems)
			if resolveErr == nil {
				weaveID = resolved
			}
		} else {
			items, err := weaves.ListMerged(root, norn.OverlayPlanningRoot(w))
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no weaves available; create a weave first")
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Summary})
			}
			selected, err := promptArtifactSelection("Select a weave to remove", artifactItems)
			if err != nil {
				return err
			}
			weaveID = selected
		}
		return weaves.Delete(root, weaveID)
	}
	return fmt.Errorf("usage: norn weaves <list|add|show|remove>")
}

func runThreads(args []string) error {
	if showHelp(threadsHelp(), args) {
		return nil
	}
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	root := norn.SharedPlanningRoot(w)
	if len(args) == 0 {
		return fmt.Errorf("usage: norn threads <list|add|show|remove>")
	}
	if args[0] == "list" {
		weaveID := ""
		if len(args) >= 2 {
			weaveID = args[1]
		} else {
			selected, err := promptWeaveSelection(w, "Select a weave to list threads")
			if err != nil {
				return err
			}
			weaveID = selected
		}
		items, err := threads.ListMerged(root, norn.OverlayPlanningRoot(w), weaveID)
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render("Threads"))
		for _, item := range items {
			logger.Print(styles.KV(item.ID, item.Title))
		}
		return nil
	}
	if len(args) >= 4 && args[0] == "add" {
		if len(args) < 4 {
			return fmt.Errorf("usage: norn threads add <weave-id> <title> <summary>")
		}
		weaveID := args[1]
		title := args[2]
		summary := strings.Join(args[3:], " ")
		doc := norn.Document{ID: slug(title), Title: title, Summary: summary, Body: threads.DefaultBody(summary)}
		return threads.SaveToSurface(norn.SharedPlanningRoot(w), weaveID, doc)
	}
	if len(args) == 1 && args[0] == "add" {
		weaveID, doc, err := promptThreadCreation(w)
		if err != nil {
			return err
		}
		return threads.SaveToSurface(norn.SharedPlanningRoot(w), weaveID, doc)
	}
	if args[0] == "show" {
		weaveID := ""
		if len(args) >= 2 {
			weaveID = args[1]
		} else {
			selected, err := promptWeaveSelection(w, "Select a weave")
			if err != nil {
				return err
			}
			weaveID = selected
		}
		threadID := ""
		if len(args) >= 3 {
			threadID = args[2]
		} else {
			items, err := threads.ListMerged(root, norn.OverlayPlanningRoot(w), weaveID)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no threads available in weave %s; create a thread first", weaveID)
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Summary})
			}
			selected, err := promptArtifactSelection("Select a thread", artifactItems)
			if err != nil {
				return err
			}
			threadID = selected
		}
		doc, err := threads.LoadMerged(root, norn.OverlayPlanningRoot(w), weaveID, threadID)
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render(doc.Title))
		logger.Print(doc.Body)
		return nil
	}
	if args[0] == "remove" {
		weaveID := ""
		if len(args) >= 2 {
			weaveID = args[1]
		} else {
			selected, err := promptWeaveSelection(w, "Select a weave")
			if err != nil {
				return err
			}
			weaveID = selected
		}
		threadID := ""
		if len(args) >= 3 {
			threadID = args[2]
		} else {
			items, err := threads.ListMerged(root, norn.OverlayPlanningRoot(w), weaveID)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return fmt.Errorf("no threads available in weave %s; create a thread first", weaveID)
			}
			artifactItems := make([]ArtifactItem, 0, len(items))
			for _, item := range items {
				artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Summary})
			}
			selected, err := promptArtifactSelection("Select a thread to remove", artifactItems)
			if err != nil {
				return err
			}
			threadID = selected
		}
		return threads.Delete(root, weaveID, threadID)
	}
	return fmt.Errorf("usage: norn threads <list|add|show|remove>")
}

func runExport(args []string) error {
	if showHelp(exportHelp(), args) {
		return nil
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: norn export --opencode [flags]")
	}
	if args[0] != "--opencode" {
		return fmt.Errorf("usage: norn export --opencode [flags]")
	}

	w, err := norn.Load(".")
	if err != nil {
		return err
	}

	opts := export.Options{Target: "opencode"}
	for _, arg := range args[1:] {
		switch {
		case arg == "--fates":
			opts.Fates = true
		case arg == "--skills":
			opts.Skills = true
		case arg == "--dry-run":
			opts.DryRun = true
		case strings.HasPrefix(arg, "--fate="):
			opts.FateName = strings.TrimPrefix(arg, "--fate=")
		case strings.HasPrefix(arg, "--skill="):
			opts.SkillName = strings.TrimPrefix(arg, "--skill=")
		}
	}

	// If no specific flags, export all
	if !opts.Fates && !opts.Skills && opts.FateName == "" && opts.SkillName == "" {
		opts.Fates = true
		opts.Skills = true
	}

	return export.Run(w, opts)
}

func runChat(args []string) error {
	if showHelp(chatHelp(), args) {
		return nil
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: norn chat <validate|status|export|assist|preview>")
	}
	switch args[0] {
	case "validate":
		if err := opencode.Validate(); err != nil {
			return err
		}
		logger.Info("opencode available")
		return nil
	case "status":
		return runChatStatus()
	case "export":
		return runChatExport(args[1:])
	case "assist":
		return runChatAssist(args[1:])
	case "preview":
		return runChatPreview(args[1:])
	default:
		return fmt.Errorf("usage: norn chat <validate|status|export|assist|preview>")
	}
}

func runChatStatus() error {
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	status := opencode.GetStatus(w)
	logger.Print(styles.Title.Render("OpenCode Status"))
	logger.Print(styles.KV("Available", fmt.Sprintf("%t", status.Available)))
	logger.Print(styles.KV("Enabled", fmt.Sprintf("%t", status.Enabled)))
	if status.Enabled {
		logger.Print(styles.KV("Provider", status.Provider))
		logger.Print(styles.KV("Model", status.Model))
		logger.Print(styles.KV("Agent", status.Agent))
		logger.Print(styles.KV("Response language", status.ResponseLang))
		logger.Print(styles.KV("Drafting mode", status.DraftingMode))
	}
	logger.Print(styles.KV("Agents path", status.AgentsPath))
	logger.Print(styles.KV("Agents generated", fmt.Sprintf("%d", status.AgentsCount)))
	if len(status.AgentNames) > 0 {
		logger.Print(styles.KV("Agent names", strings.Join(status.AgentNames, ", ")))
	}
	return nil
}

func runChatExport(args []string) error {
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	if !w.Runes.OpenCode.Enabled {
		return fmt.Errorf("opencode is not enabled; run 'norn runes edit --set opencode.enabled=true' to enable")
	}
	// Export agents
	if err := fates.ExportOpenCode(norn.FatesRoot(w), norn.ToolsRoot(w), norn.OpenCodeAgentsRoot(w)); err != nil {
		return err
	}
	// Export config
	targetDir := "."
	for _, arg := range args {
		if strings.HasPrefix(arg, "--output=") {
			targetDir = strings.TrimPrefix(arg, "--output=")
		}
	}
	if err := opencode.ExportConfig(w, targetDir); err != nil {
		return err
	}
	logger.Info("opencode exported", "agents", norn.OpenCodeAgentsRoot(w), "config", filepath.Join(targetDir, "norn-opencode.json"))
	return nil
}

func runChatAssist(args []string) error {
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	if !w.Runes.OpenCode.Enabled {
		return fmt.Errorf("opencode is not enabled; run 'norn runes edit --set opencode.enabled=true' to enable")
	}
	prompt := ""
	for _, arg := range args {
		if strings.HasPrefix(arg, "--prompt=") {
			prompt = strings.TrimPrefix(arg, "--prompt=")
		}
	}
	if prompt == "" {
		// Interactive mode
		p := ""
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewText().Title("What would you like help with?").Description("Describe what you want the AI to generate or help with.").Value(&p),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
		prompt = p
	}
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	context := fmt.Sprintf("Project: %s. Stack: %s.", w.Runes.Name, strings.Join(w.Runes.Tooling.Languages, ", "))
	logger.Info("requesting assistance", "prompt", prompt)
	assist, err := opencode.Assist(w.Runes.OpenCode, context, prompt)
	if err != nil {
		return err
	}
	// Preview what was generated
	logger.Print(styles.Title.Render("Assisted Results"))
	if len(assist.Weaves) > 0 {
		logger.Print(styles.Label.Render("Weaves:"))
		for _, item := range assist.Weaves {
			logger.Print(fmt.Sprintf("  - %s: %s", item.Title, item.Summary))
		}
	}
	if len(assist.Patterns) > 0 {
		logger.Print(styles.Label.Render("Patterns:"))
		for _, item := range assist.Patterns {
			logger.Print(fmt.Sprintf("  - %s: %s", item.Title, item.Summary))
		}
	}
	if len(assist.Skills) > 0 {
		logger.Print(styles.Label.Render("Skills:"))
		for _, item := range assist.Skills {
			logger.Print(fmt.Sprintf("  - %s: %s", item.Title, item.Summary))
		}
	}
	// Ask for approval
	confirmed := false
	confirm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().Title("Save these artifacts?").Value(&confirmed),
		),
	)
	if err := confirm.Run(); err != nil {
		return err
	}
	if !confirmed {
		logger.Info("assistance discarded")
		return nil
	}
	// Save artifacts
	for _, item := range assist.Weaves {
		item.ID = slug(item.Title)
		item.Body = weaves.DefaultBody(item.Title, item.Summary)
		if err := weaves.SaveToSurface(norn.SharedPlanningRoot(w), item); err != nil {
			logger.Warn("failed to save weave", "title", item.Title, "error", err)
		} else {
			logger.Info("weave saved", "id", item.ID)
		}
	}
	for _, item := range assist.Patterns {
		item.ID = slug(item.Title)
		if err := saveDoc("patterns", filepath.Join(norn.SharedPlanningRoot(w), "patterns"), item); err != nil {
			logger.Warn("failed to save pattern", "title", item.Title, "error", err)
		} else {
			logger.Info("pattern saved", "id", item.ID)
		}
	}
	for _, item := range assist.Skills {
		item.ID = slug(item.Title)
		if err := saveDoc("skills", filepath.Join(norn.SharedPlanningRoot(w), "skills"), item); err != nil {
			logger.Warn("failed to save skill", "title", item.Title, "error", err)
		} else {
			logger.Info("skill saved", "id", item.ID)
		}
	}
	return nil
}

func runChatPreview(args []string) error {
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	if !w.Runes.OpenCode.Enabled {
		return fmt.Errorf("opencode is not enabled; run 'norn runes edit --set opencode.enabled=true' to enable")
	}
	prompt := ""
	for _, arg := range args {
		if strings.HasPrefix(arg, "--prompt=") {
			prompt = strings.TrimPrefix(arg, "--prompt=")
		}
	}
	if prompt == "" {
		return fmt.Errorf("usage: norn chat preview --prompt=<prompt>")
	}
	context := fmt.Sprintf("Project: %s. Stack: %s.", w.Runes.Name, strings.Join(w.Runes.Tooling.Languages, ", "))
	logger.Info("generating preview", "prompt", prompt)
	assist, err := opencode.Assist(w.Runes.OpenCode, context, prompt)
	if err != nil {
		return err
	}
	logger.Print(styles.Title.Render("Preview (not saved)"))
	if len(assist.Weaves) > 0 {
		logger.Print(styles.Label.Render("Weaves:"))
		for _, item := range assist.Weaves {
			logger.Print(fmt.Sprintf("  - %s: %s", item.Title, item.Summary))
			if item.Body != "" {
				logger.Print(fmt.Sprintf("    Body: %s", strings.ReplaceAll(item.Body, "\n", "\n    ")))
			}
		}
	}
	if len(assist.Patterns) > 0 {
		logger.Print(styles.Label.Render("Patterns:"))
		for _, item := range assist.Patterns {
			logger.Print(fmt.Sprintf("  - %s: %s", item.Title, item.Summary))
			if item.Body != "" {
				logger.Print(fmt.Sprintf("    Body: %s", strings.ReplaceAll(item.Body, "\n", "\n    ")))
			}
		}
	}
	if len(assist.Skills) > 0 {
		logger.Print(styles.Label.Render("Skills:"))
		for _, item := range assist.Skills {
			logger.Print(fmt.Sprintf("  - %s: %s", item.Title, item.Summary))
			if item.Body != "" {
				logger.Print(fmt.Sprintf("    Body: %s", strings.ReplaceAll(item.Body, "\n", "\n    ")))
			}
		}
	}
	logger.Print(styles.Dimmed.Render("Use 'norn chat assist --prompt=...' to save these artifacts."))
	return nil
}

func ensureWorkspacePaths(workspace norn.Workspace) error {
	paths := []string{
		norn.SharedPlanningRoot(workspace),
		norn.OverlayPlanningRoot(workspace),
		norn.ToolsRoot(workspace),
		norn.SkillsRoot(workspace),
		norn.FatesRoot(workspace),
		norn.SpindleRoot(workspace),
		norm(filepath.Join(norn.SharedPlanningRoot(workspace), "weaves")),
		norm(filepath.Join(norn.SharedPlanningRoot(workspace), "patterns")),
		norm(filepath.Join(norn.SharedPlanningRoot(workspace), "skills")),
		norm(filepath.Join(norn.OverlayPlanningRoot(workspace), "weaves")),
		norm(norn.OpenCodeAgentsRoot(workspace)),
	}
	for _, path := range paths {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func bootstrapCommands(workspace norn.Workspace, detected norn.Detection) error {
	root := norn.ToolsRoot(workspace)
	defaults := []norn.ManagedTool{
		{ID: "git-status", Title: "Git status", Category: "inspect", Command: "git status", Pattern: "git status*", Risk: "low", Roles: []string{"keeper", "weaver", "judge", "fates"}},
		{ID: "git-diff", Title: "Git diff", Category: "inspect", Command: "git diff", Pattern: "git diff*", Risk: "low", Roles: []string{"keeper", "weaver", "judge", "fates"}},
		{ID: "git-log", Title: "Git log", Category: "inspect", Command: "git log", Pattern: "git log*", Risk: "low", Roles: []string{"keeper", "weaver", "judge", "fates"}},
	}
	for _, language := range detected.Languages {
		defaults = append(defaults, languageDefaults(language)...)
	}
	for _, tool := range detected.Tools {
		defaults = append(defaults, toolDefaults(tool)...)
	}
	seen := map[string]bool{}
	for _, item := range defaults {
		if seen[item.ID] {
			continue
		}
		seen[item.ID] = true
		if err := toolstore.Save(root, item); err != nil {
			return err
		}
	}
	return nil
}

func languageDefaults(language string) []norn.ManagedTool {
	switch language {
	case "go":
		return []norn.ManagedTool{{ID: "go-build", Title: "Go build", Category: "build", Command: "go build ./...", Pattern: "go build*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}, {ID: "go-test", Title: "Go test", Category: "test", Command: "go test ./...", Pattern: "go test*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}, {ID: "go-run", Title: "Go run", Category: "run", Command: "go run .", Pattern: "go run*", Risk: "low", Roles: []string{"weaver", "fates"}}}
	case "node":
		return []norn.ManagedTool{{ID: "npm-install", Title: "npm install", Category: "setup", Command: "npm install", Pattern: "npm install*", Risk: "medium", Roles: []string{"weaver", "fates"}}, {ID: "npm-test", Title: "npm test", Category: "test", Command: "npm test", Pattern: "npm test*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}, {ID: "npm-build", Title: "npm build", Category: "build", Command: "npm run build", Pattern: "npm run build*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	case "bun":
		return []norn.ManagedTool{{ID: "bun-install", Title: "bun install", Category: "setup", Command: "bun install", Pattern: "bun install*", Risk: "medium", Roles: []string{"weaver", "fates"}}, {ID: "bun-test", Title: "bun test", Category: "test", Command: "bun test", Pattern: "bun test*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	case "java":
		return []norn.ManagedTool{{ID: "maven-package", Title: "mvn package", Category: "build", Command: "mvn package", Pattern: "mvn package*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}, {ID: "maven-test", Title: "mvn test", Category: "test", Command: "mvn test", Pattern: "mvn test*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	case "kotlin":
		return []norn.ManagedTool{{ID: "gradle-build", Title: "gradle build", Category: "build", Command: "./gradlew build", Pattern: "./gradlew build*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	case "rust":
		return []norn.ManagedTool{{ID: "cargo-build", Title: "cargo build", Category: "build", Command: "cargo build", Pattern: "cargo build*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}, {ID: "cargo-test", Title: "cargo test", Category: "test", Command: "cargo test", Pattern: "cargo test*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	case ".net":
		return []norn.ManagedTool{{ID: "dotnet-build", Title: "dotnet build", Category: "build", Command: "dotnet build", Pattern: "dotnet build*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}, {ID: "dotnet-test", Title: "dotnet test", Category: "test", Command: "dotnet test", Pattern: "dotnet test*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	default:
		return nil
	}
}

func toolDefaults(tool string) []norn.ManagedTool {
	switch tool {
	case "make":
		return []norn.ManagedTool{{ID: "make-build", Title: "make build", Category: "build", Command: "make build", Pattern: "make build*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	case "docker":
		return []norn.ManagedTool{{ID: "docker-build", Title: "docker build", Category: "build", Command: "docker build .", Pattern: "docker build*", Risk: "medium", Roles: []string{"weaver", "fates"}}}
	case "mise":
		return []norn.ManagedTool{{ID: "mise-trust", Title: "mise trust", Category: "setup", Command: "mise trust", Pattern: "mise trust*", Risk: "medium", Roles: []string{"weaver", "fates"}}}
	default:
		return nil
	}
}

func detectWorkspaceMode(root string) norn.WorkspaceMode {
	if exists(filepath.Join(root, ".hydra.yaml")) {
		return norn.WorkspaceModeWorkspace
	}
	return norn.WorkspaceModeRepo
}

func themeOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("Tokyo Night", "tokyonight"),
		huh.NewOption("Catppuccin", "catppuccin"),
		huh.NewOption("Dracula", "dracula"),
		huh.NewOption("Nord", "nord"),
		huh.NewOption("One Dark", "onedark"),
	}
}

// ArtifactItem represents a selectable artifact for fuzzy find.
type ArtifactItem struct {
	ID      string
	Title   string
	Summary string
}

// promptArtifactSelection shows an interactive filterable list to select an artifact.
// Returns the selected ID or error if cancelled.
func promptArtifactSelection(title string, items []ArtifactItem) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no artifacts available")
	}
	// Sort alphabetically by ID
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	options := make([]huh.Option[string], 0, len(items))
	for _, item := range items {
		label := item.ID
		if item.Title != "" {
			label = fmt.Sprintf("%s — %s", item.ID, item.Title)
		}
		options = append(options, huh.NewOption(label, item.ID))
	}
	selected := items[0].ID
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Description(fmt.Sprintf("%d available. Type to filter.", len(items))).
				Options(options...).
				Value(&selected),
		),
	)
	if err := form.Run(); err != nil {
		return "", err
	}
	return selected, nil
}

// promptWeaveSelection prompts the user to select a weave from the workspace.
func promptWeaveSelection(w norn.Workspace, title string) (string, error) {
	items, err := weaves.ListMerged(norn.SharedPlanningRoot(w), norn.OverlayPlanningRoot(w))
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no weaves available; create a weave first")
	}
	artifactItems := make([]ArtifactItem, 0, len(items))
	for _, item := range items {
		artifactItems = append(artifactItems, ArtifactItem{ID: item.ID, Title: item.Title, Summary: item.Summary})
	}
	return promptArtifactSelection(title, artifactItems)
}

// resolveArtifactID attempts exact match first, then substring match.
// If exactly one substring match, asks for confirmation.
// If multiple matches, returns error with list.
func resolveArtifactID(query string, items []ArtifactItem) (string, error) {
	if query == "" {
		return "", fmt.Errorf("no artifact ID provided")
	}
	queryLower := strings.ToLower(query)

	// Exact match first
	for _, item := range items {
		if strings.ToLower(item.ID) == queryLower {
			return item.ID, nil
		}
	}

	// Substring match
	var matches []ArtifactItem
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.ID), queryLower) ||
			strings.Contains(strings.ToLower(item.Title), queryLower) {
			matches = append(matches, item)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no artifact matches %q", query)
	case 1:
		// Ask for confirmation
		confirmed := false
		preview := fmt.Sprintf("ID: %s\nTitle: %s", matches[0].ID, matches[0].Title)
		if matches[0].Summary != "" {
			preview += fmt.Sprintf("\nSummary: %s", matches[0].Summary)
		}
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().Title("Did you mean?").Description(preview),
				huh.NewConfirm().Title(fmt.Sprintf("Use '%s'?", matches[0].ID)).Value(&confirmed),
			),
		)
		if err := form.Run(); err != nil {
			return "", err
		}
		if !confirmed {
			return "", fmt.Errorf("selection cancelled")
		}
		return matches[0].ID, nil
	default:
		// Multiple matches — error with list
		var b strings.Builder
		b.WriteString(fmt.Sprintf("ambiguous match for %q; multiple artifacts found:\n", query))
		for _, m := range matches {
			b.WriteString(fmt.Sprintf("  - %s (%s)\n", m.ID, m.Title))
		}
		return "", fmt.Errorf("%s", b.String())
	}
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func listDocs(kind, root string) ([]norn.Document, error) {
	switch kind {
	case "patterns":
		return patterns.List(root)
	case "skills":
		return skills.List(root)
	default:
		return nil, fmt.Errorf("unknown kind: %s", kind)
	}
}

func loadDoc(kind, root, id string) (norn.Document, error) {
	switch kind {
	case "patterns":
		return patterns.Load(root, id)
	case "skills":
		return skills.Load(root, id)
	default:
		return norn.Document{}, fmt.Errorf("unknown kind: %s", kind)
	}
}

func saveDoc(kind, root string, doc norn.Document) error {
	switch kind {
	case "patterns":
		return patterns.Save(root, doc)
	case "skills":
		return skills.Save(root, doc)
	default:
		return fmt.Errorf("unknown kind: %s", kind)
	}
}

func deleteDoc(kind, root, id string) error {
	switch kind {
	case "patterns":
		return patterns.Delete(root, id)
	case "skills":
		return skills.Delete(root, id)
	default:
		return fmt.Errorf("unknown kind: %s", kind)
	}
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "_", "-", ":", "-")
	return replacer.Replace(value)
}

func countWeaves(w norn.Workspace) int {
	items, err := weaves.ListMerged(norn.SharedPlanningRoot(w), norn.OverlayPlanningRoot(w))
	if err != nil {
		return 0
	}
	return len(items)
}

func countThreads(w norn.Workspace) int {
	items, err := weaves.ListMerged(norn.SharedPlanningRoot(w), norn.OverlayPlanningRoot(w))
	if err != nil {
		return 0
	}
	total := 0
	for _, item := range items {
		threadsForWeave, err := threads.ListMerged(norn.SharedPlanningRoot(w), norn.OverlayPlanningRoot(w), item.ID)
		if err != nil {
			continue
		}
		total += len(threadsForWeave)
	}
	return total
}

func norm(path string) string {
	return filepath.Clean(path)
}

func saveWeave(w norn.Workspace, doc norn.Document) error {
	return weaves.SaveToSurface(norn.SharedPlanningRoot(w), doc)
}

func saveThread(w norn.Workspace, weaveID string, doc norn.Document) error {
	return threads.SaveToSurface(norn.SharedPlanningRoot(w), weaveID, doc)
}

func promptWeaveCreation(w norn.Workspace) (norn.Document, error) {
	title := ""
	id := ""
	summary := ""
	goal := ""
	userStories := ""
	scope := ""
	acceptance := ""
	form := huh.NewForm(
		huh.NewGroup(huh.NewInput().Title("Title").Description("Human-readable name — e.g., 'API Authentication'").Placeholder("API Authentication").Value(&title)),
		huh.NewGroup(huh.NewInput().Title("ID").Description("URL-friendly identifier. Leave empty to auto-generate from title.").Placeholder("api-authentication").Value(&id)),
		huh.NewGroup(huh.NewText().Title("Summary").Description("One-line description of this weave's purpose").Placeholder("Secure all API endpoints with JWT-based authentication").Value(&summary)),
		huh.NewGroup(huh.NewText().Title("Goal").Description("What this weave aims to achieve when complete").Placeholder("All API requests require valid authentication tokens").Value(&goal)),
		huh.NewGroup(huh.NewText().Title("User stories").Description("Format: As a [role], I want [goal], so that [benefit]. One per line.").Placeholder("As a user, I want secure endpoints, so that my data is protected\nAs a developer, I want clear auth errors, so that I can debug issues").Value(&userStories)),
		huh.NewGroup(huh.NewText().Title("Scope").Description("What's included. Use 'Out of scope:' to define boundaries. One per line.").Placeholder("JWT token generation and validation\nOut of scope: OAuth, SAML, session management").Value(&scope)),
		huh.NewGroup(huh.NewText().Title("Acceptance").Description("Testable criteria for completion. One per line.").Placeholder("All endpoints return 401 for missing tokens\nToken expiration returns 403 with clear message\nRate limiting applies per user, not per IP").Value(&acceptance)),
	)
	if err := form.Run(); err != nil {
		return norn.Document{}, err
	}
	if id == "" {
		id = slug(title)
	}
	body := buildArtifactBody(goal, "User Stories", userStories, "Scope", scope, "Acceptance", acceptance)
	preview := fmt.Sprintf("ID: %s\nPath: %s\n\n%s", id, weaves.ReadmePath(norn.SharedPlanningRoot(w), id), body)
	confirmed := true
	confirm := huh.NewForm(huh.NewGroup(huh.NewNote().Title("Preview").Description(preview), huh.NewConfirm().Title("Create weave with this content?").Value(&confirmed)))
	if err := confirm.Run(); err != nil {
		return norn.Document{}, err
	}
	if !confirmed {
		return norn.Document{}, fmt.Errorf("weave creation cancelled")
	}
	return norn.Document{ID: id, Title: title, Summary: summary, Body: body}, nil
}

func promptThreadCreation(w norn.Workspace) (string, norn.Document, error) {
	items, err := weaves.ListMerged(norn.SharedPlanningRoot(w), norn.OverlayPlanningRoot(w))
	if err != nil {
		return "", norn.Document{}, err
	}
	if len(items) == 0 {
		return "", norn.Document{}, fmt.Errorf("no weaves available; create a weave first")
	}
	options := make([]huh.Option[string], 0, len(items))
	selectedWeave := items[0].ID
	for _, item := range items {
		options = append(options, huh.NewOption(fmt.Sprintf("%s (%s)", item.Title, item.ID), item.ID))
	}
	title := ""
	id := ""
	summary := ""
	goal := ""
	userStory := ""
	strands := ""
	acceptance := ""
	form := huh.NewForm(
		huh.NewGroup(huh.NewSelect[string]().Title("Parent weave").Description("Which weave this thread belongs to").Options(options...).Value(&selectedWeave)),
		huh.NewGroup(huh.NewInput().Title("Title").Description("Human-readable name — e.g., 'Implement JWT middleware'").Placeholder("Implement JWT middleware").Value(&title)),
		huh.NewGroup(huh.NewInput().Title("ID").Description("URL-friendly identifier. Leave empty to auto-generate from title.").Placeholder("jwt-middleware").Value(&id)),
		huh.NewGroup(huh.NewText().Title("Summary").Description("One-line description of this thread's purpose").Placeholder("Add JWT token validation to all API endpoints").Value(&summary)),
		huh.NewGroup(huh.NewText().Title("Goal").Description("Specific objective for this thread").Placeholder("Ensure all API requests are authenticated with valid JWT tokens").Value(&goal)),
		huh.NewGroup(huh.NewText().Title("User story").Description("Format: As a [role], I want [goal], so that [benefit]").Placeholder("As an API consumer, I want secure endpoints, so that my data is protected").Value(&userStory)),
		huh.NewGroup(huh.NewText().Title("Strands").Description("Sub-tasks or implementation steps. One per line.").Placeholder("Create auth middleware\nAdd token validation\nWrite tests\nUpdate API docs").Value(&strands)),
		huh.NewGroup(huh.NewText().Title("Acceptance").Description("Testable completion criteria. One per line.").Placeholder("All API endpoints reject requests without valid JWT\nToken expiration is handled gracefully\nTests cover valid, expired, and malformed tokens").Value(&acceptance)),
	)
	if err := form.Run(); err != nil {
		return "", norn.Document{}, err
	}
	if id == "" {
		id = slug(title)
	}
	body := buildArtifactBody(goal, "User Story", userStory, "Strands", strands, "Acceptance", acceptance)
	preview := fmt.Sprintf("Weave: %s\nID: %s\nPath: %s\n\n%s", selectedWeave, id, threads.Path(norn.SharedPlanningRoot(w), selectedWeave, id), body)
	confirmed := true
	confirm := huh.NewForm(huh.NewGroup(huh.NewNote().Title("Preview").Description(preview), huh.NewConfirm().Title("Create thread with this content?").Value(&confirmed)))
	if err := confirm.Run(); err != nil {
		return "", norn.Document{}, err
	}
	if !confirmed {
		return "", norn.Document{}, fmt.Errorf("thread creation cancelled")
	}
	return selectedWeave, norn.Document{ID: id, Title: title, Summary: summary, Body: body}, nil
}

func buildArtifactBody(goal string, sections ...string) string {
	var b strings.Builder
	b.WriteString("## Goal\n\n")
	b.WriteString(strings.TrimSpace(goal))
	b.WriteString("\n\n")
	for i := 0; i+1 < len(sections); i += 2 {
		title := sections[i]
		content := strings.TrimSpace(sections[i+1])
		if content == "" {
			continue
		}
		b.WriteString("## ")
		b.WriteString(title)
		b.WriteString("\n\n")
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(title, "User") {
				b.WriteString("- ")
				b.WriteString(line)
				b.WriteString("\n")
			} else {
				b.WriteString("- ")
				b.WriteString(line)
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}
	b.WriteString("## Documentation\n\n")
	b.WriteString("- [ ] CLI --help text updated for any new commands or flags\n")
	b.WriteString("- [ ] Project docs (docs/) updated with user-facing guides\n")
	b.WriteString("- [ ] Specifications documented for AI agent consumption\n")
	b.WriteString("- [ ] Integration boundaries documented (what Norn owns vs what OpenCode owns)\n")
	b.WriteString("\n")
	b.WriteString("## Guides\n\n")
	b.WriteString("- Quick start path for users\n")
	b.WriteString("- Configuration reference (if applicable)\n")
	b.WriteString("\n")
	b.WriteString("## Specifications\n\n")
	b.WriteString("- Data formats and schemas\n")
	b.WriteString("- API/contracts for AI agent interaction\n")
	b.WriteString("\n")
	b.WriteString("## Notes\n\n- replace this template content with project-specific details\n")
	return strings.TrimSpace(b.String())
}

func fatesHelp() HelpTopic {
	return HelpTopic{
		Name:        "norn fates",
		Description: "Manage agent fates",
		Usage:       "norn fates <command>",
		Commands: []CommandHelp{
			{
				Name:        "list",
				Description: "List all fates",
				Usage:       "norn fates list",
			},
			{
				Name:        "show",
				Description: "Show a fate and its OpenCode agent definition",
				Usage:       "norn fates show <name>",
			},
			{
				Name:        "add",
				Description: "Add a new fate",
				Usage:       "norn fates add <name> <description>",
				Examples: []string{
					"norn fates add guardian \"Guards project boundaries\"",
				},
			},
			{
				Name:        "edit",
				Description: "Edit a fate interactively or non-interactively",
				Usage:       "norn fates edit <name> [--set field=value ...]",
				Flags: []FlagHelp{
					{Name: "--set field=value", Description: "Set a field value (description, model, temperature, body, allow_edit)"},
				},
				Examples: []string{
					"norn fates edit keeper --set description=\"New description\"",
					"norn fates edit keeper --set model=\"gpt-4\" --set allow_edit=true",
				},
			},
			{
				Name:        "remove",
				Description: "Remove a fate",
				Usage:       "norn fates remove <name>",
			},
		},
	}
}

func docCollectionHelp(kind string) HelpTopic {
	singular := kind
	if strings.HasSuffix(kind, "s") {
		singular = kind[:len(kind)-1]
	}
	return HelpTopic{
		Name:        fmt.Sprintf("norn %s", kind),
		Description: fmt.Sprintf("Manage %s documents", kind),
		Usage:       fmt.Sprintf("norn %s <command>", kind),
		Commands: []CommandHelp{
			{
				Name:        "list",
				Description: fmt.Sprintf("List all %s", kind),
				Usage:       fmt.Sprintf("norn %s list", kind),
			},
			{
				Name:        "add",
				Description: fmt.Sprintf("Add a new %s", singular),
				Usage:       fmt.Sprintf("norn %s add <title> <summary>", kind),
			},
			{
				Name:        "show",
				Description: fmt.Sprintf("Show a %s", singular),
				Usage:       fmt.Sprintf("norn %s show <id>", kind),
			},
			{
				Name:        "edit",
				Description: fmt.Sprintf("Edit a %s interactively or non-interactively", singular),
				Usage:       fmt.Sprintf("norn %s edit <id> [--set field=value ...]", kind),
				Flags: []FlagHelp{
					{Name: "--set field=value", Description: "Set a field value (title, summary, body)"},
				},
				Examples: []string{
					fmt.Sprintf("norn %s edit my-doc --set title=\"New Title\"", kind),
					fmt.Sprintf("norn %s edit my-doc --set summary=\"Updated summary\"", kind),
				},
			},
			{
				Name:        "remove",
				Description: fmt.Sprintf("Remove a %s", singular),
				Usage:       fmt.Sprintf("norn %s remove <id>", kind),
			},
		},
	}
}

func toolsHelp() HelpTopic {
	return HelpTopic{
		Name:        "norn tools",
		Description: "Manage tool permission definitions",
		Usage:       "norn tools <command>",
		Commands: []CommandHelp{
			{
				Name:        "list",
				Description: "List all tool permission definitions",
				Usage:       "norn tools list",
			},
			{
				Name:        "add",
				Description: "Add a new tool permission definition",
				Usage:       "norn tools add <title> <category> <command>",
				Examples: []string{
					"norn tools add lint lint \"npm run lint\"",
				},
			},
			{
				Name:        "show",
				Description: "Show tool details",
				Usage:       "norn tools show <id>",
			},
			{
				Name:        "edit",
				Description: "Edit a tool interactively or non-interactively",
				Usage:       "norn tools edit <id> [--set field=value ...]",
				Flags: []FlagHelp{
					{Name: "--set field=value", Description: "Set a field value (title, description, category, command, pattern, risk, roles)"},
				},
				Examples: []string{
					"norn tools edit lint --set command=\"npm run lint\"",
					"norn tools edit lint --set risk=low --set roles=weaver,judge",
				},
			},
			{
				Name:        "remove",
				Description: "Remove a tool",
				Usage:       "norn tools remove <id>",
			},
		},
	}
}

func exportHelp() HelpTopic {
	return HelpTopic{
		Name:        "norn export",
		Description: "Export Norn artifacts to agent tools",
		Usage:       "norn export --opencode [flags]",
		Commands: []CommandHelp{
			{
				Name:        "--opencode",
				Description: "Export to OpenCode format",
				Usage:       "norn export --opencode",
			},
			{
				Name:        "--fates",
				Description: "Export fates only",
				Usage:       "norn export --opencode --fates",
			},
			{
				Name:        "--skills",
				Description: "Export skills only",
				Usage:       "norn export --opencode --skills",
			},
			{
				Name:        "--fate=<name>",
				Description: "Export specific fate",
				Usage:       "norn export --opencode --fate=keeper",
			},
			{
				Name:        "--skill=<name>",
				Description: "Export specific skill",
				Usage:       "norn export --opencode --skill=go-api",
			},
			{
				Name:        "--dry-run",
				Description: "Preview without writing",
				Usage:       "norn export --opencode --dry-run",
			},
		},
	}
}

func weavesHelp() HelpTopic {
	return HelpTopic{
		Name:        "norn weaves",
		Description: "Manage weave planning artifacts",
		Usage:       "norn weaves <command>",
		Commands: []CommandHelp{
			{
				Name:        "list",
				Description: "List all weaves",
				Usage:       "norn weaves list",
			},
			{
				Name:        "add",
				Description: "Add a new weave",
				Usage:       "norn weaves add <title> <summary>",
				Examples: []string{
					"norn weaves add \"API Contract\" \"Document API expectations\"",
				},
			},
			{
				Name:        "show",
				Description: "Show a weave",
				Usage:       "norn weaves show <id>",
			},
			{
				Name:        "remove",
				Description: "Remove a weave",
				Usage:       "norn weaves remove <id>",
			},
		},
	}
}

func threadsHelp() HelpTopic {
	return HelpTopic{
		Name:        "norn threads",
		Description: "Manage thread planning artifacts",
		Usage:       "norn threads <command>",
		Commands: []CommandHelp{
			{
				Name:        "list",
				Description: "List threads for a weave",
				Usage:       "norn threads list <weave-id>",
			},
			{
				Name:        "add",
				Description: "Add a new thread",
				Usage:       "norn threads add <weave-id> <title> <summary>",
				Examples: []string{
					"norn threads add my-weave \"Add CLI\" \"Implement the command\"",
				},
			},
			{
				Name:        "show",
				Description: "Show a thread",
				Usage:       "norn threads show <weave-id> <thread-id>",
			},
			{
				Name:        "remove",
				Description: "Remove a thread",
				Usage:       "norn threads remove <weave-id> <thread-id>",
			},
		},
	}
}

func initHelp() HelpTopic {
	return HelpTopic{
		Name:        "norn init",
		Description: "Bootstrap a new Norn project or workspace",
		Usage:       "norn init [flags]",
		Commands: []CommandHelp{
			{
				Name:        "init",
				Description: "Bootstrap a new Norn project or workspace",
				Usage:       "norn init [flags]",
				Flags: []FlagHelp{
					{Name: "--no-interactive", Description: "Run in non-interactive mode"},
					{Name: "--enable-opencode", Description: "Enable OpenCode integration"},
					{Name: "--name=<name>", Description: "Project name"},
					{Name: "--theme=<theme>", Description: "UI theme"},
					{Name: "--languages=<list>", Description: "Comma-separated language list"},
					{Name: "--tools=<list>", Description: "Comma-separated tool list"},
					{Name: "--frameworks=<list>", Description: "Comma-separated framework list"},
					{Name: "--model=<model>", Description: "OpenCode model"},
					{Name: "--agent=<agent>", Description: "OpenCode agent"},
					{Name: "--prompt=<prompt>", Description: "OpenCode assisted init prompt"},
				},
				Examples: []string{
					"norn init",
					"norn init --no-interactive --name=my-project --enable-opencode",
				},
			},
		},
	}
}
