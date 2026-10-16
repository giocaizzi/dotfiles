This is the user's [chezmoi](https://chezmoi.io/) dotfiles configuration, managing macOS, Linux and Windows from a single source.

## Principles

- Always consider OS-specific behavior. The same config is rendered to different destinations per OS.
- Prefer symlinks over copies: every per-OS path should point back to a single source under `dot_config/` or `.chezmoidata/`.
- Check current [chezmoi docs](https://chezmoi.io/docs/) for template functions and source-state attributes.

## Architecture (single sources → per-OS destinations)

Canonical configs live under `dot_config/<xdg-path>/`. On Linux they deploy directly; on macOS/Windows they are symlink targets referenced via `{{ .chezmoi.sourceDir }}/...` from `Library/` or `AppData/` paths.

```bash
dot_config/Code/User/{settings,keybindings,mcp}.json
    → ~/.config/Code/User/         (Linux, direct)
    → Library/.../Code/User/        (macOS, symlinks)
    → AppData/Roaming/Code/User/    (Windows, symlinks)
    (Code - Insiders symlinks point at the same files)

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
dot_config/readline/inputrc                        → ~/.config/readline/ (via $INPUTRC; Unix + Git Bash)
dot_config/tmux/tmux.conf                          → ~/.config/tmux/ (Unix)
dot_config/git/personal.tmpl                       → ~/.config/git/personal (work machines only)
.chezmoitemplates/{claude,copilot}-settings.json   → merged into ~/.claude, ~/.copilot settings.json by modify_ templates
private_dot_ssh/encrypted_private_config.age       → ~/.ssh/config (age-encrypted in the repo)
```

Shell rc files (`dot_profile`, `dot_bashrc`, `dot_zshrc`, `dot_config/powershell/profile.ps1`) are thin orchestrators that source the snippets under `~/.config/shell/` (POSIX) or `~/.config/powershell/` (PS).

## Working in this repo

- `.chezmoiignore` partitions per-OS paths with three symmetric blocks: `ne darwin`, `ne linux`, `ne windows`. Symlink-target-only sources (`dot_config/windows-terminal`, `dot_config/powershell/profile.ps1`) are ignored unconditionally.
- Every alias lives in the alias registry (`.chezmoidata/aliases.toml`, personal ones in `.aliases.private.toml.age`); never hand-write one in a shell file. `make lint` validates it, `make aliases-audit` lists live aliases it does not cover. See docs/SHELL.md → Alias registry.
- Add a new cross-OS app: place the canonical file under `dot_config/<xdg-path>/`. On Linux it deploys directly. For macOS/Windows, add `symlink_*.tmpl` files under `Library/...` / `AppData/...` pointing at `{{ .chezmoi.sourceDir }}/dot_config/<xdg-path>/...`.
- Bootstrap scripts in `.chezmoiscripts/` use `run_once_` for setup and `run_onchange_` for package installs. `.sh.tmpl` for Unix, `.ps1` for Windows; ignored on the other platform via `.chezmoiignore`. `make lint` renders each `.sh.tmpl` script per OS, role and harness selection (`lint-tmpl`, needs chezmoi) and shellchecks the output.
- Per-machine behaviour comes from two `chezmoi init` flags, `light` and `work`, read in templates as `get . "light"` / `get . "work"` (missing = false). Gate role-specific config on them; keep platform workarounds on `.chezmoi.os`. See docs/OS.md → Machine roles.
- AI harness configs are gated on `data.harnesses` (chosen at `chezmoi init`). Read it only through `includeTemplate "harnesses" . | splitList " "`, which defaults a missing key to all six. A new harness file needs its ignore entry in `.chezmoiignore`. See docs/OS.md → AI harnesses.
- Files an app also writes (Claude Code, Copilot settings) are `modify_` templates that merge managed keys into the live file; never replace them wholesale.
- The repo is public: secret files go in encrypted (`chezmoi add --encrypt`), per-machine values in prompts (local config), never in plain source.
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
- `.chezmoitemplates/claude-settings.json` `statusLine` — Claude Code statusline, drawn by `hud`'s `hooks/statusline.mjs` (harness repo)
- `.chezmoitemplates/claude-settings.json` — Claude Code theme (merged into `~/.claude/settings.json` by `dot_claude/modify_settings.json`, which keeps keys Claude Code writes itself) (`"theme": "dark"`)
- `dot_config/ghostty/config` — `theme = dark:Catppuccin Mocha,light:Catppuccin Latte`
- `dot_config/windows-terminal/settings.json` — schemes + `"colorScheme": "Catppuccin Mocha"`
- `dot_config/tmux/tmux.conf` — status line and borders
- `dot_config/vim/vimrc` — `colorscheme catppuccin_mocha` (plugin fetched via `.chezmoiexternal.toml.tmpl`)
- Obsidian vault `.obsidian/plugins/obsidian-style-settings/data.json` — Minimal theme colours
- Obsidian vault `.obsidian/snippets/` — `html-example.css`, `math.css`

When adding a new tool: use the Mocha/Latte hex values above. Never introduce new accent colours.

## Docs

- [README.md](./README.md) — new-machine setup and everyday chezmoi commands.
- [docs/OS.md](./docs/OS.md) — source → destination matrix per OS, machine roles, Bitwarden, encryption, managed-Windows workarounds, hardcoded-by-design values.
- [docs/SHELL.md](./docs/SHELL.md) — shell architecture, snippets, alias registry, tools, secrets.
- [docs/CI.md](./docs/CI.md) — `make check` targets and the CI workflow; run `make check` before pushing.
- Commits follow Conventional Commits.

Keep docs short, factual, and updated with structural changes.
