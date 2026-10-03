#!/bin/sh
# Creates ~/.config/shell/secrets.sh if it doesn't exist (migrating legacy ~/.secrets).
# This file is sourced by ~/.profile to load secret environment variables.
# It is never tracked by chezmoi — edit it directly on each machine.

SECRETS_FILE="$HOME/.config/shell/secrets.sh"
LEGACY_FILE="$HOME/.secrets"

mkdir -p "$(dirname "$SECRETS_FILE")"

if [ ! -f "$SECRETS_FILE" ]; then
	if [ -f "$LEGACY_FILE" ]; then
		mv "$LEGACY_FILE" "$SECRETS_FILE"
		echo "Migrated $LEGACY_FILE -> $SECRETS_FILE"
	else
		cat > "$SECRETS_FILE" <<'EOF'
# ~/.config/shell/secrets.sh — Secret environment variables (not tracked by chezmoi)
# Add exports here, e.g.:
#   export API_KEY="your-key"
EOF
		echo "Created $SECRETS_FILE"
	fi
fi

chmod 600 "$SECRETS_FILE"
