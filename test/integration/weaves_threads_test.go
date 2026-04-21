package integration

import (
	"os"
	"path/filepath"
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
	if err := cli.Run([]string{"init", "--no-interactive", "--name=weaves-test", "--mode=folder"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"weaves", "add", "Planning Surface", "Define planning artifacts"}); err != nil {
		t.Fatalf("weaves add failed: %v", err)
	}
	for _, path := range []string{
		filepath.Join(root, "loom", "weaves", "planning-surface", "README.md"),
		filepath.Join(root, "loom", "weaves", "planning-surface", "threads.md"),
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
	if _, err := os.Stat(filepath.Join(root, "loom", "weaves", "planning-surface")); !os.IsNotExist(err) {
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
	if err := cli.Run([]string{"init", "--no-interactive", "--name=threads-test", "--mode=folder"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"weaves", "add", "Planning Surface", "Define planning artifacts"}); err != nil {
		t.Fatalf("weaves add failed: %v", err)
	}
	if err := cli.Run([]string{"threads", "add", "planning-surface", "Add Weaves CLI", "Implement the weave command surface"}); err != nil {
		t.Fatalf("threads add failed: %v", err)
	}
	threadPath := filepath.Join(root, "loom", "weaves", "planning-surface", "threads", "add-weaves-cli.md")
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

func TestOverlayArtifactsOverrideSharedOnRead(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=overlay-test", "--mode=folder"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"weaves", "add", "Planning Surface", "Shared planning artifacts"}); err != nil {
		t.Fatalf("weaves add failed: %v", err)
	}
	sharedThread := filepath.Join(root, "loom", "weaves", "planning-surface", "threads", "add-weaves-cli.md")
	if err := os.MkdirAll(filepath.Dir(sharedThread), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sharedThread, []byte("---\ntitle: Add Weaves CLI\nsummary: Shared version\nweave: planning-surface\n---\n\nshared body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	overlayThread := filepath.Join(root, ".norn", "loom", "weaves", "planning-surface", "threads", "add-weaves-cli.md")
	if err := os.MkdirAll(filepath.Dir(overlayThread), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(overlayThread, []byte("---\ntitle: Add Weaves CLI\nsummary: Local overlay version\nweave: planning-surface\n---\n\noverlay body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := threads.LoadMerged(filepath.Join(root, "loom"), filepath.Join(root, ".norn", "loom"), "planning-surface", "add-weaves-cli")
	if err != nil {
		t.Fatalf("load merged thread failed: %v", err)
	}
	if doc.Summary != "Local overlay version" || doc.Body != "overlay body" {
		t.Fatalf("expected overlay thread to win, got summary=%q body=%q", doc.Summary, doc.Body)
	}
	items, err := threads.ListMerged(filepath.Join(root, "loom"), filepath.Join(root, ".norn", "loom"), "planning-surface")
	if err != nil {
		t.Fatalf("list merged threads failed: %v", err)
	}
	if len(items) != 1 || items[0].Summary != "Local overlay version" {
		t.Fatalf("expected merged thread list to prefer overlay, got %#v", items)
	}
}

func TestWeavesAddToLocalSurface(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=local-weave-test", "--mode=folder"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"weaves", "add", "--surface=local", "Local Planning", "Local only weave"}); err != nil {
		t.Fatalf("weaves add local failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".norn", "loom", "weaves", "local-planning", "README.md")); err != nil {
		t.Fatalf("expected local weave file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "loom", "weaves", "local-planning", "README.md")); !os.IsNotExist(err) {
		t.Fatalf("expected no shared weave file, stat err=%v", err)
	}
}

func TestThreadsAddToBothSurfaces(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=both-surface-test", "--mode=folder"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"weaves", "add", "--surface=both", "Planning Surface", "Shared and local weave"}); err != nil {
		t.Fatalf("weaves add both failed: %v", err)
	}
	if err := cli.Run([]string{"threads", "add", "--surface=both", "planning-surface", "Add Weaves CLI", "Implement both-surface thread"}); err != nil {
		t.Fatalf("threads add both failed: %v", err)
	}
	for _, path := range []string{
		filepath.Join(root, "loom", "weaves", "planning-surface", "threads", "add-weaves-cli.md"),
		filepath.Join(root, ".norn", "loom", "weaves", "planning-surface", "threads", "add-weaves-cli.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected thread file at %s: %v", path, err)
		}
	}
}
