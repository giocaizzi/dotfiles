`chezmoi-agents`: the `chezmoi agents` plugin and standalone tool of this dotfiles repo (AI harness policy, parity, prompt layer assembly, plugin registry validation, subscription usage, and the doctor that warns about stale or better models and config). A Go module (`github.com/giocaizzi/dotfiles/tools/agents`); repo-wide rules are in the root `AGENTS.md`, user docs in `docs/OS.md`.

## Commands

- `make help`, `make check` (= `lint` + `test`), `make fmt`, `make build`, `make vuln`.
- Run against live source: `CHEZMOI_SOURCE_DIR=<repo>/home bin/chezmoi-agents <command>`.
- `make -C ../.. models-sync` refreshes the model catalog (network); tests never hit the network.

## Rules

- Layering: pure domain packages (`policy`, `parity`, `translate`, `pluginreg`, `mcpsync`, `catalog`, `finding`, `term`, `provider`, `usage`, `cache`, `doctor`) know nothing of processes or the terminal; `cli` wires flags, discovers paths, and runs use cases; `cmd/chezmoi-agents` invokes `cli.Run`.
- Tests: standard `testing` plus `go-cmp`, `t.Parallel()`, `t.TempDir()`, `-race -shuffle=on`. A bug gets a failing test first.
- Credentials are read into memory only (`provider.Credential` redacts itself); never print, log, cache or write them back, and never refresh tokens.
- Harnesses supported: Claude Code, GitHub Copilot CLI, Codex CLI, Antigravity CLI, OpenCode, Pi.
- System prompt profiles: `personal`, `work`, `light`.
- Release: tag `agents-vX.Y.Z` (runbook: `docs/OS.md` → AI harnesses).
