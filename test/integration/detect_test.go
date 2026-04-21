package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mssantosdev/norn/internal/detect"
)

func TestDetectGoAndMake(t *testing.T) {
	root := t.TempDir()
	for _, file := range []string{"go.mod", "Makefile", "Dockerfile"} {
		if err := os.WriteFile(filepath.Join(root, file), []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	detected, err := detect.Scan(root)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}
	assertContains(t, detected.Languages, "go")
	assertContains(t, detected.Tools, "make")
	assertContains(t, detected.Tools, "docker")
}

func assertContains(t *testing.T, items []string, expected string) {
	t.Helper()
	for _, item := range items {
		if item == expected {
			return
		}
	}
	t.Fatalf("expected %q in %v", expected, items)
}
