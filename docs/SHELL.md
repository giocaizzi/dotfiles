# Shell Configuration

Cross-shell setup for bash, zsh, and PowerShell. Aliases and FZF config live in shared snippet files; rc files are thin orchestrators.

**Rule of thumb:** environment variables go in `~/.profile` (loaded once, inherited by child processes); aliases and functions go in the snippets, which `.bashrc`/`.zshrc` load in *every* interactive shell (they aren't inherited).

## Architecture

```
.chezmoidata/aliases.toml + .aliases.private.toml.age   ← alias registry (public + encrypted)
        │  selected per shell, OS and machine role by .chezmoitemplates/aliases
        ├─ rendered into ──→  ~/.config/shell/aliases.sh          (bash + zsh + Git Bash)
        ├─ rendered into ──→  ~/.config/shell/omz-aliases.zsh     (Oh My Zsh alias sets, zsh)
        └─ rendered into ──→  ~/.config/powershell/aliases.ps1    (PowerShell)

~/.profile        sources →  ~/.config/shell/secrets.sh  (environment only, inherited)
~/.bashrc         sources →  ~/.profile (once) + ~/.config/shell/{aliases.sh, fzf.sh} + bash-specific
~/.zshrc          sources →  omz-aliases.zsh + Oh My Zsh + ~/.config/shell/{aliases.sh, fzf.sh} + zsh-specific
profile.ps1       sources →  ~/.config/powershell/{secrets.ps1, aliases.ps1, coreutils.ps1, fzf.ps1} + PSReadLine
```

## Files

History lives in `$XDG_STATE_HOME/{bash,zsh}/history`; the zsh completion dump in `$XDG_CACHE_HOME/zsh/`.

| File                                     | Role                                                                    | OS scope            |
| ---------------------------------------- | ----------------------------------------------------------------------- | ------------------- |
| `dot_profile`                            | POSIX environment only (inherited): `XDG_*`, PATH, `EDITOR`, `PAGER`/`LESS`/`MANPAGER`, `INPUTRC`, `POSH_THEME`, `JAVA_HOME`; sources `secrets.sh` | Unix + Git Bash |
| `dot_bash_profile`                       | Login-shell wrapper → sources `.bashrc` (which sources `.profile` once) | Unix                |
| `dot_bashrc`                             | Bash: loads `.profile` once, then **returns for non-interactive shells** (`ssh host cmd`, scp, rsync); snippets, history (timestamped), completion, FZF bindings, oh-my-posh | Unix + Git Bash |
| `dot_zshrc` / `dot_zprofile`             | Zsh: `.zprofile` loads `.profile`; `.zshrc` sets the Oh My Zsh alias sets, loads Oh My Zsh, then the snippets (so registry aliases win), history, FZF bindings, oh-my-posh | macOS only |
| `dot_config/powershell/profile.ps1`      | PowerShell orchestrator → PATH, PSReadLine, sources snippets            | Windows             |
| `dot_config/shell/aliases.sh.tmpl`       | POSIX aliases and refused commands (rendered from the registry), `ls` colours, `python` fallback, the edit commands (full machines), `rp5-secrets` (personal Mac) | Unix + Git Bash |
| `dot_config/shell/omz-aliases.zsh.tmpl`  | `zstyle ':omz:…' aliases` per Oh My Zsh lib/plugin (from `[aliasExternal.zsh]`) | macOS (zsh) |
| `dot_config/shell/fzf.sh`                | `FZF_*` env vars + `h()` history function                               | Unix + Git Bash     |
| `dot_config/powershell/aliases.ps1.tmpl` | PowerShell functions (rendered from the registry), the edit commands (full machines) | Windows             |
| `dot_config/powershell/coreutils.ps1`    | uutils first on PATH; removes the cmdlet aliases it shadows (`ls`, `cat`, `rm`, ...) — use `Get-ChildItem` etc. for objects | Windows |
| `dot_config/powershell/fzf.ps1`          | PSFzf + `FZF_*` env vars + `h`/`history` fuzzy history                  | Windows             |
| `.chezmoidata/aliases.toml`              | Alias registry: every alias, refused command and framework alias set    | all                 |
| `.aliases.private.toml.age`              | Personal aliases, same schema, age-encrypted; read on non-light machines with the age key | all |
| `.chezmoitemplates/aliases`              | Selector: merges both registries, filters by shell, OS and machine role | all                 |
| `dot_config/readline/inputrc`            | Readline: ↑/↓ prefix history search, case-insensitive coloured completion | Unix + Git Bash   |
| `dot_config/tmux/tmux.conf`              | tmux (prefix `C-b`; `h/j/k/l` pane moves; Catppuccin status)            | Unix                |

## Alias registry

Every alias on every machine comes from `.chezmoidata/aliases.toml` (public) and
`.aliases.private.toml.age` (personal, encrypted). One entry per alias:

```toml
[alias.gs]
cmd = "git status"                 # same command in every shell

[alias.ll]
posix = "ls -lah"                  # bash, zsh, Git Bash
powershell = "Get-ChildItem -Force"

[alias.free]
posix = "top -l 1 -s 0 | grep PhysMem"
shells = ["posix"]                 # one shell only: needs `why`
os = ["darwin"]                    # darwin, linux, windows (default: all)
machines = ["personal"]            # personal, work, light (default: all)
why = "macOS has no free; Linux ships it"

[aliasBlock.pipx]                  # refused command
message = "pipx is not used here, use uv: ..."

[aliasExternal.zsh]                # Oh My Zsh alias sets, off unless true
"plugins:docker" = true
```

- **Machine roles** come from the `light` and `work` init flags: `personal` is neither.
- **PowerShell:** each alias is a function (`function gs { git status @args }`); a
  same-named built-in alias (`gc`, `gp`, `gl`) is removed first, so the registry wins.
- **Frameworks:** every Oh My Zsh lib and plugin alias set is off unless declared
  `true` in `[aliasExternal.zsh]`; registry aliases load after Oh My Zsh and win.
- **Functions** that are more than a command stay in the shell snippets, each with a
  PowerShell twin where it applies: the edit commands `agents-edit`, `aliases-edit`,
  `ssh-edit` (full machines, over the shared `_dots_edit` / `_dots-edit` and
  `_dots_pick` / `_dots-pick` helpers, [OS.md](./OS.md#editing-your-dotfiles)),
  `rp5-secrets`, `rp5-sync-infra` (personal Mac, POSIX only), fzf `h`.
- **Checks:** `make lint` validates the registry (every alias names its shells, a
  one-shell alias says why, known OS and role values) and lints the rendered files
  per OS and role; CI also parses `aliases.ps1` with `pwsh`. `make aliases-audit`
  lists aliases live in this machine's zsh and bash that the registry does not
  account for.

To add or change an alias: `aliases-edit public` (edits the registry, runs
`make lint-aliases`, diffs and offers to apply the alias files), then open a new shell.

### Personal aliases

Same schema, never in plain text in the repo. Edit with `aliases-edit private`: it
decrypts the file into a private temp file, re-encrypts it only on change, always
removes the temp file, then lints, diffs and offers to apply like the public registry
([OS.md](./OS.md#editing-your-dotfiles)).

Light machines and machines without the age key (`~/.config/chezmoi/key.txt`) skip
the file.

## Aliases (current)

| Alias                 | bash / zsh / Git Bash                    | PowerShell |
| --------------------- | ---------------------------------------- | ---------- |
| `..` `...` `....`     | `cd ..`, `cd ../..`, `cd ../../..`       | same |
| `ll` `la`             | `ls -lah`, `ls -A`                       | `Get-ChildItem -Force` |
| `l`                   | `ls -CF`                                 | — (`Get-ChildItem`) |
| `mkdir` `df` `du`     | `mkdir -p`, `df -h`, `du -h`             | — (already parents / readable) |
| `grep`                | `grep --color=auto`                      | `rg` |
| `free` (macOS)        | `top -l 1 -s 0 \| grep PhysMem`          | — |
| `gs` `ga` `gc` `gd` `gp` `gl` `glog` | git shortcuts            | same |
| `hist` / `h`          | fzf history search                       | same |
| `cz` `cze` `czd` `cza` | `chezmoi`, `chezmoi edit --apply`, `chezmoi diff`, `chezmoi apply` | same |
| `agents-edit` `aliases-edit` `ssh-edit` (full machines) | edit an agent instruction layer, the alias registry (public or private), `~/.ssh/config`; check, diff, apply ([OS.md](./OS.md#editing-your-dotfiles)) | same |
| `openclaw` (personal Mac) | rp5-homeserver openclaw script       | — |
| `pipx`                | refuses; use `uv tool install` / `uvx`   | same |
| `python`              | `python3`, only where no `python` exists | — |
| `ls`                  | colour flag for GNU or BSD `ls`          | — (uutils `ls`) |

## Tools

| Tool          | Bash | Zsh | PowerShell | Init location                                              |
| ------------- | :--: | :-: | :--------: | ---------------------------------------------------------- |
| `fzf`         |  ✓   |  ✓  |     ✓      | env in `fzf.sh`/`fzf.ps1`; bindings in bashrc/zshrc/PSFzf  |
| `oh-my-posh`  |  ✓   |  ✓  |     ✓      | Uses `$POSH_THEME` → `~/.config/oh-my-posh/theme.omp.json`|
| `uv`          |  ✓   |  ✓  |     ✓      | Python + CLI tools in `~/.local/bin` (on PATH in `.profile` / `profile.ps1`) |
| `Oh My Zsh`   |  —   |  ✓  |     —      | `.zshrc` (plugins: `docker gh`; alias sets per `[aliasExternal.zsh]`; no `git` plugin, the registry covers git) |
| readline      |  ✓   |  —  |     —      | `~/.config/readline/inputrc` via `$INPUTRC`: ↑/↓ prefix history search, case-insensitive completion |
| `less`        |  ✓   |  ✓  |     —      | `PAGER`/`LESS`/`MANPAGER` in `.profile` (colour, quit if one screen, smart-case search) |
| `tmux`        |  ✓   |  ✓  |     —      | `~/.config/tmux/tmux.conf` (native XDG path); prefix `C-b`, `h/j/k/l` pane moves |
| `PSReadLine`  |  —   |  —  |     ✓      | `profile.ps1` — history search, predictions                |
| `PSFzf`       |  —   |  —  |     ✓      | `fzf.ps1` — Ctrl-R, Ctrl-T, `h`, and `history` via fzf     |

## Secrets

| OS      | File                               | Permissions        |
| ------- | ---------------------------------- | ------------------ |
| Unix    | `~/.config/shell/secrets.sh`       | `600`              |
| Windows | `~/.config/powershell/secrets.ps1` | current user only  |

- **Never committed.** Two modes per machine:
  - **Generated** from Bitwarden Secrets Manager when the machine has a `bws` project ID and keyring token (`dot_config/shell/private_secrets.sh.tmpl`, `dot_config/powershell/secrets.ps1.tmpl`): edit secrets in Bitwarden, then `chezmoi apply`. See [OS.md → Bitwarden](./OS.md#bitwarden).
  - **By hand** otherwise (light machines, blank project ID): ignored by chezmoi, edit the file on that machine.
- **Created** by `.chezmoiscripts/run_once_init-untracked-env.{sh,ps1}`: writes a stub if missing, then enforces owner-only permissions (on Windows the `~\.config\powershell` folder too, so regenerated files inherit it).
- **Your global git ignore has `*secret*`**, so new secrets *templates* need `git add -f`; they contain no values.
- **Sourced** at the top of `.profile` / `profile.ps1`, so values are exported to every process started from the shell.
- **Why env vars**: project `.mcp.json` files expand `${VAR}` from the environment Claude Code is launched with, so MCP tokens must be exported here. Tools with their own credential store (e.g. `~/.config/gh/`, `~/.config/greenhouse/token`) keep it there.
- **Secrets only**: non-secret env (`JAVA_HOME`, `PATH`, …) belongs in `dot_profile`. One exception: on personal machines with an OTel endpoint the generated file also exports `GUARD_OTEL_EXPORTER_OTLP_{ENDPOINT,HEADERS}` and `GUARD_OTEL_RESOURCE_ATTRIBUTES` (the headers copied from `OTEL_EXPORTER_OTLP_HEADERS`) for the harness `guard` decision log, since Claude Code and Copilot CLI strip `OTEL_*` from hook environments.
- **Secret files** (e.g. `~/.ssh/config`) are different: they're committed age-encrypted, see [OS.md](./OS.md#encrypted-files-age).
- **Claude Code** is denied Read/Edit on both files (`.chezmoitemplates/claude-settings.json`).

```sh
vim ~/.config/shell/secrets.sh               # Unix
notepad $HOME\.config\powershell\secrets.ps1 # Windows
```
