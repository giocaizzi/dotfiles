# Stop hook for Windows (Claude Code + Copilot CLI), twin of stop-make-check.sh.
# Run with -NoProfile: the uutils PATH/alias setup in the profile is not active,
# so this uses plain cmdlets only. Runs `make check` once; on failure blocks the
# stop once. stop_hook_active (a second stop) is always allowed, so it cannot loop.
$ErrorActionPreference = 'Stop'
try { $in = [Console]::In.ReadToEnd() | ConvertFrom-Json } catch { exit 0 }
if ($in.stop_hook_active -eq $true) { exit 0 }
$dir = if ($in.cwd) { $in.cwd } else { (Get-Location).Path }
if (-not (Get-Command make -ErrorAction SilentlyContinue)) { exit 0 }   # no make: skip
$ErrorActionPreference = 'Continue'
& make -C $dir -n check *> $null
if ($LASTEXITCODE -ne 0) { exit 0 }                                     # no Makefile or no check target
$out = (& make -C $dir check 2>&1 | Out-String)
if ($LASTEXITCODE -eq 0) { exit 0 }
$tail = (($out -split "`r?`n") | Select-Object -Last 40) -join "`n"
@{ decision = 'block'; reason = "make check failed; fix before finishing:`n$tail" } | ConvertTo-Json -Compress
