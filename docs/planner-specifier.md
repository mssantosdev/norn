---
title: Planner and Specifier Concept
summary: Define the optional planner/specifier specialist concept for Norn v0.0.1.
fizzy: 30
tags:
  - docs
  - planning
  - v0.0.1
---

## Goal

Define the planner/specifier concept clearly enough that users and agents can reference it without treating it as a shipped core runtime role.

## User Stories

- As a user, I want to know when planning/specification work should stay with the core fates and when a specialist planner/specifier concept may help.
- As an agent, I want explicit boundaries so I do not confuse planning/documentation help with the shipped core fate model.
- As a maintainer, I want the concept documented now without overcommitting to a full invocation workflow in v0.0.1.

## Definition

Planner/specifier is an optional specialist concept for planning, specification, and related documentation work.

It is intended to help with:

- refining weaves and threads
- drafting or tightening patterns and design docs
- improving acceptance criteria, scope notes, and user stories
- clarifying documentation boundaries before implementation

It is not a replacement for the core fate model and is not required for normal Norn operation.

## v0.0.1 Boundaries

In v0.0.1, planner/specifier is:

- documented as a concept
- allowed as a future specialist workflow or generated assistant
- useful for planning and documentation surfaces only

In v0.0.1, planner/specifier is not:

- a fifth core fate
- a required runtime actor
- a dedicated CLI command family
- a mandatory OpenCode export target
- a replacement for keeper ownership of coordination or weaver ownership of implementation

## Relationship To Core Fates

- `keeper` still coordinates planning flow, assignments, and handoffs
- `weaver` still owns implementation of assigned threads
- `judge` still reviews changes for correctness and quality
- `fates` still handles integration and release transitions

Planner/specifier may assist with artifact quality before implementation, but it does not take authority away from the core fates.

## Surfaces It May Touch

Planner/specifier is conceptually allowed to operate on:

- `loom/weaves/`
- `loom/docs/`
- `loom/patterns/`
- local planning overlays under `.norn/loom/`
- related documentation in `docs/`

It should not be treated as the owner of runtime coordination data in `.norn/spindle/`.

## Invocation Status

Invocation workflow is intentionally deferred.

That means v0.0.1 does not define:

- a dedicated `norn planner` command
- a dedicated `norn specifier` command
- automatic routing rules that select planner/specifier at runtime
- required agent generation for planner/specifier

If later added, the workflow should still follow the same Norn rules:

- interactive and non-interactive support
- readable artifact outputs
- preview before write for assisted drafting
- targeted clarification instead of broad re-questioning

## Acceptance

- the concept is documented in repo docs and planning artifacts
- boundaries with the four core fates are explicit
- v0.0.1 exclusions are explicit
- future workflow additions remain possible without changing the core fate model
