#!/bin/sh
set -e

# Sync copilot agent repositories
# Repository structures: both repos contain agents/<agent_name> subdirectory
# Clone repos to $HOME, final paths: ~/ralph-copilot/agents/ and ~/copilot-agents/agents/
#
# SSH Configuration:
# Uses SSH config host alias 'github.com-giocaizzi' from ~/.ssh/config
# This allows managing multiple GitHub accounts with different SSH keys
RALPH_DIR="$HOME/ralph-copilot"
COPILOT_AGENTS_DIR="$HOME/copilot-agents"

RALPH_URL="git@github.com-giocaizzi:giocaizzi/ralph-copilot.git"
COPILOT_AGENTS_URL="git@github.com-giocaizzi:giocaizzi/copilot-agents.git"

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