---
title: Add Threads CLI
summary: Implement the threads command family and thread storage grouped by weave.
weave: weaves-and-threads-command-surface
---

## Goal

Implement `threads list|add|show|remove`.

## User Story

As a user, I want `threads` commands to create executable work items under a weave so work can be broken down consistently and referenced by humans and agents.

## Acceptance

- thread markdown files live under `loom/weaves/<weave-id>/threads/<thread-id>.md`
- commands require a weave id
- thread artifacts are grouped by weave
- default thread template is defined and used
- integration coverage exists

## Strands

- define thread artifact wrapper
- define thread template
- add CLI handlers
- add tests

## Template Fields

The default thread template should support:

- title
- summary
- goal
- user story
- strands
- acceptance

## Related Files

- `internal/cli/cli.go`
- `internal/threads/`
- `test/integration/`
