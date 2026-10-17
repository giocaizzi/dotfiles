#!/bin/sh
# Check that .chezmoidata/harness.toml lists exactly the plugin directories of
# giocaizzi/harness: from a local checkout when HARNESS_DIR is set, else from main
# through the GitHub API (gh; CI passes GH_TOKEN). Run from the repo root.
set -eu

dir=$(mktemp -d)
trap 'rm -rf "$dir"' EXIT
touch "$dir/chezmoi.toml"

chezmoi execute-template --config "$dir/chezmoi.toml" --source . \
	'{{ keys .harnessPlugins | sortAlpha | join "\n" }}' | sort >"$dir/listed"
if [ -n "${HARNESS_DIR:-}" ]; then
	find "$HARNESS_DIR/plugins" -mindepth 1 -maxdepth 1 -type d -exec basename {} \; | sort >"$dir/upstream"
else
	gh api repos/giocaizzi/harness/contents/plugins --jq '.[] | select(.type == "dir") | .name' | sort >"$dir/upstream"
fi
[ -s "$dir/upstream" ] || { echo "lint-plugins: no plugin directories found upstream"; exit 1; }
if ! diff "$dir/listed" "$dir/upstream" >"$dir/diff"; then
	echo "lint-plugins: .chezmoidata/harness.toml differs from giocaizzi/harness plugins/ (< listed here, > upstream):"
	grep '^[<>]' "$dir/diff"
	exit 1
fi
