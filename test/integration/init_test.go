package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mssantosdev/norn/internal/cli"
	"github.com/mssantosdev/norn/internal/norn"
)

func TestInitNonInteractiveFolderMode(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=test-project", "--enable-opencode"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	w, err := norn.Load(root)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if w.Runes.Name != "test-project" {
		t.Fatalf("expected workspace name test-project, got %s", w.Runes.Name)
	}
	if w.Runes.Planning.Path != ".norn" {
		t.Fatalf("expected planning path loom, got %s", w.Runes.Planning.Path)
	}
	for _, path := range []string{".norn/weaves", ".norn/patterns", ".norn/skills", ".norn/runes.yaml", ".norn/fates", ".norn/tools", ".opencode/agents"} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}
