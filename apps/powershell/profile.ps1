# ============================================================================
# POWERSHELL PROFILE (Windows)
# ============================================================================
# Mirrors the Unix .profile + .bashrc/.zshrc setup where it makes sense.

# ----------------------------------------------------------------------------
# 1. SECRETS & ENV
# ----------------------------------------------------------------------------

# Load secrets (untracked, user-managed)
$secretsFile = Join-Path $HOME ".secrets.ps1"
if (Test-Path $secretsFile) {
    . $secretsFile
}

$env:EDITOR  = 'vim'
$env:VISUAL  = 'vim'

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

# pipx user scripts (PEP 668 path for Microsoft Store Python)
$pyUserScripts = "$env:APPDATA\Python\Python313\Scripts"
Add-ToPath $pyUserScripts
Add-ToPath "$HOME\.local\bin"

# ----------------------------------------------------------------------------
# 3. ALIASES (parity with .profile)
# ----------------------------------------------------------------------------

# Directory navigation: PowerShell already treats `..`, `...` as relative if used
# as arguments, but provide explicit shortcuts.
function .. { Set-Location .. }
function ... { Set-Location ..\.. }
function .... { Set-Location ..\..\.. }

# File listing
function ll { Get-ChildItem -Force @args }
function la { Get-ChildItem -Force -Hidden @args }

# Git shortcuts
function gs { git status @args }
function ga { git add @args }
function gc { git commit -v @args }
function gd { git diff @args }
function gp { git push @args }
function gl { git pull @args }
function glog { git log --oneline --decorate --graph @args }

# ----------------------------------------------------------------------------
# 4. PSReadLine (history & editing)
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
# 5. FZF
# ----------------------------------------------------------------------------

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

# ----------------------------------------------------------------------------
# 6. PROMPT (oh-my-posh)
# ----------------------------------------------------------------------------

if (Get-Command oh-my-posh -ErrorAction SilentlyContinue) {
    oh-my-posh init pwsh --config "$HOME\.mytheme.omp.json" | Invoke-Expression
}

