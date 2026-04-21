package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	cmdstore "github.com/mssantosdev/norn/internal/commands"
	"github.com/mssantosdev/norn/internal/detect"
	"github.com/mssantosdev/norn/internal/fates"
	"github.com/mssantosdev/norn/internal/loom"
	"github.com/mssantosdev/norn/internal/norn"
	"github.com/mssantosdev/norn/internal/opencode"
	"github.com/mssantosdev/norn/internal/patterns"
	"github.com/mssantosdev/norn/internal/skills"
	"github.com/mssantosdev/norn/internal/threads"
	"github.com/mssantosdev/norn/internal/ui/logger"
	"github.com/mssantosdev/norn/internal/ui/styles"
	"github.com/mssantosdev/norn/internal/weaves"
)

var errUsage = errors.New("usage: norn <init|status|detect|fates|patterns|skills|commands|weaves|threads|warps|runes|chat>")

func Run(args []string) error {
	if len(args) == 0 {
		logger.Print(styles.AppHeader.Render("Norn") + " " + styles.Subtitle.Render("Weave-aware multi-agent harness"))
		logger.Print("use `norn init` to bootstrap a project or workspace")
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
	case "commands":
		return runCommands(args[1:])
	case "weaves":
		return runWeaves(args[1:])
	case "threads":
		return runThreads(args[1:])
	case "warps":
		return runWarps(args[1:])
	case "runes":
		return runRunes(args[1:])
	case "chat":
		return runChat(args[1:])
	default:
		return errUsage
	}
}

func runInit(args []string) error {
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
	if opts.LocalOverlayDir == "" {
		opts.LocalOverlayDir = ".norn/loom"
	}
	if opts.Mode == "" {
		opts.Mode = norn.PlanningModeFolder
	}
	if opts.PlanningPath == "" {
		if opts.Mode == norn.PlanningModeBranch {
			opts.PlanningPath = ".loom"
		} else {
			opts.PlanningPath = "loom"
		}
	}

	workspace := norn.Workspace{
		Root: ".",
		Runes: norn.RuneFile{
			Name:        opts.Name,
			Mode:        detectWorkspaceMode("."),
			Preferences: norn.PreferencesConfig{Language: "en", Verbosity: "normal"},
			UI:          norn.UIConfig{Theme: opts.Theme},
			Planning:    norn.PlanningConfig{Mode: opts.Mode, Path: opts.PlanningPath, Branch: opts.Branch, DefaultSurface: "shared"},
			Overlay:     norn.OverlayConfig{Path: opts.LocalOverlayDir},
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
	if err := fates.ExportOpenCode(norn.FatesRoot(workspace), norn.CommandsRoot(workspace), norn.OpenCodeAgentsRoot(workspace)); err != nil {
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
	logger.Info("workspace initialized", "name", workspace.Runes.Name, "planning", workspace.Runes.Planning.Path, "mode", workspace.Runes.Planning.Mode)
	return nil
}

func parseInitArgs(args []string) (norn.InitOptions, error) {
	opts := norn.InitOptions{Mode: norn.PlanningModeFolder}
	for _, arg := range args {
		switch {
		case arg == "--no-interactive":
			opts.NonInteractive = true
		case arg == "--enable-opencode":
			opts.EnableOpenCode = true
		case arg == "--create-branch":
			opts.CreateBranch = true
		case strings.HasPrefix(arg, "--mode="):
			opts.Mode = norn.PlanningMode(strings.TrimPrefix(arg, "--mode="))
		case strings.HasPrefix(arg, "--name="):
			opts.Name = strings.TrimPrefix(arg, "--name=")
		case strings.HasPrefix(arg, "--path="):
			opts.PlanningPath = strings.TrimPrefix(arg, "--path=")
		case strings.HasPrefix(arg, "--branch="):
			opts.Branch = strings.TrimPrefix(arg, "--branch=")
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
	planningMode := string(opts.Mode)
	branchChoice := "__new__"
	branchName := "loom"
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
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Project name").Value(&projectName),
			huh.NewSelect[string]().Title("Planning mode").Options(
				huh.NewOption("Simple folder (recommended)", string(norn.PlanningModeFolder)),
				huh.NewOption("Planning branch worktree", string(norn.PlanningModeBranch)),
			).Value(&planningMode),
			huh.NewSelect[string]().Title("Select existing branch or create new").Options(branchOptions()...).Value(&branchChoice),
			huh.NewInput().Title("New planning branch name").Value(&branchName),
			huh.NewSelect[string]().Title("Initial structure").Options(
				huh.NewOption("Standard Norn structure", "standard"),
				huh.NewOption("Empty", "empty"),
				huh.NewOption("Guided", "guided"),
				huh.NewOption("Help me with OpenCode", "opencode"),
			).Value(&skeleton),
			huh.NewConfirm().Title("Enable OpenCode integration?").Value(&openCodeEnabled),
			huh.NewInput().Title("OpenCode prompt").Value(&openCodePrompt),
			huh.NewSelect[string]().Title("Theme").Options(themeOptions()...).Value(&theme),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}
	opts.Name = projectName
	opts.Mode = norn.PlanningMode(planningMode)
	opts.Skeleton = skeleton
	opts.EnableOpenCode = openCodeEnabled
	if openCodeEnabled || skeleton == "opencode" {
		opts.OpenCodePrompt = openCodePrompt
	}
	opts.Theme = theme
	if opts.Mode == norn.PlanningModeBranch {
		if branchChoice == "__new__" {
			opts.Branch = branchName
			opts.CreateBranch = true
		} else {
			opts.Branch = branchChoice
		}
	}
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
	logger.Print(styles.KV("planning mode", string(w.Runes.Planning.Mode)))
	logger.Print(styles.KV("planning path", w.Runes.Planning.Path))
	logger.Print(styles.KV("overlay path", w.Runes.Overlay.Path))
	logger.Print(styles.KV("hydra", fmt.Sprintf("%t", w.Runes.Hydra.Enabled)))
	logger.Print(styles.KV("opencode", fmt.Sprintf("%t", w.Runes.OpenCode.Enabled)))
	logger.Print(styles.KV("languages", strings.Join(w.Runes.Tooling.Languages, ", ")))
	logger.Print(styles.KV("tools", strings.Join(w.Runes.Tooling.Tools, ", ")))
	logger.Print(styles.KV("frameworks", strings.Join(w.Runes.Tooling.Frameworks, ", ")))
	logger.Print(styles.KV("fates", fmt.Sprintf("%d", norn.CountFiles(norn.FatesRoot(w), ".yaml"))))
	logger.Print(styles.KV("commands", fmt.Sprintf("%d", norn.CountFiles(norn.CommandsRoot(w), ".yaml"))))
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
	if len(args) == 2 && args[0] == "show" {
		item, err := fates.Load(norn.FatesRoot(w), args[1])
		if err != nil {
			return err
		}
		rendered, err := fates.RenderOpenCode(item, norn.CommandsRoot(w))
		if err != nil {
			return err
		}
		logger.Print(rendered)
		return nil
	}
	return fmt.Errorf("usage: norn fates <list|show <name>>")
}

func runDocCollection(kind string, args []string) error {
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
	if len(args) == 2 && args[0] == "show" {
		doc, err := loadDoc(kind, root, args[1])
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render(doc.Title))
		logger.Print(doc.Body)
		return nil
	}
	if len(args) == 2 && args[0] == "remove" {
		return deleteDoc(kind, root, args[1])
	}
	return fmt.Errorf("usage: norn %s <list|add|show|remove>", kind)
}

func runCommands(args []string) error {
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	root := norn.CommandsRoot(w)
	if len(args) == 0 || args[0] == "list" {
		items, err := cmdstore.List(root)
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render("Commands"))
		for _, item := range items {
			logger.Print(styles.KV(item.ID, fmt.Sprintf("%s [%s]", item.Command, strings.Join(item.Roles, ", "))))
		}
		return nil
	}
	if len(args) >= 4 && args[0] == "add" {
		item := norn.ManagedCommand{ID: slug(args[1]), Title: args[1], Category: args[2], Command: strings.Join(args[3:], " "), Roles: []string{"weaver"}}
		if err := cmdstore.Save(root, item); err != nil {
			return err
		}
		return fates.ExportOpenCode(norn.FatesRoot(w), root, norn.OpenCodeAgentsRoot(w))
	}
	if len(args) == 2 && args[0] == "remove" {
		if err := cmdstore.Delete(root, args[1]); err != nil {
			return err
		}
		return fates.ExportOpenCode(norn.FatesRoot(w), root, norn.OpenCodeAgentsRoot(w))
	}
	return fmt.Errorf("usage: norn commands <list|add|remove>")
}

func runWeaves(args []string) error {
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
		surface, remaining, err := parseSurfaceArgs(args[1:])
		if err != nil {
			return err
		}
		if len(remaining) < 2 {
			return fmt.Errorf("usage: norn weaves add [--surface=shared|local|both] <title> <summary>")
		}
		title := remaining[0]
		summary := strings.Join(remaining[1:], " ")
		doc := norn.Document{ID: slug(title), Title: title, Summary: summary, Body: weaves.DefaultBody(title, summary)}
		return saveWeaveToSurface(w, surface, doc)
	}
	if len(args) == 1 && args[0] == "add" {
		surface, doc, err := promptWeaveCreation(w)
		if err != nil {
			return err
		}
		return saveWeaveToSurface(w, surface, doc)
	}
	if len(args) == 2 && args[0] == "show" {
		doc, err := weaves.LoadMerged(root, norn.OverlayPlanningRoot(w), args[1])
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render(doc.Title))
		logger.Print(doc.Body)
		return nil
	}
	if len(args) == 2 && args[0] == "remove" {
		return weaves.Delete(root, args[1])
	}
	return fmt.Errorf("usage: norn weaves <list|add|show|remove>")
}

func runThreads(args []string) error {
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	root := norn.SharedPlanningRoot(w)
	if len(args) == 0 {
		return fmt.Errorf("usage: norn threads <list|add|show|remove>")
	}
	if len(args) == 2 && args[0] == "list" {
		items, err := threads.ListMerged(root, norn.OverlayPlanningRoot(w), args[1])
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
		surface, remaining, err := parseSurfaceArgs(args[1:])
		if err != nil {
			return err
		}
		if len(remaining) < 3 {
			return fmt.Errorf("usage: norn threads add [--surface=shared|local|both] <weave-id> <title> <summary>")
		}
		weaveID := remaining[0]
		title := remaining[1]
		summary := strings.Join(remaining[2:], " ")
		doc := norn.Document{ID: slug(title), Title: title, Summary: summary, Body: threads.DefaultBody(summary)}
		return saveThreadToSurface(w, surface, weaveID, doc)
	}
	if len(args) == 1 && args[0] == "add" {
		surface, weaveID, doc, err := promptThreadCreation(w)
		if err != nil {
			return err
		}
		return saveThreadToSurface(w, surface, weaveID, doc)
	}
	if len(args) == 3 && args[0] == "show" {
		doc, err := threads.LoadMerged(root, norn.OverlayPlanningRoot(w), args[1], args[2])
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render(doc.Title))
		logger.Print(doc.Body)
		return nil
	}
	if len(args) == 3 && args[0] == "remove" {
		return threads.Delete(root, args[1], args[2])
	}
	return fmt.Errorf("usage: norn threads <list <weave-id>|add <weave-id> <title> <summary>|show <weave-id> <thread-id>|remove <weave-id> <thread-id>>")
}

func runChat(args []string) error {
	if len(args) == 1 && args[0] == "validate" {
		if err := opencode.Validate(); err != nil {
			return err
		}
		logger.Info("opencode available")
		return nil
	}
	return fmt.Errorf("usage: norn chat validate")
}

func ensureWorkspacePaths(workspace norn.Workspace) error {
	paths := []string{
		norn.SharedPlanningRoot(workspace),
		norn.OverlayPlanningRoot(workspace),
		norn.CommandsRoot(workspace),
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
	root := norn.CommandsRoot(workspace)
	defaults := []norn.ManagedCommand{
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
		if err := cmdstore.Save(root, item); err != nil {
			return err
		}
	}
	return nil
}

func languageDefaults(language string) []norn.ManagedCommand {
	switch language {
	case "go":
		return []norn.ManagedCommand{{ID: "go-build", Title: "Go build", Category: "build", Command: "go build ./...", Pattern: "go build*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}, {ID: "go-test", Title: "Go test", Category: "test", Command: "go test ./...", Pattern: "go test*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}, {ID: "go-run", Title: "Go run", Category: "run", Command: "go run .", Pattern: "go run*", Risk: "low", Roles: []string{"weaver", "fates"}}}
	case "node":
		return []norn.ManagedCommand{{ID: "npm-install", Title: "npm install", Category: "setup", Command: "npm install", Pattern: "npm install*", Risk: "medium", Roles: []string{"weaver", "fates"}}, {ID: "npm-test", Title: "npm test", Category: "test", Command: "npm test", Pattern: "npm test*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}, {ID: "npm-build", Title: "npm build", Category: "build", Command: "npm run build", Pattern: "npm run build*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	case "bun":
		return []norn.ManagedCommand{{ID: "bun-install", Title: "bun install", Category: "setup", Command: "bun install", Pattern: "bun install*", Risk: "medium", Roles: []string{"weaver", "fates"}}, {ID: "bun-test", Title: "bun test", Category: "test", Command: "bun test", Pattern: "bun test*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	case "java":
		return []norn.ManagedCommand{{ID: "maven-package", Title: "mvn package", Category: "build", Command: "mvn package", Pattern: "mvn package*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}, {ID: "maven-test", Title: "mvn test", Category: "test", Command: "mvn test", Pattern: "mvn test*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	case "kotlin":
		return []norn.ManagedCommand{{ID: "gradle-build", Title: "gradle build", Category: "build", Command: "./gradlew build", Pattern: "./gradlew build*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	case "rust":
		return []norn.ManagedCommand{{ID: "cargo-build", Title: "cargo build", Category: "build", Command: "cargo build", Pattern: "cargo build*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}, {ID: "cargo-test", Title: "cargo test", Category: "test", Command: "cargo test", Pattern: "cargo test*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	case ".net":
		return []norn.ManagedCommand{{ID: "dotnet-build", Title: "dotnet build", Category: "build", Command: "dotnet build", Pattern: "dotnet build*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}, {ID: "dotnet-test", Title: "dotnet test", Category: "test", Command: "dotnet test", Pattern: "dotnet test*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	default:
		return nil
	}
}

func toolDefaults(tool string) []norn.ManagedCommand {
	switch tool {
	case "make":
		return []norn.ManagedCommand{{ID: "make-build", Title: "make build", Category: "build", Command: "make build", Pattern: "make build*", Risk: "low", Roles: []string{"weaver", "judge", "fates"}}}
	case "docker":
		return []norn.ManagedCommand{{ID: "docker-build", Title: "docker build", Category: "build", Command: "docker build .", Pattern: "docker build*", Risk: "medium", Roles: []string{"weaver", "fates"}}}
	case "mise":
		return []norn.ManagedCommand{{ID: "mise-trust", Title: "mise trust", Category: "setup", Command: "mise trust", Pattern: "mise trust*", Risk: "medium", Roles: []string{"weaver", "fates"}}}
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

func branchOptions() []huh.Option[string] {
	options := []huh.Option[string]{huh.NewOption("Create new branch", "__new__")}
	for _, branch := range gitBranches() {
		options = append(options, huh.NewOption(branch, branch))
	}
	return options
}

func gitBranches() []string {
	if !exists(".git/refs/heads") {
		return nil
	}
	var out []string
	_ = filepath.Walk(".git/refs/heads", func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(".git/refs/heads", path)
		if err == nil {
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(out)
	return out
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

func parseSurfaceArgs(args []string) (string, []string, error) {
	surface := ""
	remaining := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.HasPrefix(arg, "--surface=") {
			surface = strings.TrimPrefix(arg, "--surface=")
			continue
		}
		remaining = append(remaining, arg)
	}
	if surface == "" {
		return "", remaining, nil
	}
	switch surface {
	case "shared", "local", "both":
		return surface, remaining, nil
	default:
		return "", nil, fmt.Errorf("invalid surface %q; expected shared, local, or both", surface)
	}
}

func saveWeaveToSurface(w norn.Workspace, surface string, doc norn.Document) error {
	if surface == "" {
		surface = w.Runes.Planning.DefaultSurface
	}
	switch surface {
	case "shared":
		return weaves.SaveToSurface(norn.SharedPlanningRoot(w), doc)
	case "local":
		return weaves.SaveToSurface(norn.OverlayPlanningRoot(w), doc)
	case "both":
		if err := weaves.SaveToSurface(norn.SharedPlanningRoot(w), doc); err != nil {
			return err
		}
		return weaves.SaveToSurface(norn.OverlayPlanningRoot(w), doc)
	default:
		return fmt.Errorf("unsupported surface %q", surface)
	}
}

func saveThreadToSurface(w norn.Workspace, surface, weaveID string, doc norn.Document) error {
	if surface == "" {
		surface = w.Runes.Planning.DefaultSurface
	}
	switch surface {
	case "shared":
		return threads.SaveToSurface(norn.SharedPlanningRoot(w), weaveID, doc)
	case "local":
		return threads.SaveToSurface(norn.OverlayPlanningRoot(w), weaveID, doc)
	case "both":
		if err := threads.SaveToSurface(norn.SharedPlanningRoot(w), weaveID, doc); err != nil {
			return err
		}
		return threads.SaveToSurface(norn.OverlayPlanningRoot(w), weaveID, doc)
	default:
		return fmt.Errorf("unsupported surface %q", surface)
	}
}

func promptWeaveCreation(w norn.Workspace) (string, norn.Document, error) {
	surface := "shared"
	title := ""
	id := ""
	summary := ""
	goal := ""
	userStories := ""
	scope := ""
	acceptance := ""
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Title("Planning surface").Options(
				huh.NewOption("Shared (loom/)", "shared"),
				huh.NewOption("Local (.norn/loom/)", "local"),
				huh.NewOption("Both", "both"),
			).Value(&surface),
			huh.NewInput().Title("Title").Value(&title),
			huh.NewInput().Title("ID").Description("Leave empty to derive from title").Value(&id),
			huh.NewText().Title("Summary").Value(&summary),
			huh.NewText().Title("Goal").Value(&goal),
			huh.NewText().Title("User stories").Description("One or more user stories, one per line").Value(&userStories),
			huh.NewText().Title("Scope").Description("Scope items, one per line").Value(&scope),
			huh.NewText().Title("Acceptance").Description("Acceptance items, one per line").Value(&acceptance),
		),
	)
	if err := form.Run(); err != nil {
		return "", norn.Document{}, err
	}
	if id == "" {
		id = slug(title)
	}
	body := buildArtifactBody(goal, "User Stories", userStories, "Scope", scope, "Acceptance", acceptance)
	preview := fmt.Sprintf("Surface: %s\nID: %s\nShared path: %s\nLocal path: %s\n\n%s", surface, id, weaves.ReadmePath(norn.SharedPlanningRoot(w), id), weaves.ReadmePath(norn.OverlayPlanningRoot(w), id), body)
	confirmed := true
	confirm := huh.NewForm(huh.NewGroup(huh.NewNote().Title("Preview").Description(preview), huh.NewConfirm().Title("Create weave with this content?").Value(&confirmed)))
	if err := confirm.Run(); err != nil {
		return "", norn.Document{}, err
	}
	if !confirmed {
		return "", norn.Document{}, fmt.Errorf("weave creation cancelled")
	}
	return surface, norn.Document{ID: id, Title: title, Summary: summary, Body: body}, nil
}

func promptThreadCreation(w norn.Workspace) (string, string, norn.Document, error) {
	items, err := weaves.ListMerged(norn.SharedPlanningRoot(w), norn.OverlayPlanningRoot(w))
	if err != nil {
		return "", "", norn.Document{}, err
	}
	if len(items) == 0 {
		return "", "", norn.Document{}, fmt.Errorf("no weaves available; create a weave first")
	}
	options := make([]huh.Option[string], 0, len(items))
	selectedWeave := items[0].ID
	for _, item := range items {
		options = append(options, huh.NewOption(fmt.Sprintf("%s (%s)", item.Title, item.ID), item.ID))
	}
	surface := "shared"
	title := ""
	id := ""
	summary := ""
	goal := ""
	userStory := ""
	strands := ""
	acceptance := ""
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Title("Planning surface").Options(
				huh.NewOption("Shared (loom/)", "shared"),
				huh.NewOption("Local (.norn/loom/)", "local"),
				huh.NewOption("Both", "both"),
			).Value(&surface),
			huh.NewSelect[string]().Title("Parent weave").Options(options...).Value(&selectedWeave),
			huh.NewInput().Title("Title").Value(&title),
			huh.NewInput().Title("ID").Description("Leave empty to derive from title").Value(&id),
			huh.NewText().Title("Summary").Value(&summary),
			huh.NewText().Title("Goal").Value(&goal),
			huh.NewText().Title("User story").Value(&userStory),
			huh.NewText().Title("Strands").Description("One strand per line").Value(&strands),
			huh.NewText().Title("Acceptance").Description("One item per line").Value(&acceptance),
		),
	)
	if err := form.Run(); err != nil {
		return "", "", norn.Document{}, err
	}
	if id == "" {
		id = slug(title)
	}
	body := buildArtifactBody(goal, "User Story", userStory, "Strands", strands, "Acceptance", acceptance)
	preview := fmt.Sprintf("Surface: %s\nWeave: %s\nID: %s\nShared path: %s\nLocal path: %s\n\n%s", surface, selectedWeave, id, threads.Path(norn.SharedPlanningRoot(w), selectedWeave, id), threads.Path(norn.OverlayPlanningRoot(w), selectedWeave, id), body)
	confirmed := true
	confirm := huh.NewForm(huh.NewGroup(huh.NewNote().Title("Preview").Description(preview), huh.NewConfirm().Title("Create thread with this content?").Value(&confirmed)))
	if err := confirm.Run(); err != nil {
		return "", "", norn.Document{}, err
	}
	if !confirmed {
		return "", "", norn.Document{}, fmt.Errorf("thread creation cancelled")
	}
	return surface, selectedWeave, norn.Document{ID: id, Title: title, Summary: summary, Body: body}, nil
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
	b.WriteString("## Notes\n\n- replace this template content with project-specific details\n")
	return strings.TrimSpace(b.String())
}
