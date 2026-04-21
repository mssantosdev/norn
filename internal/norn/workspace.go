package norn

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const runesPath = ".norn/runes.yaml"
const globalRunesPath = ".config/norn/runes.yaml"
const localOverridePath = ".norn/runes.local.yaml"

var ErrWorkspaceNotFound = errors.New("norn workspace not found")

func FindRoot(start string) (string, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(current, runesPath)); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", ErrWorkspaceNotFound
		}
		current = parent
	}
}

func Load(start string) (Workspace, error) {
	root, err := FindRoot(start)
	if err != nil {
		return Workspace{}, err
	}
	runes, err := loadLayeredRunes(root)
	if err != nil {
		return Workspace{}, err
	}
	applyDefaults(&runes, root)
	return Workspace{Root: root, Runes: runes}, nil
}

func Save(workspace Workspace) error {
	applyDefaults(&workspace.Runes, workspace.Root)
	data, err := yaml.Marshal(workspace.Runes)
	if err != nil {
		return fmt.Errorf("marshal runes: %w", err)
	}
	path := filepath.Join(workspace.Root, runesPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func applyDefaults(runes *RuneFile, root string) {
	if runes.Version == "" {
		runes.Version = "0.0.1"
	}
	if runes.Name == "" {
		runes.Name = filepath.Base(root)
	}
	if runes.Mode == "" {
		if hasHydra(root) {
			runes.Mode = WorkspaceModeWorkspace
		} else {
			runes.Mode = WorkspaceModeRepo
		}
	}
	if runes.UI.Theme == "" {
		runes.UI.Theme = "tokyonight"
	}
	if runes.Preferences.Language == "" {
		runes.Preferences.Language = "en"
	}
	if runes.Preferences.Verbosity == "" {
		runes.Preferences.Verbosity = "normal"
	}
	if runes.Planning.Mode == "" {
		runes.Planning.Mode = PlanningModeFolder
	}
	if runes.Planning.Path == "" {
		if runes.Planning.Mode == PlanningModeBranch {
			runes.Planning.Path = ".loom"
		} else {
			runes.Planning.Path = "loom"
		}
	}
	if runes.Overlay.Path == "" {
		runes.Overlay.Path = ".norn/loom"
	}
	if runes.Planning.DefaultSurface == "" {
		runes.Planning.DefaultSurface = "shared"
	}
	if runes.OpenCode.Provider == "" {
		runes.OpenCode.Provider = "github-copilot"
	}
	if runes.OpenCode.Model == "" {
		runes.OpenCode.Model = "github-copilot/gpt-5.4-mini"
	}
	if runes.OpenCode.Agent == "" {
		runes.OpenCode.Agent = "build"
	}
	if runes.OpenCode.ResponseLanguage == "" {
		runes.OpenCode.ResponseLanguage = runes.Preferences.Language
	}
	if runes.OpenCode.DraftingMode == "" {
		runes.OpenCode.DraftingMode = "ask"
	}
	runes.Tooling.Languages = dedupeSorted(runes.Tooling.Languages)
	runes.Tooling.Tools = dedupeSorted(runes.Tooling.Tools)
	runes.Tooling.Frameworks = dedupeSorted(runes.Tooling.Frameworks)
	runes.Hydra.Enabled = runes.Hydra.Enabled || hasHydra(root)
}

func loadLayeredRunes(root string) (RuneFile, error) {
	var merged RuneFile
	paths := []string{globalConfigPath(), filepath.Join(root, runesPath), filepath.Join(root, localOverridePath)}
	for _, path := range paths {
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return RuneFile{}, fmt.Errorf("read runes layer %s: %w", path, err)
		}
		var layer runeLayer
		if err := yaml.Unmarshal(data, &layer); err != nil {
			return RuneFile{}, fmt.Errorf("parse runes layer %s: %w", path, err)
		}
		mergeRuneLayer(&merged, layer)
	}
	return merged, nil
}

type runeLayer struct {
	Version     string              `yaml:"version"`
	Name        string              `yaml:"name"`
	Mode        WorkspaceMode       `yaml:"mode"`
	Preferences PreferencesConfig   `yaml:"preferences"`
	UI          UIConfig            `yaml:"ui"`
	Planning    PlanningConfig      `yaml:"planning"`
	Overlay     OverlayConfig       `yaml:"overlay"`
	OpenCode    openCodeConfigLayer `yaml:"opencode"`
	Tooling     ToolingConfig       `yaml:"tooling"`
	Hydra       hydraConfigLayer    `yaml:"hydra"`
}

type openCodeConfigLayer struct {
	Enabled          *bool  `yaml:"enabled,omitempty"`
	Provider         string `yaml:"provider"`
	Model            string `yaml:"model"`
	Agent            string `yaml:"agent"`
	ResponseLanguage string `yaml:"response_language,omitempty"`
	DraftingMode     string `yaml:"drafting_mode,omitempty"`
}

type hydraConfigLayer struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

func mergeRuneLayer(dst *RuneFile, src runeLayer) {
	if src.Version != "" {
		dst.Version = src.Version
	}
	if src.Name != "" {
		dst.Name = src.Name
	}
	if src.Mode != "" {
		dst.Mode = src.Mode
	}
	mergePreferences(&dst.Preferences, src.Preferences)
	mergeUI(&dst.UI, src.UI)
	mergePlanning(&dst.Planning, src.Planning)
	mergeOverlay(&dst.Overlay, src.Overlay)
	mergeOpenCode(&dst.OpenCode, src.OpenCode)
	mergeTooling(&dst.Tooling, src.Tooling)
	if src.Hydra.Enabled != nil {
		dst.Hydra.Enabled = *src.Hydra.Enabled
	}
}

func mergePreferences(dst *PreferencesConfig, src PreferencesConfig) {
	if src.Language != "" {
		dst.Language = src.Language
	}
	if src.Verbosity != "" {
		dst.Verbosity = src.Verbosity
	}
}

func mergeUI(dst *UIConfig, src UIConfig) {
	if src.Theme != "" {
		dst.Theme = src.Theme
	}
}

func mergePlanning(dst *PlanningConfig, src PlanningConfig) {
	if src.Mode != "" {
		dst.Mode = src.Mode
	}
	if src.Path != "" {
		dst.Path = src.Path
	}
	if src.Branch != "" {
		dst.Branch = src.Branch
	}
	if src.DefaultSurface != "" {
		dst.DefaultSurface = src.DefaultSurface
	}
}

func mergeOverlay(dst *OverlayConfig, src OverlayConfig) {
	if src.Path != "" {
		dst.Path = src.Path
	}
}

func mergeOpenCode(dst *OpenCodeConfig, src openCodeConfigLayer) {
	if src.Enabled != nil {
		dst.Enabled = *src.Enabled
	}
	if src.Provider != "" {
		dst.Provider = src.Provider
	}
	if src.Model != "" {
		dst.Model = src.Model
	}
	if src.Agent != "" {
		dst.Agent = src.Agent
	}
	if src.ResponseLanguage != "" {
		dst.ResponseLanguage = src.ResponseLanguage
	}
	if src.DraftingMode != "" {
		dst.DraftingMode = src.DraftingMode
	}
}

func mergeTooling(dst *ToolingConfig, src ToolingConfig) {
	dst.Languages = append(dst.Languages, src.Languages...)
	dst.Tools = append(dst.Tools, src.Tools...)
	dst.Frameworks = append(dst.Frameworks, src.Frameworks...)
}

func globalConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, globalRunesPath)
}

func hasHydra(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".hydra.yaml"))
	return err == nil
}

func dedupeSorted(items []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func SharedPlanningRoot(w Workspace) string {
	return filepath.Join(w.Root, w.Runes.Planning.Path)
}

func OverlayPlanningRoot(w Workspace) string {
	return filepath.Join(w.Root, w.Runes.Overlay.Path)
}

func CommandsRoot(w Workspace) string {
	return filepath.Join(w.Root, ".norn", "commands")
}

func SkillsRoot(w Workspace) string {
	return filepath.Join(w.Root, ".norn", "skills")
}

func FatesRoot(w Workspace) string {
	return filepath.Join(w.Root, ".norn", "fates")
}

func OpenCodeAgentsRoot(w Workspace) string {
	return filepath.Join(w.Root, ".opencode", "agents")
}

func SpindleRoot(w Workspace) string {
	return filepath.Join(w.Root, ".norn", "spindle")
}

func CountFiles(root string, suffix string) int {
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), suffix) {
			count++
		}
	}
	return count
}
