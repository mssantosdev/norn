# Changelog

## v0.0.1

### Added

- **Core CLI** - Interactive-first CLI with Charm TUI and non-interactive support
- **Init** - `norn init` with guided and non-interactive modes
- **Status** - `norn status` showing workspace state
- **Detect** - `norn detect` for languages, tools, and frameworks
- **Config** - `norn runes show|resolve|edit` with layered scopes (global, workspace, local)
- **Fates** - `norn fates list|show|add|edit|remove` for agent role management
- **Patterns** - `norn patterns list|add|show|edit|remove` for design artifacts
- **Skills** - `norn skills list|add|show|edit|remove` for agent capabilities
- **Tools** - `norn tools list|add|show|edit|remove` for permission batch definitions
- **Weaves** - `norn weaves list|add|show|remove` for planning epics
- **Threads** - `norn threads list|add|show|remove` for work items
- **Warps** - `norn warps list|add|assign|assignment|show|remove` for runtime lanes
- **Export** - `norn export --opencode` for agent and skill export
- **OpenCode Integration** - `norn chat validate|status|assist|preview`
- **Help System** - `--help` and `--help --format=json` for all commands
- **Four Core Fates** - keeper, weaver, judge, fates

### Changed

- **Consolidated artifacts** - All Norn artifacts now live under `.norn/` (removed `loom/`)
- **Renamed commands to tools** - `norn commands` is now `norn tools` to clarify permission definitions
- **Simplified context** - `norn chat assist` uses minimal project info instead of full taxonomy

### Architecture

- Single `.norn/` directory for all artifacts
- `.norn/` is gitignored by default; users manually add what to share
- Planning branch mode and shared/local overlays deferred to v0.0.2
- Fates export to `.opencode/agents/`
- Skills export to `.opencode/skills/norn-<name>/`
- Tools generate fate permissions but are not exported

### Documentation

- `docs/README.md` - Project overview and command reference
- `docs/getting-started.md` - Quick start guide
- `docs/architecture.md` - Architecture and design
- `docs/fates.md` - Fate roles and generation
- `docs/opencode.md` - OpenCode integration guide
- `docs/opencode-integration.md` - Living specification
- `.norn/memory/` - Decision records and session bootstrap

## Deferred to v0.0.2

- Planning branch mode
- Shared vs local overlays
- SQLite backend
- Diff previews and iterative refinement
- Public/private repo sharing strategies
- External tool integration (Fizzy boards, etc.)
