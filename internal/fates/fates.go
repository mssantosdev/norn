package fates

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mssantosdev/norn/internal/norn"
	"github.com/mssantosdev/norn/internal/tools"
	"gopkg.in/yaml.v3"
)

var defaults = []norn.FateSource{
	{
		Name:        "keeper",
		Description: "Coordinates weaves, assignments, and handoffs.",
		Model:       "github-copilot/gpt-5.4-mini",
		Temperature: "0.2",
		AllowEdit:   false,
		Body:        "You are the keeper fate. Coordinate planning, assign work, and keep runtime state clear and current.",
	},
	{
		Name:        "weaver",
		Description: "Implements threads and validates owned work.",
		Model:       "github-copilot/gpt-5.4-mini",
		Temperature: "0.2",
		AllowEdit:   true,
		Body:        "You are the weaver fate. Implement assigned threads, keep strands small, and validate your own work before handoff.",
	},
	{
		Name:        "judge",
		Description: "Reviews architecture, standards, security, performance, and test coverage.",
		Model:       "github-copilot/gpt-5.4-mini",
		Temperature: "0.2",
		AllowEdit:   false,
		Body:        "You are the judge fate. Review changes for correctness, architecture, security, performance, standards, and test coverage.",
	},
	{
		Name:        "fates",
		Description: "Integrates approved work and owns merge and release transitions.",
		Model:       "github-copilot/gpt-5.4-mini",
		Temperature: "0.2",
		AllowEdit:   true,
		Body:        "You are the fates role. Integrate approved work, rerun integration validation, and handle controlled release flow.",
	},
}

func Bootstrap(root string, model string) error {
	for _, item := range defaults {
		item.Model = model
		if err := Save(root, item); err != nil {
			return err
		}
	}
	return nil
}

func Save(root string, source norn.FateSource) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(source)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, source.Name+".yaml"), data, 0o644)
}

func Load(root, name string) (norn.FateSource, error) {
	data, err := os.ReadFile(filepath.Join(root, name+".yaml"))
	if err != nil {
		return norn.FateSource{}, err
	}
	var source norn.FateSource
	if err := yaml.Unmarshal(data, &source); err != nil {
		return norn.FateSource{}, err
	}
	return source, nil
}

func List(root string) ([]norn.FateSource, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []norn.FateSource
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		item, err := Load(root, strings.TrimSuffix(entry.Name(), ".yaml"))
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func RenderOpenCode(source norn.FateSource, toolsRoot string) (string, error) {
	cmds, err := tools.List(toolsRoot)
	if err != nil {
		return "", err
	}
	allow, ask, deny := permissionSets(source.Name, cmds, source)
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("description: %s\n", yamlString(source.Description)))
	b.WriteString("mode: all\n")
	b.WriteString(fmt.Sprintf("model: %s\n", source.Model))
	b.WriteString(fmt.Sprintf("temperature: %s\n", source.Temperature))
	b.WriteString("permission:\n")
	if source.AllowEdit {
		b.WriteString("  edit: allow\n")
	} else {
		b.WriteString("  edit: deny\n")
	}
	b.WriteString("  bash:\n")
	b.WriteString("    \"*\": ask\n")
	for _, item := range allow {
		b.WriteString(fmt.Sprintf("    \"%s\": allow\n", item))
	}
	for _, item := range ask {
		b.WriteString(fmt.Sprintf("    \"%s\": ask\n", item))
	}
	for _, item := range deny {
		b.WriteString(fmt.Sprintf("    \"%s\": deny\n", item))
	}
	b.WriteString("---\n")
	b.WriteString(source.Body)
	b.WriteString("\n")
	return b.String(), nil
}

func ExportOpenCode(sourceRoot, toolsRoot, targetRoot string) error {
	items, err := List(sourceRoot)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		return err
	}
	for _, item := range items {
		rendered, err := RenderOpenCode(item, toolsRoot)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(targetRoot, item.Name+".md"), []byte(rendered), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func permissionSets(role string, cmds []norn.ManagedTool, source norn.FateSource) ([]string, []string, []string) {
	allowSet := map[string]bool{}
	askSet := map[string]bool{}
	denySet := map[string]bool{}
	for _, item := range []string{"git status*", "git diff*", "git log*"} {
		if role == "keeper" || role == "judge" || role == "weaver" || role == "fates" {
			allowSet[item] = true
		}
	}
	for _, item := range []string{"git push*", "git tag*"} {
		if role != "fates" {
			denySet[item] = true
		}
	}
	if role == "judge" {
		denySet["git commit*"] = true
	}
	if role == "keeper" {
		denySet["git commit*"] = true
	}
	for _, cmd := range cmds {
		for _, cmdRole := range cmd.Roles {
			if cmdRole != role {
				continue
			}
			switch role {
			case "judge":
				if cmd.Category == "test" || cmd.Category == "lint" || cmd.Category == "review" || cmd.Category == "build" {
					allowSet[cmd.Pattern] = true
				} else {
					askSet[cmd.Pattern] = true
				}
			case "keeper":
				askSet[cmd.Pattern] = true
			default:
				if cmd.Risk == "high" {
					askSet[cmd.Pattern] = true
				} else {
					allowSet[cmd.Pattern] = true
				}
			}
		}
	}
	for _, item := range source.ExtraAllow {
		allowSet[item] = true
	}
	for _, item := range source.ExtraAsk {
		askSet[item] = true
	}
	for _, item := range source.ExtraDeny {
		denySet[item] = true
	}
	return sortedKeys(allowSet), sortedKeys(askSet), sortedKeys(denySet)
}

func Delete(root, name string) error {
	return os.Remove(filepath.Join(root, name+".yaml"))
}

func sortedKeys(items map[string]bool) []string {
	out := make([]string, 0, len(items))
	for item := range items {
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func yamlString(value string) string {
	if strings.ContainsAny(value, ":#") {
		return fmt.Sprintf("%q", value)
	}
	return value
}
