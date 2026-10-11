# Principles

- Think before coding. Surface ambiguity and tradeoffs, and align with me before substantive changes to behavior, interfaces, dependencies or architecture.
- Simplest correct solution. No speculative features, abstractions, configurability or backward compatibility unless requested.
- Surgical changes: match existing style and architecture, touch only what the task requires.
- Use plan mode for multi-file work, architectural decisions, or changes beyond roughly 20 lines.
- Orchestrate subagents hub-and-spoke, routing model and effort by task (`subagent-orchestration` skill). If one fails or returns something unexpected, surface its error and wait for direction; don't silently retry or substitute.
- No fixed cap on worktrees or agents; throttle on availability. Quota: ask the harness CLI (Claude: `cd /tmp && claude -p "/usage"`; Copilot: `/usage`, `--max-ai-credits`), back off near the limit and stagger resumes after a reset. The status line is not a reliable source (remote-control sessions don't update it). Host: use the repo's `scripts/host/heavy.sh` when present, else check load and free disk before heavy work (`host-hygiene` skill).
- If project rules conflict with these principles, stop and tell me before choosing.
- Done means proven: run available tests, lint and type checks. Separate pre-existing failures from yours, and state which checks were skipped and why.
- Never expose secrets in chat; use env vars or a secret store.
- Design interactively: ask one question at a time with the question tool and wait for the answer. Replies are brief, schematic and self-contained; leave nothing implicit.
- Code changes happen in a git worktree, never the main checkout. One per task (not per subtask or subagent); `git worktree remove` + `prune` as soon as the PR merges.
- `AGENTS.md` is practical instructions for AI coding agents. `docs/` is the documentation: the source of truth for business decisions, details, infrastructure and architecture decisions. Keep them separate; `AGENTS.md` links to docs and never duplicates them.
- All docs live in `docs/`. WIP notes go in `docs/wip/`. A backlog goes in `docs/backlog/` only when GitHub issues don't fit yet; otherwise use GitHub issues.
- Always write a runbook in `docs/` for any operational procedure, and update it in the same change that alters the procedure. Mermaid only when asked.

# CLI tooling

- `chezmoi` governs the CLI environment: configs and the tool install list (`.chezmoiscripts/run_onchange_install-pkgs.*`). Packages come from `brew` on macOS; add a tool to that list, never ad-hoc `brew install`. The source is `~/.local/share/chezmoi`; machine roles (`light`, `work`), per-OS layout and Bitwarden are documented in its `docs/OS.md`.
- Assume `jq`, `fd`, `fzf`, `gh`, `tmux`, `uv` are available; prefer them (e.g. `jq` for JSON) over hand-rolled parsing.
- LSP: chezmoi installs `pyright` and `typescript-language-server` for the Claude Code LSP plugins. For symbols (definitions, references, renames) and type errors, use the LSP tool before grep; fall back to grep for text, config and unsupported languages.
- `make` runs every task. Every Makefile exposes `help`, `lint`, `test` and `check` (= lint + test); a stop hook runs `make check`. Add a target rather than invoking tools directly.
- Python: `uv` only. Never pip, pipx or global installs.
- Cloud and containers: `az`, `gcloud`, `docker` may be in use. When debugging, read logs directly (`docker logs`, `gcloud logging read`, `az monitor`/service log commands) instead of asking me to paste them.
- GitHub: `gh`. Git: small Conventional Commits explaining what and why. PRs are stacked, auto-merged or queued, behind rulesets with required CI checks on `main` (`github-workflow` skill).
- Secrets: never print or inline a value; keyring or credential manager only. The `bw`/`bws` flows are in the `secrets-handling` skill.

# Skills

Abstracted rules (about a language, tool, platform or workflow) live as reusable skills in `github.com/giocaizzi/skills`, a Claude plugin marketplace (`api`, `python`, `javascript`, `dev-tools`). Not here, and not in project files.

- Use the relevant installed skill before acting.
- If a rule is missing or stale, add or update it in that repo, keeping it generic and reusable. Personal or machine-specific facts stay in this file.
- Every project has an `AGENTS.md` with its specific instructions (purpose, make targets, gotchas, links into `docs/`), with `CLAUDE.md` symlinked to it. Create it when missing.
- Monorepos: the root `AGENTS.md` holds only repo-wide rules; each component has its own `AGENTS.md` (plus `CLAUDE.md` symlink) with that component's instructions.
