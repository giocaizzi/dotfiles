# ============================================================================
# FZF CONFIGURATION (PowerShell)
# ============================================================================
# Dot-sourced by $PROFILE. Mirrors ~/.config/shell/fzf.sh for parity.
# ============================================================================

$env:FZF_DEFAULT_OPTS = @'
--height=40% --layout=reverse --info=inline --border --margin=1 --padding=1
--preview-window=right:60%:wrap
--bind ctrl-u:preview-half-page-up,ctrl-d:preview-half-page-down
--color fg:252,bg:233,hl:67,fg+:252,bg+:235,hl+:81
--color info:144,prompt:161,spinner:135,pointer:135,marker:118
'@

if (Get-Command fd -ErrorAction SilentlyContinue) {
    $env:FZF_DEFAULT_COMMAND = 'fd --type f --hidden --follow --exclude .git'
    $env:FZF_CTRL_T_COMMAND  = $env:FZF_DEFAULT_COMMAND
}

# PSFzf — provides Ctrl-R (history) and Ctrl-T (files) bindings like bash/zsh
if (Get-Module -ListAvailable -Name PSFzf) {
    Import-Module PSFzf
    Set-PsFzfOption -PSReadlineChordProvider 'Ctrl+t' -PSReadlineChordReverseHistory 'Ctrl+r'
}
