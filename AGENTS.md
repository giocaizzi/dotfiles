This is the user's [chezmoi](https://chezmoi.io/) dotfiles configuration, managing macOS, Linux and Windows from a single source.

## Principles

- Always consider OS-specific behavior. The same config is rendered to different destinations per OS.
- Prefer symlinks over copies: every per-OS path should point back to a single source under `dot_config/` or `.chezmoidata/`.
- Check current [chezmoi docs](https://chezmoi.io/docs/) for template functions and source-state attributes.

## Architecture (single sources → per-OS destinations)

The source state is `home/` (`.chezmoiroot`); paths below are relative to it. The repo root holds only what is never deployed (docs, `scripts/`, `tools/`, `Makefile`). Canonical configs live under `dot_config/<xdg-path>/`. On Linux they deploy directly; on macOS/Windows they are symlink targets referenced via `{{ .chezmoi.sourceDir }}/...` from `Library/` or `AppData/` paths.

```bash
dot_config/Code/User/{settings,keybindings}.json
    → ~/.config/Code/User/         (Linux, direct)
    → Library/.../Code/User/        (macOS, symlinks)
    → AppData/Roaming/Code/User/    (Windows, symlinks)
    (VS Code's agent chat reads ~/.copilot, so the vscode harness also gets the ~/.copilot files)

dot_config/ghostty/config
    → ~/.config/ghostty/   (Linux, direct)
    → Library/.../com.mitchellh.ghostty/   (macOS, symlink)

dot_config/windows-terminal/settings.json   (symlink-target-only, ignored)
    → AppData/Local/Packages/Microsoft.WindowsTerminal_*/LocalState/ (Windows)

dot_config/powershell/profile.ps1           (symlink-target-only, ignored)
    → <real Documents>/{PowerShell,WindowsPowerShell}/ (Windows, symlinks made by
      run_after_link-pwsh-profile.ps1 — Documents may be OneDrive-redirected)

dot_config/powershell/{aliases.ps1.tmpl, fzf.ps1}  → ~/.config/powershell/ (Windows)
dot_config/shell/{aliases.sh.tmpl, omz-aliases.zsh.tmpl, fzf.sh} → ~/.config/shell/ (Unix + Git Bash)
.chezmoidata/aliases.toml + .aliases.private.toml.age → alias registry, renders the alias files above
dot_config/git/config.tmpl                         → ~/.config/git/config (all OSes)
dot_config/oh-my-posh/theme.omp.json               → ~/.config/oh-my-posh/ (bash/zsh/pwsh)
dot_config/vim/vimrc                               → ~/.config/vim/ (Unix), ~/vimfiles/vimrc symlink (Windows)
.chezmoitemplates/vim-lsp.vim                      → ~/.config/vim/plugin/lsp.vim (Unix), ~/vimfiles/plugin/lsp.vim (Windows)
    (servers from [agents.lsp.server.*]; also Copilot CLI lsp-config.json, opencode lsp)
dot_config/readline/inputrc                        → ~/.config/readline/ (via $INPUTRC; Unix + Git Bash)
dot_config/tmux/tmux.conf                          → ~/.config/tmux/ (Unix)
dot_config/git/personal.tmpl                       → ~/.config/git/personal (work machines only)
.chezmoitemplates/agents-config.toml       → ~/.config/agents/config.toml (via dot_config/agents/private_config.toml.tmpl; on Windows via AppData/Roaming/agents/private_config.toml.tmpl); edit with `dots agents/config`
dot_config/agents/instructions/{core,env}.md + age-encrypted personal/work → ~/.config/agents/instructions/ (layers; Windows: AppData/Roaming/agents/instructions/); edit with `dots agents/<layer>`
    (`agents apply`, run by a script after chezmoi apply, writes every harness's files from these; docs/OS.md → AI agent configuration)
private_dot_ssh/encrypted_private_config.age       → ~/.ssh/config (age-encrypted in the repo); edit with `dots ssh/config` (machine-only hosts: `dots ~/.ssh/config.d/<name>`, unmanaged)
```

Shell rc files (`dot_profile`, `dot_bashrc`, `dot_zshrc`, `dot_config/powershell/profile.ps1`) are thin orchestrators that source the snippets under `~/.config/shell/` (POSIX) or `~/.config/powershell/` (PS).

## Working in this repo

- No branch protection or rules is active in this repo: a single agent commits and pushes straight to `main` after every commit, as this is a configuration repo; parallel agents each use their own worktree (see the worktrees line below) and merge back to `main`.
- `home/.chezmoiignore` partitions per-OS paths with three symmetric blocks: `ne darwin`, `ne linux`, `ne windows`. Symlink-target-only sources (`dot_config/windows-terminal`, `dot_config/powershell/profile.ps1`) are ignored unconditionally.
- Every alias lives in the alias registry (`.chezmoidata/aliases.toml`, personal ones in `.aliases.private.toml.age`); never hand-write one in a shell file. `make lint` validates it, `make aliases-audit` (`chezmoi dots audit aliases`) lists live aliases it does not cover. See docs/SHELL.md → Alias registry.
- Editing goes through `chezmoi dots` (alias `dots`; standalone plugin from `giocaizzi/chezmoi-dots` installed via release): a new editable component is an entry in `.chezmoidata/edit.toml` (name, source or local folder, optional `check` as an argv list with `{file}`), never new edit code or a shortcut; `chezmoi dots lint registry` requires one for every `.chezmoidata` file and encrypted file. A new `[data]` answer or harness in `.chezmoi.toml.tmpl` needs no second list: the configuration matrix is built from that template. See docs/OS.md → Editing your dotfiles.
- Add a new cross-OS app: place the canonical file under `dot_config/<xdg-path>/`. On Linux it deploys directly. For macOS/Windows, add `symlink_*.tmpl` files under `Library/...` / `AppData/...` pointing at `{{ .chezmoi.sourceDir }}/dot_config/<xdg-path>/...`.
- Bootstrap scripts in `.chezmoiscripts/` use `run_once_` for setup and `run_onchange_` for package installs. `.sh.tmpl` for Unix, `.ps1.tmpl` for Windows; every script wraps its body in its own condition (OS, harness) and renders empty, so chezmoi skips it, where it does not apply. Scripts never appear in `.chezmoiignore`. `make lint` renders each script per OS, role and harness selection (`chezmoi dots lint templates`, needs chezmoi) and shellchecks the output.
- Per-machine behaviour comes from `chezmoi init` flags, `light` and `work` (plus the `terraform` answer, which gates its tools through `requires` in the package registry), read in templates as `get . "light"` / `get . "work"` (missing = false). Gate role-specific config on them; keep platform workarounds on `.chezmoi.os`. See docs/OS.md → Machine roles.
- Agent harness configuration is the standalone `agents` CLI's: chezmoi only renders its config (`.chezmoitemplates/agents-config.toml`: machine facts plus the user's choices) and deploys the instruction layers; `agents apply` (`run_after_zz-agents-apply.*`) writes the harness files. Edit the config with `dots agents/config`, never a harness file; harness selection (`data.harnesses`) is read only through `includeTemplate "harnesses" . | splitList " "`. Key reference: [agents docs](https://github.com/giocaizzi/agents/tree/main/docs); procedures: [Runbook: agent configuration](docs/OS.md#runbook-agent-configuration).
- Every tool a machine installs is a row of `.chezmoidata/packages.toml` (`[package.*]` via the install scripts, `[release.*]` as pinned, checksummed externals); never hand-list one in a script; `requires = "<answer>"` limits a row to machines whose init answer is true. Update with `chezmoi dots outdated|bump|upgrade`. See docs/OS.md → Packages.
- The repo's own agent skill lives in `.agents/skills/` (Codex, Copilot, VS Code, opencode, pi); `.claude/skills/` holds a symlink to it, since Claude Code reads only that. Neither is deployed: chezmoi skips source entries that start with a plain `.`.
- The repo is public: secret files go in encrypted (`chezmoi add --encrypt`), per-machine values in prompts (local config), never in plain source.
- Worktrees can live anywhere, `.claude/worktrees/` included: they sit outside `home/`, so chezmoi never reads their `.chezmoidata`. See docs/OS.md → Source layout.
- `agents usage` shows subscription usage per provider account (credentials read in memory only, never printed); `agents doctor` warns offline about model or config drift and runs after every apply on machines with a harness. See docs/OS.md → Models and subscriptions.
- Python tooling is uv only (no pyenv/pipx); a `pipx` shell function refuses.
- Externals are declared in `.chezmoiexternal.toml.tmpl` — chezmoi clones/fetches each on `apply`, auto-refreshes per `refreshPeriod`. Use them for **drop-in-place** upstreams (vim/zsh plugins, themes) where install reduces to "put this repo at this path". Use install scripts when there's a build step, PATH/Registry wiring, or package-manager registration. Use symlinks for canonical per-OS configs.

## Style system

All tools use [**Catppuccin**](https://github.com/catppuccin) (Mocha dark / Latte light, auto-switching where supported).

| Token | Mocha hex | Latte hex | Role |
|---|---|---|---|
| blue | `#89b4fa` | `#1e66f5` | headings, structural, OMP connectors |
| mauve | `#cba6f7` | `#8839ef` | main accent, bold, OMP path/session |
| red | `#f38ba8` | `#d20f39` | git, errors |
| peach | `#fab387` | `#fe640b` | italic |
| yellow | `#f9e2af` | `#df8e1d` | warnings, GCP, rate-limit ≥50% |
| green | `#a6e3a1` | `#40a02b` | battery, math blocks |
| text | `#cdd6f4` | `#4c4f69` | icons, labels |
| overlay0 | `#6c7086` | `#9ca0b0` | dim/grey |

**Font:** `JetBrainsMono Nerd Font` in Ghostty, VS Code terminal and Windows Terminal (installed by brew cask / oh-my-posh (user scope) / Linux external). These configs are symlink targets, so the name is kept identical by hand.

**Sources of truth (edit these, then `chezmoi apply`):**
- `dot_config/oh-my-posh/theme.omp.json` — oh-my-posh prompt. Same convention as the Claude status line: every part `[icon value]` (dim overlay0 brackets, one Nerd Font icon, value coloured by role), one space apart, rows led by `┌`/`├`/`└`. Segment settings go under `properties` (oh-my-posh ignores an `options` key).
- `.chezmoitemplates/agents-config.toml` `[harness.claude.settings]` — Claude Code statusline (drawn by `hud`'s `hooks/statusline.mjs`, harness repo) and theme (`"theme": "dark"`), written by `agents apply`
- `dot_config/ghostty/config` — `theme = dark:Catppuccin Mocha,light:Catppuccin Latte`
- `dot_config/windows-terminal/settings.json` — schemes + `"colorScheme": "Catppuccin Mocha"`
- `dot_config/tmux/tmux.conf` — status line and borders
- `dot_config/vim/vimrc` — `colorscheme catppuccin_mocha` (plugin fetched via `.chezmoiexternal.toml.tmpl`)
- Obsidian vault `.obsidian/plugins/obsidian-style-settings/data.json` — Minimal theme colours
- Obsidian vault `.obsidian/snippets/` — `html-example.css`, `math.css`

When adding a new tool: use the Mocha/Latte hex values above. Never introduce new accent colours.

## Docs

- [README.md](./README.md) — what this is, quickstart, everyday commands.
- [docs/SETUP.md](./docs/SETUP.md) — set up a new machine, secrets.
- [docs/OS.md](./docs/OS.md) — source → destination matrix per OS, machine roles, Bitwarden, encryption, managed-Windows workarounds, hardcoded-by-design values, the AI agent configuration and its [runbook](./docs/OS.md#runbook-agent-configuration).
- [docs/SHELL.md](./docs/SHELL.md) — shell architecture, snippets, alias registry, tools, secrets.
- Commits follow Conventional Commits.

Keep docs short, factual, and updated with structural changes.
