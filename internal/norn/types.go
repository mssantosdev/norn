package norn

type WorkspaceMode string

const (
	WorkspaceModeRepo      WorkspaceMode = "repo"
	WorkspaceModeWorkspace WorkspaceMode = "workspace"
)

type PlanningMode string

const (
	PlanningModeFolder PlanningMode = "folder"
	PlanningModeBranch PlanningMode = "branch"
)

type RuneFile struct {
	Version     string            `yaml:"version"`
	Name        string            `yaml:"name"`
	Mode        WorkspaceMode     `yaml:"mode"`
	Preferences PreferencesConfig `yaml:"preferences"`
	UI          UIConfig          `yaml:"ui"`
	Planning    PlanningConfig    `yaml:"planning"`
	Overlay     OverlayConfig     `yaml:"overlay"`
	OpenCode    OpenCodeConfig    `yaml:"opencode"`
	Tooling     ToolingConfig     `yaml:"tooling"`
	Hydra       HydraConfig       `yaml:"hydra"`
}

type PreferencesConfig struct {
	Language  string `yaml:"language,omitempty"`
	Verbosity string `yaml:"verbosity,omitempty"`
}

type UIConfig struct {
	Theme string `yaml:"theme"`
}

type PlanningConfig struct {
	Mode           PlanningMode `yaml:"mode"`
	Path           string       `yaml:"path"`
	Branch         string       `yaml:"branch,omitempty"`
	DefaultSurface string       `yaml:"default_surface,omitempty"`
}

type OverlayConfig struct {
	Path string `yaml:"path"`
}

type OpenCodeConfig struct {
	Enabled          bool   `yaml:"enabled"`
	Provider         string `yaml:"provider"`
	Model            string `yaml:"model"`
	Agent            string `yaml:"agent"`
	ResponseLanguage string `yaml:"response_language,omitempty"`
	DraftingMode     string `yaml:"drafting_mode,omitempty"`
}

type ToolingConfig struct {
	Languages  []string `yaml:"languages"`
	Tools      []string `yaml:"tools"`
	Frameworks []string `yaml:"frameworks"`
}

type HydraConfig struct {
	Enabled bool `yaml:"enabled"`
}

type Workspace struct {
	Root  string
	Runes RuneFile
}

type InitOptions struct {
	Name            string
	Mode            PlanningMode
	PlanningPath    string
	Skeleton        string
	EnableOpenCode  bool
	OpenCodeModel   string
	OpenCodeAgent   string
	Theme           string
	Languages       []string
	Tools           []string
	Frameworks      []string
	WorkspaceMode   WorkspaceMode
	HydraEnabled    bool
	NonInteractive  bool
	OpenCodePrompt  string
	LocalOverlayDir string
}

type Detection struct {
	Languages  []string
	Tools      []string
	Frameworks []string
	Locations  []string
}

type ManagedTool struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description,omitempty"`
	Category    string   `yaml:"category"`
	Command     string   `yaml:"command"`
	Pattern     string   `yaml:"pattern"`
	Risk        string   `yaml:"risk"`
	Roles       []string `yaml:"roles"`
}

type FateSource struct {
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	Model         string   `yaml:"model"`
	Temperature   string   `yaml:"temperature"`
	Body          string   `yaml:"body"`
	ExtraAllow    []string `yaml:"extra_allow,omitempty"`
	ExtraAsk      []string `yaml:"extra_ask,omitempty"`
	ExtraDeny     []string `yaml:"extra_deny,omitempty"`
	AllowEdit     bool     `yaml:"allow_edit"`
	AllowedSkills []string `yaml:"allowed_skills,omitempty"`
	AllowedTasks  []string `yaml:"allowed_tasks,omitempty"`
}

type Document struct {
	ID      string
	Title   string
	Summary string
	Body    string
}

type Warp struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	Summary   string   `yaml:"summary,omitempty"`
	Root      string   `yaml:"root,omitempty"`
	Branch    string   `yaml:"branch,omitempty"`
	Status    string   `yaml:"status,omitempty"`
	Owner     string   `yaml:"owner,omitempty"`
	WeaveIDs  []string `yaml:"weaves,omitempty"`
	ThreadIDs []string `yaml:"threads,omitempty"`
	Notes     string   `yaml:"notes,omitempty"`
}

type RuntimeAssignment struct {
	Kind   string `yaml:"kind"`
	ID     string `yaml:"id"`
	WarpID string `yaml:"warp"`
	Owner  string `yaml:"owner,omitempty"`
	State  string `yaml:"state,omitempty"`
	Notes  string `yaml:"notes,omitempty"`
}

type Status struct {
	Root          string
	Mode          WorkspaceMode
	PlanningMode  PlanningMode
	PlanningPath  string
	OverlayPath   string
	OpenCode      bool
	Hydra         bool
	Languages     []string
	Tools         []string
	Frameworks    []string
	Fates         int
	Commands      int
	Patterns      int
	Skills        int
	SharedWeaves  int
	OverlayWeaves int
}
