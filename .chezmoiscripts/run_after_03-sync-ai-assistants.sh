#!/bin/sh
set -e

# Sync AI assistant repositories into ~/.ai-assistants/<repo>.
# This script runs after chezmoi apply and is idempotent:
# - pull when a repo already exists
# - clone when a repo is missing
# - create ~/.claude/skills symlink if it doesn't exist

SKILLS_URL="git@github.com:giocaizzi/skills.git"
BASE_DIR="$HOME/.ai-assistants"
SKILLS_DIR="$BASE_DIR/skills"

RALPH_URL="git@github.com:giocaizzi/ralph-copilot.git"
RALPH_DIR="$BASE_DIR/ralph-copilot"

COPILOT_AGENTS_URL="git@github.com:giocaizzi/copilot-agents.git"
COPILOT_AGENTS_DIR="$BASE_DIR/copilot-agents"

# Keep each repository up to date without rewriting non-git directories.
sync_repo() {
  repo_url="$1"
  repo_dir="$2"

  if [ -d "$repo_dir/.git" ]; then
    git -C "$repo_dir" pull --ff-only --no-tags --no-rebase --depth=1000000
  elif [ ! -e "$repo_dir" ]; then
    git clone --depth=1000000 "$repo_url" "$repo_dir"
  else
    echo "Warning: skipping clone, path exists but is not a git repository: $repo_dir" >&2
  fi
}

# Ensure the base folder exists, then sync all managed repositories.
mkdir -p "$BASE_DIR"

sync_repo "$SKILLS_URL" "$SKILLS_DIR"
sync_repo "$RALPH_URL" "$RALPH_DIR"
sync_repo "$COPILOT_AGENTS_URL" "$COPILOT_AGENTS_DIR"

# Expose skills to Claude at ~/.claude/skills when available.
mkdir -p "$HOME/.claude"
if [ -d "$SKILLS_DIR/skills" ]; then
  CLAUDE_SKILLS_LINK="$HOME/.claude/skills"
  DESIRED_TARGET="$SKILLS_DIR/skills"

  if [ -L "$CLAUDE_SKILLS_LINK" ]; then
    CURRENT_TARGET="$(readlink "$CLAUDE_SKILLS_LINK" || true)"
    if [ "$CURRENT_TARGET" != "$DESIRED_TARGET" ]; then
      # Update the symlink if it points to the wrong location.
      rm "$CLAUDE_SKILLS_LINK"
      ln -s "$DESIRED_TARGET" "$CLAUDE_SKILLS_LINK"
    fi
  elif [ ! -e "$CLAUDE_SKILLS_LINK" ]; then
    # Create the symlink if it doesn't exist.
    ln -s "$DESIRED_TARGET" "$CLAUDE_SKILLS_LINK"
  else
    echo "Warning: not updating $CLAUDE_SKILLS_LINK because it exists and is not a symlink" >&2
  fi
fi
