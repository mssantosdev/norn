# Constitution

## Mission

Norn should define a practical multi-agent planning and coordination harness that separates durable planning truth, reusable capability artifacts, runtime coordination, and memory continuity.

## Artifact Rule

- planning artifacts define durable scope truth
- capability artifacts define reusable agent-facing capabilities
- runtime coordination artifacts define active execution truth
- memory artifacts preserve rationale and continuity across sessions

## Readability Rule

All artifact classes should remain readable, searchable, and easy to index for both humans and AI agents.

## Planning Rule

- shared planning lives in `loom/`
- local planning overlays live in `.norn/loom/`
- local overlays may refine shared work but should not silently replace shared truth

## Runtime Rule

Runtime coordination lives in `.norn/spindle/` and should be treated as operational memory rather than permanent scope truth.
