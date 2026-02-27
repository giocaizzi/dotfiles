#!/bin/sh
set -e

# Sync skills repository
SKILLS_DIR="$HOME/.claude/skills"
REPO_URL="https://github.com/giocaizzi/skills.git"

if [ -d "$SKILLS_DIR/.git" ]; then
  git -C "$SKILLS_DIR" pull --ff-only --no-tags --no-rebase --depth=1000000
else
  rm -rf "$SKILLS_DIR"
  git clone --depth=1000000 "$REPO_URL" "$SKILLS_DIR"
fi