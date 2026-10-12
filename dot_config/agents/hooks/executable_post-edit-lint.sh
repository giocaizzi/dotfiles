#!/usr/bin/env bash
# PostToolUse hook (Claude Code): after Edit/Write of a shell script,
# run shellcheck and hand findings back (exit 2 -> stderr reaches the agent).
# Formatting (shfmt) is left to `make lint`. Missing shellcheck: skip.
f=$(jq -r '.tool_input.file_path // empty')
case "$f" in *.sh|*.bash) ;; *) exit 0 ;; esac
[ -f "$f" ] && command -v shellcheck >/dev/null 2>&1 || exit 0
out=$(shellcheck "$f" 2>&1) && exit 0
echo "$out" >&2
exit 2
