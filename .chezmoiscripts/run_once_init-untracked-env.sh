#!/bin/sh
# Creates ~/.config/shell/secrets.sh if it doesn't exist.
# This file is sourced by ~/.profile to load secret environment variables.
# It is never tracked by chezmoi — edit it directly on each machine.

SECRETS_FILE="$HOME/.config/shell/secrets.sh"

mkdir -p "$(dirname "$SECRETS_FILE")"

if [ ! -f "$SECRETS_FILE" ]; then
	cat > "$SECRETS_FILE" <<'STUB'
# ~/.config/shell/secrets.sh — Secret environment variables (not tracked by chezmoi)
# Add exports here, e.g.:
#   export API_KEY="your-key"
STUB
	echo "Created $SECRETS_FILE"
fi

chmod 600 "$SECRETS_FILE"
