package loom

import (
	"os"
	"path/filepath"

	"github.com/mssantosdev/norn/internal/norn"
	"github.com/mssantosdev/norn/internal/patterns"
	"github.com/mssantosdev/norn/internal/skills"
)

func Ensure(root string, options norn.InitOptions) (string, error) {
	path := options.PlanningPath
	if path == "" {
		path = ".norn"
	}
	fullPath := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Join(fullPath, "weaves"), 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(fullPath, "patterns"), 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(fullPath, "skills"), 0o755); err != nil {
		return "", err
	}
	if options.Skeleton != "empty" {
		if err := seedSkeleton(fullPath, options); err != nil {
			return "", err
		}
	}
	return path, nil
}

func seedSkeleton(root string, options norn.InitOptions) error {
	readmePath := filepath.Join(root, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		readme := "# Loom\n\nThis directory contains shared planning artifacts for Norn.\n"
		if err := os.WriteFile(readmePath, []byte(readme), 0o644); err != nil {
			return err
		}
	}
	constitutionPath := filepath.Join(root, "constitution.md")
	if _, err := os.Stat(constitutionPath); err != nil {
		body := "# Constitution\n\n## Mission\n\nUse Norn to coordinate work through shared plans, local runtime state, and explicit handoffs.\n"
		if err := os.WriteFile(constitutionPath, []byte(body), 0o644); err != nil {
			return err
		}
	}
	if options.Skeleton == "guided" && options.Name != "" {
		if err := patterns.Save(filepath.Join(root, "patterns"), norn.Document{ID: "initial-architecture", Title: "Initial Architecture", Summary: "Seed architecture notes", Body: "Document the main components, interfaces, and constraints for the project."}); err != nil {
			return err
		}
		if err := skills.Save(filepath.Join(root, "skills"), norn.Document{ID: "project-context", Title: "Project Context", Summary: "Starter shared skill", Body: "Capture shared knowledge, review habits, and repo-specific rules here."}); err != nil {
			return err
		}
	}
	return nil
}
