package detect

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mssantosdev/norn/internal/norn"
	"gopkg.in/yaml.v3"
)

type hydraConfig struct {
	Paths struct {
		WorktreeDir string `yaml:"worktree_dir"`
	} `yaml:"paths"`
	Ecosystems map[string]map[string]string `yaml:"ecosystems"`
}

func Scan(root string) (norn.Detection, error) {
	if _, err := os.Stat(filepath.Join(root, ".hydra.yaml")); err == nil {
		return scanHydra(root)
	}
	return scanSingle(root)
}

func scanSingle(root string) (norn.Detection, error) {
	detection := norn.Detection{}
	applySignals(root, &detection)
	finalize(&detection)
	return detection, nil
}

func scanHydra(root string) (norn.Detection, error) {
	data, err := os.ReadFile(filepath.Join(root, ".hydra.yaml"))
	if err != nil {
		return norn.Detection{}, err
	}
	var cfg hydraConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return norn.Detection{}, fmt.Errorf("parse hydra config: %w", err)
	}
	worktreeDir := cfg.Paths.WorktreeDir
	if worktreeDir == "" {
		worktreeDir = "."
	}
	detection := norn.Detection{}
	for ecosystem := range cfg.Ecosystems {
		ecosystemPath := filepath.Join(root, worktreeDir, ecosystem)
		entries, err := os.ReadDir(ecosystemPath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			target := filepath.Join(ecosystemPath, entry.Name())
			if real, err := filepath.EvalSymlinks(target); err == nil {
				target = real
			}
			applySignals(target, &detection)
			detection.Locations = append(detection.Locations, target)
		}
	}
	finalize(&detection)
	return detection, nil
}

func applySignals(root string, detection *norn.Detection) {
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(root, name))
		return err == nil
	}
	if exists("go.mod") {
		detection.Languages = append(detection.Languages, "go")
	}
	if exists("package.json") {
		detection.Languages = append(detection.Languages, "node")
	}
	if exists("bun.lock") || exists("bun.lockb") || exists("bunfig.toml") {
		detection.Languages = append(detection.Languages, "bun")
	}
	if exists("pom.xml") {
		detection.Languages = append(detection.Languages, "java")
	}
	if exists("build.gradle") || exists("build.gradle.kts") {
		detection.Languages = append(detection.Languages, "java")
		detection.Frameworks = append(detection.Frameworks, "gradle")
	}
	if exists("Cargo.toml") {
		detection.Languages = append(detection.Languages, "rust")
	}
	if hasGlob(root, "*.sln") || exists("Directory.Build.props") || exists("global.json") {
		detection.Languages = append(detection.Languages, ".net")
	}
	if exists("settings.gradle.kts") || exists("settings.gradle") {
		detection.Languages = append(detection.Languages, "kotlin")
	}
	if exists("Makefile") {
		detection.Tools = append(detection.Tools, "make")
	}
	if exists("Dockerfile") || exists("docker-compose.yml") || exists("docker-compose.yaml") {
		detection.Tools = append(detection.Tools, "docker")
	}
	if exists("mise.toml") {
		detection.Tools = append(detection.Tools, "mise")
	}
	if exists(".git") {
		detection.Tools = append(detection.Tools, "git")
	}
	if exists(".github") {
		detection.Tools = append(detection.Tools, "github")
	}
	if exists("mvnw") || exists("pom.xml") {
		detection.Tools = append(detection.Tools, "maven")
	}
	if exists("gradlew") || exists("build.gradle") || exists("build.gradle.kts") {
		detection.Tools = append(detection.Tools, "gradle")
	}
	if exists("src/main/resources/application.yml") || exists("src/main/resources/application.yaml") || exists("src/main/resources/application.properties") {
		detection.Frameworks = append(detection.Frameworks, "spring")
	}
	if exists("package.json") {
		packageManager := detectPackageManager(root)
		if packageManager != "" {
			detection.Tools = append(detection.Tools, packageManager)
		}
	}
}

func detectPackageManager(root string) string {
	for _, pair := range []struct {
		name  string
		label string
	}{
		{"pnpm-lock.yaml", "pnpm"},
		{"yarn.lock", "yarn"},
		{"package-lock.json", "npm"},
	} {
		if _, err := os.Stat(filepath.Join(root, pair.name)); err == nil {
			return pair.label
		}
	}
	return "npm"
}

func finalize(detection *norn.Detection) {
	detection.Languages = unique(detection.Languages)
	detection.Tools = unique(detection.Tools)
	detection.Frameworks = unique(detection.Frameworks)
	detection.Locations = unique(detection.Locations)
}

func hasGlob(root, pattern string) bool {
	matches, err := filepath.Glob(filepath.Join(root, pattern))
	if err != nil {
		return false
	}
	return len(matches) > 0
}

func unique(items []string) []string {
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
