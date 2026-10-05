# ============================================================================
# UUTILS COREUTILS AS DEFAULT (PowerShell)
# ============================================================================
# Dot-sourced by $PROFILE. winget (uutils.coreutils) links one exe per tool into
# WinGet\Links. Put that dir first on PATH (beats System32's sort/whoami/...) and
# drop PowerShell's same-named cmdlet aliases (ls, cat, cp, ...), which outrank
# any exe on PATH.
# ============================================================================

$uutilsDir = Get-Item "$env:LOCALAPPDATA\Microsoft\WinGet\Packages\uutils.coreutils_*\*" -ErrorAction SilentlyContinue |
    Select-Object -First 1
if ($uutilsDir) {
    Add-ToPath "$env:LOCALAPPDATA\Microsoft\WinGet\Links"
    Get-ChildItem $uutilsDir.FullName -Filter *.exe | ForEach-Object {
        Remove-Item "Alias:$($_.BaseName)" -Force -ErrorAction SilentlyContinue
    }
}
