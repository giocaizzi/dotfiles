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
| `dot_config/powershell/profile.ps1` *                | —                                                           | —                                              | `~/OneDrive - Jakala SpA/Documenti/{PowerShell,WindowsPowerShell}/profile.ps1` (symlinks) |
| `dot_copilot/settings.json`                          | `~/.copilot/settings.json`                                  | `~/.copilot/settings.json`                     | `~/.copilot/settings.json`                                                         |
| `dot_claude/symlink_CLAUDE.md.tmpl`                  | `~/.claude/CLAUDE.md` → `~/.config/agents/AGENTS.md`        | same                                           | same                                                                               |
| `dot_copilot/symlink_copilot-instructions.md.tmpl`   | `~/.copilot/copilot-instructions.md` → `~/.config/agents/AGENTS.md` | same                                    | same                                                                               |
| `.chezmoidata/shortcuts.toml`                        | `~/.config/shell/aliases.sh`                                | `~/.config/shell/aliases.sh`                   | `~/.config/powershell/aliases.ps1` (+ `~/.config/shell/aliases.sh` for Git Bash)   |
| `dot_config/git/config.tmpl`                         | `~/.config/git/config`                                      | `~/.config/git/config`                         | `~/.config/git/config`                                                             |
| `dot_mytheme.omp.json`                               | `~/.mytheme.omp.json`                                       | `~/.mytheme.omp.json`                          | `~/.mytheme.omp.json`                                                              |
| `dot_profile`, `dot_bashrc`, `dot_bash_profile`      | `~/.profile`, `~/.bashrc`, `~/.bash_profile`                | same                                           | same (used by Git Bash / WSL)                                                      |
| `dot_zshrc`, `dot_zprofile`                          | `~/.zshrc`, `~/.zprofile`                                   | — (ignored)                                    | — (ignored)                                                                        |
| `dot_vimrc`, `dot_gitignore_global`, `dot_stCommitMsg`, `dot_python-version`, `dot_claude/` | all OSes                                            | all OSes                                       | all OSes                                                                           |

\* Symlink-target-only sources — never deployed as regular files anywhere (listed in `.chezmoiignore` unconditionally); only referenced via `{{ .chezmoi.sourceDir }}/...` from the Windows-specific symlinks.

## Bootstrap scripts (`.chezmoiscripts/`)

| Script                                       | macOS | Linux | Windows |
| -------------------------------------------- | :---: | :---: | :-----: |
| `run_once_init-untracked-env.sh`             |  ✓    |  ✓    |   —     |
| `run_once_init-untracked-env.ps1`            |  —    |  —    |   ✓     |
| `run_onchange_install-pkgs.sh.tmpl`          |  ✓    |  ✓    |   —     |
| `run_onchange_install-pkgs.ps1`              |  —    |  —    |   ✓     |

## Externals (`.chezmoiexternal.toml.tmpl`)

Declarative upstream sources cloned/fetched by chezmoi on `apply`, refreshed per `refreshPeriod`. Prefer over an install-script `git clone` whenever the upstream is "drop this repo at this path" (no build, no PATH wiring).

| Entry (target path)                                    | Upstream                                  | macOS | Linux | Windows |
| ------------------------------------------------------ | ----------------------------------------- | :---: | :---: | :-----: |
| `~/.vim/pack/catppuccin/start/catppuccin/`             | `github.com/catppuccin/vim`               |  ✓    |  ✓    |   —     |
| `~/vimfiles/pack/catppuccin/start/catppuccin/`         | `github.com/catppuccin/vim`               |  —    |  —    |   ✓     |
| `~/.oh-my-zsh/`                                        | `github.com/ohmyzsh/ohmyzsh`              |  ✓    |  —    |   —     |

**When NOT to use externals:** anything needing `chmod +x` on a downloaded binary with arch detection (oh-my-posh), git-clone-plus-plugins flows (pyenv + virtualenv/doctor/update plugins), Registry/PATH setup (pyenv-win), or package-manager registration (brew/apt/winget). Those stay in `.chezmoiscripts/`.

## Asymmetries (by design)

- **Ghostty** has no Windows build → Windows Terminal fills the gap with matching theme/font.
- **zsh** is ignored on Linux and Windows (bash is the Unix default; PowerShell is the Windows default).
- **PowerShell profile** is Windows-only; trivially extendable to Unix by mirroring the `dot_config/powershell/` snippet pattern.

## Testing the ignore matrix

```sh
cat .chezmoiignore | chezmoi execute-template --init \
  --promptString email=x --promptString gitUser=x --promptString gitEmail=x \
  --override-data '{"chezmoi":{"os":"darwin"}}'   # or "linux" / "windows"
```

For the live OS, `chezmoi managed` and `chezmoi ignored` show the actual split.
