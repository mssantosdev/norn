# Skald

The `skald` is the fifth core fate in Norn, named after the Norse poets and storytellers who composed structured, formal narratives.

## Role

Skald specializes in planning, specification, and pattern definition. It helps generate and refine project artifacts within the Norn Standard and project-specific conventions.

## Responsibilities

- **Plan generation**: Create weaves and threads with clear goals, user stories, scope, and acceptance criteria
- **Specification drafting**: Write ADRs, code conventions, and Norn Patterns
- **Implementation guidance**: Help weavers with solutions when stuck
- **Run assistance**: Support planning, scoping, and refinement during active development
- **Quality assurance**: Ensure all artifacts follow the documentation standard

## Interaction

Skald is primarily an OpenCode-facing agent. Users interact with it through:

- OpenCode chat: *"skald, help me plan the authentication weave"*
- Sub-agent invocation during assisted workflows

There is no dedicated `norn skald` CLI command. The skald operates through the OpenCode integration.

## Export

When OpenCode integration is enabled, the skald agent definition is exported to:

```
.opencode/agents/skald.md
```

## Relationship to Core Fates

- `keeper` still coordinates planning flow, assignments, and handoffs
- `weaver` still owns implementation of assigned threads
- `judge` still reviews changes for correctness, standards, and constraints
- `fates` still handles integration and release transitions

Skald assists with artifact quality before implementation but does not take authority away from the core fates.

## Documentation Standard

All skald-generated artifacts must include:

- CLI --help text (for commands)
- Project docs (user-facing guides)
- Specifications (for AI agent consumption)
- Integration boundaries (what Norn owns vs what OpenCode owns)
