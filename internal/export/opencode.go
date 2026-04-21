package export

import (
	"os"
	"path/filepath"

	"github.com/mssantosdev/norn/internal/fates"
	"github.com/mssantosdev/norn/internal/norn"
)

func exportFates(w norn.Workspace, opts Options) ([]Change, error) {
	changes := []Change{}
	sourceRoot := norn.FatesRoot(w)
	targetRoot := norn.OpenCodeAgentsRoot(w)

	items, err := fates.List(sourceRoot)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if opts.FateName != "" && item.Name != opts.FateName {
			continue
		}
		targetPath := filepath.Join(targetRoot, item.Name+".md")
		action := "create"
		if _, err := os.Stat(targetPath); err == nil {
			action = "update"
		}
		changes = append(changes, Change{Path: targetPath, Action: action})
	}

	return changes, nil
}

func doExportFates(w norn.Workspace, opts Options) error {
	return fates.ExportOpenCode(norn.FatesRoot(w), norn.ToolsRoot(w), norn.OpenCodeAgentsRoot(w))
}
