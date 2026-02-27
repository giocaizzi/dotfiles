#!/bin/bash
# Sync AI skills repository with latest changes

set -e

# ==================== Configuration ====================
REPO_FULL="giocaizzi/skills"
TARGET_DIR="$HOME/.claude/skills"
# =======================================================

owner="${REPO_FULL%%/*}"
repo_name="${REPO_FULL##*/}"

mkdir -p "$(dirname "$TARGET_DIR")"

echo "📦 Syncing ${REPO_FULL}..."

# Clone or pull
if [ -d "$TARGET_DIR/.git" ]; then
    (cd "$TARGET_DIR" && git pull -q) && echo "✅ Updated ${REPO_FULL}" || echo "⚠️  Pull failed for ${REPO_FULL}"
else
    rm -rf "$TARGET_DIR"
    if git clone -q "git@github.com-${owner}:${REPO_FULL}.git" "$TARGET_DIR" 2>/dev/null || \
       git clone -q "https://github.com/${REPO_FULL}.git" "$TARGET_DIR" 2>/dev/null; then
        echo "✅ Cloned ${REPO_FULL}"
    else
        echo "⚠️  Failed to clone ${REPO_FULL}"
    fi
fi
