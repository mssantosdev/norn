package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mssantosdev/norn/internal/cli"
)

func TestInitGeneratesOpenCodeFates(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/test\n\ngo 1.24.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=fate-test"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".opencode", "agents", "weaver.md"))
	if err != nil {
		t.Fatalf("read weaver agent failed: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "go build*") || !strings.Contains(text, "go test*") {
		t.Fatalf("expected go permissions in generated weaver agent, got:\n%s", text)
	}
}
