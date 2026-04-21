package export

import (
	"fmt"
	"strings"

	"github.com/mssantosdev/norn/internal/norn"
	"github.com/mssantosdev/norn/internal/ui/logger"
)

type Options struct {
	Target    string // "opencode"
	Fates     bool
	Skills    bool
	FateName  string // specific fate
	SkillName string // specific skill
	DryRun    bool
}

type Change struct {
	Path   string
	Action string // "create", "update", "delete"
}

func Run(w norn.Workspace, opts Options) error {
	if opts.Target != "opencode" {
		return fmt.Errorf("unsupported export target: %s", opts.Target)
	}

	changes := []Change{}

	if opts.Fates || opts.FateName != "" {
		fateChanges, err := exportFates(w, opts)
		if err != nil {
			return err
		}
		changes = append(changes, fateChanges...)
	}

	if opts.Skills || opts.SkillName != "" {
		skillChanges, err := exportSkills(w, opts)
		if err != nil {
			return err
		}
		changes = append(changes, skillChanges...)
	}

	if len(changes) == 0 {
		logger.Info("nothing to export")
		return nil
	}

	if opts.DryRun {
		logger.Print("Would export:")
		for _, change := range changes {
			logger.Print(fmt.Sprintf("  [%s] %s", strings.ToUpper(change.Action), change.Path))
		}
		return nil
	}

	// Confirm overwrite if files exist
	existing := []string{}
	for _, change := range changes {
		if change.Action == "update" {
			existing = append(existing, change.Path)
		}
	}

	if len(existing) > 0 {
		// For now, auto-accept in non-interactive mode
		// TODO: add interactive confirmation
		logger.Info("updating existing files", "count", len(existing))
	}

	// Execute exports
	if opts.Fates || opts.FateName != "" {
		if err := doExportFates(w, opts); err != nil {
			return err
		}
	}

	if opts.Skills || opts.SkillName != "" {
		if err := doExportSkills(w, opts); err != nil {
			return err
		}
	}

	logger.Info("export complete", "target", opts.Target, "files", len(changes))
	return nil
}
