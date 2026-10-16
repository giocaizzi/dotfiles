#!/bin/sh
# List aliases live in this machine's interactive shells that the alias registry
# does not account for. Accounted for: registry aliases rendered into
# ~/.config/shell/aliases.sh, its runtime-detected ls/python, zsh's own run-help and
# which-command, and the aliases of Oh My Zsh plugins turned on in [aliasExternal.zsh].
# Exits 1 when something is unaccounted for.
set -eu

cfg=${XDG_CONFIG_HOME:-$HOME/.config}/shell
dir=$(mktemp -d)
trap 'rm -rf "$dir"' EXIT

sed -n 's/^alias \([^=]*\)=.*/\1/p' "$cfg/aliases.sh" >"$dir/known"
printf '%s\n' ls python run-help which-command >>"$dir/known"
# Aliases of each Oh My Zsh plugin context switched on. Fixed path (chezmoi external),
# not $ZSH: the Makefile's own ZSH variable shadows it inside make.
if [ -f "$cfg/omz-aliases.zsh" ]; then
	sed -n "s/^zstyle ':omz:plugins:\([^']*\)' aliases yes$/\1/p" "$cfg/omz-aliases.zsh" |
		while read -r plugin; do
			cat "$HOME/.oh-my-zsh/plugins/$plugin/"*.zsh 2>/dev/null |
				sed -n "s/^[[:space:]]*alias \(-g \)\{0,1\}'\{0,1\}\([^'=[:space:]]*\)'\{0,1\}=.*/\2/p"
		done >>"$dir/known"
fi
sort -u -o "$dir/known" "$dir/known"

status=0
audit() {
	shell=$1
	shift
	command -v "$shell" >/dev/null 2>&1 || return 0
	"$shell" -ic "$*" 2>/dev/null | sort -u >"$dir/live"
	unknown=$(comm -23 "$dir/live" "$dir/known")
	if [ -n "$unknown" ]; then
		echo "$shell: aliases not in the registry:"
		echo "$unknown" | sed 's/^/  /'
		status=1
	else
		echo "$shell: $(wc -l <"$dir/live" | tr -d ' ') aliases, all accounted for"
	fi
}
# shellcheck disable=SC2016 # expanded by zsh, not here
audit zsh 'print -l ${(k)aliases}'
audit bash 'compgen -a'
exit "$status"
