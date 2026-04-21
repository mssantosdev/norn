---
title: Global, Workspace, and Local Config Model
summary: Define config scopes and precedence.
weave: config-scopes-and-preferences
---

## Goal

Define how Norn resolves configuration across global defaults, workspace config, and optional private local overrides.

## User Story

As a user, I want predictable config precedence so I know where to change settings and what will win when multiple values exist.

## Acceptance

- global, workspace, and local scopes are documented
- precedence order is explicit
- private local config is identified as non-shared

## Strands

- define scope locations
- define precedence rules
- document intended use for each scope
