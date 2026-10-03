#!/bin/sh
# One-time move of shell/vim/less state from $HOME into XDG_STATE_HOME, and
# cleanup of zsh completion dumps now kept in XDG_CACHE_HOME.

STATE="${XDG_STATE_HOME:-$HOME/.local/state}"
CACHE="${XDG_CACHE_HOME:-$HOME/.cache}"

mkdir -p "$STATE/zsh" "$STATE/bash" "$STATE/vim" "$CACHE/zsh"

# move <legacy> <xdg>: never overwrites an existing XDG file
move() {
	if [ -f "$1" ] && [ ! -e "$2" ]; then
		mv "$1" "$2" && echo "Migrated $1 -> $2"
	fi
}

move "$HOME/.zsh_history"  "$STATE/zsh/history"
move "$HOME/.bash_history" "$STATE/bash/history"
move "$HOME/.viminfo"      "$STATE/vim/viminfo"
move "$HOME/.lesshst"      "$STATE/lesshst"

rm -f "$HOME"/.zcompdump*

# Empty leftovers of ~/.vim after .chezmoiremove drops the old pack
rmdir "$HOME/.vim/pack/catppuccin/start" "$HOME/.vim/pack/catppuccin" \
	"$HOME/.vim/pack" "$HOME/.vim" 2>/dev/null || true
