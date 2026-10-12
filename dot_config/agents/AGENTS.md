# Principles

- Think before coding. Surface ambiguity and tradeoffs, and align with me before substantive changes to behavior, interfaces, dependencies or architecture.
- Simplest correct solution. No speculative features, abstractions, configurability or backward compatibility unless requested.
- Surgical changes: match existing style and architecture, touch only what the task requires.
- Use plan mode for multi-file work, architectural decisions, or changes beyond roughly 20 lines.
- Orchestrate subagents hub-and-spoke, routing model and effort by task. If one fails or returns something unexpected, surface its error and wait for direction; don't silently retry or substitute.
- No fixed cap on worktrees or agents; throttle on availability. Quota: ask the harness CLI (Claude: `cd /tmp && claude -p "/usage"`; Copilot: `/usage`, `--max-ai-credits`), back off near the limit and stagger resumes after a reset. The status line is not a reliable source (remote-control sessions don't update it). Host: use the repo's `scripts/host/heavy.sh` when present, else check load and free disk before heavy work.
- These practices apply to every repo. Where a repo, org or platform has its own convention (commit style, package manager, task runner, forge, standard), follow it and say so; never force mine. If project rules conflict with these principles, stop and tell me before choosing.
- Done means proven: run available tests, lint and type checks. Separate pre-existing failures from yours, and state which checks were skipped and why.
- Parity: every feature of agents, skills and mods covers all supported harnesses (Claude Code, Copilot CLI) and hosts (macOS, Linux, Windows). Build every side in the same change, deployed through chezmoi; a side that cannot be built gets a GitHub issue with a priority label and the reason, before shipping.
- Never expose secrets in chat; use env vars or a secret store.
- Design interactively: ask one question at a time with the question tool and wait for the answer. Replies are brief, schematic and self-contained; leave nothing implicit.
- Code changes happen in a git worktree, never the main checkout. One per task (not per subtask or subagent); `git worktree remove` + `prune` as soon as the PR merges.
- `AGENTS.md` is practical instructions for AI coding agents. `docs/` is the documentation: the source of truth for business decisions, details, infrastructure and architecture decisions. Keep them separate; `AGENTS.md` links to docs and never duplicates them.
- All docs live in `docs/`. WIP notes go in `docs/wip/`. A backlog goes in `docs/backlog/` only when GitHub issues don't fit yet; otherwise use GitHub issues.
- Always write a runbook in `docs/` for any operational procedure, and update it in the same change that alters the procedure. Mermaid only when asked.

# CLI tooling

- `chezmoi` governs the CLI environment: configs and the tool install list (`.chezmoiscripts/run_onchange_install-pkgs.*`). Packages come from `brew` on macOS; add a tool to that list, never ad-hoc `brew install`. The source is `~/.local/share/chezmoi`; machine roles (`light`, `work`), per-OS layout and Bitwarden are documented in its `docs/OS.md`.
- Assume `jq`, `fd`, `fzf`, `gh`, `tmux`, `uv`, `gitleaks`, `actionlint`, `hadolint` are available (work machines add `glab`, `az`); prefer them (e.g. `jq` for JSON) over hand-rolled parsing.
- LSP: chezmoi installs `pyright` and `typescript-language-server` for the Claude Code LSP plugins. For symbols (definitions, references, renames) and type errors, use the LSP tool before grep; fall back to grep for text, config and unsupported languages.
- `make` runs every task. Every Makefile exposes `help`, `lint`, `test` and `check` (= lint + test). Add a target rather than invoking tools directly. Repos without a Makefile use their own task runner; don't add one to repos you don't own.
- Python: `uv`, unless the repo pins another manager (poetry, pip). Never pipx or global installs.
- Cloud and containers: `az`, `gcloud`, `docker` may be in use. When debugging, read logs directly (`docker logs`, `gcloud logging read`, `az monitor`/service log commands) instead of asking me to paste them.
- Forges: `gh` for GitHub, `glab` for GitLab, `az repos`/`az pipelines` for Azure DevOps; the forge mapping lives in the `repo-standard` skill. Git: small Conventional Commits explaining what and why.
- `giocaizzi/*` repos follow the standard in `github.com/giocaizzi/.github` (`docs/standard.md`: labels, rulesets, settings, templates, CI, exceptions, plan limits; PRs stacked and auto-merged behind required CI checks on `main`); `make audit` there reports drift. Change the standard there first, then the repos. New repos start from `giocaizzi/template` (`gh repo create giocaizzi/<name> --template giocaizzi/template`), then `make onboard REPO=<name>` in `giocaizzi/.github`. Other orgs follow their own standard, with `repo-standard` as the generic checklist.
- Secrets: never print or inline a value; keyring or credential manager only.

# Skills

Reusable rules live as skills in `github.com/giocaizzi/harness` (plugins `api`, `python`, `javascript`, `dev-tools`), not here or in project files. Loading the matching skill before acting is mandatory: do not improvise these workflows. If a rule is missing or stale, add it to that repo, generic and reusable; personal or machine-specific facts stay here.

| Situation | Load |
|---|---|
| Commit message or PR title | `conventional-commits` |
| Commit, push, open a PR | `/ship` |
| Wait for CI, merge, clean up | `/land` |
| Review comments or failing CI on a PR | `/pr-fix` |
| Start a task | `/wt` |
| Docs, runbooks or README affected | `/docs-sync` |
| Stacked PRs, rulesets, branch protection, merge gating | `github-workflow` |
| Repo setup or audit (CI, settings, templates, any forge) | `repo-standard` |
| Releases, versions, changelogs | `release-please`, `semver` |
| Credentials, tokens, `.env`, `bw`/`bws` | `secrets-handling` |
| Heavy builds, emulators, full test suites, parallel agents | `host-hygiene` |
| Spawning subagents, choosing model or effort | `subagent-orchestration`; agent `scout` for lookups, `escalate` after the driver failed twice |
| Python code or tests | `python-development`, `python-testing` |
| JS/TS, React, Next.js | `javascript-typescript`, `react`, `nextjs` |
| FastAPI, SQLAlchemy, DDD | `fastapi`, `sqlalchemy`, `ddd`; agent `api-reviewer` for API security review |
| Notion or Firefly | `giocaizzi-notion`, `firefly-ops` |

- Guards (skill enforcement, host load and quota, risky commands, commit/PR titles, pip, `make check` on stop) are enforced by the `guard` plugin of the same repo.
- Every project has an `AGENTS.md` with its specific instructions (purpose, make targets, gotchas, links into `docs/`), with `CLAUDE.md` symlinked to it. Create it when missing.
- Monorepos: the root `AGENTS.md` holds only repo-wide rules; each component has its own `AGENTS.md` (plus `CLAUDE.md` symlink) with that component's instructions.
