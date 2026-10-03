# OS-Specific Configuration

How each managed file maps to its destination per OS. Canonical sources live under `dot_config/` in the repo; OS-specific destinations symlink to them via `{{ .chezmoi.sourceDir }}/...`.

## Destination matrix

| Canonical source (in repo)                          | macOS                                                       | Linux                                          | Windows                                                                            |
| ---------------------------------------------------- | ----------------------------------------------------------- | ---------------------------------------------- | ---------------------------------------------------------------------------------- |
| `dot_config/Code/User/{settings,keybindings,mcp}.json` | `~/Library/Application Support/Code/User/` (symlinks)      | `~/.config/Code/User/` (direct)                | `%APPDATA%\Code\User\` (symlinks)                                                |
| same + Insiders                                      | `~/Library/.../Code - Insiders/User/` (symlinks → stable)   | `~/.config/Code - Insiders/User/` (symlinks → stable) | `%APPDATA%\Code - Insiders\User\` (symlinks → stable)                       |
| `dot_config/agents/AGENTS.md`                        | `~/.config/agents/AGENTS.md`                                | `~/.config/agents/AGENTS.md`                   | `~/.config/agents/AGENTS.md`                                                       |
| `dot_config/ghostty/config`                          | `~/Library/.../com.mitchellh.ghostty/` (symlink)            | `~/.config/ghostty/config` (direct)            | — (no Windows build)                                                               |
| `dot_config/windows-terminal/settings.json` *        | —                                                           | —                                              | `%LOCALAPPDATA%\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\` (symlink) |
| `dot_config/powershell/profile.ps1` *                | —                                                           | —                                              | `<Documents>/{PowerShell,WindowsPowerShell}/profile.ps1` (symlinks, created by `run_after_link-pwsh-profile.ps1`) |
| `dot_copilot/modify_private_settings.json` (merge)   | `~/.copilot/settings.json`                                  | `~/.copilot/settings.json`                     | `~/.copilot/settings.json`                                                         |
| `dot_claude/symlink_CLAUDE.md.tmpl`                  | `~/.claude/CLAUDE.md` → `~/.config/agents/AGENTS.md`        | same                                           | same                                                                               |
| `dot_copilot/symlink_copilot-instructions.md.tmpl`   | `~/.copilot/copilot-instructions.md` → `~/.config/agents/AGENTS.md` | same                                    | same                                                                               |
| `.chezmoidata/shortcuts.toml`                        | `~/.config/shell/aliases.sh`                                | `~/.config/shell/aliases.sh`                   | `~/.config/powershell/aliases.ps1` (+ `~/.config/shell/aliases.sh` for Git Bash)   |
| `dot_config/git/config.tmpl`                         | `~/.config/git/config`                                      | `~/.config/git/config`                         | `~/.config/git/config`                                                             |
| `dot_config/oh-my-posh/theme.omp.json`               | `~/.config/oh-my-posh/theme.omp.json`                       | same                                           | same                                                                               |
| `dot_config/vim/vimrc`                               | `~/.config/vim/vimrc`                                       | same                                           | `~/vimfiles/vimrc` (symlink)                                                       |
| `dot_config/git/{ignore,commit-template}`            | `~/.config/git/`                                            | same                                           | same                                                                               |
| `dot_profile`, `dot_bashrc`, `dot_bash_profile`      | `~/.profile`, `~/.bashrc`, `~/.bash_profile`                | same                                           | same (used by Git Bash / WSL)                                                      |
| `dot_zshrc`, `dot_zprofile`                          | `~/.zshrc`, `~/.zprofile`                                   | — (ignored)                                    | — (ignored)                                                                        |
| `dot_claude/modify_settings.json` (merge)            | `~/.claude/settings.json`                                   | same                                           | same                                                                               |
| `dot_claude/statusline-command.sh`                   | `~/.claude/statusline-command.sh`                           | same                                           | same                                                                               |
| `dot_config/git/work.tmpl` (work machines only)      | `~/.config/git/work`                                        | same                                           | same                                                                               |
| `dot_config/readline/inputrc`                        | `~/.config/readline/inputrc` (via `$INPUTRC`)               | same                                           | same (Git Bash)                                                                    |
| `dot_config/tmux/tmux.conf`                          | `~/.config/tmux/tmux.conf`                                  | same                                           | — (no tmux)                                                                        |
| `private_dot_ssh/encrypted_private_config.age`       | `~/.ssh/config` (decrypted, owner-only)                     | same                                           | same                                                                               |

\* Symlink-target-only sources — never deployed as regular files anywhere (listed in `.chezmoiignore` unconditionally); only referenced via `{{ .chezmoi.sourceDir }}/...` from the Windows-specific symlinks.

## Bootstrap scripts (`.chezmoiscripts/`)

| Script                                       | macOS | Linux | Windows |
| -------------------------------------------- | :---: | :---: | :-----: |
| `run_once_init-untracked-env.sh` → `~/.config/shell/secrets.sh`             |  ✓    |  ✓    |   —     |
| `run_once_init-untracked-env.ps1` → `~/.config/powershell/secrets.ps1`      |  —    |  —    |   ✓     |
| `run_once_after_create-xdg-dirs.sh` → creates `$XDG_STATE_HOME/{zsh,bash,vim}`, `$XDG_CACHE_HOME/zsh` |  ✓    |  ✓    |   —     |
| `run_onchange_install-pkgs.sh.tmpl`          |  ✓    |  ✓    |   —     |
| `run_after_link-pwsh-profile.ps1` → resolves the real Documents folder (OneDrive/locale) and links the profile |  —    |  —    |   ✓     |
| `run_onchange_install-pkgs.ps1`              |  —    |  —    |   ✓     |

## Externals (`.chezmoiexternal.toml.tmpl`)

Declarative upstream sources cloned/fetched by chezmoi on `apply`, refreshed per `refreshPeriod`. Prefer over an install-script `git clone` whenever the upstream is "drop this repo at this path" (no build, no PATH wiring).

| Entry (target path)                                    | Upstream                                  | macOS | Linux | Windows |
| ------------------------------------------------------ | ----------------------------------------- | :---: | :---: | :-----: |
| `~/.config/vim/pack/catppuccin/start/catppuccin/`      | `github.com/catppuccin/vim`               |  ✓    |  ✓    |   —     |
| `~/vimfiles/pack/catppuccin/start/catppuccin/`         | `github.com/catppuccin/vim`               |  —    |  —    |   ✓     |
| `~/.local/share/fonts/JetBrainsMonoNerdFont/`          | Nerd Fonts `JetBrainsMono.tar.xz` release |  —    |  ✓ (not light) |   —     |
| `~/.oh-my-zsh/`                                        | `github.com/ohmyzsh/ohmyzsh`              |  ✓    |  —    |   —     |

**When NOT to use externals:** anything needing `chmod +x` on a downloaded binary with arch detection (oh-my-posh), official installer scripts (uv), or package-manager registration (brew/apt/winget). Those stay in `.chezmoiscripts/`.

## Asymmetries (by design)

- **Ghostty** has no Windows build → Windows Terminal fills the gap with matching theme/font.
- **zsh** is ignored on Linux and Windows (bash is the Unix default; PowerShell is the Windows default).
- **PowerShell profile** is Windows-only; trivially extendable to Unix by mirroring the `dot_config/powershell/` snippet pattern.

## Machine roles

Two independent yes/no answers, asked once per machine by `chezmoi init` and stored only in that machine's `~/.config/chezmoi/chezmoi.toml` (`data.light`, `data.work`). Templates read them with `get . "light"` / `get . "work"`, which treats a missing key as false.

| Flag | Prompt | When true |
|---|---|---|
| `light` | Light install (shell, git, vim, prompt only) | `.chezmoiignore` skips VS Code, Ghostty, Claude Code, Copilot and AGENTS.md; no Nerd Font download; the install script installs only `fzf git vim fd tmux` + oh-my-posh (no uv/Python). |
| `work` | Work computer (work git identity, no personal telemetry) | Asks the work repos folder and work email. Repos under that folder commit with the work email (`~/.config/git/config` → `includeIf "gitdir/i:<folder>/"` → `~/.config/git/work`); use the folder's real path, not a symlink. Claude Code sends no telemetry to the personal OTel endpoint (the keys are left out and stripped from the live file). |

Current machines: Mac = full/personal, Raspberry Pi = light/personal, Windows PC = full/work.

Change an answer later by passing the prompt text: `chezmoi init --promptBool "Work computer (work git identity, no personal telemetry)=true"`, then `chezmoi apply`.

The corporate-PC workarounds below are tied to Windows, not to `work`: they're harmless on any Windows machine.

## Bitwarden

Two Bitwarden CLIs, used per machine role:

| Role | `bw` (password vault) | `bws` (Secrets Manager) |
|---|---|---|
| full + personal (Mac) | Homebrew `bitwarden-cli`; fetches the age key at setup | own machine-account token → **personal** project |
| light (Pi) | — | — (`secrets.sh` by hand) |
| work (Windows PC) | — (personal vault stays off the work PC; age key copied by hand) | own machine-account token → **work** project |

- **Env secrets** (`~/.config/shell/secrets.sh`, `~/.config/powershell/secrets.ps1`) are rendered by chezmoi from the machine's project (`bws secret list <project> --output json`, via `.chezmoitemplates/bws-secrets.json`): each secret's name becomes the variable name. Edit secrets in Bitwarden, then `chezmoi apply`. Names that aren't valid variable names are skipped.
- **One token per machine**, created as a Bitwarden machine account with read access to one project, stored only in that machine's `~/.config/chezmoi/chezmoi.toml` (0600). Revoke a single machine's token in Bitwarden if it's lost.
- **`bws` is fetched** by `.chezmoiexternal.toml.tmpl` into `~/.local/bin` (version pinned in `.chezmoidata/bws.toml`; bump to update). The first apply on a new machine fetches it; the next apply renders the secrets.
- **Online needed:** on machines with a token, every `chezmoi status`/`diff`/`apply` calls `bws` once and fails offline.
- No Bitwarden MCP server: it has no read-only mode and would give an AI agent full vault access.
- **Claude Code and `bw`:** every `bw` command Claude runs needs your approval (an `ask` rule in the managed Claude settings; ask rules beat the blanket Bash allow, also inside pipes and in auto mode). To let Claude act, unlock in your own terminal into a private file, `(umask 077; bw unlock --raw > <file>)`; Claude runs `bw --session "$(cat <file>)" …` and prints only non-secret fields. Finish with `bw lock` and delete the file. The rule matches the command as written, so it's an approval gate, not a security boundary.

## Encrypted files (age)

Secrets such as `~/.ssh/config` are committed age-encrypted (`encrypted_` in the source name). Each machine needs the private key at `~/.config/chezmoi/key.txt` (owner-only) **before** `chezmoi apply`; copy it from another machine over a secure channel, never through the repo. The public key in `.chezmoi.toml.tmpl` is safe to publish. Add or update an encrypted file with `chezmoi add --encrypt <file>`.

## Hardcoded by design

chezmoi target paths are literal (they can't be templated), so a few values stay fixed on purpose:

- **`~/.config`, `~/.local/{share,state}`, `~/.cache`** — chezmoi deploys to these exact paths, so `.profile` pins `XDG_*` to the spec defaults. Pointing `XDG_CONFIG_HOME` elsewhere would split config from where chezmoi writes it.
- **Homebrew prefixes** `/opt/homebrew` (Apple Silicon) and `/usr/local` (Intel) — fixed by Homebrew.
- **Vendor paths** — `Library/Application Support/…`, `AppData/…`, the Windows Terminal package family name.
- **`{{ .chezmoi.homeDir }}`** in templates — rendered per machine.
- **Personal data** — plugin marketplaces, the OTel endpoint (personal machines only), project aliases. Git identity is prompted at `chezmoi init`.

Anything that genuinely differs per machine and can't be templated (e.g. a redirected Documents folder) is resolved at runtime by a script.

## Files merged, not replaced

`~/.claude/settings.json` and `~/.copilot/settings.json` are also written by the apps themselves, so chezmoi doesn't own them whole. `modify_` templates overlay `.chezmoitemplates/{claude,copilot}-settings.json` onto the live file: managed keys win, keys the app writes stay local and never reach this public repo (Claude `autoMode`, Copilot `model`; Copilot `allowedUrls` is the union of both). Removing a managed key from a template doesn't delete it from the live file; delete it there once (except the telemetry keys on work machines, which are stripped automatically).

## Managed (corporate) Windows PCs

Endpoint policies on the work PC block several defaults; the workarounds apply to every Windows machine:

| Blocked | Workaround |
|---|---|
| Machine-wide installs (winget font package) | JetBrainsMono installed per user by `oh-my-posh font install` |
| `Set-Acl` (needs SeSecurityPrivilege) | `icacls` restricts `secrets.ps1` |
| uv's `python.exe` launchers in `~\.local\bin` (unsigned exe named python.exe) | `uv python install --no-bin`; `profile.ps1` puts uv's `cpython-3.NN-*` interpreter folder on PATH |
| `New-Item -ItemType SymbolicLink` in Windows PowerShell 5.1 without admin | `mklink` (honours Developer Mode) in `run_after_link-pwsh-profile.ps1` |
| Documents redirected to OneDrive and localised | profile link target resolved at runtime with `GetFolderPath('MyDocuments')` |

## Testing the ignore matrix

```sh
chezmoi execute-template \
  --override-data '{"light":false,"work":false,"chezmoi":{"os":"darwin"}}' \
  < .chezmoiignore   # vary os (darwin/linux/windows), light, work
```

For the live OS, `chezmoi managed` and `chezmoi ignored` show the actual split.
