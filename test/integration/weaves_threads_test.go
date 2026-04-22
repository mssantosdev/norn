package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mssantosdev/norn/internal/cli"
	"github.com/mssantosdev/norn/internal/threads"
)

func TestWeavesCRUD(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=weaves-test"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"weaves", "add", "Planning Surface", "Define planning artifacts"}); err != nil {
		t.Fatalf("weaves add failed: %v", err)
	}
	for _, path := range []string{
		filepath.Join(root, ".norn", "weaves", "planning-surface", "README.md"),
		filepath.Join(root, ".norn", "weaves", "planning-surface", "threads.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}
	if err := cli.Run([]string{"weaves", "show", "planning-surface"}); err != nil {
		t.Fatalf("weaves show failed: %v", err)
	}
	if err := cli.Run([]string{"weaves", "remove", "planning-surface"}); err != nil {
		t.Fatalf("weaves remove failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".norn", "weaves", "planning-surface")); !os.IsNotExist(err) {
		t.Fatalf("expected weave to be removed, stat err=%v", err)
	}
}

func TestThreadsCRUD(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=threads-test"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"weaves", "add", "Planning Surface", "Define planning artifacts"}); err != nil {
		t.Fatalf("weaves add failed: %v", err)
	}
	if err := cli.Run([]string{"threads", "add", "planning-surface", "Add Weaves CLI", "Implement the weave command surface"}); err != nil {
		t.Fatalf("threads add failed: %v", err)
	}
	threadPath := filepath.Join(root, ".norn", "weaves", "planning-surface", "threads", "add-weaves-cli.md")
	if _, err := os.Stat(threadPath); err != nil {
		t.Fatalf("expected thread file: %v", err)
	}
	if err := cli.Run([]string{"threads", "show", "planning-surface", "add-weaves-cli"}); err != nil {
		t.Fatalf("threads show failed: %v", err)
	}
	if err := cli.Run([]string{"threads", "remove", "planning-surface", "add-weaves-cli"}); err != nil {
		t.Fatalf("threads remove failed: %v", err)
	}
	if _, err := os.Stat(threadPath); !os.IsNotExist(err) {
		t.Fatalf("expected thread to be removed, stat err=%v", err)
	}
}

func TestReadThread(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=read-thread-test"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"weaves", "add", "Planning Surface", "Test weave"}); err != nil {
		t.Fatalf("weaves add failed: %v", err)
	}
	threadPath := filepath.Join(root, ".norn", "weaves", "planning-surface", "threads", "test-thread.md")
	if err := os.MkdirAll(filepath.Dir(threadPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(threadPath, []byte("---\ntitle: Test Thread\nsummary: Test summary\nweave: planning-surface\n---\n\ntest body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := threads.Load(filepath.Join(root, ".norn"), "planning-surface", "test-thread")
	if err != nil {
		t.Fatalf("load thread failed: %v", err)
	}
	if doc.Summary != "Test summary" || doc.Body != "test body" {
		t.Fatalf("expected thread content, got summary=%q body=%q", doc.Summary, doc.Body)
	}
}

func TestWeavesAddSimple(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=simple-weave-test"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"weaves", "add", "Simple Planning", "Simple weave"}); err != nil {
		t.Fatalf("weaves add failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".norn", "weaves", "simple-planning", "README.md")); err != nil {
		t.Fatalf("expected weave file: %v", err)
	}
}

func TestWeavesShowWithoutID(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=show-test"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	// Without artifacts, show should return error (can't prompt in test)
	err := cli.Run([]string{"weaves", "show"})
	if err == nil {
		t.Fatal("expected error when no weaves exist")
	}
	if !contains(err.Error(), "no weaves available") {
		t.Fatalf("expected 'no weaves available' error, got: %v", err)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
