package weaves

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mssantosdev/norn/internal/norn"
	"gopkg.in/yaml.v3"
)

type frontmatter struct {
	Title   string   `yaml:"title"`
	Summary string   `yaml:"summary,omitempty"`
	Fizzy   int      `yaml:"fizzy,omitempty"`
	Tags    []string `yaml:"tags,omitempty"`
}

func Root(sharedPlanningRoot string) string {
	return filepath.Join(sharedPlanningRoot, "weaves")
}

func WeaveRoot(sharedPlanningRoot, id string) string {
	return filepath.Join(Root(sharedPlanningRoot), id)
}

func ReadmePath(sharedPlanningRoot, id string) string {
	return filepath.Join(WeaveRoot(sharedPlanningRoot, id), "README.md")
}

func ThreadsLedgerPath(sharedPlanningRoot, id string) string {
	return filepath.Join(WeaveRoot(sharedPlanningRoot, id), "threads.md")
}

func ThreadsRoot(sharedPlanningRoot, id string) string {
	return filepath.Join(WeaveRoot(sharedPlanningRoot, id), "threads")
}

func Save(sharedPlanningRoot string, doc norn.Document) error {
	return SaveToSurface(sharedPlanningRoot, doc)
}

func SaveToSurface(planningRoot string, doc norn.Document) error {
	if doc.ID == "" {
		doc.ID = slug(doc.Title)
	}
	if err := os.MkdirAll(ThreadsRoot(planningRoot, doc.ID), 0o755); err != nil {
		return err
	}
	fm, err := yaml.Marshal(frontmatter{Title: doc.Title, Summary: doc.Summary})
	if err != nil {
		return err
	}
	var body strings.Builder
	body.WriteString("---\n")
	body.Write(fm)
	body.WriteString("---\n\n")
	body.WriteString(strings.TrimSpace(doc.Body))
	body.WriteString("\n")
	if err := os.WriteFile(ReadmePath(planningRoot, doc.ID), []byte(body.String()), 0o644); err != nil {
		return err
	}
	if _, err := os.Stat(ThreadsLedgerPath(planningRoot, doc.ID)); err != nil {
		ledger := "# Threads\n\nActive threads for this weave:\n\n"
		if err := os.WriteFile(ThreadsLedgerPath(planningRoot, doc.ID), []byte(ledger), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func Load(sharedPlanningRoot, id string) (norn.Document, error) {
	data, err := os.ReadFile(ReadmePath(sharedPlanningRoot, id))
	if err != nil {
		return norn.Document{}, err
	}
	parts := strings.SplitN(string(data), "---\n", 3)
	if len(parts) < 3 {
		return norn.Document{ID: id, Title: id, Body: string(data)}, nil
	}
	var fm frontmatter
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return norn.Document{}, err
	}
	return norn.Document{ID: id, Title: fm.Title, Summary: fm.Summary, Body: strings.TrimSpace(parts[2])}, nil
}

func List(sharedPlanningRoot string) ([]norn.Document, error) {
	entries, err := os.ReadDir(Root(sharedPlanningRoot))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []norn.Document
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		doc, err := Load(sharedPlanningRoot, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("load weave %s: %w", entry.Name(), err)
		}
		out = append(out, doc)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func ListMerged(sharedPlanningRoot, overlayPlanningRoot string) ([]norn.Document, error) {
	merged := map[string]norn.Document{}
	shared, err := List(sharedPlanningRoot)
	if err != nil {
		return nil, err
	}
	for _, item := range shared {
		merged[item.ID] = item
	}
	overlay, err := List(overlayPlanningRoot)
	if err != nil {
		return nil, err
	}
	for _, item := range overlay {
		merged[item.ID] = item
	}
	out := make([]norn.Document, 0, len(merged))
	for _, item := range merged {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func LoadMerged(sharedPlanningRoot, overlayPlanningRoot, id string) (norn.Document, error) {
	if doc, err := Load(overlayPlanningRoot, id); err == nil {
		return doc, nil
	}
	return Load(sharedPlanningRoot, id)
}

func Delete(sharedPlanningRoot, id string) error {
	return os.RemoveAll(WeaveRoot(sharedPlanningRoot, id))
}

func DefaultBody(title, summary string) string {
	return strings.TrimSpace(fmt.Sprintf(`## Goal

%s

## User Stories

- As a user, I want to manage this weave through the CLI so the plan structure is created consistently.

## Scope

- define the initial scope for this weave

## Acceptance

- define the acceptance criteria for this weave

## Notes

- replace this template content with project-specific details
`, summary))
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "_", "-", ":", "-")
	return replacer.Replace(value)
}
