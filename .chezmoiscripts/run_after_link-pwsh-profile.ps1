$ErrorActionPreference = "Stop"

# Links the canonical PowerShell profile into PowerShell's real Documents
# folder. chezmoi targets are literal paths, but Documents may be redirected
# (OneDrive known-folder move) and localised, so it is resolved at runtime.
# Idempotent: runs on every apply and only touches links that are wrong.

$target = [IO.Path]::GetFullPath((Join-Path $env:CHEZMOI_SOURCE_DIR "dot_config\powershell\profile.ps1"))
$docs   = [Environment]::GetFolderPath("MyDocuments")

foreach ($dir in "PowerShell", "WindowsPowerShell") {
    $link = Join-Path $docs "$dir\profile.ps1"
    $item = Get-Item -LiteralPath $link -Force -ErrorAction SilentlyContinue

    # Normalise: chezmoi writes link targets with forward slashes
    if ($item -and $item.LinkType -eq "SymbolicLink" -and
        [IO.Path]::GetFullPath(@($item.Target)[0]) -eq $target) {
        continue
    }
    if ($item -and -not $item.LinkType) {
        Move-Item -LiteralPath $link -Destination "$link.bak" -Force
        Write-Host "Backed up $link -> $link.bak"
    }

    New-Item -ItemType Directory -Force -Path (Split-Path $link) | Out-Null
    if ($item -and $item.LinkType) { Remove-Item -LiteralPath $link -Force }
    # mklink honours Developer Mode's unprivileged symlinks; Windows
    # PowerShell 5.1's New-Item -ItemType SymbolicLink needs admin.
    cmd /c mklink "$link" "$target" | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "mklink failed for $link (enable Developer Mode or run as admin)" }
    Write-Host "Linked $link -> $target"
}
