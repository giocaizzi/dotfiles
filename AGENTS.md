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
    → OneDrive/Documenti/{PowerShell,WindowsPowerShell}/ (Windows)

dot_config/powershell/{aliases.ps1.tmpl, fzf.ps1}  → ~/.config/powershell/ (Windows)
dot_config/shell/{aliases.sh.tmpl, fzf.sh}         → ~/.config/shell/     (Unix + Git Bash)
.chezmoidata/shortcuts.toml                        → renders the two aliases files above
dot_config/git/config.tmpl                         → ~/.config/git/config (all OSes)
dot_mytheme.omp.json                               → ~/.mytheme.omp.json  (bash/zsh/pwsh)
```

Shell rc files (`dot_profile`, `dot_bashrc`, `dot_zshrc`, `dot_config/powershell/profile.ps1`) are thin orchestrators that source the snippets under `~/.config/shell/` (POSIX) or `~/.config/powershell/` (PS).

## Working in this repo

- `.chezmoiignore` partitions per-OS paths with three symmetric blocks: `ne darwin`, `ne linux`, `ne windows`. Symlink-target-only sources (`dot_config/windows-terminal`, `dot_config/powershell/profile.ps1`) are ignored unconditionally.
- Add a new git alias once in `.chezmoidata/shortcuts.toml` → it renders into both shells on `chezmoi apply`.
- Add a new cross-OS app: place the canonical file under `dot_config/<xdg-path>/`. On Linux it deploys directly. For macOS/Windows, add `symlink_*.tmpl` files under `Library/...` / `AppData/...` pointing at `{{ .chezmoi.sourceDir }}/dot_config/<xdg-path>/...`.
- Bootstrap scripts in `.chezmoiscripts/` use `run_once_` for setup and `run_onchange_` for package installs. `.sh.tmpl` for Unix, `.ps1` for Windows; ignored on the other platform via `.chezmoiignore`.

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

**Sources of truth (edit these, then `chezmoi apply`):**
- `dot_mytheme.omp.json` — oh-my-posh prompt
- `dot_claude/statusline-command.sh` — Claude Code statusline
- `dot_claude/settings.json.tmpl` — Claude Code theme (`"theme": "dark"`)
- `dot_config/ghostty/config` — `theme = dark:Catppuccin Mocha,light:Catppuccin Latte`
- `dot_config/windows-terminal/settings.json` — schemes + `"colorScheme": "Catppuccin Mocha"`
- Obsidian vault `.obsidian/plugins/obsidian-style-settings/data.json` — Minimal theme colours
- Obsidian vault `.obsidian/snippets/` — `html-example.css`, `math.css`

When adding a new tool: use the Mocha/Latte hex values above. Never introduce new accent colours.

## Docs

- [docs/OS.md](./docs/OS.md) — source → destination matrix per OS.
- [docs/SHELL.md](./docs/SHELL.md) — shell architecture, snippets, shortcuts.toml workflow.
- [docs/GIT.md](./docs/GIT.md) — commit conventions.
- [README.md](./README.md) — everyday chezmoi commands.

Keep docs short, factual, and updated with structural changes.
