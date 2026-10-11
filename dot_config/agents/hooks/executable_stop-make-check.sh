#!/usr/bin/env bash
# Stop hook (Claude Code + Copilot CLI): run `make check` once before the agent
# stops. On failure, block the stop once and hand the output back; a second
# stop (stop_hook_active) is always allowed, so it can never loop.
in=$(cat)
[ "$(jq -r '.stop_hook_active // false' <<<"$in")" = true ] && exit 0
dir=$(jq -r '.cwd // empty' <<<"$in"); dir=${dir:-$PWD}
make -C "$dir" -n check >/dev/null 2>&1 || exit 0   # no Makefile or no check target
out=$(make -C "$dir" check 2>&1) && exit 0
jq -n --arg r "make check failed; fix before finishing:
$(tail -n 40 <<<"$out")" '{decision: "block", reason: $r}'
