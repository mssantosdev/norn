# Norn Agent Rules

Read this file before doing non-trivial work in this repository.

## Core Context

- Norn is a multi-agent planning and coordination harness.
- Shared planning lives in `loom/`.
- Local-only overlays live in `.norn/loom/`.
- Runtime coordination lives in `.norn/spindle/`.
- Durable handoff and decision memory lives in `.norn/loom/memory/`.

Start with:

1. `docs/README.md`
2. `docs/architecture.md`
3. `.norn/loom/memory/decisions.md`
4. `.norn/loom/memory/conversation-summary.md`
5. `.norn/loom/memory/roadmap-v0.0.1.md`
6. `.norn/loom/memory/patterns.md`
7. `.norn/loom/memory/next-session-bootstrap.md`

## Fizzy Awareness

This project also uses Fizzy as a shared task tracker for Norn development.

- Board name: `Norn`
- Use Fizzy to understand active roadmap cards, deferred spikes, and shared execution context.
- Fizzy is a tracking surface for Norn development work. It is not the architectural source of truth.
- Architecture, standards, and local handoff context remain in repository files.

Read-first expectation:

- Before substantial implementation, check relevant Fizzy cards when the work appears roadmap-related or when the user references board work.
- Prefer read-only Fizzy actions by default: list cards, show cards, read comments, inspect steps.

## Fizzy Write Policy

Do not create, edit, move, close, postpone, tag, assign, or comment on Fizzy cards unless the user explicitly asks or clearly authorizes it.

Default behavior:

- read-only Fizzy usage is allowed for context
- write actions to Fizzy require explicit user instruction or approval

## Current Relevant Cards

These are the primary roadmap cards currently relevant to implementation:

- `#17` `[weave] v0.0.1 bootstrap foundation`
- `#18` `[weave] weaves and threads command surface`
- `#28` `[weave] warps and runtime coordination views`
- `#25` `[weave] interactive editors for managed artifacts`
- `#22` `[weave] OpenCode export and assisted workflows`
- `#23` `[weave] judge and fates review integration flow`
- `#19` `[thread] planning branch UX refinement`
- `#21` `[release] v0.0.1 readiness`
- `#30` `[thread] planner/specifier concept for v0.0.1`
- `#31` `[thread] config scopes and preferences for v0.0.1`

## Resume Guidance

If resuming from a new machine or a fresh chat session:

- read `.norn/loom/memory/next-session-bootstrap.md`
- re-run `go test ./...`
- re-run `go build ./cmd/norn`
- check Fizzy cards `#25`, `#28`, `#30`, and `#32`

Current recommended next implementation target:

- `#25` interactive editors for managed artifacts

Already completed recently:

- `#32` config scopes and preferences for v0.0.1
- `#30` planner/specifier concept for v0.0.1
- the first major slices of `#28` warps and runtime coordination views

Deferred cards are intentionally in `Not Now`.

## Working Rules

- Keep changes minimal and aligned with the current v0.0.1 roadmap.
- Prefer integration and e2e-style verification over low-value unit tests.
- Keep command output and user-facing messaging aligned with Charm log usage.
- When in doubt, update repository memory files before relying on Fizzy comments as the only record.
- For implementation planning, treat `loom/` and `.norn/loom/` as the real planning surfaces. Fizzy cards are tracking context, not full implementation plans.
- Treat `.norn/fates/`, `.norn/skills/`, and `.norn/commands/` as capability surfaces rather than planning artifacts.
