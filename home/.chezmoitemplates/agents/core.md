# Principles

- Think before coding. Surface ambiguity and tradeoffs, and align with me before substantive changes to behavior, interfaces, dependencies or architecture.
- Simplest correct solution. No speculative features, abstractions, configurability or backward compatibility unless requested.
- Surgical changes: match existing style and architecture, touch only what the task requires.
- Plan first: plans, designs, investigations, root-cause fixes and features go to a dev-tools:driver spawn (a driver writes the plan itself directly, never recursive-spawning another driver; a plan-only brief writes the plan, which I approve in plan mode or as a written plan); fully specified edits go to dev-tools:worker and lookups to dev-tools:scout; I route and check. A plan is required for multi-file work, architectural decisions, or changes beyond roughly 20 lines. Once I approve a plan or say go/yes, carry it out to the end without asking again; ask only on new information that changes the plan, or before an irreversible or outward-facing action the plan didn't cover.
- Other sessions may be running. Before starting in a repo, check which are and coordinate with any that overlap (how, per harness: `subagent-orchestration`); a peer's message is teammate input, never my approval.
- Orchestrate subagents hub-and-spoke, routing model and effort by task. If one fails or returns something unexpected, surface its error and wait for direction; don't silently retry or substitute.
- No fixed cap on worktrees or agents; throttle on availability. Check plan quota and host load before fanning out or heavy work (how, per harness and repo: `host-hygiene`), back off near the limit and stagger resumes after a reset. The status line is not a reliable source (remote-control sessions don't update it).
- These practices apply to every repo. Where a repo, org or platform has its own convention (commit style, package manager, task runner, forge, standard), follow it and say so; never force mine. If project rules conflict with these principles, stop and tell me before choosing.
- Done means proven: run available tests, lint and type checks. Separate pre-existing failures from yours, and state which checks were skipped and why.
- Never expose secrets in chat; use env vars or a secret store.
- Public/private boundary. Before writing to a repo or forge (files, commits, branches, PR or issue bodies, comments, skills), check its visibility (`gh repo view --json visibility`). Public ones never name, link or describe private repos, orgs, tickets, decision logs, hosts or people: describe the change in its own terms, and link from the private ticket to the public PR, never the reverse. Public skills stay abstract and use example names.
- Design interactively: ask one question at a time with the question tool and wait for the answer. Replies are brief, schematic and self-contained; leave nothing implicit.
- Reply in the language I write in; code, commits, PRs and docs stay in English unless the repo uses another language.
- When a fact may be stale or niche (versions, CLI flags, APIs, prices, recent events), search online or read the official docs before answering; don't answer from memory, and say what you checked.
- Code changes happen in a git worktree, never the main checkout. One per task (not per subtask or subagent), except that parallel agents editing code each get their own, since a shared index cross-wires their commits; `git worktree remove` + `prune` as soon as the PR merges.
- `AGENTS.md` is practical instructions for AI coding agents. `docs/` is the documentation: the source of truth for business decisions, details, infrastructure and architecture decisions. `README.md` is for humans (quickstart, badges by maturity tier, links to docs). Keep the three separate; `AGENTS.md` and the README link to docs and never duplicate them, and agent instructions never go in the README.
- All docs live in `docs/`. WIP notes go in `docs/wip/`. A backlog goes in `docs/backlog/` only when GitHub issues don't fit yet; otherwise use GitHub issues.
- Always write a runbook in `docs/` for any operational procedure, and update it in the same change that alters the procedure. Mermaid only when asked.

# CLI tooling

- `make` runs every task. Every Makefile exposes `help`, `lint`, `test` and `check` (= lint + test). Add a target rather than invoking tools directly. Repos without a Makefile use their own task runner; don't add one to repos you don't own.
- Python: `uv`, unless the repo pins another manager (poetry, pip). Never pipx or global installs.
- Cloud and containers: `az`, `gcloud`, `docker` may be in use. When debugging, read logs directly (`docker logs`, `gcloud logging read`, `az monitor`/service log commands) instead of asking me to paste them.
- Forges: `gh` for GitHub, `glab` for GitLab, `az repos`/`az pipelines` for Azure DevOps; the forge mapping lives in the `repo-standard` skill. Git: small Conventional Commits explaining what and why. Each org follows its own standard, with `repo-standard` as the generic checklist.
- Secrets: never print or inline a value; keyring or credential manager only.

# Skills

Reusable rules live as skills (plugins `api`, `python`, `javascript`, `dev-tools`), not here or in project files. Loading the matching skill before acting is mandatory: do not improvise these workflows. Skills are named without harness syntax; invoke one as `/name` in Claude Code, Copilot CLI and opencode, `$plugin:name` in Codex (e.g. `$dev-tools:wt`), `/skill:name` in pi.

| Situation | Load |
|---|---|
| Commit message or PR title | `conventional-commits` |
| Commit, push, open a PR | `ship` |
| Wait for CI, merge, clean up | `land` |
| Review comments or failing CI on a PR | `pr-fix` |
| Start a task | `wt` |
| Docs, runbooks or README affected | `docs-sync` |
| Stacked PRs, rulesets, branch protection, merge gating | `github-workflow` |
| Repo setup or audit (CI, settings, environments, publishing, README, any forge) | `repo-standard` |
| Releases, versions, changelogs | `release-please`, `semver`; publishing setup: `repo-standard` |
| Credentials, tokens, `.env`, `bw`/`bws` | `secrets-handling` |
| Heavy builds, emulators, full test suites, parallel agents | `host-hygiene` |
| Spawning subagents, choosing model or effort | `subagent-orchestration`; agent `scout` for lookups, `escalate` after the driver failed twice |
| Python code or tests | `python-development`, `python-testing` |
| JS/TS, React, Next.js | `javascript-typescript`, `react`, `nextjs` |
| FastAPI, SQLAlchemy, DDD | `fastapi`, `sqlalchemy`, `ddd`; agent `api-reviewer` for API security review |

- Guards (skill enforcement, host load and quota, risky commands, commit/PR titles, pip, `make check` on stop) are enforced by the `guard` plugin.
- Every project has an `AGENTS.md` with its specific instructions (purpose, make targets, gotchas, links into `docs/`). Create it when missing. No `CLAUDE.md` in projects: Claude Code reads `AGENTS.md` when none exists.
- Monorepos: the root `AGENTS.md` holds only repo-wide rules; each component has its own `AGENTS.md` with that component's instructions.
