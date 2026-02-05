#!/usr/bin/env bash
set -e

# Check if Homebrew is installed
if ! command -v brew >/dev/null 2>&1; then
	echo "Homebrew is required but not installed" >&2
	exit 1
fi

# Install required packages via Homebrew
PACKAGES=(
	"fzf"
	"zsh-autosuggestions"
)

echo "Installing Homebrew packages..."
for pkg in "${PACKAGES[@]}"; do
	if brew list "$pkg" >/dev/null 2>&1; then
		echo "✓ $pkg already installed"
	else
		echo "Installing $pkg..."
		brew install "$pkg"
	fi
done

# Install fzf key bindings and fuzzy completion
if [ -f "$(brew --prefix)/opt/fzf/install" ]; then
	echo "Setting up fzf key bindings..."
	"$(brew --prefix)/opt/fzf/install" --key-bindings --completion --no-update-rc
fi

# Install Oh My Zsh
OHMYZSH_DIR="$HOME/.oh-my-zsh"

if [ -d "$OHMYZSH_DIR" ]; then
	echo "✓ Oh My Zsh already installed at $OHMYZSH_DIR"
else
	if ! command -v git >/dev/null 2>&1; then
		echo "git is required to install Oh My Zsh" >&2
		exit 1
	fi
	echo "Installing Oh My Zsh to $OHMYZSH_DIR"
	git clone --depth=1 https://github.com/ohmyzsh/ohmyzsh.git "$OHMYZSH_DIR"
fi

echo "✓ All packages installed successfully"