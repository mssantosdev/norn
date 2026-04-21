# Norn

Norn is a multi-agent planning and coordination harness built to work with simple repositories, manual git worktree layouts, and Hydra-managed workspaces.

## v0.0.1 scope

- interactive-first CLI with Charm-based UI
- non-interactive pathways for automation and agent usage
- root-level `.norn/` runtime coordination
- all artifacts consolidated under `.norn/`
- four core fates: `keeper`, `weaver`, `judge`, and `fates`

## Current commands

- `norn init`
- `norn status`
- `norn detect`
- `norn fates list|show|add|edit|remove`
- `norn patterns list|add|show|edit|remove`
- `norn skills list|add|show|edit|remove`
- `norn tools list|add|show|edit|remove`
- `norn weaves list|add|show|remove`
- `norn threads list|add|show|remove`
- `norn warps list|add|assign|assignment|show|remove`
- `norn runes show|resolve|edit`
- `norn export --opencode`
- `norn chat validate|status|assist|preview`

All commands support `--help` and `--help --format=json` for discovery.

Interactive creation is available by default for:

- `norn weaves add`
- `norn threads add`

These flows currently provide:

- Charm TUI forms
- template-backed goal, user story, scope/strands, and acceptance sections
- preview before write

`norn init` is interactive by default and supports non-interactive flags such as `--no-interactive`, `--mode=folder|branch`, `--branch=<name>`, `--create-branch`, `--enable-opencode`, `--languages=...`, and `--tools=...`.

## Configuration

- Global user config: `~/.config/norn/runes.yaml`
- Workspace config: `.norn/runes.yaml`
- Private local override: `.norn/runes.local.yaml`
- Precedence: local override, workspace, global, built-in defaults

Config commands support both interactive and non-interactive usage:

- `norn runes show`
- `norn runes show --scope=global|workspace|local`
- `norn runes resolve`
- `norn runes resolve --format=table|yaml`
- `norn runes edit`
- `norn runes edit --scope=global|workspace|local --set path=value`
- `norn runes edit --scope=global|workspace|local --unset path`

`norn runes resolve` returns effective config with full origin metadata per field.

- default format is a human-friendly table
- `--format=yaml` returns the full structured resolution payload
- `resolve` works both inside a workspace and outside one, using global config plus defaults when no workspace is present

## Planning Artifacts

All Norn artifacts live under `.norn/`:

- `.norn/weaves/<weave-id>/README.md`
- `.norn/weaves/<weave-id>/threads.md`
- `.norn/weaves/<weave-id>/threads/<thread-id>.md`
- `.norn/patterns/`
- `.norn/skills/`
- `.norn/fates/`
- `.norn/tools/`

Runtime coordination lives in `.norn/spindle/`.

- `norn warps list`
- `norn warps list --view=runtime`
- `norn warps add`
- `norn warps add --status=... --owner=... --root=... --branch=... --weaves=a,b --threads=a,b <title> <summary>`
- `norn warps assign --kind=weave|thread --id=<artifact-id> --warp=<warp-id> [--owner=...] [--state=...] [--notes=...]`
- `norn warps assignment show <weave|thread> <id>`
- `norn warps assignment remove <weave|thread> <id>`
- `norn warps show <warp-id>`
- `norn warps remove <warp-id>`

Capability artifacts:

- `.norn/fates/` - Agent fate definitions
- `.norn/skills/` - Agent capability documents
- `.norn/tools/` - Tool permission batch definitions

All artifact classes should remain readable and easy to index for both humans and AI agents.

## Status

The repository has a working v0.0.1 bootstrap foundation with init, detection, managed tools, generated fates, skill export, and OpenCode-compatible agent export.

The optional planner/specifier specialist is documented as a concept for v0.0.1, but workflow and invocation are still deferred.

The runtime slice now includes local spindle-backed warp records and direct runtime assignment management for weave/thread ownership.

## Documentation

- `docs/getting-started.md` - Quick start guide
- `docs/architecture.md` - Architecture and design
- `docs/fates.md` - Fate roles and generation
- `docs/opencode.md` - OpenCode integration guide
- `docs/opencode-integration.md` - Living specification
- `docs/decisions.md` - Key design decisions
