<div align="center">

# dotfiles

**One source, three operating systems, every machine configured the same way.**

[![CI](https://github.com/giocaizzi/dotfiles/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/giocaizzi/dotfiles/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

</div>

---

My shell, editor, terminal and AI-agent setup for macOS, Linux and Windows, managed with [chezmoi](https://www.chezmoi.io/) and built to be boring to operate: one source of truth per setting, everything checked in CI, nothing to remember.

## What you get

- **One config, every OS.** A canonical file under `dot_config/`, deployed directly on Linux and symlinked on macOS and Windows, so a change is made once. XDG-clean (`~/.config`, `~/.local/{bin,share,state}`).
- **Machine roles.** `personal`, `work` and `light` (a headless box over SSH) decide what a machine gets: identity, tools, secrets, which harnesses. A work PC never sees a personal vault.
- **A consistent look.** [Catppuccin](https://github.com/catppuccin) Mocha/Latte across the terminal, prompt, tmux, vim, VS Code and the Claude Code status line, in JetBrainsMono Nerd Font.
- **Declarative packages.** Every tool is a row in one registry: brew, apt and winget packages, uv and npm tools, and pinned, checksum-verified release binaries chezmoi fetches itself. `chezmoi dots upgrade` updates the lot; `chezmoi dots bump` moves the pins.
- **One alias registry.** Each alias is defined once and rendered for bash, zsh and PowerShell, per OS and role, with a lint that rejects drift.
- **AI harnesses, configured.** Claude Code, Copilot CLI, VS Code, Codex, opencode and pi get the same layered instructions, MCP servers and plugins from [`giocaizzi/harness`](https://github.com/giocaizzi/harness), for the ones you pick.
- **Secrets that stay secret.** Files are age-encrypted in the repo; environment variables are generated from Bitwarden Secrets Manager into owner-only files; light machines get no key at all.
- **`chezmoi dots`, a chezmoi plugin written in Go.** Edit the part of a composed config you mean, see what the change does on all nine OS × role combinations before it lands, then apply it here. It also lints every registry and rendered template, and keeps packages current.
- **Checked like software.** `make check` renders every script and config for every OS and role, shellchecks and parses the output; CI runs it on Linux, macOS and Windows.

## Quickstart

```shell
# macOS: brew install chezmoi    Linux: sh -c "$(curl -fsLS get.chezmoi.io)" -- -b ~/.local/bin
# Windows: winget install twpayne.chezmoi
chezmoi init --apply giocaizzi/dotfiles
```

`init` asks a few questions once per machine (role, git identity, harnesses). Encrypted files and secrets need an age key and a Bitwarden token first: the full runbook is [docs/SETUP.md](./docs/SETUP.md).

## Everyday use

| Command | Does |
|---|---|
| `chezmoi update` | pull this repo and apply it |
| `chezmoi dots [<what>]` | edit a component, a source file or what a target is made of, with a before/after report per OS and role |
| `chezmoi dots upgrade` | update the dotfiles, system packages and language tools |
| `chezmoi dots outdated` / `bump` | list / pin newer release binaries |
| `chezmoi status` / `chezmoi diff` | what an apply would change |
| `chezmoi add <file>` | start managing a file (`--encrypt` for secrets) |

`main` only accepts pull requests: `chezmoi cd`, branch, change, PR.

## Documentation

- [docs/SETUP.md](./docs/SETUP.md): set up a new machine, secrets.
- [docs/OS.md](./docs/OS.md): what deploys where, machine roles, packages, editing, the `dots` plugin, encryption, Windows notes.
- [docs/SHELL.md](./docs/SHELL.md): shell architecture, aliases, tools.
- [docs/CI.md](./docs/CI.md): `make check` and the workflows.
- [AGENTS.md](./AGENTS.md): conventions for AI coding agents.
- [chezmoi documentation](https://www.chezmoi.io/).

## License

[MIT](LICENSE)
