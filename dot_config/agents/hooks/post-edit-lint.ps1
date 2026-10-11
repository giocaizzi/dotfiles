# PostToolUse hook for Windows (Claude Code), twin of post-edit-lint.sh.
# Run with -NoProfile: plain cmdlets only. After Edit/Write of a .sh/.bash file
# runs shellcheck, of a .ps1 file PSScriptAnalyzer; findings go to stderr with
# exit 2 so the agent sees them. Missing tool: skip.
try { $in = [Console]::In.ReadToEnd() | ConvertFrom-Json } catch { exit 0 }
$f = $in.tool_input.file_path
if (-not $f -or -not (Test-Path -LiteralPath $f)) { exit 0 }
$out = $null
switch -Regex ($f) {
    '\.(sh|bash)$' {
        if (Get-Command shellcheck -ErrorAction SilentlyContinue) {
            $out = (& shellcheck $f 2>&1 | Out-String)
            if ($LASTEXITCODE -eq 0) { $out = $null }
        }
    }
    '\.ps1$' {
        if (Get-Module -ListAvailable -Name PSScriptAnalyzer) {
            $r = Invoke-ScriptAnalyzer -Path $f -Severity Error, Warning
            if ($r) { $out = ($r | Format-List RuleName, Line, Message | Out-String) }
        }
    }
}
if ($out) { [Console]::Error.WriteLine($out); exit 2 }
