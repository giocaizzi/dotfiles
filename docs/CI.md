# CI and `make check`

[GitHub standard](https://github.com/giocaizzi/.github/blob/main/docs/standard.md): one `ci.yml`, required status check `check`, CI runs `make check`.

## Locally

```bash
make help    # list targets
make check   # lint + test; run before pushing
```

Needs `go`, `chezmoi`, `shellcheck`, `jq`, `bash`, `zsh` (all in the chezmoi install list on macOS). `pwsh` is optional locally (the PowerShell alias parse is skipped without it) and present on CI runners. `make` runs `chezmoi dots lint` ([giocaizzi/chezmoi-dots](https://github.com/giocaizzi/chezmoi-dots)) and `agents check` ([giocaizzi/agents](https://github.com/giocaizzi/agents)) against this checkout.
On Windows, `make` uses Git for Windows' POSIX shell; zsh syntax checks run when
zsh is installed and are otherwise skipped because zsh is not a Windows dependency.
ShellCheck also ignores working-tree CRLF diagnostics on Windows; tracked shell
files remain LF through `.gitattributes`.

| Target | Does |
|---|---|
| `tools` | not in `check`: installs the pinned `chezmoi`, `chezmoi-dots` and `agents` (linux/amd64) into `TOOLS_DIR` (default `~/.local/bin`), as CI does (below) |
| `lint` | `shellcheck` on the bash rc files, the global git hooks and `tests/git-hooks.sh`; `jq empty` on tracked JSON; `chezmoi dots lint` (below); `agents check` and `agents prompt check` on the rendered agents config (layers listed by `agents config show`; encrypted ones are empty stand-ins) |
| `test-syntax` | `bash -n` on the same shell files, `zsh -n` on `dot_zshrc`, `dot_zprofile` |
| `test-git-hooks` | `tests/git-hooks.sh`: the global git hooks against throwaway repos, with a stub `gitleaks` |
| `test` | `test-syntax` + `test-git-hooks` |
| `aliases-audit` | not in `check`: lists aliases live in this machine's zsh and bash that the registry does not account for |
| `check` | `lint` + `test` |
| `check-fast` | `lint` + `test-syntax`: the agent stop hook (guard plugin) runs it instead of `check`; run `make check` before pushing |

`chezmoi dots lint` runs these checks (name them to run a subset):

| Check | Verifies |
|---|---|
| `registry` | the edit registry: every entry has a `desc`, an existing `source` or a home-relative `local` folder, a `check` that is a command list, and only known keys; every `.chezmoidata` file and every encrypted file has an entry |
| `aliases` | the alias registry (the encrypted part too where the age key exists): known keys, every alias names its shells, a one-shell alias says why, known OS and role values |
| `packages` | the package and release tables, including a sha256 for every release target and a well-formed `versionCmd` |
| `matrix` | the configuration matrix built from `.chezmoi.toml.tmpl` lists the harnesses it offers |
| `templates` | renders every script template, `aliases.sh`, `omz-aliases.zsh`, and `aliases.ps1` per OS × role × harness selection; `shellcheck` the shell output, `zsh -n`, parse `aliases.ps1` with `pwsh` (and Windows PowerShell 5.1 on Windows), `jq empty` the JSON (each where installed) |

Skipped by the JSON check: JSON containing `{{` (rendered by `templates`) and `dot_config/Code/User/settings.json` (JSONC). No `shfmt`: existing scripts mix tab and space indents.

## CI (`.github/workflows/ci.yml`)

On every PR and push to `main`:

1. `gitleaks` scans the git history for secrets.
2. `check` (the required check) runs `make tools`, then `make check`.

### Pinned tools

`make tools` installs exactly the pinned versions, so CI never follows `latest`:

- `agents` and `chezmoi-dots`: version, URL and linux/amd64 sha256 from `[release.*]` in `home/.chezmoidata/packages.toml` (the same rows machines install; `chezmoi dots bump` moves them). The download is checked against the sha256 and fails on a mismatch.
- `chezmoi`: machines get it from their bootstrap, so the registry has no row. Its CI version is `[tools] chezmoi` in `.github/tools.toml`, installed with the official installer (which verifies its own checksum). Bump it by hand.

`chezmoi dots` is run as `chezmoi-dots` with `CHEZMOI_SOURCE_DIR` set (the `DOTS` variable in the `Makefile`): chezmoi's plugin dispatch ignores `--source` placed before `dots` and always uses the default source directory, so it cannot target a CI checkout or a worktree.

A red `check`: open the run, read the failing step, reproduce with `make check`, fix, push. A gitleaks hit: rotate the secret first, then remove it from the change (encrypted file or `bws`, see [OS.md](./OS.md)).

Other workflows: `bump-releases` opens the weekly PR that pins newer release binaries ([OS.md](./OS.md#packages)); it needs a `BUMP_TOKEN` secret for CI to run on its PR. Its checkout keeps no credential (`persist-credentials: false`); only the final step that pushes and opens the PR gets the token.
