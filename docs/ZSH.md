# ZSH Configuration Features & Keybindings

This document describes the features, keybindings, and functionality configured in the zsh setup.

**Performance:** Optimized for fast startup with lazy loading and Homebrew integration.

---

## 📋 Table of Contents

- [Performance Optimizations](#-performance-optimizations)
- [History Features](#-history-features)
- [Completion System](#-completion-system)
- [Key Bindings](#-key-bindings)
- [Directory Navigation](#-directory-navigation)
- [Autosuggestions](#-autosuggestions)
- [Aliases](#-aliases)
- [ZSH Options](#-zsh-options)

---

### Maintenance
```bash
# Rebuild completions if needed
rm -f ~/.zcompdump && exec zsh

# Update Oh My Zsh manually
omz update

# Re-initialize pyenv (if using Python)
eval "$(pyenv init - zsh)"
```

---

## 📚 History Features

### Configuration
- **50,000** entries stored in history
- History file: `~/.zsh_history`
- Shared across all terminal sessions

### Behavior
- `EXTENDED_HISTORY` - Timestamps recorded for each command
- `HIST_IGNORE_ALL_DUPS` - Automatically removes duplicate entries
- `HIST_FIND_NO_DUPS` - Search results show unique commands only
- `HIST_IGNORE_SPACE` - Commands starting with space aren't saved
- `HIST_REDUCE_BLANKS` - Extra whitespace removed before saving
- `SHARE_HISTORY` - Real-time history sharing between sessions
- `HIST_VERIFY` - Shows command before executing from history

### Usage
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

### Visual Feedback
- **Green** - Available completions descriptions
- **Yellow** - Spelling corrections with error count
- **Purple** - System messages
- **Red** - No matches found warnings

### Features
- **Case-insensitive matching** - `cd documents` matches `Documents`
- **Smart matching** - Partial matches from anywhere in the name
- **Menu selection** - Navigate completions with arrow keys
- **Cache enabled** - Faster repeated completions

### Navigation
```bash
Tab                 # Trigger completion
Tab Tab             # Show all options
↑/↓/←/→             # Navigate completion menu
Enter               # Accept selection
```

### Advanced Completion
- Corrections offered for typos (with error count)
- Completion from both ends of word
- Extensions and approximate matching
- Color-coded file types (from LS_COLORS)

---

## ⌨️ Key Bindings

### History Search
| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl+R` | fzf-history-widget | **Interactive fuzzy search** - shows scrollable list of all history |
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

### Auto-CD
```bash
# Just type the directory name - no 'cd' needed!
~/Documents         # Automatically changes to Documents directory
..                  # Up one level
```

### Directory Stack
The system maintains a stack of visited directories:

```bash
# View directory stack
d                   # Shows numbered list of recent directories

# Jump to directory by number
1                   # Jump to 1st directory in stack
2                   # Jump to 2nd directory in stack
... up to 9

# Built-in navigation
cd -                # Toggle between last two directories
pushd <dir>         # Push directory onto stack
popd                # Pop directory from stack
```

### Quick Navigation Aliases
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

Powered by `zsh-autosuggestions` plugin via Homebrew.

### How It Works
As you type, suggestions appear in **gray text** based on:
1. **Command history** (primary)
2. **Completion system** (secondary)

### Accepting Suggestions
```bash
→                   # Accept entire suggestion
Ctrl+→              # Accept one word at a time
Esc                 # Clear suggestion
```

### Visual Style
- Suggestions appear in dim gray (`fg=8`)
- Doesn't interfere with normal typing
- Updates in real-time as you type

---

## 🔧 Aliases

### Directory Navigation
```bash
..                  # cd ..
...                 # cd ../..
....                # cd ../../..
d                   # Show directory stack (dirs -v)
1-9                 # Jump to directory number in stack
```

### File Operations
```bash
ls                  # ls with colors (-G)
ll                  # Detailed list (ls -lah)
la                  # All files (ls -A)
l                   # Compact format (ls -CF)
mkdir               # Always create parent dirs (-p)
```

### System Info
```bash
h                   # View history
df                  # Disk space in human-readable format
du                  # Disk usage in human-readable format
free                # Memory usage on macOS
```

### Search
```bash
grep                # With color output
fgrep               # Fast grep with color
egrep               # Extended grep with color
```

### Git
Most git aliases provided by **Oh My Zsh git plugin**:
- `ga` - git add
- `gc` - git commit --verbose
- `gd` - git diff
- `gp` - git push
- `gst` - git status
- `glog` - git log --oneline --decorate --graph
- `gs` - git status (custom alias)

**Note:** The git plugin provides 100+ aliases. Run `alias | grep git` to see all.

### Python
```bash
python              # Points to python3
```

---

## ⚙️ ZSH Options

### Globbing & Pattern Matching
- `EXTENDED_GLOB` - Advanced pattern matching (`**`, `^`, `~`)
- `GLOB_DOTS` - Include hidden files in glob results
- `NO_CASE_GLOB` - Case-insensitive filename matching
- `NUMERIC_GLOB_SORT` - Sort files numerically (file1, file2, file10)

### Input/Output
- `CORRECT` - Suggests corrections for mistyped commands
- `INTERACTIVE_COMMENTS` - Allow `#` comments in command line
- `RC_QUOTES` - Use `''` to represent a literal single quote

### Job Control
- `AUTO_CONTINUE` - Automatically continue stopped jobs when disowning
- `LONG_LIST_JOBS` - Detailed job information
- `NOTIFY` - Immediate notification when background job finishes

### Completion Behavior
- `COMPLETE_IN_WORD` - Completion works from cursor position
- `ALWAYS_TO_END` - Cursor moves to end after completion
- `LIST_PACKED` - Compact completion list display

---

## 🎨 Plugins

### Oh My Zsh Plugins
- **git** - Git aliases and prompt integration
- **docker** - Docker command completions
- **python** - Python development helpers
- **pip** - pip command completions
- **npm** - npm command completions
- **gh** - GitHub CLI integration

### External Plugins
- **fzf** - Fuzzy finder for history, files, and more (Ctrl+R for history search)
- **zsh-autosuggestions** - Fish-like command suggestions
- **oh-my-posh** - Customizable prompt theme

---

## 🚀 Tips & Tricks

### History Power User
```bash
# Fuzzy search with fzf (best method!)
Ctrl+R              # Opens interactive list - type to filter anything
                    # Navigate with ↑/↓, select with Enter

# Prefix search with arrows
git log             # Type partial command
↑                   # Find last matching command starting with "git log"
Enter               # Execute it

# Don't save sensitive commands
 echo $PASSWORD     # Leading space = not saved to history
```

### Tab Completion Ninja
```bash
cd ~/D<Tab>         # Completes to ~/Documents or ~/Downloads
git co<Tab><Tab>    # Shows: commit, config, checkout, etc.
```

### Directory Stack Master
```bash
cd ~/project1       # Work in project 1
cd ~/project2       # Work in project 2
d                   # See stack
1                   # Instantly back to project 1
2                   # Back to project 2
```

### Case-Insensitive Everything
```bash
cd documents        # Works even if it's 'Documents'
ls *.PDF            # Matches .pdf files too
```

---

## 📖 Further Reading

- [Zsh Documentation](https://zsh.sourceforge.io/Doc/)
- [Oh My Zsh](https://ohmyz.sh/)
- [zsh-autosuggestions](https://github.com/zsh-users/zsh-autosuggestions)
- [oh-my-posh](https://ohmyposh.dev/)

---

**Last Updated:** February 2026
