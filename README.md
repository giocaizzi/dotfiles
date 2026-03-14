# dotfiles

My configuration files backup and syncronization system.

Using [chezmoi](https://www.chezmoi.io/)

## Init chezmoi on a machine

Will be prompted for local data, like name and email.

```shell
chezmoi init gh:giocaizzi/dotfiles.git
```

## File structure

- chezmoi repository: `~/.local/share/chezmoi`
- chezmoi local configuration and state: `~/.config/chezmoi`

## Usage

- Navigate to chezmoi repository for file ops.

```shell
chezmoi cd
```

- Ensure that files are in the target state, applying changes if necessary.

```shell
chezmoi update
chezmoi apply
```

- Add new files to be managed by chezmoi.

```shell
chezmoi add <file>
```

- Edit files directly in the chezmoi repository.

```shell
chezmoi edit <file>
```

- View the status of files managed by chezmoi.

```shell
chezmoi status
```

- View differences between source and target states.

```shell
chezmoi diff
```

- Merge changes from the source state into the target state.

```shell
chezmoi merge
chezmoi merge --all
```

## Scripts

Scripts in [`.chezmoiscripts/`](./.chezmoiscripts) run during `chezmoi apply` based on filename prefixes.

### Execution Order

1. Read source/destination → compute changes → `run_before_*` → update files → `run_after_*`  
2. Scripts execute **alphabetically** within each phase (use numeric prefixes like `10-`, `20-` for explicit ordering)

### Script Prefixes

| Prefix | When Executed | Use For |
|--------|--------------|---------|
| `run_` | Every `chezmoi apply` | Environment setup, always-run tasks |
| `run_once_` | Once per unique content | One-time initialization |
| `run_onchange_` | When content changes | Package installs, syncing |
| `run_before_` | Before file updates | Pre-setup, validation |
| `run_after_` | After file updates | Services restart, post-setup |

**Combine prefixes:** `run_once_before_`, `run_onchange_after_`. All scripts must be idempotent.

### Templates & Conditionals

Scripts with `.tmpl` suffix access chezmoi variables: `{{ .chezmoi.os }}`, `{{ .chezmoi.arch }}`, `{{ .chezmoi.homeDir }}`. Empty templates don't execute.

**Prefer simplicity:** Use `.chezmoiignore` for conditional execution and plain scripts (`.sh`) over templates (`.tmpl`) when possible. Templates add complexity—only use when dynamic content generation is required.

### Current Scripts

**MacOS/Linux:**
- `run_onchange_env.sh` - Environment variables (when changed)
- `run_onchange_install-pkgs.sh.tmpl` - Package management (when list changes)
- `run_after_03-sync-ai-assistants.sh` - Clone/pull `~/.ai-assistants/skills`, `~/.ai-assistants/ralph-copilot`, and `~/.ai-assistants/copilot-agents`; symlink `~/.claude/skills` to `~/.ai-assistants/skills/skills`

**Windows:**
- `run_onchange_env.ps1` - PowerShell environment
- `run_after_03-sync-ai-assistants.ps1` - Clone/pull `~/.ai-assistants/skills`, `~/.ai-assistants/ralph-copilot`, and `~/.ai-assistants/copilot-agents`; create `~/.claude/skills` junction to `~/.ai-assistants/skills/skills`

### State Management

Reset script execution tracking:
```bash
chezmoi state delete-bucket --bucket=entryState   # run_onchange_
chezmoi state delete-bucket --bucket=scriptState  # run_once_
```

### SSH Configuration

Sync scripts use SSH host aliases from `~/.ssh/config` to manage multiple GitHub accounts. This configuration allows using different SSH keys for different GitHub accounts.

**Example `~/.ssh/config` entry:**
```ssh
Host github.com
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_github_giocaizzi
  IdentitiesOnly yes
```

The scripts use `git@github.com:giocaizzi/<repo>.git` URLs, which automatically match the SSH config entry above.

### Best Practices

✅ Idempotent, numeric ordering, `set -e`, user feedback  
❌ Don't modify source/dest state, avoid externals in `run_before_`, no exit codes for warnings

## Applications

Using symlinks, the [apps](./apps/) folder contains configuration files for specific applications, which are linked to their expected locations in the filesystem.

## Ignore files

The [.chezmoiignore](./.chezmoiignore) file is used to exclude certain files or directories from being managed by `chezmoi`, ensuring that only the intended configuration files are tracked and applied. This will consider OS and other variables.

## Documentation

- [docs/OS.md](./docs/OS.md) - Operating-system specific behavior and testing.
- [docs/SHELL.md](./docs/SHELL.md) - Shell configuration architecture and usage.
- [docs/GIT.md](./docs/GIT.md) - Git commit conventions (Conventional Commits 1.0.0).
- Official chezmoi docs: [chezmoi documentation](https://www.chezmoi.io/docs/).