---
title: Config Command Surface
summary: Define the initial config-facing CLI behavior for inspection and editing.
weave: config-scopes-and-preferences
---

## Goal

Define how users, power users, and agents inspect and edit configuration through the CLI using both interactive and non-interactive flows.

## User Story

As a user, I want to inspect effective config values and understand where they come from before changing them.

## Acceptance

- the desired command surface is documented
- inspection and resolution behaviors are identified
- all config commands support both interactive and non-interactive use
- `resolve` exposes full origin data for effective values
- `edit` supports scope selection and scope-local unsetting

## Strands

- define read commands such as `show` and `resolve`
- define write/edit expectations for `global`, `workspace`, and `local` scopes
- define how full field origin is exposed

## Command Surface

- `norn runes show`
- `norn runes show --scope=global|workspace|local`
- `norn runes resolve`
- `norn runes edit`
- `norn runes edit --scope=global|workspace|local`
- `norn runes edit --scope=... --set path=value`
- `norn runes edit --scope=... --unset path`

## Notes

- `show` without scope returns the effective merged configuration
- `show --scope=...` returns the raw contents of the selected scope file
- `resolve` returns effective values plus origin metadata for each field
- `edit` without flags is interactive and prompts for scope selection
- `edit` writes only the selected scope file and does not copy inherited values into that scope
- blank fields in interactive edit mean inherit/unset for that scope where applicable
