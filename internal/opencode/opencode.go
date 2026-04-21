package opencode

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mssantosdev/norn/internal/norn"
)

type Assistance struct {
	Weaves   []norn.Document `json:"weaves"`
	Patterns []norn.Document `json:"patterns"`
	Skills   []norn.Document `json:"skills"`
}

func Validate() error {
	if _, err := exec.LookPath("opencode"); err != nil {
		return fmt.Errorf("opencode not found in PATH")
	}
	return nil
}

func AssistInit(cfg norn.OpenCodeConfig, prompt string) (Assistance, error) {
	if err := Validate(); err != nil {
		return Assistance{}, err
	}
	rawPrompt := strings.Join([]string{
		"Return JSON only.",
		"Schema: {\"weaves\": [{\"title\": string, \"summary\": string, \"body\": string}], \"patterns\": [{\"title\": string, \"summary\": string, \"body\": string}], \"skills\": [{\"title\": string, \"summary\": string, \"body\": string}]}",
		"Keep the lists short and immediately useful for a new project harness.",
		prompt,
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

func extractJSONObject(data []byte) ([]byte, error) {
	text := string(data)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || end < start {
		return nil, fmt.Errorf("opencode output did not contain a JSON object")
	}
	return []byte(text[start : end+1]), nil
}
