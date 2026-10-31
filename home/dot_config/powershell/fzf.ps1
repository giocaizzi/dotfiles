# ============================================================================
# FZF CONFIGURATION (PowerShell)
# ============================================================================
# Dot-sourced by $PROFILE. Mirrors ~/.config/shell/fzf.sh for parity.
# ============================================================================

$env:FZF_DEFAULT_OPTS = @'
--height=40% --layout=reverse --info=inline --border --margin=1 --padding=1
--preview-window=right:60%:wrap
--bind ctrl-u:preview-half-page-up,ctrl-d:preview-half-page-down
--color fg:#cdd6f4,bg:-1,hl:#89b4fa,fg+:#cdd6f4,bg+:-1,hl+:#89dceb
--color info:#94e2d5,prompt:#f5c2e7,spinner:#f5c2e7,pointer:#f38ba8,marker:#a6e3a1
--color border:#45475a,header:#f9e2af,gutter:-1
'@

if (Get-Command fd -ErrorAction SilentlyContinue) {
    $env:FZF_DEFAULT_COMMAND = 'fd --type f --hidden --follow --exclude .git'
    $env:FZF_CTRL_T_COMMAND  = $env:FZF_DEFAULT_COMMAND
}

# Mirror the Unix h/history shortcuts through PSFzf's history picker.
function Invoke-FzfHistory {
    if (Get-Command Invoke-FuzzyHistory -ErrorAction SilentlyContinue) {
        Invoke-FuzzyHistory
        return
    }

    Write-Error 'PSFzf is not available, so fzf history search cannot run.'
}

if (Test-Path Alias:history) {
    Remove-Item Alias:history -Force
}

if (Test-Path Alias:h) {
    Remove-Item Alias:h -Force
}

function history {
    Invoke-FzfHistory @args
}

function h {
    Invoke-FzfHistory @args
}

# PSFzf — provides Ctrl-R (history) and Ctrl-T (files) bindings like bash/zsh
if (Get-Module -ListAvailable -Name PSFzf) {
    Import-Module PSFzf
    Set-PsFzfOption -PSReadlineChordProvider 'Ctrl+t' -PSReadlineChordReverseHistory 'Ctrl+r'
}
