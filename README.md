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
- `run_env.sh` - Environment variables (every apply)
- `run_onchange_install-pkgs.sh.tmpl` - Package management (when list changes)
- `run_after_10-sync-skills.sh` - Clone/pull `~/.claude/skills`
- `run_after_11-sync-copilot-agents.sh` - Clone/pull copilot agents to `$HOME`

**Windows:** `run_env.ps1` - PowerShell environment

### State Management

Reset script execution tracking:
```bash
chezmoi state delete-bucket --bucket=entryState   # run_onchange_
chezmoi state delete-bucket --bucket=scriptState  # run_once_
```

### Best Practices

✅ Idempotent, numeric ordering, `set -e`, user feedback  
❌ Don't modify source/dest state, avoid externals in `run_before_`, no exit codes for warnings

## Applications

Using symlinks, the [apps](./apps/) folder contains configuration files for specific applications, which are linked to their expected locations in the filesystem.

## Ignore files

The [.chezmoiignore](./.chezmoiignore) file is used to exclude certain files or directories from being managed by `chezmoi`, ensuring that only the intended configuration files are tracked and applied. This will consider OS and other variables.

## Documentation

See the official [chezmoi documentation](https://www.chezmoi.io/docs/) for more details and advanced usage.