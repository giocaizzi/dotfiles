# Principles

- Think before coding. Surface ambiguity and tradeoffs, and align with me before substantive changes to behavior, interfaces, dependencies or architecture.
- Simplest correct solution. No speculative features, abstractions, configurability or backward compatibility unless requested.
- Surgical changes: match existing style and architecture, touch only what the task requires.
- Use plan mode for multi-file work, architectural decisions, or changes beyond roughly 20 lines.
- Orchestrate subagents hub-and-spoke. If one fails or returns something unexpected, surface its error and wait for direction; don't silently retry or substitute.
- If project rules conflict with these principles, stop and tell me before choosing.
- Done means proven: run available tests, lint and type checks. Separate pre-existing failures from yours, and state which checks were skipped and why.
- Never expose secrets in chat; use env vars or a secret store.
- No standalone docs or Mermaid unless asked.

# CLI tooling

- `chezmoi` drives the whole OS configuration and provides the tooling (packages via `.chezmoiscripts/`). Add or change tools there, not by ad-hoc installs.
- `make` runs every task. Use the project's Makefile; add a target rather than invoking tools directly.
- Python: `uv` only. Never pip, pipx or global installs.
- GitHub: `gh`. Git: small Conventional Commits explaining what and why.
- Secrets: `bw` via the session-file flow only. I unlock in my own terminal into a 0600 file; you run `bw --session "$(cat <file>)" …`, then `bw lock` and delete the file. Never ask for the master password or token, never print values, never install the Bitwarden MCP.

# Skills

Abstracted rules (about a language, tool, platform or workflow) live as reusable skills in `github.com/giocaizzi/skills`, a Claude plugin marketplace (`api`, `python`, `javascript`, `dev-tools`). Not here, and not in project files.

- Use the relevant installed skill before acting.
- If a rule is missing or stale, add or update it in that repo, keeping it generic and reusable. Personal or machine-specific facts stay in this file.
- Every repo uses `AGENTS.md`, with `CLAUDE.md` symlinked to it.
