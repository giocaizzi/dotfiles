$ErrorActionPreference = "Stop"

# Creates ~\.config\powershell\secrets.ps1 if it doesn't exist.
# This file is dot-sourced by the PowerShell profile to load secret environment variables.
# It is never tracked by chezmoi — edit it directly on each machine.

$secretsFile = Join-Path $HOME ".config\powershell\secrets.ps1"

New-Item -ItemType Directory -Force -Path (Split-Path $secretsFile) | Out-Null

if (-not (Test-Path $secretsFile)) {
    @'
# ~\.config\powershell\secrets.ps1 — Secret environment variables (not tracked by chezmoi)
# Add exports here, e.g.:
#   $env:API_KEY = "your-key"
'@ | Set-Content -Path $secretsFile -Encoding UTF8
    Write-Host "Created $secretsFile"
}

# Restrict access to current user only. icacls edits the DACL alone; Set-Acl
# also rewrites the SACL, which needs SeSecurityPrivilege (absent on managed PCs).
$me = [System.Security.Principal.WindowsIdentity]::GetCurrent().Name
icacls $secretsFile /inheritance:r /grant:r "${me}:(F)" | Out-Null
if ($LASTEXITCODE -ne 0) { throw "icacls failed to restrict $secretsFile" }
