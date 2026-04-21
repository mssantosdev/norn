---
title: Config Scopes and Preferences
summary: Define the minimal configuration scope model and preference fields for Norn v0.0.1.
fizzy: 31
tags:
  - weave
  - v0.0.1
  - cli
  - ui
  - opencode
  - docs
---

## Goal

Define and implement a minimal but extensible config model for Norn that supports global defaults, workspace configuration, and optional private local overrides.

## User Stories

- As a user, I want Norn to remember my preferred language, theme, and drafting preferences so I do not have to repeat them in every workspace.
- As a workspace owner, I want shared config defaults for a project so collaborators and agents behave consistently.
- As an agent, I want config resolution to be explicit and readable so I can understand which values are in effect before acting.

## Scope

- define global, workspace, and optional private-local config scopes
- define precedence rules between scopes
- define minimal v0.0.1 preference fields
- document how config should influence CLI, TUI, and OpenCode-assisted responses

## Acceptance

- config scope model is documented
- v0.0.1 preference fields are defined
- config inspection and editing support both interactive and non-interactive flows
- config resolution exposes full origin data for effective values
- planned config work is tracked in repo planning and Fizzy
- future advanced config work is explicitly deferred

## Related Paths

- `docs/architecture.md`
- `.norn/runes.yaml`
- `.norn/loom/memory/roadmap-v0.0.1.md`

## Notes

This weave defines the config model first. Implementation details can remain split into focused threads.
