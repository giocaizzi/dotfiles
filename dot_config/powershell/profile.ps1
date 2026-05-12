# ============================================================================
# POWERSHELL PROFILE (Windows)
# ============================================================================
# Thin orchestrator. Aliases live in ~/.config/powershell/aliases.ps1
# (rendered from .chezmoidata/shortcuts.toml). FZF setup in fzf.ps1.
# ============================================================================

# ----------------------------------------------------------------------------
# 1. SECRETS & ENV
# ----------------------------------------------------------------------------

$secretsFile = Join-Path $HOME ".secrets.ps1"
if (Test-Path $secretsFile) { . $secretsFile }

$env:EDITOR     = 'vim'
$env:VISUAL     = 'vim'
$env:POSH_THEME = Join-Path $HOME '.mytheme.omp.json'

# ----------------------------------------------------------------------------
# 2. PATH
# ----------------------------------------------------------------------------

function Add-ToPath([string]$dir) {
    if ((Test-Path $dir) -and ($env:Path -split ';' -notcontains $dir)) {
        $env:Path = "$dir;$env:Path"
    }
}

# pyenv-win
$env:PYENV_ROOT = "$HOME\.pyenv\pyenv-win"
Add-ToPath "$env:PYENV_ROOT\bin"
Add-ToPath "$env:PYENV_ROOT\shims"

# pipx
Add-ToPath "$env:APPDATA\Python\Python313\Scripts"
Add-ToPath "$HOME\.local\bin"

# ----------------------------------------------------------------------------
# 3. PSReadLine
# ----------------------------------------------------------------------------

if (Get-Module -ListAvailable -Name PSReadLine) {
    Import-Module PSReadLine
    Set-PSReadLineOption -HistorySaveStyle SaveIncrementally
    Set-PSReadLineOption -HistoryNoDuplicates
    Set-PSReadLineOption -PredictionSource History
    Set-PSReadLineOption -EditMode Windows
    Set-PSReadLineKeyHandler -Key UpArrow   -Function HistorySearchBackward
    Set-PSReadLineKeyHandler -Key DownArrow -Function HistorySearchForward
}

# ----------------------------------------------------------------------------
# 3b. FILE-SYSTEM COLORS (PS 7.2+)
# ----------------------------------------------------------------------------
# Override defaults that use background colors (unreadable on dark terminals).
# Use foreground-only ANSI so Get-ChildItem stays legible.

if ($PSStyle) {
    $PSStyle.FileInfo.Directory    = "`e[1;34m"   # bold blue
    $PSStyle.FileInfo.SymbolicLink = "`e[1;36m"   # bold cyan
    $PSStyle.FileInfo.Executable   = "`e[1;32m"   # bold green
}

# ----------------------------------------------------------------------------
# 4. SHARED SNIPPETS
# ----------------------------------------------------------------------------

$aliasesFile = Join-Path $HOME '.config\powershell\aliases.ps1'
if (Test-Path $aliasesFile) { . $aliasesFile }

$fzfFile = Join-Path $HOME '.config\powershell\fzf.ps1'
if (Test-Path $fzfFile) { . $fzfFile }

# ----------------------------------------------------------------------------
# 5. PROMPT (oh-my-posh)
# ----------------------------------------------------------------------------

if (Get-Command oh-my-posh -ErrorAction SilentlyContinue) {
    oh-my-posh init pwsh --config $env:POSH_THEME | Invoke-Expression
}
