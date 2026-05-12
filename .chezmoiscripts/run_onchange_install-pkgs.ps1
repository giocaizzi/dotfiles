#Requires -Version 5.1
$ErrorActionPreference = "Stop"

# ============================================================================
# CONFIGURATION
# ============================================================================

$Verbose = $true

# ============================================================================
# PACKAGE DEFINITIONS (Single Source of Truth)
# Mirrors .chezmoiscripts/run_onchange_install-pkgs.sh.tmpl
# Each entry: @{ Name = '<display>'; Id = '<winget id>'; Cmd = '<command to probe>' }
# ============================================================================

$CorePackages = @(
    @{ Name = 'fzf';     Id = 'junegunn.fzf';        Cmd = 'fzf' }
    @{ Name = 'git';     Id = 'Git.Git';             Cmd = 'git' }
    @{ Name = 'vim';     Id = 'vim.vim';             Cmd = 'vim' }
    @{ Name = 'fd';      Id = 'sharkdp.fd';          Cmd = 'fd' }
    @{ Name = 'gh';      Id = 'GitHub.cli';          Cmd = 'gh' }
    @{ Name = 'pwsh';    Id = 'Microsoft.PowerShell';Cmd = 'pwsh' }  # PowerShell 7+
)

$AdditionalTools = @(
    @{ Name = 'oh-my-posh'; Id = 'JanDeDobbeleer.OhMyPosh'; Cmd = 'oh-my-posh' }
    @{ Name = 'pyenv-win';  Id = $null;                     Cmd = 'pyenv' }
    @{ Name = 'pipx';       Id = $null;                     Cmd = 'pipx' }
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
        'pyenv-win' {
            Write-Info "Installing pyenv-win..."
            $installer = "$env:TEMP\install-pyenv-win.ps1"
            Invoke-WebRequest -UseBasicParsing `
                -Uri "https://raw.githubusercontent.com/pyenv-win/pyenv-win/master/pyenv-win/install-pyenv-win.ps1" `
                -OutFile $installer
            & $installer
        }
        'pipx' {
            if (-not (Test-Command 'python')) {
                Write-Warning "python not found; skipping pipx install"
                return
            }
            Write-Info "Installing pipx..."
            python -m pip install --user --upgrade pipx
            python -m pipx ensurepath
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
