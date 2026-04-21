# Architecture

## Core Terms

- `weave` - epic or large objective
- `thread` - work item
- `strand` - step inside a thread
- `fate` - agent role
- `pattern` - specification or design artifact
- `skill` - reusable capability
- `rune` - configuration
- `warp` - worktree lane
- `loom` - durable planning surface
- `spindle` - runtime coordination surface

## Planning Model

Shared planning lives in `loom/` by default.

Local-only planning can live in `.norn/loom/` and is overlaid with the shared plan set.

When branch mode is selected, Norn creates or reuses a planning branch and checks it out as a dedicated worktree, defaulting to `.loom/`.

## Artifact Taxonomy

Norn separates artifacts into four classes.

### Planning Artifacts

Planning artifacts are durable scope truth. They define what should be built, why it matters, and how acceptance is evaluated.

Examples:

- `constitution`
- `weave`
- `thread`
- `pattern`
- durable reference and design docs

Locations:

- shared: `loom/`
- local overlay: `.norn/loom/`

### Capability Artifacts

Capability artifacts define how agents operate and what reusable capabilities Norn manages for them.

Examples:

- `fate`
- `skill`
- `command`

Locations:

- `.norn/fates/`
- `.norn/skills/`
- `.norn/commands/`

These artifacts should remain plain-text or structured-text and easy to search, index, and render.

An optional planner/specifier specialist may later operate on planning artifacts and related documentation, but it is not a core fate in v0.0.1.

Planner/specifier is a documentation and planning concept in v0.0.1, not a shipped runtime role.

Its intended scope is:

- improving weave and thread quality
- drafting or refining patterns and planning docs
- tightening user stories, scope notes, and acceptance criteria

Its explicit v0.0.1 exclusions are:

- no fifth core fate
- no dedicated command family
- no required OpenCode export target
- no automatic invocation or routing workflow

If introduced later as a workflow, it should remain subordinate to the core fate model rather than replacing it.

### Runtime Coordination Artifacts

Runtime coordination artifacts hold active execution truth.

Examples:

- current owner
- active state
- blockers
- handoffs
- review notes
- warp/runtime metadata

Location:

- `.norn/spindle/`

Runtime coordination records are operational memory, not durable scope truth.

### Memory Artifacts

Memory artifacts preserve compacted history, rationale, and handoff continuity across sessions.

Location:

- `.norn/loom/memory/`

## Default Directory Standard

### Shared Planning

- `loom/README.md`
- `loom/constitution.md`
- `loom/weaves/<weave-id>/README.md`
- `loom/weaves/<weave-id>/threads.md`
- `loom/weaves/<weave-id>/threads/<thread-id>.md`
- `loom/patterns/<pattern-id>.md`
- `loom/docs/<doc-id>.md`

### Local Planning Overlay

- `.norn/loom/weaves/<weave-id>/README.md`
- `.norn/loom/weaves/<weave-id>/threads.md`
- `.norn/loom/weaves/<weave-id>/threads/<thread-id>.md`
- `.norn/loom/patterns/<pattern-id>.md`
- `.norn/loom/docs/<doc-id>.md`

### Capability Artifacts

- `.norn/fates/<fate-id>.yaml`
- `.norn/skills/<skill-id>.md`
- `.norn/commands/<command-id>.yaml`

### Runtime Coordination

- `.norn/spindle/weaves/<weave-id>.yaml`
- `.norn/spindle/threads/<thread-id>.yaml`
- `.norn/spindle/warps/<warp-id>.yaml`

The first implemented runtime slice is warp lane records stored as readable yaml files under `.norn/spindle/warps/`.

The initial warp command surface is:

- `norn warps list`
- `norn warps list --view=runtime`
- `norn warps add`
- `norn warps add --status=... --owner=... --root=... --branch=... --weaves=a,b --threads=a,b <title> <summary>`
- `norn warps assign --kind=weave|thread --id=<artifact-id> --warp=<warp-id> [--owner=...] [--state=...] [--notes=...]`
- `norn warps assignment show <weave|thread> <id>`
- `norn warps assignment remove <weave|thread> <id>`
- `norn warps show <warp-id>`
- `norn warps remove <warp-id>`

Like other Norn command families, warp management supports both interactive and non-interactive creation.

The first cross-warp runtime view is an ownership index derived from local spindle assignment records under:

- `.norn/spindle/weaves/*.yaml`
- `.norn/spindle/threads/*.yaml`

These records summarize which warp currently owns active weave or thread work and in what state.

Direct runtime assignment management currently includes:

- create through `norn warps assign`
- inspect through `norn warps assignment show <weave|thread> <id>`
- remove through `norn warps assignment remove <weave|thread> <id>`

### Memory

- `.norn/loom/memory/decisions.md`
- `.norn/loom/memory/conversation-summary.md`
- `.norn/loom/memory/roadmap-v0.0.1.md`
- `.norn/loom/memory/patterns.md`

### Configuration

Configuration should support layered scopes:

- global user defaults in `~/.config/norn/runes.yaml`
- workspace configuration in `.norn/runes.yaml`
- optional private local override in `.norn/runes.local.yaml` for machine- or user-specific settings

Preferred precedence should be:

- private local override
- workspace config
- global user config
- built-in defaults

Config management commands should support both interactive and non-interactive flows.

The initial command surface is:

- `norn runes show`
- `norn runes show --scope=global|workspace|local`
- `norn runes resolve`
- `norn runes resolve --format=table|yaml`
- `norn runes edit`
- `norn runes edit --scope=... --set path=value`
- `norn runes edit --scope=... --unset path`

`resolve` should expose the effective value plus full origin data for each field.

Default resolve output should be optimized for humans, while a structured YAML mode remains available for machines and tooling.

## Template Rule

Every managed artifact type should eventually have:

- a standard location
- a standard schema or template
- a standard CLI create/edit/read workflow

All artifacts should remain readable and easy to index and reference for both AI agents and humans.

## Planning Template Rule

Feature-facing planning and specification artifacts should include user stories when they materially improve understanding of the feature or workflow.

At minimum, weave, thread, and feature-facing pattern templates should support:

- title
- summary
- goal
- user stories
- scope
- acceptance

## AI-Assisted Creation Rule

AI-assisted artifact creation should be structured and contract-based.

Default behavior:

- start with one primary prompt per artifact
- ask targeted clarification questions only when required sections remain incomplete or weak
- preview generated sections before writing files
- preserve user-provided content and avoid re-asking for sections the user already made clear

The CLI/OpenCode interaction should use a standardized request/response contract so clarification and insertion remain predictable.

### Preferred AI Draft Contract

The first prompt should attempt to draft as much of the artifact as possible, typically including:

- title
- summary
- goal
- scope
- user stories
- acceptance
- optional suggested child artifacts such as default threads

If follow-up is needed, clarification should be field-specific rather than broad or repetitive.

## Runtime Model

Runtime coordination lives under `.norn/spindle/` and is local to the root execution surface for the repo or workspace.

Warp records are intentionally local operational memory. They should not be treated as durable planning truth and do not replace weave/thread planning artifacts.

Spindle assignment records follow the same rule: they reflect active execution ownership and state, not durable scope truth.

## Workspace Modes

- `repo` mode is used for a single repository root.
- `workspace` mode is auto-selected when `.hydra.yaml` is present at the root.

In workspace mode, Norn coordinates from the workspace root and ignores nested `.norn/` directories inside child worktrees for v0.0.1.

## Generated Assets

- `.norn/runes.yaml` - workspace configuration
- `.norn/fates/*.yaml` - Norn-managed fate sources
- `.norn/commands/*.yaml` - command registry used for permission generation
- `.opencode/agents/*.md` - generated OpenCode-compatible fate agents

## Pending Config Scope

Planned v0.0.1 config fields should include at least:

- preferred language
- theme
- verbosity
- default planning surface
- AI response language
- drafting and confirmation preferences
