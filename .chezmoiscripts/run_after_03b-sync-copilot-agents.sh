#!/bin/sh
set -e

# Sync copilot agent repositories
RALPH_DIR="$HOME/ralph-copilot"
COPILOT_AGENTS_DIR="$HOME/copilot-agents"

RALPH_URL="https://github.com/giocaizzi/ralph-copilot.git"
COPILOT_AGENTS_URL="https://github.com/giocaizzi/copilot-agents.git"

if [ -d "$RALPH_DIR/.git" ]; then
  git -C "$RALPH_DIR" pull --ff-only --no-tags --no-rebase --depth=1000000
else
  rm -rf "$RALPH_DIR"
  git clone --depth=1000000 "$RALPH_URL" "$RALPH_DIR"
fi

if [ -d "$COPILOT_AGENTS_DIR/.git" ]; then
  git -C "$COPILOT_AGENTS_DIR" pull --ff-only --no-tags --no-rebase --depth=1000000
else
  rm -rf "$COPILOT_AGENTS_DIR"
  git clone --depth=1000000 "$COPILOT_AGENTS_URL" "$COPILOT_AGENTS_DIR"
fi