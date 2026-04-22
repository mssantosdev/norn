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

func TestDetectMonoRepo(t *testing.T) {
	root := t.TempDir()

	// Create nested projects
	authDir := filepath.Join(root, "kyro-auth")
	coreDir := filepath.Join(root, "kyro-core")
	os.MkdirAll(authDir, 0o755)
	os.MkdirAll(coreDir, 0o755)

	// Auth project: TypeScript/Bun
	os.WriteFile(filepath.Join(authDir, "package.json"), []byte("{}"), 0o644)
	os.WriteFile(filepath.Join(authDir, "bun.lock"), []byte(""), 0o644)

	// Core project: Rust
	os.WriteFile(filepath.Join(coreDir, "Cargo.toml"), []byte("[package]"), 0o644)

	// Root: Makefile
	os.WriteFile(filepath.Join(root, "Makefile"), []byte("test"), 0o644)

	detected, err := detect.Scan(root)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}

	// Should detect all languages
	assertContains(t, detected.Languages, "node")
	assertContains(t, detected.Languages, "bun")
	assertContains(t, detected.Languages, "rust")

	// Should detect root tools
	assertContains(t, detected.Tools, "make")

	// Should have both project locations
	if len(detected.Locations) < 2 {
		t.Fatalf("expected at least 2 locations, got %d: %v", len(detected.Locations), detected.Locations)
	}
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
