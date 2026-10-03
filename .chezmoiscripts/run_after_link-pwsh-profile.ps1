$ErrorActionPreference = "Stop"

# Links the canonical PowerShell profile into PowerShell's real Documents
# folder. chezmoi targets are literal paths, but Documents may be redirected
# (OneDrive known-folder move) and localised, so it is resolved at runtime.
# Idempotent: runs on every apply and only touches links that are wrong.

$target = Join-Path $env:CHEZMOI_SOURCE_DIR "dot_config\powershell\profile.ps1"
$docs   = [Environment]::GetFolderPath("MyDocuments")

foreach ($dir in "PowerShell", "WindowsPowerShell") {
    $link = Join-Path $docs "$dir\profile.ps1"
    $item = Get-Item -LiteralPath $link -Force -ErrorAction SilentlyContinue

    if ($item -and $item.LinkType -eq "SymbolicLink" -and $item.Target -eq $target) {
        continue
    }
    if ($item -and -not $item.LinkType) {
        Move-Item -LiteralPath $link -Destination "$link.bak" -Force
        Write-Host "Backed up $link -> $link.bak"
    }

    New-Item -ItemType Directory -Force -Path (Split-Path $link) | Out-Null
    New-Item -ItemType SymbolicLink -Force -Path $link -Target $target | Out-Null
    Write-Host "Linked $link -> $target"
}
