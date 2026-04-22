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
	Weaves   []norn.Document    `json:"weaves"`
	Patterns []norn.Document    `json:"patterns"`
	Skills   []norn.Document    `json:"skills"`
	Tools    []norn.ManagedTool `json:"tools"`
}

type AssistRequest struct {
	Type    string // "weaves", "patterns", "skills", "tools", "all"
	Context string
	Prompt  string
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

type Provider struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Configured bool   `json:"configured"`
}

func GetProviders() ([]Provider, error) {
	if err := Validate(); err != nil {
		return nil, err
	}
	authPath := filepath.Join(os.Getenv("HOME"), ".local", "share", "opencode", "auth.json")
	data, err := os.ReadFile(authPath)
	if err != nil {
		return nil, fmt.Errorf("no opencode auth found; run 'opencode providers login' first")
	}
	var auth map[string]any
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil, fmt.Errorf("parse auth.json: %w", err)
	}
	var providers []Provider
	for name, info := range auth {
		if infoMap, ok := info.(map[string]any); ok {
			pType := "unknown"
			if t, ok := infoMap["type"].(string); ok {
				pType = t
			}
			providers = append(providers, Provider{Name: name, Type: pType, Configured: true})
		}
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers configured; run 'opencode providers login <provider>'")
	}
	return providers, nil
}

func GetModels(provider string) ([]string, error) {
	if err := Validate(); err != nil {
		return nil, err
	}
	cmd := exec.Command("opencode", "models", provider)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %s", strings.TrimSpace(string(output)))
	}
	var models []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "Error") {
			models = append(models, line)
		}
	}
	return models, nil
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

func Assist(req AssistRequest) (Assistance, error) {
	if err := Validate(); err != nil {
		return Assistance{}, err
	}

	artifactType := req.Type
	if artifactType == "" {
		artifactType = InferType(req.Prompt)
	}

	schema := getSchemaForType(artifactType)
	rawPrompt := strings.Join([]string{
		"Return JSON only.",
		"Schema:", schema,
		"Context:", req.Context,
		"Request:", req.Prompt,
	}, " ")
	cmd := exec.Command("opencode", "run", "--model", "github-copilot/gpt-5.4-mini", "--agent", "build", rawPrompt)
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

	// Validate requested type was returned
	if err := validateResult(result, artifactType); err != nil {
		return result, err
	}

	return result, nil
}

func InferType(prompt string) string {
	promptLower := strings.ToLower(prompt)
	typeKeywords := map[string][]string{
		"tools":    {"tools", "commands", "permissions", "tool", "command"},
		"weaves":   {"weaves", "plan", "features", "feature", "planning", "weave"},
		"patterns": {"patterns", "conventions", "pattern", "convention"},
		"skills":   {"skills", "knowledge", "skill", "know-how"},
	}
	for t, keywords := range typeKeywords {
		for _, kw := range keywords {
			if strings.Contains(promptLower, kw) {
				return t
			}
		}
	}
	return "all"
}

func getSchemaForType(artifactType string) string {
	schemas := map[string]string{
		"tools":    `{"tools": [{"id": string, "title": string, "command": string, "category": string, "risk": string, "roles": [string]}]}`,
		"weaves":   `{"weaves": [{"title": string, "summary": string, "body": string}]}`,
		"patterns": `{"patterns": [{"title": string, "summary": string, "body": string}]}`,
		"skills":   `{"skills": [{"title": string, "summary": string, "body": string}]}`,
		"all":      `{"weaves": [{"title": string, "summary": string, "body": string}], "patterns": [{"title": string, "summary": string, "body": string}], "skills": [{"title": string, "summary": string, "body": string}], "tools": [{"id": string, "title": string, "command": string, "category": string, "risk": string, "roles": [string]}]}`,
	}
	if schema, ok := schemas[artifactType]; ok {
		return schema
	}
	return schemas["all"]
}

func validateResult(result Assistance, artifactType string) error {
	switch artifactType {
	case "tools":
		if len(result.Tools) == 0 {
			return fmt.Errorf("requested tools but AI returned no tools; try a more specific prompt")
		}
	case "weaves":
		if len(result.Weaves) == 0 {
			return fmt.Errorf("requested weaves but AI returned no weaves; try a more specific prompt")
		}
	case "patterns":
		if len(result.Patterns) == 0 {
			return fmt.Errorf("requested patterns but AI returned no patterns; try a more specific prompt")
		}
	case "skills":
		if len(result.Skills) == 0 {
			return fmt.Errorf("requested skills but AI returned no skills; try a more specific prompt")
		}
	}
	return nil
}

func AssistInit(prompt string) (Assistance, error) {
	return Assist(AssistRequest{
		Type:    "all",
		Context: "Help bootstrap a new project harness. Keep the lists short and immediately useful.",
		Prompt:  prompt,
	})
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
