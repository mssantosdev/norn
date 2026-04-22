# Artifact Autocomplete and Selection

## Overview

Norn provides multiple ways to select artifacts (weaves, threads, warps, fates, tools, patterns, skills) without memorizing exact IDs:

1. **Interactive fuzzy find** — filterable list when no ID is provided
2. **Shell completions** — tab-complete IDs in bash/zsh/fish
3. **Partial match** — substring matching with confirmation

## Interactive Fuzzy Find

When a command expects an artifact ID but receives none, Norn presents an interactive filterable list:

```bash
$ norn weaves show
Select a weave:
> [1] api-auth-weave    — Secure all API endpoints with JWT
  [2] rate-limit-weave  — Implement request rate limiting
  [3] caching-weave     — Add Redis caching layer
  Type to filter...
```

**Supported commands:**
- `norn weaves show` — select from all weaves
- `norn weaves remove` — select from all weaves
- `norn threads list` — select weave, then list threads
- `norn threads show` — select weave, then thread
- `norn threads remove` — select weave, then thread
- `norn warps show` — select from all warps
- `norn warps remove` — select from all warps
- `norn fates show` — select from all fates
- `norn fates edit` — select from all fates
- `norn fates remove` — select from all fates
- `norn tools show` — select from all tools
- `norn tools edit` — select from all tools
- `norn tools remove` — select from all tools
- `norn patterns show` — select from all patterns
- `norn patterns edit` — select from all patterns
- `norn patterns remove` — select from all patterns
- `norn skills show` — select from all skills
- `norn skills edit` — select from all skills
- `norn skills remove` — select from all skills

## Shell Completions

Generate completion scripts for your shell:

```bash
# Bash
norn completion bash > ~/.config/norn/completions.bash
echo 'source ~/.config/norn/completions.bash' >> ~/.bashrc

# Zsh
norn completion zsh > ~/.config/norn/completions.zsh
echo 'source ~/.config/norn/completions.zsh' >> ~/.zshrc

# Fish
norn completion fish > ~/.config/fish/completions/norn.fish
```

Completions include:
- Commands and flags
- Dynamic artifact IDs from the current workspace

## Partial Match

For quick CLI usage, provide a partial ID:

```bash
$ norn weaves show auth
# Matches api-auth-weave (if unique)

$ norn threads show auth jwt
# Matches api-auth-weave + jwt-middleware
```

If the match is ambiguous, Norn shows the matching options and asks for confirmation.

## Implementation

- `internal/cli/cli.go` — `promptArtifactSelection()` helper
- `internal/cli/cli.go` — `promptWeaveSelection()` helper
- `internal/cli/completion.go` — shell completion generation (Phase 2)
