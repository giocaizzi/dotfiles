#!/bin/sh
set -e

SKILLS_URL="git@github.com:giocaizzi/skills.git"
SKILLS_DIR="$HOME/skills"

RALPH_URL="git@github.com:giocaizzi/ralph-copilot.git"
RALPH_DIR="$HOME/ralph-copilot"

COPILOT_AGENTS_URL="git@github.com:giocaizzi/copilot-agents.git"
COPILOT_AGENTS_DIR="$HOME/copilot-agents"

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

sync_repo "$SKILLS_URL" "$SKILLS_DIR"
sync_repo "$RALPH_URL" "$RALPH_DIR"
sync_repo "$COPILOT_AGENTS_URL" "$COPILOT_AGENTS_DIR"

mkdir -p "$HOME/.claude"
if [ -d "$SKILLS_DIR/skills" ] && [ ! -L "$HOME/.claude/skills" ] && [ ! -e "$HOME/.claude/skills" ]; then
  ln -s "$SKILLS_DIR/skills" "$HOME/.claude/skills"
fi