# CI and `make check`

[GitHub standard](https://github.com/giocaizzi/.github/blob/main/docs/standard.md): one `ci.yml`, required status check `check`, CI runs `make check`.

## Locally

```bash
make help    # list targets
make check   # lint + test; run before pushing
```

Needs `chezmoi`, `shellcheck`, `jq`, `bash`, `zsh` (all in the chezmoi install list on macOS);
`pwsh` is optional locally (the PowerShell alias parse is skipped without it) and present on CI runners.
On Windows, `make` uses Git for Windows' POSIX shell; zsh syntax checks run when
zsh is installed and are otherwise skipped because zsh is not a Windows dependency.
ShellCheck also ignores working-tree CRLF diagnostics on Windows; tracked shell
files remain LF through `.gitattributes`.

| Target | Does |
|---|---|
| `lint` | `lint-tmpl` + `lint-aliases`, then `shellcheck` on tracked `*.sh` and the bash rc files (`dot_bashrc`, `dot_bash_profile`, `dot_profile`, as bash); `jq empty` on tracked JSON |
| `lint-tmpl` | renders `.chezmoiscripts/*.sh.tmpl` per OS × role × harness selection, `shellcheck` the output |
| `lint-aliases` | validates the alias registry (`scripts/aliases-lint.sh`, plus the encrypted part where the age key exists), renders `aliases.sh` per OS × role (`shellcheck` as bash), `omz-aliases.zsh` (`zsh -n`), `aliases.ps1` (parsed by `pwsh`) |
| `aliases-audit` | not in `check`: lists aliases live in this machine's zsh and bash that the registry does not account for |
| `test` | `bash -n` on the same shell files, `zsh -n` on `dot_zshrc`, `dot_zprofile` |
| `check` | `lint` + `test` |

Skipped: other `*.tmpl` files and JSON containing `{{` (need chezmoi to render), `dot_config/Code/User/settings.json` (JSONC). No `shfmt`: existing scripts mix tab and space indents.

## CI (`.github/workflows/ci.yml`)

On every PR and push to `main`:

1. `gitleaks` scans the git history for secrets.
2. `check` (the required check) fails unless `gitleaks` succeeded, installs `zsh`, runs `make check`.

A red `check`: open the run, read the failing step, reproduce with `make check`, fix, push. A gitleaks hit: rotate the secret first, then remove it from the change (encrypted file or `bws`, see [OS.md](./OS.md)).

Dependabot (`.github/dependabot.yml`) bumps action versions weekly.
