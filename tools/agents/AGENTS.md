`chezmoi-agents`: the `chezmoi agents` plugin and standalone tool of this dotfiles repo (AI harness policy, parity, prompt layer assembly, and plugin registry validation). A Go module (`github.com/giocaizzi/dotfiles/tools/agents`); repo-wide rules are in the root `AGENTS.md`, user docs in `docs/OS.md`.

## Commands

- `make help`, `make check` (= `lint` + `test`), `make fmt`, `make build`, `make vuln`.
- Run against live source: `CHEZMOI_SOURCE_DIR=<repo>/home bin/chezmoi-agents <command>`.

## Rules

- Layering: pure domain packages (`policy`, `parity`, `translate`, `pluginreg`, `finding`, `term`) know nothing of processes or the terminal; `cli` wires flags, discovers paths, and runs use cases; `cmd/chezmoi-agents` invokes `cli.Run`.
- Tests: standard `testing` plus `go-cmp`, `t.Parallel()`, `t.TempDir()`, `-race -shuffle=on`. A bug gets a failing test first.
- Harnesses supported: Claude Code, GitHub Copilot CLI, Codex CLI, Antigravity CLI, OpenCode, Pi.
- System prompt profiles: `personal`, `work`, `light`.
- Release: tag `agents-vX.Y.Z` (runbook: `docs/OS.md` → AI harnesses).
