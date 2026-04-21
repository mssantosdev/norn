package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mssantosdev/norn/internal/norn"
	"github.com/mssantosdev/norn/internal/skills"
)

func exportSkills(w norn.Workspace, opts Options) ([]Change, error) {
	changes := []Change{}
	sourceRoot := norn.SkillsRoot(w)
	targetRoot := filepath.Join(w.Root, ".opencode", "skills")

	items, err := skills.List(sourceRoot)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if opts.SkillName != "" && slug(item.ID) != slug(opts.SkillName) {
			continue
		}
		skillName := "norn-" + slug(item.ID)
		targetPath := filepath.Join(targetRoot, skillName, "SKILL.md")
		action := "create"
		if _, err := os.Stat(targetPath); err == nil {
			action = "update"
		}
		changes = append(changes, Change{Path: targetPath, Action: action})
	}

	return changes, nil
}

func doExportSkills(w norn.Workspace, opts Options) error {
	sourceRoot := norn.SkillsRoot(w)
	targetRoot := filepath.Join(w.Root, ".opencode", "skills")

	items, err := skills.List(sourceRoot)
	if err != nil {
		return err
	}

	for _, item := range items {
		if opts.SkillName != "" && slug(item.ID) != slug(opts.SkillName) {
			continue
		}
		skillName := "norn-" + slug(item.ID)
		skillDir := filepath.Join(targetRoot, skillName)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			return err
		}

		content := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n%s\n", skillName, item.Summary, item.Body)
		targetPath := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "_", "-", ":", "-")
	return replacer.Replace(value)
}
