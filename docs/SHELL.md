# Shell Configuration

Cross-shell setup for bash, zsh, and PowerShell. Aliases and FZF config live in shared snippet files; rc files are thin orchestrators.

## Architecture

```
.chezmoidata/shortcuts.toml          ← single source of truth (git aliases)
        │
        ├─ rendered into ──→  ~/.config/shell/aliases.sh          (bash + zsh)
        └─ rendered into ──→  ~/.config/powershell/aliases.ps1    (PowerShell)

~/.profile        sources →  ~/.config/shell/secrets.sh  (environment only, inherited)
~/.bashrc         sources →  ~/.profile (once) + ~/.config/shell/{aliases.sh, fzf.sh} + bash-specific
~/.zshrc          sources →  Oh My Zsh + ~/.config/shell/{aliases.sh, fzf.sh} + zsh-specific
profile.ps1       sources →  ~/.config/powershell/{secrets.ps1, aliases.ps1, fzf.ps1} + PSReadLine
```

## Files

History lives in `$XDG_STATE_HOME/{bash,zsh}/history`; the zsh completion dump in `$XDG_CACHE_HOME/zsh/`.

| File                                     | Role                                                                    | OS scope            |
| ---------------------------------------- | ----------------------------------------------------------------------- | ------------------- |
| `dot_profile`                            | POSIX core: `XDG_*`, PATH, env vars (`EDITOR`, `POSH_THEME`, `JAVA_HOME`), sources snippets | Unix + Git Bash |
| `dot_bash_profile`                       | Login-shell wrapper → sources `.bashrc` (which sources `.profile` once) | Unix                |
| `dot_bashrc`                             | Bash: history, completion, FZF bindings, oh-my-posh init                | Unix                |
| `dot_zshrc` / `dot_zprofile`             | Zsh: Oh My Zsh, history, plugins, FZF bindings, oh-my-posh init         | macOS only          |
| `dot_config/powershell/profile.ps1`      | PowerShell orchestrator → PATH, PSReadLine, sources snippets            | Windows             |
| `dot_config/shell/aliases.sh.tmpl`       | POSIX aliases (rendered from `shortcuts.toml`)                          | Unix + Git Bash     |
| `dot_config/shell/fzf.sh`                | `FZF_*` env vars + `h()` history function                               | Unix + Git Bash     |
| `dot_config/powershell/aliases.ps1.tmpl` | PowerShell functions (rendered from `shortcuts.toml`)                   | Windows             |
| `dot_config/powershell/fzf.ps1`          | PSFzf + `FZF_*` env vars + `h`/`history` fuzzy history                  | Windows             |
| `.chezmoidata/shortcuts.toml`            | Single source for cross-shell shortcuts                                 | all                 |

## Adding a new shortcut

Edit `.chezmoidata/shortcuts.toml`:

```toml
[shortcuts.git]
gco = "git checkout"
```

Then `chezmoi apply`. Both `~/.config/shell/aliases.sh` and `~/.config/powershell/aliases.ps1` get the new entry on next shell launch.

## Aliases (current)

POSIX shell (`aliases.sh`):

| Alias                 | Command                                  |
| --------------------- | ---------------------------------------- |
| `..` `...` `....`     | `cd ..`, `cd ../..`, `cd ../../..`       |
| `ll` `la` `l`         | `ls -lah`, `ls -A`, `ls -CF`             |
| `mkdir`               | `mkdir -p`                               |
| `df` `du`             | `df -h`, `du -h`                         |
| `grep` `fgrep` `egrep`| `--color=auto`                           |
| `python`              | `python3` (if available)                 |
| `free` (macOS only)   | `top -l 1 -s 0 \| grep PhysMem`          |
| `gs` `ga` `gc` `gd` `gp` `gl` `glog` | git shortcuts (from `shortcuts.toml`) |
| `hist` / `h`          | fzf history search                       |

PowerShell (`aliases.ps1`) mirrors the same shortcuts as functions (e.g. `function gs { git status @args }`).

## Tools

| Tool          | Bash | Zsh | PowerShell | Init location                                              |
| ------------- | :--: | :-: | :--------: | ---------------------------------------------------------- |
| `fzf`         |  ✓   |  ✓  |     ✓      | env in `fzf.sh`/`fzf.ps1`; bindings in bashrc/zshrc/PSFzf  |
| `oh-my-posh`  |  ✓   |  ✓  |     ✓      | Uses `$POSH_THEME` → `~/.config/oh-my-posh/theme.omp.json`|
| `uv`          |  ✓   |  ✓  |     ✓      | Python + CLI tools in `~/.local/bin` (on PATH in `.profile` / `profile.ps1`) |
| `nvm`         |  ✓   |  —  |     —      | `.bashrc` only                                             |
| `Oh My Zsh`   |  —   |  ✓  |     —      | `.zshrc` (plugins: `git docker gh`)                        |
| `PSReadLine`  |  —   |  —  |     ✓      | `profile.ps1` — history search, predictions                |
| `PSFzf`       |  —   |  —  |     ✓      | `fzf.ps1` — Ctrl-R, Ctrl-T, `h`, and `history` via fzf     |

## Secrets

| OS      | File                               | Permissions        |
| ------- | ---------------------------------- | ------------------ |
| Unix    | `~/.config/shell/secrets.sh`       | `600`              |
| Windows | `~/.config/powershell/secrets.ps1` | current user only  |

- **Never tracked**: listed in `.chezmoiignore`; edit on each machine separately.
- **Created / migrated** by `.chezmoiscripts/run_once_init-untracked-env.{sh,ps1}`: moves a legacy `~/.secrets` / `~\.secrets.ps1` into place if present, otherwise writes a stub, then enforces permissions.
- **Sourced** at the top of `.profile` / `profile.ps1`, so values are exported to every process started from the shell.
- **Why env vars**: project `.mcp.json` files expand `${VAR}` from the environment Claude Code is launched with, so MCP tokens must be exported here. Tools with their own credential store (e.g. `~/.config/gh/`, `~/.config/greenhouse/token`) keep it there.
- **Secrets only**: non-secret env (`JAVA_HOME`, `PATH`, …) belongs in `dot_profile`.
- **Claude Code** is denied Read/Edit on both files (`.chezmoitemplates/claude-settings.json`).

```sh
vim ~/.config/shell/secrets.sh               # Unix
notepad $HOME\.config\powershell\secrets.ps1 # Windows
```
