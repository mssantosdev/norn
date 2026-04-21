package docs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mssantosdev/norn/internal/norn"
	"gopkg.in/yaml.v3"
)

type frontmatter struct {
	Title   string `yaml:"title"`
	Summary string `yaml:"summary,omitempty"`
}

func Save(root string, doc norn.Document) error {
	if doc.ID == "" {
		doc.ID = slug(doc.Title)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	fm, err := yaml.Marshal(frontmatter{Title: doc.Title, Summary: doc.Summary})
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fm)
	buf.WriteString("---\n\n")
	buf.WriteString(strings.TrimSpace(doc.Body))
	buf.WriteString("\n")
	return os.WriteFile(filepath.Join(root, doc.ID+".md"), buf.Bytes(), 0o644)
}

func Load(root, id string) (norn.Document, error) {
	data, err := os.ReadFile(filepath.Join(root, id+".md"))
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

func List(root string) ([]norn.Document, error) {
	entries, err := os.ReadDir(root)
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
		doc, err := Load(root, id)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", id, err)
		}
		out = append(out, doc)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func Delete(root, id string) error {
	return os.Remove(filepath.Join(root, id+".md"))
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "_", "-", ":", "-")
	return replacer.Replace(value)
}
