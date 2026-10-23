# CI and `make check`

[GitHub standard](https://github.com/giocaizzi/.github/blob/main/docs/standard.md): one `ci.yml`, required status check `check`, CI runs `make check`.

## Locally

```bash
make help    # list targets
make check   # lint + test; run before pushing
```

Needs `chezmoi`, `shellcheck`, `jq`, `bash`, `zsh` (all in the chezmoi install list on macOS);
`pwsh` is optional locally (the PowerShell alias parse and edit command tests are skipped without it) and present on CI runners; on Windows, Windows PowerShell 5.1 is parsed and tested too. The edit command tests also need `ssh`.
On Windows, `make` uses Git for Windows' POSIX shell; zsh syntax checks run when
zsh is installed and are otherwise skipped because zsh is not a Windows dependency.
ShellCheck also ignores working-tree CRLF diagnostics on Windows; tracked shell
files remain LF through `.gitattributes`.

| Target | Does |
|---|---|
| `lint` | `lint-tmpl` + `lint-aliases` + `lint-plugins` + `lint-edit`, then `shellcheck` on tracked `*.sh` and the bash rc files (`dot_bashrc`, `dot_bash_profile`, `dot_profile`, as bash); `jq empty` on tracked JSON and on the renders of `.chezmoitemplates/{vscode-mcp,claude-settings,copilot-settings}.json` (personal and work) |
| `lint-tmpl` | renders `.chezmoiscripts/*.sh.tmpl` per OS × role × harness selection, `shellcheck` the output |
| `lint-plugins` | `.chezmoidata/harness.toml` lists exactly the `plugins/` directories of `giocaizzi/harness` (`scripts/harness-plugins-lint.sh`): main through `gh api` (CI passes `GH_TOKEN`), or a local checkout with `HARNESS_DIR=<path>`; every `harnessPersonalPlugins` entry is one of them |
| `lint-aliases` | validates the alias registry (`scripts/aliases-lint.sh`, plus the encrypted part where the age key exists), renders `aliases.sh` per OS × role (`shellcheck` as bash), `omz-aliases.zsh` (`zsh -n`), `aliases.ps1` (parsed by `pwsh` and, on Windows, Windows PowerShell 5.1) |
| `lint-edit` | validates the edit registry (`scripts/edit-lint.sh`): every entry has a `desc`, an existing `source` and only known keys; every `.chezmoidata` file and every encrypted file has an entry; the `dots-edit` matrix lists the same harnesses and `[data]` answers as `.chezmoi.toml.tmpl` |
| `aliases-audit` | not in `check`: lists aliases live in this machine's zsh and bash that the registry does not account for |
| `test-syntax` | `bash -n` on the same shell files, `zsh -n` on `dot_zshrc`, `dot_zprofile` |
| `test` | `test-syntax`, then `scripts/dots-edit-test.sh`: `dots-edit` (edit, per-configuration report, render failure, target and source lookup, `--view`, `--as`, `--matrix`, the live pane through a fake tmux or Windows Terminal) and `agents-edit`, `aliases-edit`, `ssh-edit` in bash, zsh and pwsh, on Windows (Git Bash) in bash, Windows PowerShell 5.1 and pwsh (each where installed) against a throwaway source copy with dummy encrypted files, a throwaway age key and a temp destination |
| `check` | `lint` + `test` |
| `check-fast` | `lint` + `test-syntax`: the agent stop hook (guard plugin) runs it instead of `check`, since the edit command tests take minutes per shell (about 8 on Windows); run `make check` before pushing |

Skipped: other `*.tmpl` files and JSON containing `{{` (need chezmoi to render), `dot_config/Code/User/settings.json` (JSONC). No `shfmt`: existing scripts mix tab and space indents.

## CI (`.github/workflows/ci.yml`)

On every PR and push to `main`:

1. `gitleaks` scans the git history for secrets.
2. `windows` (`windows-latest`) installs chezmoi, `make` and `shellcheck` (Chocolatey) and runs `scripts/dots-edit-test.sh` in Git Bash: the edit commands in bash, Windows PowerShell 5.1 and pwsh.
3. `check` (the required check) fails unless `gitleaks` and `windows` succeeded, installs `zsh`, runs `make check`.

A red `check`: open the run, read the failing step, reproduce with `make check`, fix, push. A red `windows` has no local reproduction off Windows: read its log; on a Windows machine, `sh scripts/dots-edit-test.sh` from Git Bash. A gitleaks hit: rotate the secret first, then remove it from the change (encrypted file or `bws`, see [OS.md](./OS.md)).

Dependabot (`.github/dependabot.yml`) bumps action versions weekly.
