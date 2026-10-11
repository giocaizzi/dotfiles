#!/usr/bin/env bash
# preToolUse hook (Copilot CLI): deny pip/pipx; Python tooling is uv only.
# Copilot has no deny list, unlike Claude (see Bash(pip:*) in claude-settings.json).
in=$(cat)
jq -e '(.toolName // .tool_name) | test("^(bash|Bash)$")' <<<"$in" >/dev/null 2>&1 || exit 0
if jq -e '(.toolArgs // .tool_input) | (if type == "string" then . else tostring end)
  | test("(?<!uv )(?<![\\w-])(pipx|pip3?|python3? -m pip)\\s")' <<<"$in" >/dev/null 2>&1; then
  echo "Use uv (uv add / uv run / uvx); pip and pipx are not allowed." >&2
  exit 2
fi
