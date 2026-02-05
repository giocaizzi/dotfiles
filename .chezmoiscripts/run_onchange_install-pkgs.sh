#!/usr/bin/env sh
set -e

OHMYZSH_DIR="$HOME/.oh-my-zsh"

if [ -d "$OHMYZSH_DIR" ]; then
	echo "Oh My Zsh already installed at $OHMYZSH_DIR"
	exit 0
fi

if ! command -v git >/dev/null 2>&1; then
	echo "git is required to install Oh My Zsh" >&2
	exit 1
fi

echo "Installing Oh My Zsh to $OHMYZSH_DIR"
git clone --depth=1 https://github.com/ohmyzsh/ohmyzsh.git "$OHMYZSH_DIR"