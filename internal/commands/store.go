package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mssantosdev/norn/internal/norn"
	"gopkg.in/yaml.v3"
)

func Path(root, id string) string {
	return filepath.Join(root, id+".yaml")
}

func Save(root string, item norn.ManagedCommand) error {
	if item.ID == "" {
		item.ID = slug(item.Title)
	}
	if item.Pattern == "" {
		item.Pattern = item.Command + "*"
	}
	if item.Risk == "" {
		item.Risk = "medium"
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(item)
	if err != nil {
		return err
	}
	return os.WriteFile(Path(root, item.ID), data, 0o644)
}

func Load(root, id string) (norn.ManagedCommand, error) {
	data, err := os.ReadFile(Path(root, id))
	if err != nil {
		return norn.ManagedCommand{}, err
	}
	var item norn.ManagedCommand
	if err := yaml.Unmarshal(data, &item); err != nil {
		return norn.ManagedCommand{}, err
	}
	return item, nil
}

func List(root string) ([]norn.ManagedCommand, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var items []norn.ManagedCommand
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".yaml")
		item, err := Load(root, id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func Delete(root, id string) error {
	if err := os.Remove(Path(root, id)); err != nil {
		return fmt.Errorf("remove command: %w", err)
	}
	return nil
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "_", "-", ":", "-")
	return replacer.Replace(value)
}
