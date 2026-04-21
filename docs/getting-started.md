# Getting Started

## Initialize

Interactive mode:

```bash
norn init
```

Non-interactive folder mode:

```bash
norn init --no-interactive --name=my-project --mode=folder --enable-opencode
```

Non-interactive planning branch mode:

```bash
norn init --no-interactive --mode=branch --branch=loom --create-branch
```

## Inspect

```bash
norn status
norn detect
norn fates list
norn commands list
norn runes show
norn runes resolve
norn runes resolve --format=yaml
norn warps list
norn warps list --view=runtime
```

## Edit config

Interactive:

```bash
norn runes edit
```

Non-interactive:

```bash
norn runes edit --scope=workspace --set preferences.language=pt-BR
norn runes edit --scope=local --unset opencode.response_language
```

## Add shared planning artifacts

```bash
norn patterns add "API Contract" "Document the API expectations"
norn skills add "Deploy Flow" "Document the deployment path"
norn commands add lint lint "npm run lint"
```

## Validate OpenCode availability

```bash
norn chat validate
```

## Runtime coordination

Create a warp:

```bash
norn warps add --status=active --owner=marcus "API Warp" "Runtime coordination for API lane"
```

Assign active work to a warp:

```bash
norn warps assign --kind=thread --id=add-weaves-cli --warp=api-warp --owner=marcus --state=active
norn warps assignment show thread add-weaves-cli
norn warps list --view=runtime
```
