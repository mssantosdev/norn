package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mssantosdev/norn/internal/cli"
)

func TestBasicWorkflow(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=e2e-project", "--mode=folder"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"patterns", "add", "API Contract", "Shared API expectations"}); err != nil {
		t.Fatalf("pattern add failed: %v", err)
	}
	if err := cli.Run([]string{"skills", "add", "Deploy Flow", "How deployments work"}); err != nil {
		t.Fatalf("skill add failed: %v", err)
	}
	if err := cli.Run([]string{"commands", "add", "lint", "lint", "npm run lint"}); err != nil {
		t.Fatalf("command add failed: %v", err)
	}
	if err := cli.Run([]string{"weaves", "add", "Planning Surface", "Define planning artifacts"}); err != nil {
		t.Fatalf("weave add failed: %v", err)
	}
	if err := cli.Run([]string{"threads", "add", "planning-surface", "Add Weaves CLI", "Implement the weave command surface"}); err != nil {
		t.Fatalf("thread add failed: %v", err)
	}
	for _, path := range []string{
		filepath.Join(root, "loom", "patterns", "api-contract.md"),
		filepath.Join(root, "loom", "skills", "deploy-flow.md"),
		filepath.Join(root, ".norn", "commands", "lint.yaml"),
		filepath.Join(root, "loom", "weaves", "planning-surface", "README.md"),
		filepath.Join(root, "loom", "weaves", "planning-surface", "threads", "add-weaves-cli.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected artifact %s: %v", path, err)
		}
	}
}
