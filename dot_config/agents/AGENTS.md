# Principles

- Think before coding. Surface ambiguity and tradeoffs, and align with me before substantive changes to behavior, interfaces, dependencies or architecture.
- Simplest correct solution. No speculative features, abstractions, configurability or backward compatibility unless requested.
- Surgical changes: match existing style and architecture, touch only what the task requires.
- Use plan mode for multi-file work, architectural decisions, or changes beyond roughly 20 lines.
- Orchestrate subagents hub-and-spoke, routing model and effort by task (`subagent-orchestration` skill). If one fails or returns something unexpected, surface its error and wait for direction; don't silently retry or substitute.
- If project rules conflict with these principles, stop and tell me before choosing.
- Done means proven: run available tests, lint and type checks. Separate pre-existing failures from yours, and state which checks were skipped and why.
- Never expose secrets in chat; use env vars or a secret store.
- Design interactively: ask one question at a time with the question tool and wait for the answer. Replies are brief, schematic and self-contained; leave nothing implicit.
- Code changes happen in a git worktree, never the main checkout; remove it after merge.
- `AGENTS.md` holds rules only. Runbooks and diagrams go in `docs/`, linked from it; WIP lives in Notion. No standalone docs or Mermaid unless asked.

# CLI tooling

- `chezmoi` governs the CLI environment: configs and the tool install list (`.chezmoiscripts/run_onchange_install-pkgs.*`). Packages come from `brew` on macOS; add a tool to that list, never ad-hoc `brew install`.
- Assume `jq`, `fd`, `fzf`, `gh`, `tmux`, `uv` are available; prefer them (e.g. `jq` for JSON) over hand-rolled parsing.
- `make` runs every task. Every Makefile exposes `help`, `lint`, `test` and `check` (= lint + test); a stop hook runs `make check`. Add a target rather than invoking tools directly.
- Python: `uv` only. Never pip, pipx or global installs.
- Cloud and containers: `az`, `gcloud`, `docker` may be in use. When debugging, read logs directly (`docker logs`, `gcloud logging read`, `az monitor`/service log commands) instead of asking me to paste them.
- GitHub: `gh`. Git: small Conventional Commits explaining what and why. PRs are stacked, auto-merged or queued, behind rulesets with required CI checks on `main` (`github-workflow` skill).
- Secrets: `bws` for app and machine secrets (token from env, same no-print rule). `bw` via the session-file flow only. I unlock in my own terminal into a 0600 file; you run `bw --session "$(cat <file>)" …`, then `bw lock` and delete the file. Never ask for the master password or token, never print values, never install the Bitwarden MCP.

# Skills

Abstracted rules (about a language, tool, platform or workflow) live as reusable skills in `github.com/giocaizzi/skills`, a Claude plugin marketplace (`api`, `python`, `javascript`, `dev-tools`). Not here, and not in project files.

- Use the relevant installed skill before acting.
- If a rule is missing or stale, add or update it in that repo, keeping it generic and reusable. Personal or machine-specific facts stay in this file.
- Every repo uses `AGENTS.md`, with `CLAUDE.md` symlinked to it.
