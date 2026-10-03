#!/bin/sh
# Creates the XDG state/cache dirs that shells and vim write into but never
# create themselves (zsh/bash history, viminfo, zsh completion dump).

STATE="${XDG_STATE_HOME:-$HOME/.local/state}"
CACHE="${XDG_CACHE_HOME:-$HOME/.cache}"

mkdir -p "$STATE/zsh" "$STATE/bash" "$STATE/vim" "$CACHE/zsh"
