package threads

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mssantosdev/norn/internal/norn"
	"github.com/mssantosdev/norn/internal/weaves"
	"gopkg.in/yaml.v3"
)

type frontmatter struct {
	Title   string `yaml:"title"`
	Summary string `yaml:"summary,omitempty"`
	Weave   string `yaml:"weave"`
}

func Root(sharedPlanningRoot, weaveID string) string {
	return weaves.ThreadsRoot(sharedPlanningRoot, weaveID)
}

func Path(sharedPlanningRoot, weaveID, threadID string) string {
	return filepath.Join(Root(sharedPlanningRoot, weaveID), threadID+".md")
}

func Save(sharedPlanningRoot, weaveID string, doc norn.Document) error {
	return SaveToSurface(sharedPlanningRoot, weaveID, doc)
}

func SaveToSurface(planningRoot, weaveID string, doc norn.Document) error {
	if _, err := os.Stat(weaves.ReadmePath(planningRoot, weaveID)); err != nil {
		return fmt.Errorf("weave %s does not exist", weaveID)
	}
	if doc.ID == "" {
		doc.ID = slug(doc.Title)
	}
	if err := os.MkdirAll(Root(planningRoot, weaveID), 0o755); err != nil {
		return err
	}
	fm, err := yaml.Marshal(frontmatter{Title: doc.Title, Summary: doc.Summary, Weave: weaveID})
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fm)
	buf.WriteString("---\n\n")
	buf.WriteString(strings.TrimSpace(doc.Body))
	buf.WriteString("\n")
	if err := os.WriteFile(Path(planningRoot, weaveID, doc.ID), buf.Bytes(), 0o644); err != nil {
		return err
	}
	return syncLedger(planningRoot, weaveID)
}

func Load(sharedPlanningRoot, weaveID, threadID string) (norn.Document, error) {
	data, err := os.ReadFile(Path(sharedPlanningRoot, weaveID, threadID))
	if err != nil {
		return norn.Document{}, err
	}
	parts := strings.SplitN(string(data), "---\n", 3)
	if len(parts) < 3 {
		return norn.Document{ID: threadID, Title: threadID, Body: string(data)}, nil
	}
	var fm frontmatter
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return norn.Document{}, err
	}
	return norn.Document{ID: threadID, Title: fm.Title, Summary: fm.Summary, Body: strings.TrimSpace(parts[2])}, nil
}

func List(sharedPlanningRoot, weaveID string) ([]norn.Document, error) {
	entries, err := os.ReadDir(Root(sharedPlanningRoot, weaveID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []norn.Document
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		doc, err := Load(sharedPlanningRoot, weaveID, id)
		if err != nil {
			return nil, fmt.Errorf("load thread %s: %w", id, err)
		}
		out = append(out, doc)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func ListMerged(sharedPlanningRoot, overlayPlanningRoot, weaveID string) ([]norn.Document, error) {
	merged := map[string]norn.Document{}
	shared, err := List(sharedPlanningRoot, weaveID)
	if err != nil {
		return nil, err
	}
	for _, item := range shared {
		merged[item.ID] = item
	}
	overlay, err := List(overlayPlanningRoot, weaveID)
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

func LoadMerged(sharedPlanningRoot, overlayPlanningRoot, weaveID, threadID string) (norn.Document, error) {
	if doc, err := Load(overlayPlanningRoot, weaveID, threadID); err == nil {
		return doc, nil
	}
	return Load(sharedPlanningRoot, weaveID, threadID)
}

func Delete(sharedPlanningRoot, weaveID, threadID string) error {
	if err := os.Remove(Path(sharedPlanningRoot, weaveID, threadID)); err != nil {
		return err
	}
	return syncLedger(sharedPlanningRoot, weaveID)
}

func DefaultBody(summary string) string {
	return strings.TrimSpace(fmt.Sprintf(`## Goal

%s

## User Story

- As a user, I want this thread to be explicit and readable so implementation can be carried out consistently.

## Strands

- define the concrete implementation strands for this thread

## Acceptance

- define the acceptance criteria for this thread

## Documentation

- [ ] CLI --help text updated for any new commands or flags
- [ ] Project docs (docs/) updated with user-facing guides
- [ ] Specifications documented for AI agent consumption
- [ ] Integration boundaries documented (what Norn owns vs what OpenCode owns)

## Guides

- Quick start path for users
- Configuration reference (if applicable)

## Specifications

- Data formats and schemas
- API/contracts for AI agent interaction

## Notes

- replace this template content with project-specific details
`, summary))
}

func syncLedger(sharedPlanningRoot, weaveID string) error {
	items, err := List(sharedPlanningRoot, weaveID)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString("# Threads\n\n")
	if len(items) == 0 {
		buf.WriteString("Active threads for this weave:\n\n- none yet\n")
	} else {
		buf.WriteString("Active threads for this weave:\n\n")
		for _, item := range items {
			buf.WriteString(fmt.Sprintf("- `%s` - %s\n", item.ID, item.Title))
		}
	}
	return os.WriteFile(weaves.ThreadsLedgerPath(sharedPlanningRoot, weaveID), buf.Bytes(), 0o644)
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "_", "-", ":", "-")
	return replacer.Replace(value)
}
