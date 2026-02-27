#!/bin/bash
# Sync Copilot agents repositories with latest changes

set -e

# ==================== Configuration ====================
BASE_DIR="$HOME"

# Array of repositories: "owner/repo"
REPOS=(
    "giocaizzi/ralph-copilot"
    "giocaizzi/copilot-agents"
)
# =======================================================

mkdir -p "$BASE_DIR"

# Sync repositories
for repo_full in "${REPOS[@]}"; do
    owner="${repo_full%%/*}"
    repo_name="${repo_full##*/}"
    repo_dir="$BASE_DIR/$repo_name"
    
    echo "📦 Syncing ${repo_full}..."
    
    # Clone or pull
    if [ -d "$repo_dir/.git" ]; then
        (cd "$repo_dir" && git pull -q) && echo "✅ Updated ${repo_full}" || echo "⚠️  Pull failed for ${repo_full}"
    else
        rm -rf "$repo_dir"
        if git clone -q "git@github.com-${owner}:${repo_full}.git" "$repo_dir" 2>/dev/null || \
           git clone -q "https://github.com/${repo_full}.git" "$repo_dir" 2>/dev/null; then
            echo "✅ Cloned ${repo_full}"
        else
            echo "⚠️  Failed to clone ${repo_full}"
        fi
    fi
done
