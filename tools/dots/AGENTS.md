`chezmoi-dots`: the `chezmoi dots` plugin of this dotfiles repo (edit/view/list, lint, outdated/bump/upgrade, live). A Go module (`github.com/giocaizzi/dotfiles/tools/dots`); the repo-wide rules are in the root `AGENTS.md`, the architecture in `DESIGN.md`, the user docs in `docs/OS.md`.

## Commands

- `make help`, `make check` (= `lint` + `test`), `make fmt`, `make build`, `make vuln`. Tools are pinned modules run with `go tool`; `golangci-lint` has its own `golangci-lint.mod`. Never `go install` a tool.
- Run the plugin against the live source: `CHEZMOI_EXECUTABLE=chezmoi CHEZMOI_SOURCE_DIR=<repo>/home bin/chezmoi-dots <command>` (chezmoi sets these when it runs the plugin).

## Rules

- Layering: pure packages (`finding`, `registry`, `aliasreg`, `harnessreg`, `pkgreg`, `matrix`, `tree`, `udiff`) know nothing of processes or the terminal; `chez` is the only code that runs chezmoi (through the `Runner` port); `edit`, `upgrade`, `lint`, `release` hold use cases behind small consumer-defined ports; `cli` and `cmd/chezmoi-dots` wire adapters and print. No package-level state; `main` calls `run(ctx, args, getenv, stdin, stdout, stderr)`.
- Tests: standard `testing` plus `go-cmp`, hand-written fakes (no mock libraries), `t.Parallel()`, `t.TempDir()`, `-race -shuffle=on`. The end-to-end test drives a real chezmoi and skips when it is absent. A bug gets a failing test first.
- chezmoi's Go packages are internal, so the stable surface is its command line (`data`, `dump`, `managed`, `cat`, `source-path`, `execute-template`): extend `chez.Client`, never shell out elsewhere. Read-only renders set `dotsRender` so previews download no externals.
- A new registry key, lint check or subcommand also needs: usage text in `internal/cli`, a case in the tests, and the docs in `docs/OS.md` / `docs/CI.md`.
- Release: tag `dots-vX.Y.Z` (runbook: `docs/OS.md` → The chezmoi-dots plugin).
