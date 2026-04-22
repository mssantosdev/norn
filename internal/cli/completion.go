package cli

import (
	"fmt"
)

func runCompletion(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: norn completion <bash|zsh|fish>")
	}
	shell := args[0]
	switch shell {
	case "bash":
		fmt.Print(bashCompletion())
	case "zsh":
		fmt.Print(zshCompletion())
	case "fish":
		fmt.Print(fishCompletion())
	default:
		return fmt.Errorf("unknown shell %q; expected bash, zsh, or fish", shell)
	}
	return nil
}

func bashCompletion() string {
	return `#!/bin/bash
_norn() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Top-level commands
    local commands="init status detect fates patterns skills tools weaves threads warps runes chat export completion"
    
    # If completing first argument
    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
        return 0
    fi
    
    local cmd="${COMP_WORDS[1]}"
    local subcmd=""
    if [ $COMP_CWORD -ge 3 ]; then
        subcmd="${COMP_WORDS[2]}"
    fi
    
    # Dynamic artifact completion
    local artifact_ids=""
    if command -v norn &> /dev/null; then
        case "${cmd}_${subcmd}" in
            weaves_show|weaves_remove)
                artifact_ids=$(norn weaves list 2>/dev/null | tail -n +2 | awk '{print $1}')
                ;;
            threads_list|threads_show|threads_remove)
                artifact_ids=$(norn weaves list 2>/dev/null | tail -n +2 | awk '{print $1}')
                ;;
            warps_show|warps_remove)
                artifact_ids=$(norn warps list 2>/dev/null | tail -n +2 | awk '{print $1}')
                ;;
            fates_show|fates_edit|fates_remove)
                artifact_ids=$(norn fates list 2>/dev/null | tail -n +2 | awk '{print $1}')
                ;;
            tools_show|tools_edit|tools_remove)
                artifact_ids=$(norn tools list 2>/dev/null | tail -n +2 | awk '{print $1}')
                ;;
            patterns_show|patterns_edit|patterns_remove)
                artifact_ids=$(norn patterns list 2>/dev/null | tail -n +2 | awk '{print $1}')
                ;;
            skills_show|skills_edit|skills_remove)
                artifact_ids=$(norn skills list 2>/dev/null | tail -n +2 | awk '{print $1}')
                ;;
        esac
    fi
    
    if [ -n "${artifact_ids}" ]; then
        COMPREPLY=( $(compgen -W "${artifact_ids}" -- ${cur}) )
        return 0
    fi
    
    # Flag completion
    local flags=""
    case "${cmd}" in
        init)
            flags="--no-interactive --enable-opencode --name --theme --languages --tools --frameworks --model --agent --skeleton --prompt"
            ;;
        runes)
            flags="--scope --set --unset"
            ;;
        export)
            flags="--opencode --output"
            ;;
        chat)
            flags="--prompt"
            ;;
        weaves|threads|warps|fates|tools|patterns|skills)
            flags="--help --format"
            ;;
    esac
    
    if [[ ${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "${flags}" -- ${cur}) )
        return 0
    fi
}

complete -F _norn norn
`
}

func zshCompletion() string {
	return `#compdef norn

_norn() {
    local curcontext="$curcontext" state line
    typeset -A opt_args
    
    _arguments -C \
        '1: :_norn_commands' \
        '*::arg:->args'
    
    case $line[1] in
        init)
            _arguments \
                '--no-interactive[Run in non-interactive mode]' \
                '--enable-opencode[Enable OpenCode integration]' \
                '--name[Project name]:name:' \
                '--theme[UI theme]:theme:' \
                '--languages[Comma-separated language list]:languages:' \
                '--tools[Comma-separated tool list]:tools:' \
                '--frameworks[Comma-separated framework list]:frameworks:' \
                '--model[OpenCode model]:model:' \
                '--agent[OpenCode agent]:agent:' \
                '--skeleton[Project scaffold]:skeleton:(standard empty)' \
                '--prompt[OpenCode assisted init prompt]:prompt:'
            ;;
        weaves)
            _norn_weaves
            ;;
        threads)
            _norn_threads
            ;;
        warps)
            _norn_warps
            ;;
        fates)
            _norn_fates
            ;;
        tools)
            _norn_tools
            ;;
        patterns)
            _norn_patterns
            ;;
        skills)
            _norn_skills
            ;;
        runes)
            _arguments \
                '--scope[Config scope]:scope:(global workspace local)' \
                '--set[Set config value]:path=value:' \
                '--unset[Unset config value]:path:'
            ;;
        completion)
            _arguments '1: :_norn_shells'
            ;;
    esac
}

_norn_commands() {
    local commands=(
        "init:Bootstrap a new Norn project"
        "status:Show workspace status"
        "detect:Detect project stack"
        "fates:Manage agent fates"
        "patterns:Manage patterns"
        "skills:Manage skills"
        "tools:Manage tools"
        "weaves:Manage weaves"
        "threads:Manage threads"
        "warps:Manage warps"
        "runes:Manage configuration"
        "chat:OpenCode integration"
        "export:Export artifacts"
        "completion:Generate shell completions"
    )
    _describe -t commands 'norn command' commands
}

_norn_shells() {
    local shells=(bash zsh fish)
    _describe -t shells 'shell' shells
}

_norn_weaves() {
    local curcontext="$curcontext" state line
    typeset -A opt_args
    
    _arguments -C \
        '1: :_norn_weave_commands' \
        '*:: :->args'
    
    case $line[1] in
        show|remove)
            _norn_weave_ids
            ;;
    esac
}

_norn_weave_commands() {
    local commands=(list add show remove)
    _describe -t commands 'weave command' commands
}

_norn_weave_ids() {
    local ids
    ids=(${(f)"$(norn weaves list 2>/dev/null | tail -n +2 | awk '{print $1}')"})
    _describe -t ids 'weave id' ids
}

_norn_threads() {
    local curcontext="$curcontext" state line
    typeset -A opt_args
    
    _arguments -C \
        '1: :_norn_thread_commands' \
        '*:: :->args'
    
    case $line[1] in
        list|show|remove)
            _norn_weave_ids
            ;;
    esac
}

_norn_thread_commands() {
    local commands=(list add show remove)
    _describe -t commands 'thread command' commands
}

_norn_warps() {
    local curcontext="$curcontext" state line
    typeset -A opt_args
    
    _arguments -C \
        '1: :_norn_warp_commands' \
        '*:: :->args'
    
    case $line[1] in
        show|remove)
            _norn_warp_ids
            ;;
    esac
}

_norn_warp_commands() {
    local commands=(list add assign assignment show remove)
    _describe -t commands 'warp command' commands
}

_norn_warp_ids() {
    local ids
    ids=(${(f)"$(norn warps list 2>/dev/null | tail -n +2 | awk '{print $1}')"})
    _describe -t ids 'warp id' ids
}

_norn_fates() {
    local curcontext="$curcontext" state line
    typeset -A opt_args
    
    _arguments -C \
        '1: :_norn_fate_commands' \
        '*:: :->args'
    
    case $line[1] in
        show|edit|remove)
            _norn_fate_names
            ;;
    esac
}

_norn_fate_commands() {
    local commands=(list show add edit remove)
    _describe -t commands 'fate command' commands
}

_norn_fate_names() {
    local names
    names=(${(f)"$(norn fates list 2>/dev/null | tail -n +2 | awk '{print $1}')"})
    _describe -t names 'fate name' names
}

_norn_tools() {
    local curcontext="$curcontext" state line
    typeset -A opt_args
    
    _arguments -C \
        '1: :_norn_tool_commands' \
        '*:: :->args'
    
    case $line[1] in
        show|edit|remove)
            _norn_tool_ids
            ;;
    esac
}

_norn_tool_commands() {
    local commands=(list add show edit remove)
    _describe -t commands 'tool command' commands
}

_norn_tool_ids() {
    local ids
    ids=(${(f)"$(norn tools list 2>/dev/null | tail -n +2 | awk '{print $1}')"})
    _describe -t ids 'tool id' ids
}

_norn_patterns() {
    local curcontext="$curcontext" state line
    typeset -A opt_args
    
    _arguments -C \
        '1: :_norn_pattern_commands' \
        '*:: :->args'
    
    case $line[1] in
        show|edit|remove)
            _norn_pattern_ids
            ;;
    esac
}

_norn_pattern_commands() {
    local commands=(list add show edit remove)
    _describe -t commands 'pattern command' commands
}

_norn_pattern_ids() {
    local ids
    ids=(${(f)"$(norn patterns list 2>/dev/null | tail -n +2 | awk '{print $1}')"})
    _describe -t ids 'pattern id' ids
}

_norn_skills() {
    local curcontext="$curcontext" state line
    typeset -A opt_args
    
    _arguments -C \
        '1: :_norn_skill_commands' \
        '*:: :->args'
    
    case $line[1] in
        show|edit|remove)
            _norn_skill_ids
            ;;
    esac
}

_norn_skill_commands() {
    local commands=(list add show edit remove)
    _describe -t commands 'skill command' commands
}

_norn_skill_ids() {
    local ids
    ids=(${(f)"$(norn skills list 2>/dev/null | tail -n +2 | awk '{print $1}')"})
    _describe -t ids 'skill id' ids
}

compdef _norn norn
`
}

func fishCompletion() string {
	return `# Fish completion for norn

# Top-level commands
complete -c norn -f
complete -c norn -n '__fish_use_subcommand' -a 'init' -d 'Bootstrap a new Norn project'
complete -c norn -n '__fish_use_subcommand' -a 'status' -d 'Show workspace status'
complete -c norn -n '__fish_use_subcommand' -a 'detect' -d 'Detect project stack'
complete -c norn -n '__fish_use_subcommand' -a 'fates' -d 'Manage agent fates'
complete -c norn -n '__fish_use_subcommand' -a 'patterns' -d 'Manage patterns'
complete -c norn -n '__fish_use_subcommand' -a 'skills' -d 'Manage skills'
complete -c norn -n '__fish_use_subcommand' -a 'tools' -d 'Manage tools'
complete -c norn -n '__fish_use_subcommand' -a 'weaves' -d 'Manage weaves'
complete -c norn -n '__fish_use_subcommand' -a 'threads' -d 'Manage threads'
complete -c norn -n '__fish_use_subcommand' -a 'warps' -d 'Manage warps'
complete -c norn -n '__fish_use_subcommand' -a 'runes' -d 'Manage configuration'
complete -c norn -n '__fish_use_subcommand' -a 'chat' -d 'OpenCode integration'
complete -c norn -n '__fish_use_subcommand' -a 'export' -d 'Export artifacts'
complete -c norn -n '__fish_use_subcommand' -a 'completion' -d 'Generate shell completions'

# Subcommands
complete -c norn -n '__fish_seen_subcommand_from weaves' -a 'list add show remove'
complete -c norn -n '__fish_seen_subcommand_from threads' -a 'list add show remove'
complete -c norn -n '__fish_seen_subcommand_from warps' -a 'list add assign assignment show remove'
complete -c norn -n '__fish_seen_subcommand_from fates' -a 'list show add edit remove'
complete -c norn -n '__fish_seen_subcommand_from tools' -a 'list add show edit remove'
complete -c norn -n '__fish_seen_subcommand_from patterns' -a 'list add show edit remove'
complete -c norn -n '__fish_seen_subcommand_from skills' -a 'list add show edit remove'
complete -c norn -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'

# Dynamic artifact IDs
function __norn_weave_ids
    norn weaves list 2>/dev/null | tail -n +2 | awk '{print $1}'
end

function __norn_warp_ids
    norn warps list 2>/dev/null | tail -n +2 | awk '{print $1}'
end

function __norn_fate_names
    norn fates list 2>/dev/null | tail -n +2 | awk '{print $1}'
end

function __norn_tool_ids
    norn tools list 2>/dev/null | tail -n +2 | awk '{print $1}'
end

function __norn_pattern_ids
    norn patterns list 2>/dev/null | tail -n +2 | awk '{print $1}'
end

function __norn_skill_ids
    norn skills list 2>/dev/null | tail -n +2 | awk '{print $1}'
end

complete -c norn -n '__fish_seen_subcommand_from weaves; and __fish_seen_subcommand_from show remove' -a '(__norn_weave_ids)'
complete -c norn -n '__fish_seen_subcommand_from threads; and __fish_seen_subcommand_from list show remove' -a '(__norn_weave_ids)'
complete -c norn -n '__fish_seen_subcommand_from warps; and __fish_seen_subcommand_from show remove' -a '(__norn_warp_ids)'
complete -c norn -n '__fish_seen_subcommand_from fates; and __fish_seen_subcommand_from show edit remove' -a '(__norn_fate_names)'
complete -c norn -n '__fish_seen_subcommand_from tools; and __fish_seen_subcommand_from show edit remove' -a '(__norn_tool_ids)'
complete -c norn -n '__fish_seen_subcommand_from patterns; and __fish_seen_subcommand_from show edit remove' -a '(__norn_pattern_ids)'
complete -c norn -n '__fish_seen_subcommand_from skills; and __fish_seen_subcommand_from show edit remove' -a '(__norn_skill_ids)'

# Flags
complete -c norn -n '__fish_seen_subcommand_from init' -l no-interactive
complete -c norn -n '__fish_seen_subcommand_from init' -l enable-opencode
complete -c norn -n '__fish_seen_subcommand_from init' -l name
complete -c norn -n '__fish_seen_subcommand_from init' -l theme
complete -c norn -n '__fish_seen_subcommand_from runes' -l scope -a 'global workspace local'
complete -c norn -n '__fish_seen_subcommand_from runes' -l set
complete -c norn -n '__fish_seen_subcommand_from runes' -l unset
complete -c norn -n '__fish_seen_subcommand_from export' -l opencode
complete -c norn -n '__fish_seen_subcommand_from export' -l output
`
}
