# OpenCode Integration Specification

## Current State

Norn generates OpenCode-compatible agent files during initialization. The integration is one-way: Norn writes agent definitions based on its internal fate and command models.

## Target State

Bi-directional integration where Norn can:
- Generate agents from fate definitions
- Invoke OpenCode for assisted planning
- Export configuration for OpenCode consumption
- Validate OpenCode availability and configuration

## Data Formats

### Fate Source Format

Stored in `.norn/fates/<name>.yaml`:

```yaml
name: string
model: string
temperature: string
body: string
allow_edit: bool
extra_allow: []string
extra_ask: []string
extra_deny: []string
```

### Agent Export Format

Rendered to `.opencode/agents/<name>.md`:

```markdown
---
description: string
mode: all
model: string
temperature: string
permission:
  edit: allow|deny
  bash:
    "<pattern>": allow|ask|deny
---
<body>
```

### Config Format

Stored in `.norn/runes.yaml` under `opencode` key:

```yaml
opencode:
  enabled: bool
  provider: string
  model: string
  agent: string
  response_language: string
  drafting_mode: ask|auto
```

## CLI Contracts

### Validation

```bash
norn chat validate
```

Returns 0 if opencode binary is available, non-zero otherwise.

### Assisted Init

```bash
norn init --enable-opencode --prompt="Generate starter patterns for a Go project"
```

Calls `opencode run` with a JSON schema prompt and saves results to planning artifacts.

### Status

```bash
norn chat status
```

Returns JSON or text with: availability, enabled state, model, agent, response language, drafting mode, agents path, agents count, agent names.

### Export

```bash
norn chat export [--output=<dir>]
```

Exports agents and config snapshot. Returns 0 on success.

### Assist

```bash
norn chat assist [--prompt=<prompt>]
```

Interactive or non-interactive AI assistance. Prompts for approval before saving artifacts.

### Preview

```bash
norn chat preview --prompt=<prompt>
```

Generates artifacts but does not save. Returns 0 with preview output.

## API/Contracts for AI Agent Interaction

### Request Contract

When Norn calls OpenCode for assistance, it sends:
- A structured prompt with expected JSON schema
- Model and agent configuration
- Context about the current project

### Response Contract

OpenCode should return:
- JSON matching the expected schema
- Lists of weaves, patterns, and skills
- Each item has title, summary, and body fields

## Acceptance Criteria

- [x] Agent files are generated during init
- [x] Config is manageable through `norn runes`
- [x] Validation command exists
- [x] Assisted init works with JSON prompt
- [x] Export behavior beyond agent generation (`norn chat export`)
- [x] Assisted editing for existing artifacts (`norn chat assist`)
- [x] Preview and approval flows (`norn chat preview`)
- [ ] Diff previews
- [ ] Iterative refinement workflows

## Related Paths

- `internal/opencode/opencode.go`
- `internal/fates/fates.go`
- `internal/cli/cli.go`
- `docs/opencode.md`
