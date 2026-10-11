# preToolUse hook for Copilot CLI on Windows, twin of block-pip.sh: deny pip/pipx
# (Python tooling is uv only). -NoProfile: plain cmdlets, no uutils. Exit 2 denies.
try { $in = [Console]::In.ReadToEnd() | ConvertFrom-Json } catch { exit 0 }
$tool = if ($in.toolName) { $in.toolName } else { $in.tool_name }
if ($tool -notmatch '^(bash|powershell)$') { exit 0 }                  # -match is case-insensitive
$a = if ($null -ne $in.toolArgs) { $in.toolArgs } else { $in.tool_input }
$blob = if ($a -is [string]) { $a } else { $a | ConvertTo-Json -Compress -Depth 10 }
if ($blob -match '(?<!uv )(?<![\w-])(pipx|pip3?|python3? -m pip)\s') {
    [Console]::Error.WriteLine('Use uv (uv add / uv run / uvx); pip and pipx are not allowed.')
    exit 2
}
