package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/huh"
	"github.com/mssantosdev/norn/internal/norn"
	"github.com/mssantosdev/norn/internal/ui/logger"
	"github.com/mssantosdev/norn/internal/ui/styles"
	"gopkg.in/yaml.v3"
)

func runRunes(args []string) error {
	if showHelp(runesHelp(), args) {
		return nil
	}
	if len(args) == 0 {
		return runRunesInteractive()
	}
	switch args[0] {
	case "show":
		return runRunesShow(args[1:])
	case "resolve":
		return runRunesResolve(args[1:])
	case "edit":
		return runRunesEdit(args[1:])
	default:
		return fmt.Errorf("usage: norn runes <show|resolve|edit>")
	}
}

func runesHelp() HelpTopic {
	return HelpTopic{
		Name:        "norn runes",
		Description: "Manage Norn configuration",
		Usage:       "norn runes <command> [flags]",
		Commands: []CommandHelp{
			{
				Name:        "show",
				Description: "Show configuration for a scope",
				Usage:       "norn runes show [--scope=global|workspace|local]",
				Flags: []FlagHelp{
					{Name: "--scope=global|workspace|local", Description: "Config scope to show"},
				},
				Examples: []string{
					"norn runes show",
					"norn runes show --scope=global",
				},
			},
			{
				Name:        "resolve",
				Description: "Show effective configuration with origin metadata",
				Usage:       "norn runes resolve [--format=table|yaml]",
				Flags: []FlagHelp{
					{Name: "--format=table|yaml", Description: "Output format (default: table)"},
				},
				Examples: []string{
					"norn runes resolve",
					"norn runes resolve --format=yaml",
				},
			},
			{
				Name:        "edit",
				Description: "Edit configuration interactively or non-interactively",
				Usage:       "norn runes edit [--scope=global|workspace|local] [--set path=value] [--unset path]",
				Flags: []FlagHelp{
					{Name: "--scope=global|workspace|local", Description: "Config scope to edit"},
					{Name: "--set path=value", Description: "Set a config value"},
					{Name: "--unset path", Description: "Unset a config value"},
				},
				Examples: []string{
					"norn runes edit",
					"norn runes edit --scope=workspace --set preferences.language=pt-BR",
					"norn runes edit --scope=local --unset opencode.response_language",
				},
			},
		},
	}
}

func runRunesInteractive() error {
	action := "show"
	view := "effective"
	scope := string(norn.RuneScopeWorkspace)
	hasWorkspace := true
	if _, err := norn.FindRoot("."); err != nil {
		hasWorkspace = false
		scope = string(norn.RuneScopeGlobal)
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Title("Action").Options(
				huh.NewOption("Show config", "show"),
				huh.NewOption("Resolve config origins", "resolve"),
				huh.NewOption("Edit config", "edit"),
			).Value(&action),
		),
		huh.NewGroup(
			huh.NewSelect[string]().Title("Show").Options(runeShowOptions(hasWorkspace)...).Value(&view),
		).WithHideFunc(func() bool { return action != "show" }),
		huh.NewGroup(
			huh.NewSelect[string]().Title("Scope").Options(runeScopeOptions(hasWorkspace)...).Value(&scope),
		).WithHideFunc(func() bool { return action != "edit" }),
	)
	if err := form.Run(); err != nil {
		return err
	}
	if action == "show" {
		if view == "effective" {
			return runRunesShow(nil)
		}
		return runRunesShow([]string{"--scope=" + view})
	}
	if action == "resolve" {
		return runRunesResolve(nil)
	}
	return runRunesEdit([]string{"--scope=" + scope})
}

func runRunesShow(args []string) error {
	scope, err := parseRuneScopeArg(args)
	if err != nil {
		return err
	}
	if scope == "" {
		root, err := requireWorkspaceRoot()
		if err != nil {
			return err
		}
		resolution, err := norn.ResolveRunes(root)
		if err != nil {
			return err
		}
		data, err := yaml.Marshal(resolution.Effective)
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render("Runes"))
		logger.Print(string(data))
		return nil
	}
	root, err := rootForScope(scope)
	if err != nil {
		return err
	}
	layer, err := norn.LoadScopeMap(root, scope)
	if err != nil {
		return err
	}
	data, err := norn.MarshalScopeMap(layer)
	if err != nil {
		return err
	}
	logger.Print(styles.Title.Render(fmt.Sprintf("Runes (%s)", scope)))
	logger.Print(string(data))
	return nil
}

func runRunesResolve(args []string) error {
	format, err := parseRuneFormatArg(args)
	if err != nil {
		return err
	}
	root, err := currentRuneRoot()
	if err != nil {
		return err
	}
	resolution, err := norn.ResolveRunes(root)
	if err != nil {
		return err
	}
	logger.Print(styles.Title.Render("Resolved Runes"))
	if format == "table" {
		logger.Print(renderRuneResolutionTable(resolution))
		return nil
	}
	data, err := yaml.Marshal(resolution)
	if err != nil {
		return err
	}
	logger.Print(string(data))
	return nil
}

func runRunesEdit(args []string) error {
	scope, sets, unsets, err := parseRuneEditArgs(args)
	if err != nil {
		return err
	}
	if scope == "" {
		hasWorkspace := true
		if _, err := norn.FindRoot("."); err != nil {
			hasWorkspace = false
		}
		selected, err := promptRuneScope(hasWorkspace)
		if err != nil {
			return err
		}
		scope = selected
	}
	root, err := rootForScope(scope)
	if err != nil {
		return err
	}
	if len(sets) == 0 && len(unsets) == 0 {
		return runRunesEditInteractive(root, scope)
	}
	layer, err := norn.LoadScopeMap(root, scope)
	if err != nil {
		return err
	}
	for _, path := range unsets {
		if err := norn.ValidateRuneEditPath(path); err != nil {
			return err
		}
		layer = norn.UnsetScopeValue(layer, path)
	}
	for _, item := range sets {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --set %q; expected path=value", item)
		}
		path := strings.TrimSpace(parts[0])
		value, err := norn.ParseRuneEditValue(path, parts[1])
		if err != nil {
			return err
		}
		layer = norn.SetScopeValue(layer, path, value)
	}
	if err := norn.SaveScopeMap(root, scope, layer); err != nil {
		return err
	}
	logger.Info("runes updated", "scope", scope, "path", norn.ScopeDisplayPath(root, scope))
	return nil
}

func runRunesEditInteractive(root string, scope norn.RuneScope) error {
	layer, err := norn.LoadScopeMap(root, scope)
	if err != nil {
		return err
	}
	resolution, err := norn.ResolveRunes(root)
	if err != nil {
		return err
	}
	name := scopeStringValue(layer, "name")
	language := scopeStringValue(layer, "preferences.language")
	verbosity := scopeEnumValue(layer, "preferences.verbosity")
	theme := scopeEnumValue(layer, "ui.theme")
	planningPath := scopeStringValue(layer, "planning.path")
	openCodeEnabled := scopeTriStateValue(layer, "opencode.enabled")
	openCodeProvider := scopeStringValue(layer, "opencode.provider")
	openCodeModel := scopeStringValue(layer, "opencode.model")
	openCodeAgent := scopeStringValue(layer, "opencode.agent")
	responseLanguage := scopeStringValue(layer, "opencode.response_language")
	draftingMode := scopeEnumValue(layer, "opencode.drafting_mode")
	languages := scopeCSVValue(layer, "tooling.languages")
	tools := scopeCSVValue(layer, "tooling.tools")
	frameworks := scopeCSVValue(layer, "tooling.frameworks")
	hydraEnabled := scopeTriStateValue(layer, "hydra.enabled")

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title("Editing scope").Description(fmt.Sprintf("%s\nBlank values inherit from lower precedence layers or defaults.", norn.ScopeDisplayPath(root, scope))),
		),
		huh.NewGroup(
			huh.NewInput().Title("Name").Description(fieldDescription(resolution, "name")).Value(&name),
			huh.NewInput().Title("Preferred language").Description(fieldDescription(resolution, "preferences.language")).Value(&language),
			huh.NewSelect[string]().Title("Verbosity").Description(fieldDescription(resolution, "preferences.verbosity")).Options(enumOptions("<inherit>", "", []string{"quiet", "normal", "loud"})...).Value(&verbosity),
			huh.NewSelect[string]().Title("Theme").Description(fieldDescription(resolution, "ui.theme")).Options(enumOptions("<inherit>", "", []string{"tokyonight", "catppuccin", "dracula", "nord", "onedark"})...).Value(&theme),
		),
		huh.NewGroup(
			huh.NewInput().Title("Planning path").Description(fieldDescription(resolution, "planning.path")).Value(&planningPath),
		),
		huh.NewGroup(
			huh.NewSelect[string]().Title("OpenCode enabled").Description(fieldDescription(resolution, "opencode.enabled")).Options(triStateOptions()...).Value(&openCodeEnabled),
			huh.NewInput().Title("OpenCode provider").Description(fieldDescription(resolution, "opencode.provider")).Value(&openCodeProvider),
			huh.NewInput().Title("OpenCode model").Description(fieldDescription(resolution, "opencode.model")).Value(&openCodeModel),
			huh.NewInput().Title("OpenCode agent").Description(fieldDescription(resolution, "opencode.agent")).Value(&openCodeAgent),
			huh.NewInput().Title("OpenCode response language").Description(fieldDescription(resolution, "opencode.response_language")).Value(&responseLanguage),
			huh.NewSelect[string]().Title("OpenCode drafting mode").Description(fieldDescription(resolution, "opencode.drafting_mode")).Options(enumOptions("<inherit>", "", []string{"ask", "auto"})...).Value(&draftingMode),
		),
		huh.NewGroup(
			huh.NewInput().Title("Languages").Description(fieldDescription(resolution, "tooling.languages")+" Blank to inherit. Use comma-separated values.").Value(&languages),
			huh.NewInput().Title("Tools").Description(fieldDescription(resolution, "tooling.tools")+" Blank to inherit. Use comma-separated values.").Value(&tools),
			huh.NewInput().Title("Frameworks").Description(fieldDescription(resolution, "tooling.frameworks")+" Blank to inherit. Use comma-separated values.").Value(&frameworks),
		),
		huh.NewGroup(
			huh.NewSelect[string]().Title("Hydra enabled").Description(fieldDescription(resolution, "hydra.enabled")).Options(triStateOptions()...).Value(&hydraEnabled),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}
	updated := map[string]any{}
	setString(updated, "name", name)
	setString(updated, "preferences.language", language)
	setString(updated, "preferences.verbosity", verbosity)
	setString(updated, "ui.theme", theme)
	setString(updated, "planning.path", planningPath)
	setTriState(updated, "opencode.enabled", openCodeEnabled)
	setString(updated, "opencode.provider", openCodeProvider)
	setString(updated, "opencode.model", openCodeModel)
	setString(updated, "opencode.agent", openCodeAgent)
	setString(updated, "opencode.response_language", responseLanguage)
	setString(updated, "opencode.drafting_mode", draftingMode)
	setCSV(updated, "tooling.languages", languages)
	setCSV(updated, "tooling.tools", tools)
	setCSV(updated, "tooling.frameworks", frameworks)
	setTriState(updated, "hydra.enabled", hydraEnabled)

	if err := norn.SaveScopeMap(root, scope, updated); err != nil {
		return err
	}
	logger.Info("runes updated", "scope", scope, "path", norn.ScopeDisplayPath(root, scope))
	return nil
}

func parseRuneScopeArg(args []string) (norn.RuneScope, error) {
	var scope norn.RuneScope
	for _, arg := range args {
		if strings.HasPrefix(arg, "--scope=") {
			parsed, err := norn.ParseRuneScope(strings.TrimPrefix(arg, "--scope="))
			if err != nil {
				return "", err
			}
			scope = parsed
			continue
		}
		return "", fmt.Errorf("unknown argument: %s", arg)
	}
	return scope, nil
}

func parseRuneEditArgs(args []string) (norn.RuneScope, []string, []string, error) {
	var scope norn.RuneScope
	var sets []string
	var unsets []string
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--scope="):
			parsed, err := norn.ParseRuneScope(strings.TrimPrefix(arg, "--scope="))
			if err != nil {
				return "", nil, nil, err
			}
			scope = parsed
		case strings.HasPrefix(arg, "--set="):
			sets = append(sets, strings.TrimPrefix(arg, "--set="))
		case strings.HasPrefix(arg, "--unset="):
			unsets = append(unsets, strings.TrimPrefix(arg, "--unset="))
		default:
			return "", nil, nil, fmt.Errorf("unknown edit argument: %s", arg)
		}
	}
	return scope, sets, unsets, nil
}

func requireWorkspaceRoot() (string, error) {
	return norn.FindRoot(".")
}

func currentRuneRoot() (string, error) {
	if root, err := norn.FindRoot("."); err == nil {
		return root, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Abs(wd)
}

func rootForScope(scope norn.RuneScope) (string, error) {
	if scope == norn.RuneScopeGlobal {
		return currentRuneRoot()
	}
	return requireWorkspaceRoot()
}

func parseRuneFormatArg(args []string) (string, error) {
	format := "table"
	for _, arg := range args {
		if strings.HasPrefix(arg, "--format=") {
			format = strings.TrimPrefix(arg, "--format=")
			continue
		}
		return "", fmt.Errorf("unknown argument: %s", arg)
	}
	switch format {
	case "table", "yaml":
		return format, nil
	default:
		return "", fmt.Errorf("invalid format %q; expected table or yaml", format)
	}
}

func renderRuneResolutionTable(resolution norn.RuneResolution) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, styles.TableHeader.Render("Field")+"\t"+styles.TableHeader.Render("Value")+"\t"+styles.TableHeader.Render("Origin"))
	for _, path := range norn.ResolvedRunePaths() {
		value := norn.FieldValueString(norn.ResolvedFieldValue(resolution, path))
		origin := strings.Join(resolution.Origins[path], "; ")
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", path, value, origin)
	}
	_ = w.Flush()
	return b.String()
}

func promptRuneScope(hasWorkspace bool) (norn.RuneScope, error) {
	selected := string(norn.RuneScopeGlobal)
	if hasWorkspace {
		selected = string(norn.RuneScopeWorkspace)
	}
	form := huh.NewForm(huh.NewGroup(huh.NewSelect[string]().Title("Scope").Options(runeScopeOptions(hasWorkspace)...).Value(&selected)))
	if err := form.Run(); err != nil {
		return "", err
	}
	return norn.ParseRuneScope(selected)
}

func runeScopeOptions(hasWorkspace bool) []huh.Option[string] {
	options := []huh.Option[string]{huh.NewOption("Global (~/.config/norn/runes.yaml)", string(norn.RuneScopeGlobal))}
	if hasWorkspace {
		options = append(options,
			huh.NewOption("Workspace (.norn/runes.yaml)", string(norn.RuneScopeWorkspace)),
			huh.NewOption("Local (.norn/runes.local.yaml)", string(norn.RuneScopeLocal)),
		)
	}
	return options
}

func runeShowOptions(hasWorkspace bool) []huh.Option[string] {
	options := []huh.Option[string]{huh.NewOption("Effective resolved config", "effective"), huh.NewOption("Global scope", string(norn.RuneScopeGlobal))}
	if hasWorkspace {
		options = append(options,
			huh.NewOption("Workspace scope", string(norn.RuneScopeWorkspace)),
			huh.NewOption("Local scope", string(norn.RuneScopeLocal)),
		)
	}
	return options
}

func enumOptions(emptyLabel, emptyValue string, values []string) []huh.Option[string] {
	options := []huh.Option[string]{huh.NewOption(emptyLabel, emptyValue)}
	for _, value := range values {
		options = append(options, huh.NewOption(value, value))
	}
	return options
}

func triStateOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("<inherit>", "inherit"),
		huh.NewOption("true", "true"),
		huh.NewOption("false", "false"),
	}
}

func fieldDescription(resolution norn.RuneResolution, path string) string {
	value := norn.FieldValueString(norn.ResolvedFieldValue(resolution, path))
	origins := strings.Join(resolution.Origins[path], ", ")
	if origins == "" {
		origins = "unknown"
	}
	return fmt.Sprintf("Effective: %s [%s]", value, origins)
}

func scopeStringValue(layer map[string]any, path string) string {
	value, ok := lookupScopeValue(layer, path)
	if !ok {
		return ""
	}
	return fmt.Sprint(value)
}

func scopeEnumValue(layer map[string]any, path string) string {
	return scopeStringValue(layer, path)
}

func scopeTriStateValue(layer map[string]any, path string) string {
	value, ok := lookupScopeValue(layer, path)
	if !ok {
		return "inherit"
	}
	if typed, ok := value.(bool); ok {
		if typed {
			return "true"
		}
		return "false"
	}
	return "inherit"
}

func scopeCSVValue(layer map[string]any, path string) string {
	value, ok := lookupScopeValue(layer, path)
	if !ok {
		return ""
	}
	return strings.Join(asScopeStrings(value), ",")
}

func lookupScopeValue(layer map[string]any, path string) (any, bool) {
	current := any(layer)
	for _, part := range strings.Split(path, ".") {
		mapping, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		value, ok := mapping[part]
		if !ok {
			return nil, false
		}
		current = value
	}
	return current, true
}

func asScopeStrings(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			out = append(out, fmt.Sprint(item))
		}
		return out
	default:
		return nil
	}
}

func setString(layer map[string]any, path, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	norn.SetScopeValue(layer, path, value)
}

func setTriState(layer map[string]any, path, value string) {
	switch value {
	case "true":
		norn.SetScopeValue(layer, path, true)
	case "false":
		norn.SetScopeValue(layer, path, false)
	}
}

func setCSV(layer map[string]any, path, value string) {
	items := splitCSV(value)
	if len(items) == 0 {
		return
	}
	norn.SetScopeValue(layer, path, items)
}
