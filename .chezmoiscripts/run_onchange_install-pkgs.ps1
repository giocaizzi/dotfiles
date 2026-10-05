#Requires -Version 5.1
$ErrorActionPreference = "Stop"

# ============================================================================
# CONFIGURATION
# ============================================================================

$Verbose = $true
$PythonVersion = '3.13'  # global python, installed by uv

# ============================================================================
# PACKAGE DEFINITIONS (Single Source of Truth)
# Mirrors .chezmoiscripts/run_onchange_install-pkgs.sh.tmpl
# Each entry: @{ Name = '<display>'; Id = '<winget id>'; Cmd = '<command to probe>' }
# ============================================================================

$CorePackages = @(
    @{ Name = 'fzf';     Id = 'junegunn.fzf';        Cmd = 'fzf' }
    @{ Name = 'jq';      Id = 'jqlang.jq';           Cmd = 'jq' }
    @{ Name = 'git';     Id = 'Git.Git';             Cmd = 'git' }
    @{ Name = 'vim';     Id = 'vim.vim';             Cmd = 'vim' }
    @{ Name = 'fd';      Id = 'sharkdp.fd';          Cmd = 'fd' }
    @{ Name = 'gh';      Id = 'GitHub.cli';          Cmd = 'gh' }
    @{ Name = 'pwsh';    Id = 'Microsoft.PowerShell';Cmd = 'pwsh' }  # PowerShell 7+
    @{ Name = 'coreutils'; Id = 'uutils.coreutils';  Cmd = 'ls.exe' }  # exe, not the ls alias
)

$AdditionalTools = @(
    @{ Name = 'oh-my-posh'; Id = 'JanDeDobbeleer.OhMyPosh'; Cmd = 'oh-my-posh' }
    @{ Name = 'uv';         Id = $null;                     Cmd = 'uv' }
)

# Fonts — installed per user by oh-my-posh (no admin; managed PCs block the
# machine-scope winget package). Probed by one of the font's files.
$Fonts = @(
    @{ Name = 'JetBrainsMono Nerd Font'; Archive = 'JetBrainsMono'; File = 'JetBrainsMonoNerdFont-Regular.ttf' }
)

# PowerShell modules (parity with bash/zsh shell features)
$PSModules = @(
    @{ Name = 'PSReadLine'; MinVersion = '2.2.0' }  # History/edit experience
    @{ Name = 'PSFzf';      MinVersion = '2.5.0' }  # Ctrl-R / Ctrl-T fzf bindings
)

# ============================================================================
# HELPERS
# ============================================================================

function Write-Step([string]$msg)    { Write-Host "==> $msg" -ForegroundColor Blue }
function Write-Ok([string]$msg)      { if ($Verbose) { Write-Host "  OK  $msg" -ForegroundColor Green } }
function Write-Info([string]$msg)    { if ($Verbose) { Write-Host "  ..  $msg" -ForegroundColor Yellow } }

function Test-Command([string]$cmd) {
    return [bool](Get-Command -Name $cmd -ErrorAction SilentlyContinue)
}

function Install-WingetPackage([string]$name, [string]$id) {
    Write-Info "Installing $name ($id) via winget..."
    winget install --id $id --exact --silent --scope user `
        --accept-package-agreements --accept-source-agreements --source winget | Out-Null
}

function Install-Custom([string]$name) {
    switch ($name) {
        'uv' {
            # Official installer → ~\.local\bin, user-only (no admin; winget's
            # user scope misplaces portable packages on managed PCs).
            # UV_NO_MODIFY_PATH: profile.ps1 already puts ~\.local\bin on PATH.
            Write-Info "Installing uv..."
            $env:UV_NO_MODIFY_PATH = '1'
            Invoke-RestMethod https://astral.sh/uv/install.ps1 | Invoke-Expression
        }
        default {
            Write-Warning "No custom installer defined for $name"
        }
    }
}

# ============================================================================
# PRE-FLIGHT
# ============================================================================

Write-Step "Installing required packages..."

if (-not (Test-Command 'winget')) {
    Write-Error "winget is required but not installed. Install 'App Installer' from the Microsoft Store."
    exit 1
}

# ============================================================================
# INSTALL CORE PACKAGES (winget)
# ============================================================================

Write-Step "Core packages (winget)"
foreach ($pkg in $CorePackages) {
    if (Test-Command $pkg.Cmd) {
        Write-Ok "$($pkg.Name) already installed"
    } else {
        Install-WingetPackage $pkg.Name $pkg.Id
    }
}

# ============================================================================
# INSTALL ADDITIONAL TOOLS
# ============================================================================

Write-Step "Additional tools"
foreach ($tool in $AdditionalTools) {
    if (Test-Command $tool.Cmd) {
        Write-Ok "$($tool.Name) already installed"
        continue
    }
    if ($tool.Id) {
        Install-WingetPackage $tool.Name $tool.Id
    } else {
        Install-Custom $tool.Name
    }
}

# ============================================================================
# PYTHON (uv)
# ============================================================================

Write-Step "Python $PythonVersion (uv)"
$uv = (Get-Command uv -ErrorAction SilentlyContinue).Source
if (-not $uv) { $uv = Join-Path $HOME '.local\bin\uv.exe' }
if (Test-Path $uv) {
    # --no-bin: uv's python.exe launchers in ~\.local\bin are blocked on managed
    # PCs; profile.ps1 puts the interpreter folder on PATH instead.
    & $uv python install $PythonVersion --no-bin
} else {
    Write-Warning "uv not found; skipping Python install"
}

# ============================================================================
# INSTALL FONTS
# ============================================================================

Write-Step "Fonts (oh-my-posh, user scope)"
foreach ($font in $Fonts) {
    $installed = (Test-Path "$env:LOCALAPPDATA\Microsoft\Windows\Fonts\$($font.File)") -or
                 (Test-Path "$env:WINDIR\Fonts\$($font.File)")
    if ($installed) {
        Write-Ok "$($font.Name) already installed"
    } elseif (Test-Command 'oh-my-posh') {
        Write-Info "Installing $($font.Name) via oh-my-posh..."
        oh-my-posh font install $font.Archive
        if ($LASTEXITCODE -ne 0) { Write-Warning "Font install failed: $($font.Name)" }
    } else {
        Write-Warning "oh-my-posh not found; skipping $($font.Name)"
    }
}

# ============================================================================
# INSTALL POWERSHELL MODULES
# ============================================================================

Write-Step "PowerShell modules"

foreach ($mod in $PSModules) {
    $installed = Get-Module -ListAvailable -Name $mod.Name |
        Where-Object { $_.Version -ge [version]$mod.MinVersion } |
        Select-Object -First 1
    if ($installed) {
        Write-Ok "$($mod.Name) >= $($mod.MinVersion) already installed"
    } else {
        Write-Info "Installing PowerShell module $($mod.Name) (>= $($mod.MinVersion))..."
        Install-Module -Name $mod.Name -Scope CurrentUser -Force -AllowClobber -MinimumVersion $mod.MinVersion
    }
}

Write-Host ""
Write-Host "All packages installed successfully" -ForegroundColor Green
Write-Host "Note: open a new shell session for PATH changes to take effect." -ForegroundColor Yellow
