---
title: Add Tests and Docs
summary: Cover the new command families with integration/e2e tests and update docs.
weave: weaves-and-threads-command-surface
---

## Goal

Document and validate the new planning command surface.

## User Story

As a maintainer, I want the weave/thread command surface to be documented and tested so future contributors and agents can trust the standard artifact model.

## Acceptance

- integration tests cover weaves CRUD
- integration tests cover threads CRUD
- e2e flow includes creating a weave and a thread
- docs/README.md and docs/architecture.md reflect the new commands and artifact locations

## Strands

- add integration coverage
- extend e2e workflow
- update docs
- update memory and handoff references if needed
