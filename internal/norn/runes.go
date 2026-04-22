package norn

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type RuneScope string

const (
	RuneScopeGlobal    RuneScope = "global"
	RuneScopeWorkspace RuneScope = "workspace"
	RuneScopeLocal     RuneScope = "local"
)

type RuneResolution struct {
	Effective RuneFile             `yaml:"effective"`
	Origins   map[string][]string  `yaml:"origins"`
	Scopes    map[RuneScope]string `yaml:"scopes"`
}

var resolvedRunePaths = []string{
	"version",
	"name",
	"mode",
	"preferences.language",
	"preferences.verbosity",
	"ui.theme",
	"planning.mode",
	"planning.path",
	"planning.branch",
	"planning.default_surface",
	"overlay.path",
	"opencode.enabled",
	"opencode.provider",
	"opencode.model",
	"opencode.agent",
	"opencode.response_language",
	"opencode.drafting_mode",
	"tooling.languages",
	"tooling.tools",
	"tooling.frameworks",
	"hydra.enabled",
}

var editableRunePaths = map[string]string{
	"name":                       "string",
	"preferences.language":       "string",
	"preferences.verbosity":      "enum",
	"ui.theme":                   "enum",
	"planning.mode":              "enum",
	"planning.path":              "string",
	"planning.branch":            "string",
	"planning.default_surface":   "enum",
	"overlay.path":               "string",
	"opencode.enabled":           "bool",
	"opencode.provider":          "string",
	"opencode.model":             "string",
	"opencode.agent":             "string",
	"opencode.response_language": "string",
	"opencode.drafting_mode":     "enum",
	"tooling.languages":          "list",
	"tooling.tools":              "list",
	"tooling.frameworks":         "list",
	"hydra.enabled":              "bool",
}

var runePathEnums = map[string][]string{
	"preferences.verbosity":  {"quiet", "normal", "loud"},
	"ui.theme":               {"tokyonight", "catppuccin", "dracula", "nord", "onedark"},
	"opencode.drafting_mode": {"ask", "auto"},
}

func ParseRuneScope(value string) (RuneScope, error) {
	scope := RuneScope(strings.TrimSpace(value))
	if !scope.Valid() {
		return "", fmt.Errorf("invalid scope %q; expected global, workspace, or local", value)
	}
	return scope, nil
}

func (s RuneScope) Valid() bool {
	switch s {
	case RuneScopeGlobal, RuneScopeWorkspace, RuneScopeLocal:
		return true
	default:
		return false
	}
}

func ScopePath(root string, scope RuneScope) (string, error) {
	switch scope {
	case RuneScopeGlobal:
		path := globalConfigPath()
		if path == "" {
			return "", fmt.Errorf("global config path unavailable")
		}
		return path, nil
	case RuneScopeWorkspace:
		if strings.TrimSpace(root) == "" {
			return "", fmt.Errorf("workspace root required for workspace scope")
		}
		return filepath.Join(root, runesPath), nil
	case RuneScopeLocal:
		if strings.TrimSpace(root) == "" {
			return "", fmt.Errorf("workspace root required for local scope")
		}
		return filepath.Join(root, localOverridePath), nil
	default:
		return "", fmt.Errorf("unsupported scope %q", scope)
	}
}

func ScopeDisplayPath(root string, scope RuneScope) string {
	switch scope {
	case RuneScopeGlobal:
		return "~/.config/norn/runes.yaml"
	case RuneScopeWorkspace:
		return ".norn/runes.yaml"
	case RuneScopeLocal:
		return ".norn/runes.local.yaml"
	default:
		if path, err := ScopePath(root, scope); err == nil {
			return path
		}
		return string(scope)
	}
}

func LoadScopeMap(root string, scope RuneScope) (map[string]any, error) {
	path, err := ScopePath(root, scope)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("read %s scope: %w", scope, err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return map[string]any{}, nil
	}
	var layer map[string]any
	if err := yaml.Unmarshal(data, &layer); err != nil {
		return nil, fmt.Errorf("parse %s scope: %w", scope, err)
	}
	if layer == nil {
		return map[string]any{}, nil
	}
	return normalizeMap(layer), nil
}

func SaveScopeMap(root string, scope RuneScope, layer map[string]any) error {
	path, err := ScopePath(root, scope)
	if err != nil {
		return err
	}
	layer = normalizeMap(layer)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if len(layer) == 0 && scope != RuneScopeWorkspace {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	data, err := yaml.Marshal(layer)
	if err != nil {
		return fmt.Errorf("marshal %s scope: %w", scope, err)
	}
	if len(layer) == 0 && scope == RuneScopeWorkspace {
		data = []byte("{}\n")
	}
	return os.WriteFile(path, data, 0o644)
}

func MarshalScopeMap(layer map[string]any) ([]byte, error) {
	layer = normalizeMap(layer)
	if len(layer) == 0 {
		return []byte("{}\n"), nil
	}
	return yaml.Marshal(layer)
}

func ResolveRunes(root string) (RuneResolution, error) {
	runes, err := loadLayeredRunes(root)
	if err != nil {
		return RuneResolution{}, err
	}
	applyDefaults(&runes, root)
	layers, err := loadScopeMaps(root)
	if err != nil {
		return RuneResolution{}, err
	}
	return RuneResolution{
		Effective: runes,
		Origins:   buildRuneOrigins(root, layers, runes),
		Scopes: map[RuneScope]string{
			RuneScopeGlobal:    ScopeDisplayPath(root, RuneScopeGlobal),
			RuneScopeWorkspace: ScopeDisplayPath(root, RuneScopeWorkspace),
			RuneScopeLocal:     ScopeDisplayPath(root, RuneScopeLocal),
		},
	}, nil
}

func EditableRunePaths() []string {
	out := make([]string, 0, len(editableRunePaths))
	for path := range editableRunePaths {
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

func ResolvedRunePaths() []string {
	out := make([]string, len(resolvedRunePaths))
	copy(out, resolvedRunePaths)
	return out
}

func ValidateRuneEditPath(path string) error {
	kind, ok := editableRunePaths[path]
	if !ok {
		return fmt.Errorf("unsupported config path %q", path)
	}
	if kind == "enum" {
		return nil
	}
	return nil
}

func ParseRuneEditValue(path, raw string) (any, error) {
	kind, ok := editableRunePaths[path]
	if !ok {
		return nil, fmt.Errorf("unsupported config path %q", path)
	}
	raw = strings.TrimSpace(raw)
	switch kind {
	case "string":
		if raw == "" {
			return nil, fmt.Errorf("empty value is not allowed for %s; use --unset", path)
		}
		return raw, nil
	case "bool":
		switch raw {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid boolean %q for %s; expected true or false", raw, path)
		}
	case "list":
		values := parseCSV(raw)
		if len(values) == 0 {
			return nil, fmt.Errorf("empty list is not allowed for %s; use --unset", path)
		}
		return values, nil
	case "enum":
		for _, value := range runePathEnums[path] {
			if raw == value {
				return raw, nil
			}
		}
		return nil, fmt.Errorf("invalid value %q for %s; expected one of %s", raw, path, strings.Join(runePathEnums[path], ", "))
	default:
		return nil, fmt.Errorf("unsupported config path %q", path)
	}
}

func FieldValueString(value any) string {
	switch typed := value.(type) {
	case nil:
		return "<unset>"
	case []string:
		return strings.Join(typed, ", ")
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			parts = append(parts, fmt.Sprint(item))
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprint(typed)
	}
}

func ResolvedFieldValue(r RuneResolution, path string) any {
	switch path {
	case "version":
		return r.Effective.Version
	case "name":
		return r.Effective.Name
	case "mode":
		return string(r.Effective.Mode)
	case "preferences.language":
		return r.Effective.Preferences.Language
	case "preferences.verbosity":
		return r.Effective.Preferences.Verbosity
	case "ui.theme":
		return r.Effective.UI.Theme
	case "planning.path":
		return r.Effective.Planning.Path
	case "opencode.enabled":
		return r.Effective.OpenCode.Enabled
	case "opencode.provider":
		return r.Effective.OpenCode.Provider
	case "opencode.model":
		return r.Effective.OpenCode.Model
	case "opencode.agent":
		return r.Effective.OpenCode.Agent
	case "opencode.response_language":
		return r.Effective.OpenCode.ResponseLanguage
	case "opencode.drafting_mode":
		return r.Effective.OpenCode.DraftingMode
	case "tooling.languages":
		return r.Effective.Tooling.Languages
	case "tooling.tools":
		return r.Effective.Tooling.Tools
	case "tooling.frameworks":
		return r.Effective.Tooling.Frameworks
	case "hydra.enabled":
		return r.Effective.Hydra.Enabled
	default:
		return nil
	}
}

func loadScopeMaps(root string) (map[RuneScope]map[string]any, error) {
	layers := map[RuneScope]map[string]any{}
	for _, scope := range []RuneScope{RuneScopeGlobal, RuneScopeWorkspace, RuneScopeLocal} {
		layer, err := LoadScopeMap(root, scope)
		if err != nil {
			if scope != RuneScopeGlobal && strings.Contains(err.Error(), "workspace root required") {
				layer = map[string]any{}
			} else {
				return nil, err
			}
		}
		layers[scope] = layer
	}
	return layers, nil
}

func buildRuneOrigins(root string, layers map[RuneScope]map[string]any, runes RuneFile) map[string][]string {
	origins := map[string][]string{}
	for _, path := range resolvedRunePaths {
		origins[path] = fieldOrigins(root, layers, runes, path, origins)
	}
	return origins
}

func fieldOrigins(root string, layers map[RuneScope]map[string]any, runes RuneFile, path string, known map[string][]string) []string {
	if path == "hydra.enabled" && runes.Hydra.Enabled && hasHydra(root) {
		return []string{"workspace detection (.hydra.yaml)"}
	}
	if path == "mode" {
		if origin := lastExplicitOrigin(root, layers, path); len(origin) > 0 {
			return origin
		}
		if hasHydra(root) {
			return []string{"workspace detection (.hydra.yaml)"}
		}
		return []string{"built-in default"}
	}
	if path == "name" {
		if origin := lastExplicitOrigin(root, layers, path); len(origin) > 0 {
			return origin
		}
		return []string{"derived from workspace root name"}
	}
	if path == "opencode.response_language" {
		if origin := lastExplicitOrigin(root, layers, path); len(origin) > 0 {
			return origin
		}
		if prefOrigins := known["preferences.language"]; len(prefOrigins) > 0 {
			return []string{fmt.Sprintf("derived from preferences.language [%s]", strings.Join(prefOrigins, ", "))}
		}
	}
	if strings.HasPrefix(path, "tooling.") {
		if origins := listOrigins(root, layers, path); len(origins) > 0 {
			return origins
		}
		return []string{"built-in default"}
	}
	if origin := lastExplicitOrigin(root, layers, path); len(origin) > 0 {
		return origin
	}
	return []string{"built-in default"}
}

func lastExplicitOrigin(root string, layers map[RuneScope]map[string]any, path string) []string {
	for _, scope := range []RuneScope{RuneScopeLocal, RuneScopeWorkspace, RuneScopeGlobal} {
		if _, ok := lookupPath(layers[scope], path); ok {
			return []string{fmt.Sprintf("%s (%s)", scope, ScopeDisplayPath(root, scope))}
		}
	}
	return nil
}

func listOrigins(root string, layers map[RuneScope]map[string]any, path string) []string {
	origins := []string{}
	for _, scope := range []RuneScope{RuneScopeGlobal, RuneScopeWorkspace, RuneScopeLocal} {
		value, ok := lookupPath(layers[scope], path)
		if !ok {
			continue
		}
		items := asStringSlice(value)
		if len(items) == 0 {
			continue
		}
		origins = append(origins, fmt.Sprintf("%s (%s)", scope, ScopeDisplayPath(root, scope)))
	}
	return origins
}

func normalizeMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	out := map[string]any{}
	for key, value := range input {
		switch typed := value.(type) {
		case map[string]any:
			normalized := normalizeMap(typed)
			if len(normalized) > 0 {
				out[key] = normalized
			}
		case map[any]any:
			child := map[string]any{}
			for childKey, childValue := range typed {
				child[fmt.Sprint(childKey)] = childValue
			}
			normalized := normalizeMap(child)
			if len(normalized) > 0 {
				out[key] = normalized
			}
		case []any:
			if len(typed) > 0 {
				out[key] = typed
			}
		case []string:
			if len(typed) > 0 {
				out[key] = typed
			}
		case nil:
			continue
		default:
			out[key] = typed
		}
	}
	return out
}

func lookupPath(layer map[string]any, path string) (any, bool) {
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

func SetScopeValue(layer map[string]any, path string, value any) map[string]any {
	if layer == nil {
		layer = map[string]any{}
	}
	parts := strings.Split(path, ".")
	current := layer
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return layer
		}
		next, ok := current[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[part] = next
		}
		current = next
	}
	return layer
}

func UnsetScopeValue(layer map[string]any, path string) map[string]any {
	if layer == nil {
		return map[string]any{}
	}
	unsetScopeValue(layer, strings.Split(path, "."))
	return normalizeMap(layer)
}

func unsetScopeValue(layer map[string]any, parts []string) bool {
	if len(parts) == 0 {
		return len(layer) == 0
	}
	part := parts[0]
	if len(parts) == 1 {
		delete(layer, part)
		return len(layer) == 0
	}
	next, ok := layer[part].(map[string]any)
	if !ok {
		return len(layer) == 0
	}
	if unsetScopeValue(next, parts[1:]) {
		delete(layer, part)
	}
	return len(layer) == 0
}

func asStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return dedupeSorted(typed)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			out = append(out, fmt.Sprint(item))
		}
		return dedupeSorted(out)
	default:
		return nil
	}
}

func parseCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return dedupeSorted(parts)
}
