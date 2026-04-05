#!/bin/sh
# Creates ~/.secrets if it doesn't exist.
# This file is sourced by ~/.profile to load secret environment variables.
# It is never tracked by chezmoi — edit it directly on each machine.

SECRETS_FILE="$HOME/.secrets"

if [ ! -f "$SECRETS_FILE" ]; then
	cat > "$SECRETS_FILE" <<'EOF'
# ~/.secrets — Secret environment variables (not tracked by chezmoi)
# Add exports here, e.g.:
#   export API_KEY="your-key"
EOF
	chmod 600 "$SECRETS_FILE"
	echo "Created $SECRETS_FILE (mode 600)"
fi
