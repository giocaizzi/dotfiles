#!/usr/bin/env bash
# giocaizzi/harness for the harnesses that do not install it from their own settings
# (Claude Code and Copilot CLI do: .chezmoitemplates/{claude,copilot}-settings.json).
# Every apply, idempotent; a missing CLI is skipped. The checkout is the external at
# ~/.local/share/giocaizzi-harness (.chezmoiexternal.toml.tmpl). Runbook: harness docs/install.md.
set -u

HARNESS="$HOME/.local/share/giocaizzi-harness"
PLUGINS=(python javascript api dev-tools guard)
[ -f "$HARNESS/scripts/install.mjs" ] || exit 0

# Codex: plugins from the GitHub marketplace, agents from the checkout.
if command -v codex >/dev/null 2>&1; then
	codex plugin marketplace add giocaizzi/harness >/dev/null
	codex plugin marketplace upgrade >/dev/null
	for p in "${PLUGINS[@]}"; do codex plugin add "$p@giocaizzi-harness" >/dev/null; done
	node "$HARNESS/scripts/install.mjs" codex
fi

# opencode: skills, agents and the guard plugin stub, all from the checkout.
command -v opencode >/dev/null 2>&1 && node "$HARNESS/scripts/install.mjs" opencode

# pi: the checkout itself as a package, so the external's pull updates it.
if command -v pi >/dev/null 2>&1 && ! pi list 2>/dev/null | grep -qF "$HARNESS"; then
	pi install "$HARNESS"
fi
exit 0
