---
name: chezmoi
description: Manages dotfiles across machines with chezmoi. Use when working in this chezmoi source directory to add/edit dotfiles, write Go templates, configure scripts, handle OS-specific config, or manage source state attributes.
---

# Chezmoi Dotfiles Management

Source dir: `~/.local/share/chezmoi` (this repo) | Destination: `~`  
Docs: https://www.chezmoi.io | Template engine: Go `text/template` + sprig + chezmoi functions

## Source State File Naming

File names encode behavior. Prefixes are processed **in order**; suffixes come last.

| Prefix | Effect |
|--------|--------|
| `dot_` | Rename to `.name` in destination (`dot_zshrc` → `~/.zshrc`) |
| `symlink_` | Create a symlink instead of a regular file |
| `run_` | Execute as script during `chezmoi apply` |
| `run_once_` | Run only once per unique content hash |
| `run_onchange_` | Run whenever content hash changes |
| `run_before_` | Run before file updates |
| `run_after_` | Run after file updates |
| `private_` | Remove group/world permissions |
| `readonly_` | Remove write permissions |
| `exact_` | Remove unmanaged entries from the directory |
| `create_` | Create only if destination doesn't exist |
| `modify_` | Script that modifies existing destination file |
| `empty_` | Preserve empty files (default removes them) |
| `encrypted_` | Encrypt in source (age or gpg) |

| Suffix | Effect |
|--------|--------|
| `.tmpl` | Process contents as Go template |

**Script naming pattern:**
```
run_[once_|onchange_][before_|after_][NN_]description.sh[.tmpl]
```
Use numeric ordering prefixes (`10_`, `20_`) for explicit ordering within a phase.

**Examples:**
- `dot_zshrc.tmpl` → `~/.zshrc` (template)
- `symlink_dot_config.tmpl` → `~/.config` (symlink, template)
- `run_once_before_10_install-brew.sh` → one-time pre-update script

## Templates

### Key Variables

| Variable | Type | Description |
|----------|------|-------------|
| `.chezmoi.os` | string | `darwin`, `linux`, `windows` |
| `.chezmoi.arch` | string | `amd64`, `arm64`, etc. |
| `.chezmoi.hostname` | string | Hostname (up to first `.`) |
| `.chezmoi.username` | string | Current user |
| `.chezmoi.homeDir` | string | Home directory path |
| `.chezmoi.sourceDir` | string | Source directory path |
| `.chezmoi.config` | object | Full chezmoi config |
| `.chezmoi.kernel` | object | Kernel info (Linux only) |
| `.chezmoi.osRelease` | object | `/etc/os-release` data (Linux only) |

Custom variables: defined in `~/.config/chezmoi/chezmoi.toml` under `[data]`, accessed as `.varname`.

Debug: `chezmoi data` | `chezmoi execute-template "{{ .chezmoi.os }}"`

### Common Patterns

**OS-specific blocks:**
```
{{- if eq .chezmoi.os "darwin" }}
# macOS only
{{- else if eq .chezmoi.os "linux" }}
# Linux only
{{- end }}
```

**Machine-specific:**
```
{{- if eq .chezmoi.hostname "work-laptop" }}
export WORK=true
{{- end }}
```

**Custom data** (`chezmoi.toml` → `[data]` → `email = "me@example.com"`):
```
{{ .email }}
```

**Shared template** (file in `.chezmoitemplates/`):
```
{{- template "shared.tmpl" . -}}
```

**Whitespace:** `{{-` trims leading whitespace; `-}}` trims trailing whitespace.

## Special Directories

| Directory | Purpose |
|-----------|---------|
| `.chezmoiscripts/` | Scripts run during `chezmoi apply` |
| `.chezmoitemplates/` | Shared templates reusable with `{{ template "name" . }}` |
| `.chezmoidata/` | Additional TOML/JSON/YAML files merged into template data |
| `.chezmoiexternals/` | External archives/files managed by chezmoi |

## Scripts

Scripts in `.chezmoiscripts/` run during `chezmoi apply`.

**Execution order:**
1. `run_before_*` (alpha) → 2. File updates → 3. `run_after_*` (alpha)

| Prefix | Condition |
|--------|-----------|
| `run_` | Every `chezmoi apply` |
| `run_once_` | Once per unique content hash |
| `run_onchange_` | When content changes |
| `run_before_` | Before file updates |
| `run_after_` | After file updates |

Combining: `run_once_before_10_install-tools.sh.tmpl`

Scripts MUST be idempotent. Use `.tmpl` suffix for OS-conditional logic.

## Special Files

| File | Purpose |
|------|---------|
| `.chezmoiignore` | Glob patterns to exclude (template-aware) |
| `.chezmoiremove` | Files to remove from destination |
| `.chezmoiexternal.toml` | External archives/files to download and manage |
| `.chezmoidata.toml` | Supplement template data |
| `.chezmoi.toml.tmpl` | Generate chezmoi config on `chezmoi init` |

### .chezmoiignore

Logic is **inverted**: list paths to *ignore* (not install). Supports templates.

```
README.md
{{- if ne .chezmoi.os "darwin" }}
Library        # ignore on non-macOS
{{- end }}
{{- if ne .chezmoi.os "windows" }}
AppData        # ignore on non-Windows
{{- end }}
```

Check ignored files: `chezmoi ignored`

### .chezmoiexternal.toml

```toml
[".oh-my-zsh"]
  type = "archive"
  url = "https://github.com/ohmyzsh/ohmyzsh/archive/master.tar.gz"
  stripComponents = 1
```

## This Repository's Cross-Platform Structure

Manages macOS, Linux, and Windows config in one source:

```
apps/            ← Actual config files (vscode, ghostty, etc.)
Library/         ← macOS app config paths (symlinks via templates)
AppData/         ← Windows app config paths (symlinks via templates)
.chezmoiignore   ← Excludes OS-specific dirs on wrong platform
.chezmoiscripts/ ← Setup scripts
```

Symlink templates (e.g. `Library/Application Support/Code/User/symlink_settings.json.tmpl`) point to `apps/vscode/settings.json` for centralized management. The `.chezmoiignore` excludes `Library/` on non-macOS and `AppData/` on non-Windows.

## Daily Operations

```bash
chezmoi add <file>              # Start tracking a file
chezmoi add --template <file>   # Add and convert to template
chezmoi edit <file>             # Edit source state
chezmoi apply                   # Apply source → destination
chezmoi apply -v -n             # Dry run (verbose, no changes made)
chezmoi diff                    # Show pending changes
chezmoi update                  # git pull + apply
chezmoi re-add                  # Sync home dir changes back to source
chezmoi status                  # File status (M=modified, A=added)
chezmoi managed                 # List all managed files
chezmoi data                    # Show all template variables
chezmoi cd                      # cd to source directory
chezmoi git -- <args>           # Run git in source directory
chezmoi doctor                  # Diagnose issues
```

## Common Workflows

### Add and commit a new dotfile

```bash
chezmoi add ~/.newfile
chezmoi edit ~/.newfile         # edit in source
chezmoi apply
chezmoi git -- add -A && chezmoi git -- commit -m "feat: add newfile"
```

### Make an existing file OS-conditional

```bash
chezmoi chattr +template ~/.file   # adds .tmpl suffix to source file
chezmoi edit ~/.file               # add {{ if eq .chezmoi.os "darwin" }}...{{ end }}
chezmoi apply -v -n                # verify before applying
```

### Add a one-time setup script

Create `.chezmoiscripts/run_once_before_10_setup.sh.tmpl`:

```bash
#!/bin/bash
{{- if eq .chezmoi.os "darwin" }}
# macOS setup (e.g. brew install ...)
{{- else if eq .chezmoi.os "linux" }}
# Linux setup (e.g. apt install ...)
{{- end }}
```

### Add config for a new OS-specific app path

1. Place config in `apps/myapp/config`
2. Create the OS-specific symlink source (e.g. `Library/Application Support/MyApp/symlink_config.tmpl`)
3. Add exclusion to `.chezmoiignore` for other OSes
4. Run `chezmoi apply -v -n` to verify