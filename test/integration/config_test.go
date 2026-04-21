package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mssantosdev/norn/internal/cli"
	"github.com/mssantosdev/norn/internal/norn"
	uilogger "github.com/mssantosdev/norn/internal/ui/logger"
)

func TestLoadAppliesLayeredConfigPrecedence(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	workspace := filepath.Join(root, "workspace")
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()

	if err := os.MkdirAll(filepath.Join(home, ".config", "norn"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	globalRunes := `preferences:
  language: pt-BR
  verbosity: quiet
ui:
  theme: nord
planning:
  default_surface: local
opencode:
  enabled: true
  model: global-model
tooling:
  languages: [go]
  tools: [git]
`
	if err := os.WriteFile(filepath.Join(home, ".config", "norn", "runes.yaml"), []byte(globalRunes), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(workspace); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=config-test", "--mode=folder"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	workspaceRunes := `name: config-test
preferences:
  verbosity: loud
ui:
  theme: dracula
planning:
  default_surface: both
opencode:
  enabled: false
  provider: workspace-provider
tooling:
  languages: [rust]
  tools: [docker]
`
	if err := os.WriteFile(filepath.Join(workspace, ".norn", "runes.yaml"), []byte(workspaceRunes), 0o644); err != nil {
		t.Fatal(err)
	}

	localRunes := `preferences:
  language: es
opencode:
  drafting_mode: auto
tooling:
  languages: [go, node]
  tools: [git, make]
`
	if err := os.WriteFile(filepath.Join(workspace, ".norn", "runes.local.yaml"), []byte(localRunes), 0o644); err != nil {
		t.Fatal(err)
	}

	w, err := norn.Load(workspace)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if w.Runes.Preferences.Language != "es" {
		t.Fatalf("expected local language override, got %q", w.Runes.Preferences.Language)
	}
	if w.Runes.Preferences.Verbosity != "loud" {
		t.Fatalf("expected workspace verbosity override, got %q", w.Runes.Preferences.Verbosity)
	}
	if w.Runes.UI.Theme != "dracula" {
		t.Fatalf("expected workspace theme override, got %q", w.Runes.UI.Theme)
	}
	if w.Runes.Planning.DefaultSurface != "both" {
		t.Fatalf("expected workspace default surface override, got %q", w.Runes.Planning.DefaultSurface)
	}
	if w.Runes.OpenCode.Enabled {
		t.Fatalf("expected workspace layer to disable opencode")
	}
	if w.Runes.OpenCode.Provider != "workspace-provider" {
		t.Fatalf("expected workspace provider override, got %q", w.Runes.OpenCode.Provider)
	}
	if w.Runes.OpenCode.Model != "global-model" {
		t.Fatalf("expected global model fallback, got %q", w.Runes.OpenCode.Model)
	}
	if w.Runes.OpenCode.ResponseLanguage != "es" {
		t.Fatalf("expected response language to follow resolved language default, got %q", w.Runes.OpenCode.ResponseLanguage)
	}
	if w.Runes.OpenCode.DraftingMode != "auto" {
		t.Fatalf("expected local drafting mode override, got %q", w.Runes.OpenCode.DraftingMode)
	}
	if got, want := joinCSV(w.Runes.Tooling.Languages), "go,node,rust"; got != want {
		t.Fatalf("expected merged languages %q, got %q", want, got)
	}
	if got, want := joinCSV(w.Runes.Tooling.Tools), "docker,git,make"; got != want {
		t.Fatalf("expected merged tools %q, got %q", want, got)
	}
}

func TestLoadAppliesBuiltInDefaultsWithoutOptionalLayers(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	workspace := filepath.Join(root, "workspace")
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()

	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	if err := os.Chdir(workspace); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=defaults-test", "--mode=folder"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	minimalRunes := `name: defaults-test
planning:
  mode: folder
`
	if err := os.WriteFile(filepath.Join(workspace, ".norn", "runes.yaml"), []byte(minimalRunes), 0o644); err != nil {
		t.Fatal(err)
	}

	w, err := norn.Load(workspace)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if w.Runes.Preferences.Language != "en" {
		t.Fatalf("expected default language en, got %q", w.Runes.Preferences.Language)
	}
	if w.Runes.Preferences.Verbosity != "normal" {
		t.Fatalf("expected default verbosity normal, got %q", w.Runes.Preferences.Verbosity)
	}
	if w.Runes.UI.Theme != "tokyonight" {
		t.Fatalf("expected default theme tokyonight, got %q", w.Runes.UI.Theme)
	}
	if w.Runes.Planning.Path != "loom" {
		t.Fatalf("expected default planning path loom, got %q", w.Runes.Planning.Path)
	}
	if w.Runes.Planning.DefaultSurface != "shared" {
		t.Fatalf("expected default planning surface shared, got %q", w.Runes.Planning.DefaultSurface)
	}
	if w.Runes.OpenCode.ResponseLanguage != "en" {
		t.Fatalf("expected default response language en, got %q", w.Runes.OpenCode.ResponseLanguage)
	}
	if w.Runes.OpenCode.DraftingMode != "ask" {
		t.Fatalf("expected default drafting mode ask, got %q", w.Runes.OpenCode.DraftingMode)
	}
}

func TestRunesShowScopeAndResolve(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	workspace := filepath.Join(root, "workspace")
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()

	if err := os.MkdirAll(filepath.Join(home, ".config", "norn"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	if err := os.WriteFile(filepath.Join(home, ".config", "norn", "runes.yaml"), []byte("preferences:\n  language: pt-BR\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(workspace); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=runes-show-test", "--mode=folder"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	minimalWorkspaceRunes := `name: runes-show-test
planning:
  mode: folder
`
	if err := os.WriteFile(filepath.Join(workspace, ".norn", "runes.yaml"), []byte(minimalWorkspaceRunes), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, ".norn", "runes.local.yaml"), []byte("preferences:\n  language: es\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	showOutput := captureStdout(t, func() error {
		return cli.Run([]string{"runes", "show", "--scope=global"})
	})
	if !strings.Contains(showOutput, "language: pt-BR") {
		t.Fatalf("expected global scope output, got:\n%s", showOutput)
	}

	resolveOutput := captureStdout(t, func() error {
		return cli.Run([]string{"runes", "resolve"})
	})
	if !strings.Contains(resolveOutput, "Field") || !strings.Contains(resolveOutput, "preferences.language") || !strings.Contains(resolveOutput, "local (.norn/runes.local.yaml)") {
		t.Fatalf("expected resolve output to include full origin data, got:\n%s", resolveOutput)
	}
	if !strings.Contains(resolveOutput, "es") {
		t.Fatalf("expected resolved effective language, got:\n%s", resolveOutput)
	}
	if !strings.Contains(resolveOutput, "derived from preferences.language") {
		t.Fatalf("expected derived response language origin, got:\n%s", resolveOutput)
	}

	resolveYAML := captureStdout(t, func() error {
		return cli.Run([]string{"runes", "resolve", "--format=yaml"})
	})
	if !strings.Contains(resolveYAML, "effective:") || !strings.Contains(resolveYAML, "origins:") {
		t.Fatalf("expected yaml resolve output, got:\n%s", resolveYAML)
	}
}

func TestRunesEditSetAndUnsetByScope(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	workspace := filepath.Join(root, "workspace")
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()

	if err := os.MkdirAll(filepath.Join(home, ".config", "norn"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	if err := os.Chdir(workspace); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=runes-edit-test", "--mode=folder"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if err := cli.Run([]string{"runes", "edit", "--scope=global", "--set=preferences.language=pt-BR", "--set=opencode.enabled=true"}); err != nil {
		t.Fatalf("global edit failed: %v", err)
	}
	if err := cli.Run([]string{"runes", "edit", "--scope=workspace", "--set=preferences.language=en", "--set=planning.default_surface=both"}); err != nil {
		t.Fatalf("workspace edit failed: %v", err)
	}
	if err := cli.Run([]string{"runes", "edit", "--scope=workspace", "--unset=opencode.enabled"}); err != nil {
		t.Fatalf("workspace unset failed: %v", err)
	}
	if err := cli.Run([]string{"runes", "edit", "--scope=local", "--set=preferences.language=es", "--set=tooling.languages=go,node"}); err != nil {
		t.Fatalf("local edit failed: %v", err)
	}

	w, err := norn.Load(workspace)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if w.Runes.Preferences.Language != "es" {
		t.Fatalf("expected local override language, got %q", w.Runes.Preferences.Language)
	}
	if w.Runes.Planning.DefaultSurface != "both" {
		t.Fatalf("expected workspace default surface, got %q", w.Runes.Planning.DefaultSurface)
	}
	if !w.Runes.OpenCode.Enabled {
		t.Fatalf("expected global opencode enabled to apply")
	}
	if got, want := joinCSV(w.Runes.Tooling.Languages), "go,node"; got != want {
		t.Fatalf("expected local tooling languages %q, got %q", want, got)
	}

	if err := cli.Run([]string{"runes", "edit", "--scope=local", "--unset=preferences.language", "--unset=tooling.languages"}); err != nil {
		t.Fatalf("local unset failed: %v", err)
	}

	w, err = norn.Load(workspace)
	if err != nil {
		t.Fatalf("load after unset failed: %v", err)
	}
	if w.Runes.Preferences.Language != "en" {
		t.Fatalf("expected workspace language after unset, got %q", w.Runes.Preferences.Language)
	}
	if got, want := joinCSV(w.Runes.Tooling.Languages), ""; got != want {
		t.Fatalf("expected tooling languages to be unset, got %q", got)
	}

	localShow := captureStdout(t, func() error {
		return cli.Run([]string{"runes", "show", "--scope=local"})
	})
	if strings.Contains(localShow, "preferences:") || strings.Contains(localShow, "tooling:") {
		t.Fatalf("expected local file to no longer include removed fields, got:\n%s", localShow)
	}
}

func TestRunesResolveWorksOutsideWorkspace(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	standalone := filepath.Join(root, "standalone")
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()

	if err := os.MkdirAll(filepath.Join(home, ".config", "norn"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(standalone, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	globalRunes := `preferences:
  language: fr
opencode:
  enabled: true
`
	if err := os.WriteFile(filepath.Join(home, ".config", "norn", "runes.yaml"), []byte(globalRunes), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(standalone); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() error {
		return cli.Run([]string{"runes", "resolve"})
	})
	if !strings.Contains(output, "preferences.language") || !strings.Contains(output, "fr") {
		t.Fatalf("expected global-only resolve output, got:\n%s", output)
	}
	if !strings.Contains(output, "global (~/.config/norn/runes.yaml)") {
		t.Fatalf("expected global origin in resolve output, got:\n%s", output)
	}

	showOutput := captureStdout(t, func() error {
		return cli.Run([]string{"runes", "show", "--scope=global"})
	})
	if !strings.Contains(showOutput, "language: fr") {
		t.Fatalf("expected global show outside workspace, got:\n%s", showOutput)
	}
}

func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	uilogger.Logger().SetOutput(w)
	defer uilogger.Logger().SetOutput(os.Stderr)

	err = fn()
	_ = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, readErr := buf.ReadFrom(r); readErr != nil {
		t.Fatal(readErr)
	}
	return buf.String()
}

func joinCSV(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0] + mapJoin(values[1:])
}

func mapJoin(values []string) string {
	if len(values) == 0 {
		return ""
	}
	joined := ""
	for _, value := range values {
		joined += "," + value
	}
	return joined
}
