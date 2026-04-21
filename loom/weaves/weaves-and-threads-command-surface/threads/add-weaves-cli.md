---
title: Add Weaves CLI
summary: Implement the weaves command family and shared weave artifact storage.
weave: weaves-and-threads-command-surface
---

## Goal

Implement `weaves list|add|show|remove`.

## User Story

As a user, I want `weaves` commands to create and manage durable weave artifacts so planning can be created through the CLI instead of handwritten directory setup.

## Acceptance

- weave markdown files live under `loom/weaves/<weave-id>/README.md`
- CLI uses configured planning roots rather than hardcoded literals
- default weave template is defined and used
- integration coverage exists

## Strands

- define weave artifact wrapper
- define weave template
- add CLI handlers
- add tests

## Template Fields

The default weave template should support:

- title
- summary
- goal
- user stories
- scope
- acceptance

## Related Files

- `internal/cli/cli.go`
- `internal/weaves/`
- `test/integration/`
