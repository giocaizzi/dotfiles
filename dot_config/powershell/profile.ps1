# ============================================================================
# POWERSHELL PROFILE (Windows)
# ============================================================================
# Thin orchestrator. Aliases live in ~/.config/powershell/aliases.ps1
# (rendered from .chezmoidata/shortcuts.toml). FZF setup in fzf.ps1.
# ============================================================================

# ----------------------------------------------------------------------------
# 1. SECRETS & ENV
# ----------------------------------------------------------------------------

$secretsFile = Join-Path $HOME ".config\powershell\secrets.ps1"
if (Test-Path $secretsFile) { . $secretsFile }

$env:EDITOR     = 'vim'
$env:VISUAL     = 'vim'
$env:POSH_THEME = Join-Path $HOME '.config\oh-my-posh\theme.omp.json'

# ----------------------------------------------------------------------------
# 2. PATH
# ----------------------------------------------------------------------------

# Prepend $dir, moving it to the front if it's already on PATH (like the
# POSIX .profile), so user tools win over entries from the Windows PATH setting.
function Add-ToPath([string]$dir) {
    if (-not (Test-Path $dir)) { return }
    $rest = $env:Path -split ';' | Where-Object { $_ -and $_.TrimEnd('\') -ne $dir.TrimEnd('\') }
    $env:Path = (@($dir) + $rest) -join ';'
}

# uv and uv tools
Add-ToPath "$HOME\.local\bin"

# uv-managed Python: put the interpreter's own folder first. uv's small
# python.exe launchers are blocked on managed PCs ("Access is denied": an
# unsigned exe named python.exe), so they aren't installed on Windows.
# cpython-3.NN-* (no patch) is uv's per-minor folder that tracks the latest patch.
Get-Item "$env:APPDATA\uv\python\cpython-3.??-windows-*" -ErrorAction SilentlyContinue |
    Sort-Object Name -Descending | Select-Object -First 1 |
    ForEach-Object { Add-ToPath $_.FullName }

# ----------------------------------------------------------------------------
# 3. PSReadLine
# ----------------------------------------------------------------------------

if (Get-Module -ListAvailable -Name PSReadLine) {
    Import-Module PSReadLine
    Set-PSReadLineOption -HistorySaveStyle SaveIncrementally
    Set-PSReadLineOption -HistoryNoDuplicates
    # Throws when the console is redirected (non-interactive sessions).
    try { Set-PSReadLineOption -PredictionSource History } catch {}
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

$coreutilsFile = Join-Path $HOME '.config\powershell\coreutils.ps1'
if (Test-Path $coreutilsFile) { . $coreutilsFile }

$fzfFile = Join-Path $HOME '.config\powershell\fzf.ps1'
if (Test-Path $fzfFile) { . $fzfFile }

# ----------------------------------------------------------------------------
# 5. PROMPT (oh-my-posh)
# ----------------------------------------------------------------------------

if (Get-Command oh-my-posh -ErrorAction SilentlyContinue) {
    oh-my-posh init pwsh --config $env:POSH_THEME | Invoke-Expression
}
