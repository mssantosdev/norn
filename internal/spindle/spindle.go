package spindle

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mssantosdev/norn/internal/norn"
	"github.com/mssantosdev/norn/internal/warps"
	"gopkg.in/yaml.v3"
)

func WeavesRoot(spindleRoot string) string {
	return filepath.Join(spindleRoot, "weaves")
}

func ThreadsRoot(spindleRoot string) string {
	return filepath.Join(spindleRoot, "threads")
}

func assignmentRoot(spindleRoot, kind string) (string, error) {
	switch kind {
	case "weave":
		return WeavesRoot(spindleRoot), nil
	case "thread":
		return ThreadsRoot(spindleRoot), nil
	default:
		return "", fmt.Errorf("unsupported assignment kind %q", kind)
	}
}

func assignmentPath(spindleRoot, kind, id string) (string, error) {
	root, err := assignmentRoot(spindleRoot, kind)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, id+".yaml"), nil
}

func SaveAssignment(spindleRoot string, item norn.RuntimeAssignment) error {
	if item.Kind == "" || item.ID == "" {
		return fmt.Errorf("assignment kind and id are required")
	}
	path, err := assignmentPath(spindleRoot, item.Kind, item.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(item)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadAssignment(spindleRoot, kind, id string) (norn.RuntimeAssignment, error) {
	path, err := assignmentPath(spindleRoot, kind, id)
	if err != nil {
		return norn.RuntimeAssignment{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return norn.RuntimeAssignment{}, err
	}
	var item norn.RuntimeAssignment
	if err := yaml.Unmarshal(data, &item); err != nil {
		return norn.RuntimeAssignment{}, err
	}
	return item, nil
}

func LoadAssignments(spindleRoot string, kind string) ([]norn.RuntimeAssignment, error) {
	root, err := assignmentRoot(spindleRoot, kind)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []norn.RuntimeAssignment
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		var item norn.RuntimeAssignment
		if err := yaml.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func DeleteAssignment(spindleRoot, kind, id string) error {
	path, err := assignmentPath(spindleRoot, kind, id)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

type WarpView struct {
	Warp    norn.Warp
	Weaves  []norn.RuntimeAssignment
	Threads []norn.RuntimeAssignment
}

func BuildWarpViews(spindleRoot string) ([]WarpView, error) {
	items, err := warps.List(spindleRoot)
	if err != nil {
		return nil, err
	}
	weavesAssignments, err := LoadAssignments(spindleRoot, "weave")
	if err != nil {
		return nil, err
	}
	threadAssignments, err := LoadAssignments(spindleRoot, "thread")
	if err != nil {
		return nil, err
	}
	index := map[string]*WarpView{}
	views := make([]WarpView, 0, len(items))
	for _, item := range items {
		views = append(views, WarpView{Warp: item})
		index[item.ID] = &views[len(views)-1]
	}
	for _, assignment := range weavesAssignments {
		view, ok := index[assignment.WarpID]
		if !ok {
			continue
		}
		view.Weaves = append(view.Weaves, assignment)
	}
	for _, assignment := range threadAssignments {
		view, ok := index[assignment.WarpID]
		if !ok {
			continue
		}
		view.Threads = append(view.Threads, assignment)
	}
	return views, nil
}
