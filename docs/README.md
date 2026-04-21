# Norn

Norn is a multi-agent planning and coordination harness built to work with simple repositories, manual git worktree layouts, and Hydra-managed workspaces.

## v0.0.1 scope

- interactive-first CLI with Charm-based UI
- non-interactive pathways for automation and agent usage
- root-level `.norn/` runtime coordination
- flexible planning sources through a shared `loom/` directory or planning branch mode
- four core fates: `keeper`, `weaver`, `judge`, and `fates`

## Current commands

- `norn init`
- `norn status`
- `norn detect`
- `norn fates list|show`
- `norn patterns list|add|show|remove`
- `norn skills list|add|show|remove`
- `norn commands list|add|remove`
- `norn weaves list|add|show|remove`
- `norn threads list|add|show|remove`
- `norn warps list|add|assign|assignment|show|remove`
- `norn runes show|resolve|edit`
- `norn chat validate`

Write-surface selection is supported for planning creation commands:

- `--surface=shared`
- `--surface=local`
- `--surface=both`

Interactive creation is available by default for:

- `norn weaves add`
- `norn threads add`

These flows currently provide:

- Charm TUI forms
- surface selection
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

## Planning surfaces

- Shared plans live in `loom/` by default.
- Local-only plans live in `.norn/loom/`.
- Runtime coordination lives in `.norn/spindle/`.
- Planning branch mode uses a dedicated git worktree, defaulting to `.loom/`.

Runtime warp records are local and readable under `.norn/spindle/warps/`.

- `norn warps list`
- `norn warps list --view=runtime`
- `norn warps add`
- `norn warps add --status=... --owner=... --root=... --branch=... --weaves=a,b --threads=a,b <title> <summary>`
- `norn warps assign --kind=weave|thread --id=<artifact-id> --warp=<warp-id> [--owner=...] [--state=...] [--notes=...]`
- `norn warps assignment show <weave|thread> <id>`
- `norn warps assignment remove <weave|thread> <id>`
- `norn warps show <warp-id>`
- `norn warps remove <warp-id>`

Planning and documentation artifacts live in `loom/` and `.norn/loom/`.

Current shared weave/thread planning layout:

- `loom/weaves/<weave-id>/README.md`
- `loom/weaves/<weave-id>/threads.md`
- `loom/weaves/<weave-id>/threads/<thread-id>.md`

Current read behavior for `weaves` and `threads`:

- reads merge shared `loom/` and local `.norn/loom/`
- local overlay artifacts win when the same artifact id exists in both locations
- writes default to the shared planning root unless `--surface=local` or `--surface=both` is provided

Capability artifacts live under `.norn/`:

- `.norn/fates/`
- `.norn/skills/`
- `.norn/commands/`

All artifact classes should remain readable and easy to index for both humans and AI agents.

## Status

The repository has a working v0.0.1 bootstrap foundation with init, detection, managed commands, generated fates, and OpenCode-compatible agent export.

The optional planner/specifier specialist is documented as a concept for v0.0.1, but workflow and invocation are still deferred.

The runtime slice now includes local spindle-backed warp records and direct runtime assignment management for weave/thread ownership.

## Session Handoff

For a fresh session or a new machine, start with:

1. `AGENTS.md`
2. `.norn/loom/memory/next-session-bootstrap.md`
3. the memory files under `.norn/loom/memory/`
4. Fizzy cards `#25`, `#28`, `#30`, and `#32`

## Shared Tracking

Norn development is also tracked in the Fizzy board `Norn`.

- Use Fizzy for shared roadmap and task visibility.
- Use repository files for architecture, standards, and handoff memory.
- Agents should treat Fizzy as read-only unless the user explicitly asks for board updates.
