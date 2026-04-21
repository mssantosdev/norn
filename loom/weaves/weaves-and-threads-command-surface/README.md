---
title: Weaves and Threads Command Surface
summary: Add the missing weaves and threads command families so Norn can manage its core planning model directly from the CLI.
fizzy: 18
tags:
  - weave
  - v0.0.1
  - cli
  - testing
---

## Goal

Add `weaves` and `threads` command families that operate on Norn planning artifacts instead of relying only on patterns, skills, and commands.

## User Stories

- As a user, I want to create and navigate weave and thread planning artifacts through the CLI so I do not have to build the file structure manually.
- As an agent, I want weave and thread artifacts to be readable and consistently structured so I can reference, execute, and hand off work with minimal ambiguity.
- As a maintainer, I want shared planning artifacts to live in a durable location and local working notes to live in an overlay so collaboration and experimentation can coexist safely.

## Scope

- add `weaves list|add|show|remove`
- add `threads list|add|show|remove`
- store shared planning artifacts under `loom/weaves/`
- define default templates for weave and thread artifacts
- align CLI, docs, and tests with the Norn planning model

## Acceptance

- weave artifacts can be created, listed, shown, and removed
- thread artifacts can be created, listed, shown, and removed
- thread artifacts are grouped under a weave
- docs reflect the artifact model
- integration and e2e tests cover the basic flow

## Related Patterns

- planning vs capability artifact taxonomy
- shared vs local planning overlay model

## Related Paths

- `docs/architecture.md`
- `.norn/loom/memory/decisions.md`
- `.norn/loom/memory/roadmap-v0.0.1.md`

## Notes

This weave establishes planning artifact structure first. Runtime coordination remains a later slice.

## Template Expectations

The default CLI-generated weave and thread artifacts in this slice should support:

- title
- summary
- goal
- user stories
- scope or strands
- acceptance

## AI Assistance Expectations

When Norn later offers AI-assisted weave or thread creation, the first prompt should attempt to produce the full structured draft for the artifact. Clarification should be limited to missing or weak sections.
