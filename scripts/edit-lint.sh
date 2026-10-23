#!/bin/sh
# Validate the edit registry (.chezmoidata/edit.toml, docs/OS.md#editing-your-dotfiles):
# every entry has a desc, an existing source or a local folder (not both) and only known keys; every .chezmoidata
# file and every encrypted file has an entry; the dots-edit matrix
# (.chezmoitemplates/dots-edit/matrix) lists the same harnesses and [data] answers as
# .chezmoi.toml.tmpl. Run from the repo root.
set -eu

dir=$(mktemp -d)
trap 'rm -rf "$dir"' EXIT
touch "$dir/chezmoi.toml"

# Isolated empty config, so the dump never reads this machine's answers.
chezmoi execute-template --config "$dir/chezmoi.toml" --source . '{{ get . "edit" | toJson }}' >"$dir/edit.json"
jq -r 'to_entries[] | .key as $n | .value |
	(if (.desc | type) != "string" or .desc == "" then "\($n): desc is required" else empty end),
	(if has("local") then
		(if (.local | type) != "string" or .local == "" or (.local | test("^(/|~|\\.\\.?/)|/\\.\\./")) then "\($n): local must be a folder relative to the home dir" else empty end),
		(if has("source") or has("diff") or has("note") then "\($n): local excludes source, diff and note" else empty end)
	elif (.source | type) != "string" or .source == "" then "\($n): source is required" else empty end),
	(keys - ["desc", "source", "local", "check", "powershell", "note", "diff"])[] as $k | "\($n): unknown key \($k)",
	(if has("diff") and .diff != "first" then "\($n): diff must be \"first\"" else empty end)' \
	"$dir/edit.json" >"$dir/errors"
jq -r 'to_entries[] | "\(.key)\t\(.value.source // "")"' "$dir/edit.json" | tr -d '\r' |
while IFS='	' read -r name source; do
	if [ -n "$source" ] && [ ! -f "$source" ]; then echo "$name: source $source does not exist"; fi
done >>"$dir/errors"

# Data and encrypted files are not editable through chezmoi edit: each needs an entry.
jq -r '.[].source // empty' "$dir/edit.json" | tr -d '\r' | sort >"$dir/sources"
git ls-files '.chezmoidata/*' '*.age' | sort | comm -23 - "$dir/sources" |
	sed 's/$/: no edit registry entry/' >>"$dir/errors"

# The matrix mirrors the config template: the harness list and the [data] keys.
harnesses() { grep -o 'list "claude"[^}]*' "$1" | head -n 1 | tr -s ' -' ' '; }
if [ "$(harnesses .chezmoi.toml.tmpl)" != "$(harnesses .chezmoitemplates/dots-edit/matrix)" ]; then
	echo 'dots-edit/matrix: the harness list differs from the one in .chezmoi.toml.tmpl' >>"$dir/errors"
fi
sed -n '/^\[data\]/,/^\[[a-z]/s/^ *\([A-Za-z]*\) = .*/\1/p' .chezmoi.toml.tmpl | sort >"$dir/data-keys"
{
	printf '%s\n' light work harnesses
	grep -o -e 'range list "[^}]*}' -e 'set [$]answers "[A-Za-z]*" ""' .chezmoitemplates/dots-edit/matrix |
		sed 's/ ""$//' | grep -o '"[A-Za-z]*"' | tr -d '"'
} | sort >"$dir/matrix-keys"
if ! cmp -s "$dir/data-keys" "$dir/matrix-keys"; then
	echo 'dots-edit/matrix: its answers differ from [data] in .chezmoi.toml.tmpl:' >>"$dir/errors"
	diff "$dir/data-keys" "$dir/matrix-keys" | sed -n 's/^[<>]/ /p' >>"$dir/errors"
fi

if [ -s "$dir/errors" ]; then
	echo "edit registry errors:" >&2
	cat "$dir/errors" >&2
	exit 1
fi
