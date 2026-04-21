---
title: Warps and Runtime Coordination
summary: Add minimal warp records and runtime coordination views under the spindle surface.
fizzy: 28
tags:
  - weave
  - runtime
  - v0.0.1
  - cli
  - docs
---

## Goal

Add a minimal but useful runtime coordination slice so Norn can track active warp lanes locally under the spindle surface.

## User Stories

- As a user, I want to register active warp lanes so I can see which execution surfaces exist for the current repo or workspace.
- As an agent, I want runtime warp records to be local, readable, and easy to inspect before I coordinate work across lanes.

## Scope

- define the initial warp record shape
- store warp records under `.norn/spindle/warps/`
- add a CLI surface for list, add, show, and remove
- support both interactive and non-interactive warp creation
- add a first cross-warp runtime ownership view for weave and thread assignments
- make runtime assignment records directly manageable from the CLI

## Acceptance

- `norn warps` command surface exists
- warp records are stored under `.norn/spindle/warps/`
- runtime assignment views can show active weave/thread ownership by warp
- runtime assignment records can be shown and removed directly
- docs explain the local runtime nature of warp records
- integration tests cover the first CRUD flow

## Related Paths

- `.norn/spindle/`
- `internal/warps/`
- `internal/cli/warps.go`
