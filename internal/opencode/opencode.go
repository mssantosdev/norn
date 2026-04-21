package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mssantosdev/norn/internal/norn"
)

type Assistance struct {
	Weaves   []norn.Document `json:"weaves"`
	Patterns []norn.Document `json:"patterns"`
	Skills   []norn.Document `json:"skills"`
}

type Status struct {
	Available    bool     `json:"available"`
	Enabled      bool     `json:"enabled"`
	Provider     string   `json:"provider,omitempty"`
	Model        string   `json:"model,omitempty"`
	Agent        string   `json:"agent,omitempty"`
	ResponseLang string   `json:"response_language,omitempty"`
	DraftingMode string   `json:"drafting_mode,omitempty"`
	AgentsPath   string   `json:"agents_path,omitempty"`
	AgentsCount  int      `json:"agents_count,omitempty"`
	AgentNames   []string `json:"agent_names,omitempty"`
}

func Validate() error {
	if _, err := exec.LookPath("opencode"); err != nil {
		return fmt.Errorf("opencode not found in PATH")
	}
	return nil
}

func GetStatus(workspace norn.Workspace) Status {
	s := Status{
		AgentsPath: norn.OpenCodeAgentsRoot(workspace),
	}
	if err := Validate(); err == nil {
		s.Available = true
	}
	s.Enabled = workspace.Runes.OpenCode.Enabled
	s.Provider = workspace.Runes.OpenCode.Provider
	s.Model = workspace.Runes.OpenCode.Model
	s.Agent = workspace.Runes.OpenCode.Agent
	s.ResponseLang = workspace.Runes.OpenCode.ResponseLanguage
	s.DraftingMode = workspace.Runes.OpenCode.DraftingMode
	if entries, err := os.ReadDir(s.AgentsPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				s.AgentsCount++
				s.AgentNames = append(s.AgentNames, strings.TrimSuffix(entry.Name(), ".md"))
			}
		}
	}
	return s
}

func Assist(cfg norn.OpenCodeConfig, context, prompt string) (Assistance, error) {
	if err := Validate(); err != nil {
		return Assistance{}, err
	}
	rawPrompt := strings.Join([]string{
		"Return JSON only.",
		"Schema: {\"weaves\": [{\"title\": string, \"summary\": string, \"body\": string}], \"patterns\": [{\"title\": string, \"summary\": string, \"body\": string}], \"skills\": [{\"title\": string, \"summary\": string, \"body\": string}]}",
		"Context:", context,
		"Request:", prompt,
	}, " ")
	cmd := exec.Command("opencode", "run", "--model", cfg.Model, "--agent", cfg.Agent, rawPrompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return Assistance{}, fmt.Errorf("opencode run failed: %s", strings.TrimSpace(string(output)))
	}
	cleaned, err := extractJSONObject(output)
	if err != nil {
		return Assistance{}, err
	}
	var result Assistance
	if err := json.Unmarshal(cleaned, &result); err != nil {
		return Assistance{}, fmt.Errorf("parse opencode output: %w", err)
	}
	return result, nil
}

func AssistInit(cfg norn.OpenCodeConfig, prompt string) (Assistance, error) {
	return Assist(cfg, "Help bootstrap a new project harness. Keep the lists short and immediately useful.", prompt)
}

func ExportConfig(workspace norn.Workspace, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	cfg := workspace.Runes.OpenCode
	data := map[string]interface{}{
		"version":           "1",
		"enabled":           cfg.Enabled,
		"provider":          cfg.Provider,
		"model":             cfg.Model,
		"agent":             cfg.Agent,
		"response_language": cfg.ResponseLanguage,
		"drafting_mode":     cfg.DraftingMode,
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(targetDir, "norn-opencode.json"), b, 0o644)
}

func extractJSONObject(data []byte) ([]byte, error) {
	text := string(data)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || end < start {
		return nil, fmt.Errorf("opencode output did not contain a JSON object")
	}
	return []byte(text[start : end+1]), nil
}
