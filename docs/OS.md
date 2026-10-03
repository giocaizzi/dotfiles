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
| `dot_claude/`                                        | all OSes                                            | all OSes                                       | all OSes                                                                           |

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
| `~/.local/share/fonts/JetBrainsMonoNerdFont/`          | Nerd Fonts `JetBrainsMono.tar.xz` release |  —    |  ✓    |   —     |
| `~/.oh-my-zsh/`                                        | `github.com/ohmyzsh/ohmyzsh`              |  ✓    |  —    |   —     |

**When NOT to use externals:** anything needing `chmod +x` on a downloaded binary with arch detection (oh-my-posh), official installer scripts (uv), or package-manager registration (brew/apt/winget). Those stay in `.chezmoiscripts/`.

## Asymmetries (by design)

- **Ghostty** has no Windows build → Windows Terminal fills the gap with matching theme/font.
- **zsh** is ignored on Linux and Windows (bash is the Unix default; PowerShell is the Windows default).
- **PowerShell profile** is Windows-only; trivially extendable to Unix by mirroring the `dot_config/powershell/` snippet pattern.

## Light machines

`chezmoi init` asks once per machine whether it is a light install (stored as `data.light`). Light machines (e.g. the Raspberry Pi over SSH) get the shell, git, vim and the oh-my-posh prompt only: `.chezmoiignore` skips VS Code, Ghostty, Claude/Copilot and fonts, and the install script installs `fzf git vim fd` + oh-my-posh. Change it later with `chezmoi init --promptBool "Light install (shell, git, vim, prompt only)=true"` (the flag matches the prompt text, not the key).

## Work git identity

`chezmoi init` asks once per machine for a work repos folder and email (blank = none), stored only in that machine's `~/.config/chezmoi/chezmoi.toml`. Where set, `~/.config/git/config` includes `~/.config/git/work` for repos under that folder (`includeIf "gitdir/i:…"`), so they commit with the work email. Use the folder's real path, not a symlink.

## Encrypted files (age)

Secrets such as `~/.ssh/config` are committed age-encrypted (`encrypted_` in the source name). Each machine needs the private key at `~/.config/chezmoi/key.txt` (owner-only) **before** `chezmoi apply`; copy it from another machine over a secure channel, never through the repo. The public key in `.chezmoi.toml.tmpl` is safe to publish. Add or update an encrypted file with `chezmoi add --encrypt <file>`.

## Hardcoded by design

chezmoi target paths are literal (they can't be templated), so a few values stay fixed on purpose:

- **`~/.config`, `~/.local/{share,state}`, `~/.cache`** — chezmoi deploys to these exact paths, so `.profile` pins `XDG_*` to the spec defaults. Pointing `XDG_CONFIG_HOME` elsewhere would split config from where chezmoi writes it.
- **Homebrew prefixes** `/opt/homebrew` (Apple Silicon) and `/usr/local` (Intel) — fixed by Homebrew.
- **Vendor paths** — `Library/Application Support/…`, `AppData/…`, the Windows Terminal package family name.
- **`{{ .chezmoi.homeDir }}`** in templates — rendered per machine.
- **Personal data** — plugin marketplaces, OTel endpoint, project aliases. Git identity is prompted at `chezmoi init`.

`~/.claude/settings.json` and `~/.copilot/settings.json` are merged, not replaced: `modify_` templates overlay `.chezmoitemplates/{claude,copilot}-settings.json` onto the live file, so keys the app writes itself (Claude `autoMode`, Copilot `model`, approved `allowedUrls`) stay local. Removing a managed key from the template doesn't delete it from the live file; delete it there once.

Anything that genuinely differs per machine and can't be templated (e.g. a redirected Documents folder) is resolved at runtime by a script.

## Testing the ignore matrix

```sh
cat .chezmoiignore | chezmoi execute-template --init \
  --promptString email=x --promptString gitUser=x --promptString gitEmail=x \
  --override-data '{"chezmoi":{"os":"darwin"}}'   # or "linux" / "windows"
```

For the live OS, `chezmoi managed` and `chezmoi ignored` show the actual split.
