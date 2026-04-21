package warps

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mssantosdev/norn/internal/norn"
	"gopkg.in/yaml.v3"
)

func Root(spindleRoot string) string {
	return filepath.Join(spindleRoot, "warps")
}

func Path(spindleRoot, warpID string) string {
	return filepath.Join(Root(spindleRoot), warpID+".yaml")
}

func Save(spindleRoot string, warp norn.Warp) error {
	if warp.ID == "" {
		warp.ID = slug(warp.Title)
	}
	if err := os.MkdirAll(Root(spindleRoot), 0o755); err != nil {
		return err
	}
	warp.WeaveIDs = dedupeSorted(warp.WeaveIDs)
	warp.ThreadIDs = dedupeSorted(warp.ThreadIDs)
	data, err := yaml.Marshal(warp)
	if err != nil {
		return err
	}
	return os.WriteFile(Path(spindleRoot, warp.ID), data, 0o644)
}

func Load(spindleRoot, warpID string) (norn.Warp, error) {
	data, err := os.ReadFile(Path(spindleRoot, warpID))
	if err != nil {
		return norn.Warp{}, err
	}
	var warp norn.Warp
	if err := yaml.Unmarshal(data, &warp); err != nil {
		return norn.Warp{}, err
	}
	return warp, nil
}

func List(spindleRoot string) ([]norn.Warp, error) {
	entries, err := os.ReadDir(Root(spindleRoot))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []norn.Warp
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		item, err := Load(spindleRoot, strings.TrimSuffix(entry.Name(), ".yaml"))
		if err != nil {
			return nil, fmt.Errorf("load warp %s: %w", entry.Name(), err)
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func Delete(spindleRoot, warpID string) error {
	return os.Remove(Path(spindleRoot, warpID))
}

func DefaultNotes(summary string) string {
	return strings.TrimSpace(fmt.Sprintf("status: define the current runtime state for this warp\nsummary: %s", summary))
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "_", "-", ":", "-")
	return replacer.Replace(value)
}

func dedupeSorted(items []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}
