# OS-Specific Configuration

This document describes which files are managed by chezmoi on each operating system.

## ✅ Configuration Summary

### 🍎 **macOS**

- Bash: `.bash_profile`, `.bashrc`, `.profile`
- Zsh: `.zshrc`, `.zprofile` ✨ (default shell on macOS)
- VSCode: `Library/Application Support/Code/User/settings.json`
- Scripts: `.chezmoiscripts/run_env.sh`, `.chezmoiscripts/run_onchange_install-pkgs.sh`
- Shared: `.gitconfig`, `.vimrc`, `.mytheme.omp.json`, `.python-version`, `.claude/`

**Files ignored:** AppData, OneDrive, PowerShell scripts

---

### 🐧 **Linux (Raspberry Pi)**

- Bash: `.bash_profile`, `.bashrc`, `.profile` ✨ (default shell on Linux)
- Scripts: `.chezmoiscripts/run_env.sh`, `.chezmoiscripts/run_onchange_install-pkgs.sh`
- Shared: `.gitconfig`, `.vimrc`, `.mytheme.omp.json`, `.python-version`, `.claude/`

**Files ignored:** `.zshrc`, `.zprofile`, Library, AppData, OneDrive, PowerShell scripts

---

### 🪟 **Windows**

- Bash: `.bash_profile`, `.bashrc`, `.profile` (for Git Bash/WSL)
- VSCode: `AppData/Roaming/Code/User/settings.json`
- PowerShell: `OneDrive - Jakala SpA/Documenti/WindowsPowerShell/profile.ps1`
- Scripts: `.chezmoiscripts/run_env.ps1`
- Shared: `.gitconfig`, `.vimrc`, `.mytheme.omp.json`, `.python-version`, `.claude/`

**Files ignored:** `.zshrc`, `.zprofile`, Library, bash scripts

---

## 🎯 Testing Commands

Test the template for different operating systems:

```bash
# Test template for macOS
cat .chezmoiignore | chezmoi execute-template --override-data '{"chezmoi":{"os":"darwin"}}'

# Test template for Linux
cat .chezmoiignore | chezmoi execute-template --override-data '{"chezmoi":{"os":"linux"}}'

# Test template for Windows
cat .chezmoiignore | chezmoi execute-template --override-data '{"chezmoi":{"os":"windows"}}'
```
