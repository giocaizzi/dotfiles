$ErrorActionPreference = "Stop"

if ([Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT) {
    Write-Output "Skipping PowerShell profile link tests: Windows only."
    exit 0
}

$repo = Split-Path -Parent $PSScriptRoot
$template = Join-Path $repo "home\.chezmoiscripts\run_after_link-pwsh-profile.ps1.tmpl"
$rendered = & chezmoi --source $repo execute-template --file $template
if ($LASTEXITCODE -ne 0) { throw "chezmoi could not render the profile script." }
$scriptText = ($rendered -join [Environment]::NewLine).Replace(
    '[Environment]::GetFolderPath("MyDocuments")', '$documents'
)
if ($scriptText -eq ($rendered -join [Environment]::NewLine)) {
    throw "Could not replace the runtime Documents lookup."
}

function Assert($Condition, [string]$Message) {
    if (-not $Condition) { throw $Message }
}

function New-Link {
    [CmdletBinding(SupportsShouldProcess = $true)]
    param([string]$Path, [string]$Target)

    if (-not $PSCmdlet.ShouldProcess($Path, "Create fixture symbolic link")) { return }
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Path) | Out-Null
    & $env:ComSpec /c ('mklink "{0}" "{1}"' -f $Path, $Target) | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "Could not create fixture symbolic link: $Path" }
}

function Resolve-Link([string]$Path) {
    $target = @((Get-Item -LiteralPath $Path -Force).Target)[0]
    if (-not [IO.Path]::IsPathRooted($target)) {
        $target = Join-Path (Split-Path -Parent $Path) $target
    }
    [IO.Path]::GetFullPath($target)
}

function Assert-ProfileLink($Fixture) {
    foreach ($edition in "PowerShell", "WindowsPowerShell") {
        $path = Join-Path $Fixture.Documents "$edition\profile.ps1"
        $item = Get-Item -LiteralPath $path -Force
        Assert ($item.LinkType -eq "SymbolicLink") "Missing all-hosts link: $path"
        Assert ((Resolve-Link $path) -eq [IO.Path]::GetFullPath($Fixture.Canonical)) `
            "Incorrect profile target: $path"
    }
}

$fixture = $null
$originalSource = $env:CHEZMOI_SOURCE_DIR
try {
    $root = Join-Path ([IO.Path]::GetTempPath()) ("pwsh-profile-" + [Guid]::NewGuid().ToString("N"))
    $source = Join-Path $root "source"
    $documents = Join-Path $root "docs"
    $canonical = Join-Path $source "dot_config\powershell\profile.ps1"
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $canonical), $documents | Out-Null
    Set-Content -LiteralPath $canonical -Value "# fixture profile"
    $fixture = [PSCustomObject]@{ Root = $root; Documents = $documents; Canonical = $canonical }
    $env:CHEZMOI_SOURCE_DIR = $source
    $script:documents = $documents
    $run = { & ([ScriptBlock]::Create($scriptText)) }

    # Create both edition links; a second apply is a no-op.
    & $run
    Assert-ProfileLink $fixture
    foreach ($edition in "PowerShell", "WindowsPowerShell") {
        Assert (-not (Test-Path (Join-Path $documents "$edition\Microsoft.PowerShell_profile.ps1"))) `
            "The legacy profile name should not be created."
    }
    $repeatOutput = @(& $run)
    Assert ($repeatOutput.Count -eq 0) "Reapplying changed an already-correct profile link."

    # Remove only legacy links to the canonical target, including a relative target.
    New-Link (Join-Path $documents "PowerShell\Microsoft.PowerShell_profile.ps1") $canonical -Confirm:$false
    New-Link (Join-Path $documents "WindowsPowerShell\Microsoft.PowerShell_profile.ps1") `
        '..\..\source\dot_config\powershell\profile.ps1' -Confirm:$false
    & $run
    foreach ($edition in "PowerShell", "WindowsPowerShell") {
        Assert (-not (Test-Path (Join-Path $documents "$edition\Microsoft.PowerShell_profile.ps1"))) `
            "An owned legacy link remains in $edition."
    }

    # Back up an existing all-hosts file; leave unrelated files, links and backups intact.
    $allHosts = Join-Path $documents "PowerShell\profile.ps1"
    Remove-Item -LiteralPath $allHosts -Force
    Set-Content -LiteralPath $allHosts -Value "user profile"
    Set-Content -LiteralPath "$allHosts.bak" -Value "old backup"
    Set-Content -LiteralPath "$allHosts.notes.bak" -Value "unrelated backup"
    $legacy = Join-Path $documents "WindowsPowerShell\Microsoft.PowerShell_profile.ps1"
    Set-Content -LiteralPath $legacy -Value "user legacy profile"
    $otherTarget = Join-Path $root "other-profile.ps1"
    Set-Content -LiteralPath $otherTarget -Value "# unrelated link target"
    $unrelatedLink = Join-Path $documents "PowerShell\Microsoft.PowerShell_profile.ps1"
    New-Link $unrelatedLink $otherTarget -Confirm:$false
    & $run
    Assert-ProfileLink $fixture
    Assert ((Get-Content -LiteralPath "$allHosts.bak" -Raw).Trim() -eq "user profile") `
        "The existing all-hosts profile was not backed up."
    Assert ((Get-Content -LiteralPath "$allHosts.notes.bak" -Raw).Trim() -eq "unrelated backup") `
        "An unrelated backup was changed."
    Assert ((Get-Content -LiteralPath $legacy -Raw).Trim() -eq "user legacy profile") `
        "An unrelated legacy profile file was changed."
    Assert ((Resolve-Link $unrelatedLink) -eq [IO.Path]::GetFullPath($otherTarget)) `
        "An unrelated legacy symbolic link was changed."

    Write-Output "PowerShell profile link tests passed."
}
finally {
    if ($null -eq $originalSource) {
        Remove-Item Env:\CHEZMOI_SOURCE_DIR -ErrorAction SilentlyContinue
    } else {
        $env:CHEZMOI_SOURCE_DIR = $originalSource
    }
    if ($fixture -and (Test-Path -LiteralPath $fixture.Root)) {
        Remove-Item -LiteralPath $fixture.Root -Recurse -Force
    } elseif ($root -and (Test-Path -LiteralPath $root)) {
        Remove-Item -LiteralPath $root -Recurse -Force
    }
}
