$ErrorActionPreference = "Stop"

# Creates ~\.config\powershell\secrets.ps1 if it doesn't exist (migrating legacy ~\.secrets.ps1).
# This file is dot-sourced by the PowerShell profile to load secret environment variables.
# It is never tracked by chezmoi — edit it directly on each machine.

$secretsFile = Join-Path $HOME ".config\powershell\secrets.ps1"
$legacyFile  = Join-Path $HOME ".secrets.ps1"

New-Item -ItemType Directory -Force -Path (Split-Path $secretsFile) | Out-Null

if (-not (Test-Path $secretsFile)) {
    if (Test-Path $legacyFile) {
        Move-Item -Path $legacyFile -Destination $secretsFile
        Write-Host "Migrated $legacyFile -> $secretsFile"
    } else {
        @'
# ~\.config\powershell\secrets.ps1 — Secret environment variables (not tracked by chezmoi)
# Add exports here, e.g.:
#   $env:API_KEY = "your-key"
'@ | Set-Content -Path $secretsFile -Encoding UTF8
        Write-Host "Created $secretsFile"
    }
}

# Restrict access to current user only
$acl = Get-Acl $secretsFile
$acl.SetAccessRuleProtection($true, $false)
$rule = New-Object System.Security.AccessControl.FileSystemAccessRule(
    [System.Security.Principal.WindowsIdentity]::GetCurrent().Name,
    "FullControl",
    "Allow"
)
$acl.AddAccessRule($rule)
Set-Acl -Path $secretsFile -AclObject $acl
