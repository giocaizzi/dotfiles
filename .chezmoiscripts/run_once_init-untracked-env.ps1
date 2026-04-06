$ErrorActionPreference = "Stop"

# Creates ~\.secrets.ps1 if it doesn't exist.
# This file is dot-sourced by the PowerShell profile to load secret environment variables.
# It is never tracked by chezmoi — edit it directly on each machine.

$secretsFile = Join-Path $HOME ".secrets.ps1"

if (-not (Test-Path $secretsFile)) {
    @'
# ~\.secrets.ps1 — Secret environment variables (not tracked by chezmoi)
# Add exports here, e.g.:
#   $env:API_KEY = "your-key"
'@ | Set-Content -Path $secretsFile -Encoding UTF8

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

    Write-Host "Created $secretsFile (current user only)"
}
