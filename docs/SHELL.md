# Shell Configuration Guide (Bash & Zsh)

This document describes the centralized shell configuration shared between bash and zsh, along with shell-specific features.

**Architecture:** Core POSIX config in `.profile`, shell-specific features in `.bashrc` and `.zshrc`.

---

## 📋 Table of Contents

- [Configuration Structure](#-configuration-structure)
- [Core Configuration (.profile)](#-core-configuration-profile)
- [History Features](#-history-features)
- [Completion System](#-completion-system)
- [Key Bindings](#-key-bindings)
- [Directory Navigation](#-directory-navigation)
- [Autosuggestions](#-autosuggestions)
- [Aliases](#-aliases)
- [Shell Options](#-shell-options)
- [Tools & Plugins](#-tools--plugins)
- [Python Version Management (pyenv)](#-python-version-management-pyenv)
- [Tips & Tricks](#-tips--tricks)

---

## 🏗️ Configuration Structure

### File Hierarchy
```
~/
├── .profile          # POSIX core (PATH, env vars, universal aliases)
├── .bashrc           # Bash-specific features
├── .zshrc            # Zsh-specific features
└── .bash_profile     # macOS/login shell wrapper
```

### What Goes Where

**`.profile`** (POSIX-compliant, sourced by both shells):
- PATH management (Homebrew, pipx, pyenv, Android SDK)
- Environment variables (`EDITOR`, `ANDROID_HOME`)
- Universal aliases (`ll`, `la`, `..`, grep colors)
- Cross-shell compatible functions

**`.bashrc`** (Bash-specific):
- Bash history settings
- Bash completion
- fzf key bindings for bash
- Bash prompt (oh-my-posh)
- nvm (Node Version Manager)

**`.zshrc`** (Zsh-specific):
- Oh My Zsh framework
- Zsh history settings
- Zsh completion system with styling
- Zsh plugins (autosuggestions)
- Zsh options (`setopt`)
- fzf key bindings for zsh

### Maintenance
```bash
# Bash
source ~/.bashrc

# Zsh
rm -f ~/.zcompdump && exec zsh
omz update

# Re-initialize pyenv
eval "$(pyenv init - zsh)"  # zsh
eval "$(pyenv init - bash)" # bash
```

---

## 🌐 Core Configuration (.profile)

Shared POSIX-compliant configuration sourced by both bash and zsh.

### Environment Variables
- `EDITOR=vim` - Default text editor
- `PYENV_ROOT` - Python version manager
- `ANDROID_HOME` - Android SDK location
- `HOMEBREW_PREFIX` - Homebrew installation path

### Universal Aliases
```bash
..          # cd ..
...         # cd ../..
....        # cd ../../..
ll          # ls -lah
la          # ls -A
l           # ls -CF
df          # df -h (human-readable)
du          # du -h (human-readable)
```

---

## 📚 History Features

Both shells maintain comprehensive command history with similar features.

### Bash Configuration
- **50,000** entries stored in memory and file
- History file: `~/.bash_history`
- `HISTCONTROL=ignoreboth:erasedups` - No duplicates, ignore commands with leading space
- `histappend` - Append to history (don't overwrite)
- Saved after each command

### Zsh Configuration
- **50,000** entries stored in history
- History file: `~/.zsh_history`
- Shared across all terminal sessions in real-time
- `EXTENDED_HISTORY` - Timestamps recorded
- `HIST_IGNORE_ALL_DUPS` - No duplicate entries
- `SHARE_HISTORY` - Real-time sharing between sessions

### Usage (Both Shells)
```bash
# Interactive fuzzy history browser (custom function)
h                   # Opens fzf to browse and select from entire history
h search-term       # Pre-filter history with search term

# Quick history search
Ctrl+R              # fzf interactive search (shows preview)
                    # Ctrl+Y copies command to clipboard

# Prefix search with arrows
↑/↓                 # Navigate through history matching current input prefix

# View history
history             # Show all history
history | grep foo  # Search history for specific commands
```

**Note:** History browsing uses [fzf](https://github.com/junegunn/fzf), installed via `run_onchange_install-pkgs.sh`.

---

## 🎯 Completion System

### Bash Completion
- Provided by `bash-completion` package (Homebrew)
- Automatic completion for common commands
- Case-insensitive with `nocaseglob`
- Extended pattern matching with `extglob`
- Recursive globbing with `**` (globstar)

### Zsh Completion
- Advanced visual feedback with colors:
  - **Green** - Available completions descriptions
  - **Yellow** - Spelling corrections with error count
  - **Purple** - System messages
  - **Red** - No matches found warnings
- Menu selection with arrow key navigation
- Smart matching from anywhere in name
- Completion cache for faster performance

### Common Features (Both Shells)
- **Case-insensitive matching** - `cd documents` matches `Documents`
- Tab to trigger completion, Tab Tab to show all options
- Corrections offered for typos
- Color-coded file types

### Usage
```bash
# Bash
Tab                 # Trigger completion
Tab Tab             # Show all options

# Zsh (additional features)
Tab                 # Trigger completion with menu
↑/↓/←/→             # Navigate completion menu
Enter               # Accept selection
```

---

## ⌨️ Key Bindings

### History Search (Both Shells)
| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl+R` | fzf-history-widget | **Interactive fuzzy search** - scrollable history list with fzf |
| `↑` | history-beginning-search-backward | Search history for commands starting with current input |
| `↓` | history-beginning-search-forward | Navigate forward through matching history |

**Examples:** 
- Press `Ctrl+R` to open fzf with your entire history - type to filter, arrows to navigate, Enter to select
- Type `git` then press `↑` to cycle through all git commands in history

### Line Navigation
| Key | Action | Description |
|-----|--------|-------------|
| `Home` | beginning-of-line | Jump to start of line |
| `End` | end-of-line | Jump to end of line |
| `Ctrl+A` | beginning-of-line | Jump to start of line (emacs style) |
| `Ctrl+E` | end-of-line | Jump to end of line (emacs style) |
| `Delete` | delete-char | Delete character under cursor |

### Word Navigation
| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl+→` | forward-word | Move forward one word |
| `Ctrl+←` | backward-word | Move backward one word |
| `Alt+F` | forward-word | Move forward one word (emacs style) |
| `Alt+B` | backward-word | Move backward one word (emacs style) |

### Editing
| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl+W` | backward-kill-word | Delete word before cursor |
| `Ctrl+U` | backward-kill-line | Delete from cursor to start of line |

**Tip:** Combine with Shift to select text in compatible terminals.

---

## 📁 Directory Navigation

### Auto-CD (Zsh only)
```bash
# Just type the directory name - no 'cd' needed!
~/Documents         # Automatically changes to Documents directory
..                  # Up one level
```

### Directory Stack (Zsh with Oh My Zsh)
Zsh maintains a stack of visited directories:

```bash
# View directory stack
d                   # Shows numbered list of recent directories

# Jump to directory by number (aliases 1-9)
1                   # Jump to 1st directory in stack
2                   # Jump to 2nd directory in stack

# Built-in navigation (both shells)
cd -                # Toggle between last two directories
pushd <dir>         # Push directory onto stack
popd                # Pop directory from stack
```

### Quick Navigation Aliases (Both Shells)
```bash
..                  # cd ..
...                 # cd ../..
....                # cd ../../..
```

### Smart Features
- `AUTO_PUSHD` - Automatically builds directory stack as you navigate
- `PUSHD_IGNORE_DUPS` - Keeps stack clean (no duplicates)
- `PUSHD_SILENT` - Stack operations don't clutter output

---

## 💡 Autosuggestions

### Zsh Only
Powered by `zsh-autosuggestions` plugin via Homebrew.

### How It Works
As you type in zsh, suggestions appear in **gray text** based on:
1. **Command history** (primary)
2. **Completion system** (secondary)

### Accepting Suggestions
```bash
→                   # Accept entire suggestion
Ctrl+→              # Accept one word at a time
Esc                 # Clear suggestion
```

---

## 🔧 Aliases

### Universal Aliases (From .profile)
```bash
..                  # cd ..
...                 # cd ../..
....                # cd ../../..
ll                  # ls -lah
la                  # ls -A
l                   # ls -CF
df                  # df -h
du                  # du -h
mkdir               # mkdir -p (always create parents)
grep/fgrep/egrep    # With color output
```

### Bash-Specific Aliases
```bash
dirs                # Directory stack (verbose)
gs/ga/gc/gd/gp     # Git shortcuts
```

### Zsh-Specific Aliases (Oh My Zsh)
```bash
d                   # dirs -v (directory stack)
1-9                 # Jump to directory in stack
# Git plugin provides 100+ aliases:
ga                  # git add
gc                  # git commit --verbose
gd                  # git diff
gp                  # git push
gst                 # git status
glog                # git log --oneline --decorate --graph
```

---

## ⚙️ Shell Options

### Bash Options
- `histappend` - Append to history file
- `cmdhist` - Save multi-line commands as one entry
- `checkwinsize` - Update terminal size after each command
- `nocaseglob` - Case-insensitive globbing
- `extglob` - Extended pattern matching
- `globstar` - Recursive globbing with **
- `cdspell/dirspell` - Correct minor spelling errors

### Zsh Options
**Globbing & Pattern Matching:**
- `EXTENDED_GLOB` - Advanced pattern matching (`**`, `^`, `~`)
- `GLOB_DOTS` - Include hidden files in glob results
- `NO_CASE_GLOB` - Case-insensitive filename matching
- `NUMERIC_GLOB_SORT` - Sort files numerically (file1, file2, file10)

**Directory Navigation:**
- `AUTO_CD` - cd by typing directory name
- `AUTO_PUSHD` - Automatically push directories onto stack
- `PUSHD_IGNORE_DUPS` - Don't push duplicate directories
- `PUSHD_SILENT` - Don't print directory stack

**Input/Output:**
- `CORRECT` - Suggests corrections for mistyped commands
- `INTERACTIVE_COMMENTS` - Allow `#` comments in command line
- `RC_QUOTES` - Use `''` to represent a literal single quote

**Job Control:**
- `AUTO_CONTINUE` - Automatically continue stopped jobs when disowning
- `LONG_LIST_JOBS` - Detailed job information
- `NOTIFY` - Immediate notification when background job finishes

**Completion:**
- `COMPLETE_IN_WORD` - Completion works from cursor position
- `ALWAYS_TO_END` - Cursor moves to end after completion
- `LIST_PACKED` - Compact completion list display

---

## 🎨 Tools & Plugins

### Shared Tools (Both Shells)
- **pyenv** - Python version manager for managing multiple Python versions
- **fzf** - Fuzzy finder for history, files (Ctrl+R for history search)
- **oh-my-posh** - Customizable prompt theme

### Bash-Specific
- **nvm** - Node Version Manager
- **bash-completion** - Enhanced command completion

### Zsh-Specific (Oh My Zsh)
**Plugins:**
- **git** - Git aliases and prompt integration
- **docker** - Docker command completions
- **gh** - GitHub CLI integration

**External:**
- **zsh-autosuggestions** - Fish-like command suggestions

---

## � Python Version Management (pyenv)

### Configuration
pyenv is initialized in both bash and zsh to manage Python versions:
- PATH setup in `.profile` with `pyenv init --path`
- Full initialization in `.bashrc` and `.zshrc` with `pyenv init - <shell>`
- Global Python version set via `~/.python-version` file (currently 3.13)

### Common Commands
```bash
# List installed versions
pyenv versions

# Install a new Python version
pyenv install 3.13.0

# Set global Python version
pyenv global 3.13

# Set local project version (creates .python-version in current dir)
pyenv local 3.11.0

# Show current active version
pyenv version

# List available versions to install
pyenv install --list
```

### Version Priority
pyenv checks for Python versions in this order:
1. `PYENV_VERSION` environment variable
2. `.python-version` file in current directory (local project)
3. `~/.python-version` file (global default - managed by chezmoi)
4. System Python

### Updating pyenv
```bash
# via Homebrew (recommended on macOS)
brew upgrade pyenv

# Or update Python build definitions
cd $(pyenv root) && git pull
```

---

## �🚀 Tips & Tricks

### History Power User (Both Shells)
```bash
# Fuzzy search with fzf (best method!)
Ctrl+R              # Opens interactive list - type to filter
                    # Navigate with ↑/↓, select with Enter

# Prefix search with arrows
git log<↑>          # Find last command starting with "git log"

# Don't save sensitive commands
 echo $PASSWORD     # Leading space = not saved to history
```

### Tab Completion (Both Shells)
```bash
cd ~/D<Tab>         # Completes to ~/Documents or ~/Downloads
git co<Tab><Tab>    # Shows: commit, config, checkout, etc.
```

### Directory Stack (Zsh)
```bash
cd ~/project1       # Work in project 1
cd ~/project2       # Work in project 2
d                   # See stack
1                   # Instantly back to project 1
2                   # Back to project 2
```

### Case-Insensitive (Both Shells)
```bash
cd documents        # Works even if it's 'Documents'
ls *.PDF            # Matches .pdf files too (zsh and bash with nocaseglob)
```

---

## 📖 Further Reading

- [Bash Manual](https://www.gnu.org/software/bash/manual/)
- [Zsh Documentation](https://zsh.sourceforge.io/Doc/)
- [Oh My Zsh](https://ohmyz.sh/)
- [fzf](https://github.com/junegunn/fzf)
- [zsh-autosuggestions](https://github.com/zsh-users/zsh-autosuggestions)
- [oh-my-posh](https://ohmyposh.dev/)

---

**Last Updated:** February 2026
