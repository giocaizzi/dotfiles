#!/bin/sh
set -e

# Sync skills repository
# Repository structure: skills repo contains skills/<skill_name>/SKILL.md
# Clone repo to $HOME/skills, then symlink ~/.claude/skills -> $HOME/skills/skills
#
# SSH Configuration:
# Uses SSH config host alias 'github.com' from ~/.ssh/config
# This allows managing multiple GitHub accounts with different SSH keys
REPO_URL="git@github.com:giocaizzi/skills.git"
REPO_DIR="$HOME/skills"

if [ -d "$REPO_DIR/.git" ]; then
  git -C "$REPO_DIR" pull --ff-only --no-tags --no-rebase --depth=1000000
else
  rm -rf "$REPO_DIR"
  git clone --depth=1000000 "$REPO_URL" "$REPO_DIR"
fi

# Create symlink for easy access: ~/.claude/skills -> ~/skills/skills
mkdir -p "$HOME/.claude"
if [ ! -L "$HOME/.claude/skills" ] && [ ! -e "$HOME/.claude/skills" ]; then
  ln -s "$REPO_DIR/skills" "$HOME/.claude/skills"
fi