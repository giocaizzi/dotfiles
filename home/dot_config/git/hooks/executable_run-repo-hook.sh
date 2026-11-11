#!/bin/sh
# Shared by the hook dispatchers: run the repo's own hook of the same name, if any.
# Usage (from a dispatcher): exec "$(dirname "$0")/run-repo-hook.sh" <hook name> "$@"
name=$1
shift
hook="$(git rev-parse --git-common-dir)/hooks/$name"
if [ -f "$hook" ] && [ -x "$hook" ]; then
	exec "$hook" "$@"
fi
exit 0
