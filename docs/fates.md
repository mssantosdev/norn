# Fates

Norn v0.0.1 ships four core fates.

Planner/specifier is not a core fate in v0.0.1. It is an optional planning/specification concept documented for future workflow expansion.

## keeper

Coordinates planning, assignments, and handoffs.

## weaver

Implements threads and validates owned work.

## judge

Reviews changes for:

- architecture
- code standards
- security
- performance
- test coverage

## fates

Integrates approved work and owns merge/release transitions.

## Generation Model

Norn stores fate source data in `.norn/fates/*.yaml` and renders OpenCode-compatible agent files into `.opencode/agents/`.

Fate permissions are derived from the tool registry (`.norn/tools/*.yaml`).

Planner/specifier is currently outside this generated core fate set.
