---
title: Minimal Warp Records
summary: Define and implement the first runtime warp record flow for v0.0.1.
weave: warps-and-runtime-coordination
---

## Goal

Ship the smallest useful warp runtime record flow before moving into richer cross-warp coordination views.

## User Story

As a user, I want simple local warp records so I can register active worktree lanes and inspect their current state.

## Strands

- define a readable warp yaml structure
- add `warps list|add|show|remove`
- support both interactive and non-interactive creation
- keep records local under `.norn/spindle/`
- add a first runtime ownership index for weave/thread assignments by warp
- add direct assignment management for show/remove flows

## Acceptance

- warp records are readable yaml files
- `warps add` works with flags and with interactive prompts
- `warps list` and `warps show` expose the saved state
- runtime view can summarize assigned weave/thread work across warps
- runtime assignment records can be inspected and removed directly
- tests cover the initial flow
