# Decisions

## ADR-001: Name

The project is named `Norn` and uses Norse weaving terminology for core concepts.

## ADR-002: UI

The CLI is interactive-first and uses Charm tooling. Non-interactive pathways remain available through explicit flags and parameters.

## ADR-003: Planning Surfaces

Shared plans live in `loom/` by default. Local-only plans live in `.norn/loom/`. Runtime coordination lives in `.norn/spindle/`.

## ADR-004: Roles

The v0.0.1 fate set is `keeper`, `weaver`, `judge`, and `fates`.

## ADR-005: Config Scopes

Norn uses layered config scopes in this order: global user config, workspace config, local private override, then built-in defaults when unset.

## ADR-006: Runtime Coordination

Runtime coordination lives under `.norn/spindle/` and is local operational memory.

Warp lane records live under `.norn/spindle/warps/`.

Weave/thread ownership records live under `.norn/spindle/weaves/` and `.norn/spindle/threads/`.

These runtime records are not durable planning truth.
