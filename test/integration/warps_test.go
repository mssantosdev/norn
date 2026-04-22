package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mssantosdev/norn/internal/cli"
)

func TestWarpsCRUD(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=warps-test"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"warps", "add", "--status=active", "--owner=marcus", "--root=./worktrees/api", "--branch=feature/api", "--weaves=planning-surface", "--threads=add-weaves-cli", "API Warp", "Runtime coordination for API lane"}); err != nil {
		t.Fatalf("warps add failed: %v", err)
	}
	warpPath := filepath.Join(root, ".norn", "spindle", "warps", "api-warp.yaml")
	if _, err := os.Stat(warpPath); err != nil {
		t.Fatalf("expected warp file: %v", err)
	}
	showOutput := captureStdout(t, func() error {
		return cli.Run([]string{"warps", "show", "api-warp"})
	})
	if !strings.Contains(showOutput, "title: API Warp") || !strings.Contains(showOutput, "status: active") {
		t.Fatalf("expected warp show output, got:\n%s", showOutput)
	}
	listOutput := captureStdout(t, func() error {
		return cli.Run([]string{"warps", "list"})
	})
	if !strings.Contains(listOutput, "api-warp") || !strings.Contains(listOutput, "API Warp [active]") {
		t.Fatalf("expected warp list output, got:\n%s", listOutput)
	}
	if err := cli.Run([]string{"warps", "remove", "api-warp"}); err != nil {
		t.Fatalf("warps remove failed: %v", err)
	}
	if _, err := os.Stat(warpPath); !os.IsNotExist(err) {
		t.Fatalf("expected warp to be removed, stat err=%v", err)
	}
}

func TestWarpRuntimeView(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=warp-runtime-test"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"warps", "add", "--status=active", "--owner=marcus", "API Warp", "Runtime coordination for API lane"}); err != nil {
		t.Fatalf("warps add failed: %v", err)
	}
	if err := cli.Run([]string{"warps", "assign", "--kind=weave", "--id=planning-surface", "--warp=api-warp", "--owner=marcus", "--state=review"}); err != nil {
		t.Fatalf("warp weave assignment failed: %v", err)
	}
	if err := cli.Run([]string{"warps", "assign", "--kind=thread", "--id=add-weaves-cli", "--warp=api-warp", "--owner=marcus", "--state=active"}); err != nil {
		t.Fatalf("warp thread assignment failed: %v", err)
	}
	output := captureStdout(t, func() error {
		return cli.Run([]string{"warps", "list", "--view=runtime"})
	})
	if !strings.Contains(output, "API Warp [active]") || !strings.Contains(output, "planning-surface") || !strings.Contains(output, "add-weaves-cli") {
		t.Fatalf("expected runtime warp view output, got:\n%s", output)
	}
	if !strings.Contains(output, "owner=marcus") || !strings.Contains(output, "state=review") {
		t.Fatalf("expected runtime assignment metadata, got:\n%s", output)
	}
}

func TestWarpAssignmentShowAndRemove(t *testing.T) {
	root := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init", "--no-interactive", "--name=warp-assignment-test"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := cli.Run([]string{"warps", "add", "API Warp", "Runtime coordination for API lane"}); err != nil {
		t.Fatalf("warps add failed: %v", err)
	}
	if err := cli.Run([]string{"warps", "assign", "--kind=thread", "--id=add-weaves-cli", "--warp=api-warp", "--owner=marcus", "--state=blocked", "--notes=waiting on review"}); err != nil {
		t.Fatalf("warp assignment failed: %v", err)
	}
	showOutput := captureStdout(t, func() error {
		return cli.Run([]string{"warps", "assignment", "show", "thread", "add-weaves-cli"})
	})
	if !strings.Contains(showOutput, "warp: api-warp") || !strings.Contains(showOutput, "state: blocked") {
		t.Fatalf("expected assignment show output, got:\n%s", showOutput)
	}
	assignmentPath := filepath.Join(root, ".norn", "spindle", "threads", "add-weaves-cli.yaml")
	if _, err := os.Stat(assignmentPath); err != nil {
		t.Fatalf("expected assignment file: %v", err)
	}
	if err := cli.Run([]string{"warps", "assignment", "remove", "thread", "add-weaves-cli"}); err != nil {
		t.Fatalf("assignment remove failed: %v", err)
	}
	if _, err := os.Stat(assignmentPath); !os.IsNotExist(err) {
		t.Fatalf("expected assignment to be removed, stat err=%v", err)
	}
}
