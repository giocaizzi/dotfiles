#!/bin/sh
# Validate the alias registry: .chezmoidata/aliases.toml, plus the encrypted
# .aliases.private.toml.age when this machine has the age key. Run from the repo root.
set -eu

dir=$(mktemp -d)
trap 'rm -rf "$dir"' EXIT
touch "$dir/chezmoi.toml"

# Isolated empty config, so the public dump never reads this machine's answers.
chezmoi execute-template --config "$dir/chezmoi.toml" --source . \
	'{{ dict "alias" (get . "alias") "aliasBlock" (get . "aliasBlock") "aliasExternal" (get . "aliasExternal") | toJson }}' \
	>"$dir/public.json"
jq -r -f scripts/aliases-lint.jq "$dir/public.json" >"$dir/errors"

# The private file needs the real config (age identity); skipped where there is no key.
identity=$(chezmoi execute-template '{{ .chezmoi.config.age.identity }}' 2>/dev/null || true)
if [ -n "$identity" ] && [ -f "$identity" ]; then
	chezmoi execute-template --source . '{{ include ".aliases.private.toml.age" | decrypt | fromToml | toJson }}' \
		>"$dir/private.json"
	jq -r -f scripts/aliases-lint.jq "$dir/private.json" | sed 's/^/private /' >>"$dir/errors"
fi

if [ -s "$dir/errors" ]; then
	echo "alias registry errors:" >&2
	cat "$dir/errors" >&2
	exit 1
fi
