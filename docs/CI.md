# CI and `make check`

[GitHub standard](https://github.com/giocaizzi/.github/blob/main/docs/standard.md): one `ci.yml`, required status check `check`, CI runs `make check`.

## Locally

```bash
make help    # list targets
make check   # lint + test; run before pushing
```

Needs `shellcheck`, `jq`, `bash`, `zsh` (all in the chezmoi install list on macOS).
On Windows, `make` uses Git for Windows' POSIX shell; zsh syntax checks run when
zsh is installed and are otherwise skipped because zsh is not a Windows dependency.
ShellCheck also ignores working-tree CRLF diagnostics on Windows; tracked shell
files remain LF through `.gitattributes`.

| Target | Does |
|---|---|
| `lint` | `shellcheck` on tracked `*.sh` and the bash rc files (`dot_bashrc`, `dot_bash_profile`, `dot_profile`, as bash); `jq empty` on tracked JSON |
| `test` | `bash -n` on the same shell files, `zsh -n` on `dot_zshrc`, `dot_zprofile` |
| `check` | `lint` + `test` |

Skipped: `*.tmpl` files and JSON containing `{{` (need chezmoi to render), `dot_config/Code/User/settings.json` (JSONC). No `shfmt`: existing scripts mix tab and space indents.

## CI (`.github/workflows/ci.yml`)

On every PR and push to `main`:

1. `gitleaks` scans the git history for secrets.
2. `check` (the required check) fails unless `gitleaks` succeeded, installs `zsh`, runs `make check`.

A red `check`: open the run, read the failing step, reproduce with `make check`, fix, push. A gitleaks hit: rotate the secret first, then remove it from the change (encrypted file or `bws`, see [OS.md](./OS.md)).

Dependabot (`.github/dependabot.yml`) bumps action versions weekly.
