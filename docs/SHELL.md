# Shell Configuration

Cross-shell setup for bash, zsh, and PowerShell. Aliases and FZF config live in shared snippet files; rc files are thin orchestrators.

## Architecture

```
.chezmoidata/shortcuts.toml          ← single source of truth (git aliases)
        │
        ├─ rendered into ──→  ~/.config/shell/aliases.sh          (bash + zsh)
        └─ rendered into ──→  ~/.config/powershell/aliases.ps1    (PowerShell)

~/.profile        sources →  ~/.config/shell/{aliases.sh, fzf.sh}
~/.bashrc         sources →  ~/.profile + bash-specific (history, completion, FZF bindings)
~/.zshrc          sources →  ~/.profile + Oh My Zsh + zsh-specific
profile.ps1       sources →  ~/.config/powershell/{aliases.ps1, fzf.ps1} + PSReadLine
```

## Files

| File                                     | Role                                                                    | OS scope            |
| ---------------------------------------- | ----------------------------------------------------------------------- | ------------------- |
| `dot_profile`                            | POSIX core: PATH, env vars (`EDITOR`, `POSH_THEME`), sources snippets   | Unix + Git Bash     |
| `dot_bash_profile`                       | Login-shell wrapper → sources `.profile` + `.bashrc`                    | Unix                |
| `dot_bashrc`                             | Bash: history, completion, FZF bindings, oh-my-posh init                | Unix                |
| `dot_zshrc` / `dot_zprofile`             | Zsh: Oh My Zsh, history, plugins, FZF bindings, oh-my-posh init         | macOS only          |
| `dot_config/powershell/profile.ps1`      | PowerShell orchestrator → PATH, PSReadLine, sources snippets            | Windows             |
| `dot_config/shell/aliases.sh.tmpl`       | POSIX aliases (rendered from `shortcuts.toml`)                          | Unix + Git Bash     |
| `dot_config/shell/fzf.sh`                | `FZF_*` env vars + `h()` history function                               | Unix + Git Bash     |
| `dot_config/powershell/aliases.ps1.tmpl` | PowerShell functions (rendered from `shortcuts.toml`)                   | Windows             |
| `dot_config/powershell/fzf.ps1`          | PSFzf + `FZF_*` env vars                                                | Windows             |
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
| `oh-my-posh`  |  ✓   |  ✓  |     ✓      | Uses `$POSH_THEME` → `~/.mytheme.omp.json`                 |
| `pyenv`       |  ✓   |  ✓  |     ✓      | PATH in `.profile` / `profile.ps1`; `pyenv init` in rc     |
| `nvm`         |  ✓   |  —  |     —      | `.bashrc` only                                             |
| `Oh My Zsh`   |  —   |  ✓  |     —      | `.zshrc` (plugins: `git docker gh`)                        |
| `PSReadLine`  |  —   |  —  |     ✓      | `profile.ps1` — history search, predictions                |
| `PSFzf`       |  —   |  —  |     ✓      | `fzf.ps1` — Ctrl-R, Ctrl-T parity with bash/zsh            |

## Secrets

`~/.secrets` (Unix) and `~/.secrets.ps1` (Windows) are created on first apply with restrictive permissions and **never tracked**. Sourced at the top of `.profile` / `profile.ps1`. Edit on each machine separately:

```sh
vim ~/.secrets             # Unix
notepad $HOME\.secrets.ps1 # Windows
```
