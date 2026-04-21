# OpenCode Integration

Norn can generate OpenCode-compatible agent files from its fate definitions and optionally invoke OpenCode for assisted planning during initialization.

## Quick Start

Enable OpenCode when initializing a project:

```bash
norn init --enable-opencode
```

Or interactively:

```bash
norn init
# Select "Enable OpenCode integration?" → Yes
```

Validate that OpenCode is available:

```bash
norn chat validate
```

## What Gets Generated

During `norn init --enable-opencode`:

- `.opencode/agents/keeper.md`
- `.opencode/agents/weaver.md`
- `.opencode/agents/judge.md`
- `.opencode/agents/fates.md`

These files contain the agent definitions with permissions derived from the tool registry.

## Configuration

OpenCode settings live under the `opencode` key in `.norn/runes.yaml`:

```yaml
opencode:
  enabled: true
  provider: github-copilot
  model: github-copilot/gpt-5.4-mini
  agent: build
  response_language: en
  drafting_mode: ask
```

Manage settings:

```bash
norn runes edit --scope=workspace --set opencode.enabled=true
norn runes show --scope=workspace
```

## Manual Setup

If you prefer manual integration:

1. Ensure `opencode` is on your PATH
2. Create `.opencode/agents/` directory
3. Copy agent definitions from Norn or write your own
4. Agent format is documented in the specification section

## Agent Format

Agent files are Markdown with YAML frontmatter:

```markdown
---
description: Coordinates weaves, assignments, and handoffs.
mode: all
model: github-copilot/gpt-5.4-mini
temperature: "0.2"
permission:
  edit: deny
  bash:
    "*": ask
    "git status*": allow
---
You are the keeper fate. Coordinate planning, assign work, and keep runtime state clear and current.
```

## Integration Boundaries

**Norn owns:**
- Fate source definitions (`.norn/fates/*.yaml`)
- Tool permission registry (`.norn/tools/*.yaml`)
- Agent file generation (`.opencode/agents/*.md`)
- Config management (`.norn/runes.yaml`)

**OpenCode owns:**
- Agent execution
- Model selection and inference
- Prompt handling and response generation

**Shared:**
- Agent format specification
- Permission semantics

## CLI Commands

### Validate

Check if OpenCode is installed:

```bash
norn chat validate
```

### Status

Show current OpenCode integration status:

```bash
norn chat status
```

Shows: availability, enabled state, model, agent, agents generated.

### Export

Export agents and skills:

```bash
norn export --opencode
norn export --opencode --dry-run
norn export --opencode --fates
norn export --opencode --skills
norn fate export --opencode
norn skill export --opencode
```

Exports:
- `.opencode/agents/*.md` (regenerated from current fates)
- `.opencode/skills/norn-<skill>/SKILL.md` (exported from `.norn/skills/`)

### Assist

Get AI assistance for planning artifacts:

```bash
norn chat assist
norn chat assist --prompt="Generate starter patterns for a Go API"
```

Interactive mode prompts for input, then shows generated artifacts for approval before saving.

### Preview

Preview AI-generated artifacts without saving:

```bash
norn chat preview --prompt="Generate deployment skills"
```

Use this to review what the AI would generate before committing with `norn chat assist`.

## Troubleshooting

- `norn chat validate` reports "opencode not found in PATH" → Install opencode CLI
- Agents not generated → Check `opencode.enabled` in runes config
- Permission changes not reflected → Re-run `norn chat export` or `norn init`
- `norn chat assist` fails → Ensure `opencode.enabled=true` and binary is on PATH

## Future Work

- Diff previews and iterative refinement workflows
- Merge behavior for partial updates to existing artifacts

## Related

- `docs/getting-started.md`
- `docs/fates.md`
- `docs/architecture.md`
- `docs/opencode-integration.md`
