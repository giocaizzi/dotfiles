# CI and `make check`

[GitHub standard](https://github.com/giocaizzi/.github/blob/main/docs/standard.md): one `ci.yml`, required status check `check`, CI runs `make check`.

## Locally

```bash
make help    # list targets
make check   # lint + test; run before pushing
```

Needs `go`, `chezmoi`, `shellcheck`, `jq`, `bash`, `zsh` (all in the chezmoi install list on macOS). `pwsh` is optional locally (PowerShell checks are skipped without it) and present on CI runners. `make` runs `chezmoi dots lint` ([giocaizzi/chezmoi-dots](https://github.com/giocaizzi/chezmoi-dots)) and `agents check` ([giocaizzi/agents](https://github.com/giocaizzi/agents)) against this checkout.
On Windows, `make` uses Git for Windows' POSIX shell; zsh syntax checks run when
zsh is installed and are otherwise skipped because zsh is not a Windows dependency.
ShellCheck also ignores working-tree CRLF diagnostics on Windows; tracked shell
files remain LF through `.gitattributes`.

| Target | Does |
|---|---|
| `tools` | not in `check`: installs the pinned `chezmoi`, `chezmoi-dots` and `agents` (linux/amd64) into `TOOLS_DIR` (default `.tools/bin`, gitignored; CI passes `TOOLS_DIR="$HOME/.local/bin"`) |
| `lint` | `shellcheck` on the bash rc files, the global git hooks and the `tests/` scripts; `jq empty` on tracked JSON; `chezmoi dots lint` (below); `agents check` and `agents prompt check` on the rendered agents config (layers listed by `agents config show`; encrypted ones are empty stand-ins) |
| `test-syntax` | `bash -n` on the same shell files, `zsh -n` on `dot_zshrc`, `dot_zprofile` |
| `test-git-hooks` | `tests/git-hooks.sh`: the global git hooks against throwaway repos, with a stub `gitleaks` |
| `test-secrets-gate` | `tests/secrets-gate.sh`: when `.chezmoiignore` lets chezmoi manage the env secrets files (project ID, `bws`, `chezmoi-dots` version), checked read-only with `chezmoi managed` and stub binaries; a case marked `xfail` is a known gap and fails the run once it passes (skipped on Windows) |
| `test-pwsh-profile` | `tests/pwsh-profile.ps1`: render the Windows profile-link script and test creation, idempotence, owned duplicate cleanup and preservation in an isolated Documents fixture (Windows only; skips on Unix) |
| `test` | `test-syntax` + `test-git-hooks` + `test-secrets-gate` + `test-pwsh-profile` |
| `aliases-audit` | not in `check`: lists aliases live in this machine's zsh and bash that the registry does not account for |
| `check` | `lint` + `test` |
| `check-fast` | `lint` + `test-syntax`: the agent stop hook (guard plugin) runs it instead of `check`; run `make check` before pushing |

`chezmoi dots lint` runs these checks (name them to run a subset):

| Check | Verifies |
|---|---|
| `registry` | the edit registry: every entry has a `desc`, an existing `source` or a home-relative `local` folder, a `check` that is a command list, and only known keys; every `.chezmoidata` file and every encrypted file has an entry |
| `aliases` | the alias registry (the encrypted part too where the age key exists): known keys, every alias names its shells, a one-shell alias says why, known OS and role values |
| `packages` | the package and release tables, including a sha256 for every release target and a well-formed `versionCmd` |
| `secrets` | `.chezmoidata/secrets.toml`: known keys and kinds, a project source for each secret manager ([OS.md](./OS.md#bitwarden)) |
| `matrix` | `.chezmoidata/matrix.toml`: the declared profiles, seed answers and variants are well formed and name known harnesses ([OS.md](./OS.md#the-configuration-matrix)) |
| `templates` | renders every script template, `aliases.sh`, `omz-aliases.zsh`, and `aliases.ps1` for every OS × profile (plus the matrix variants); `shellcheck` the shell output, `zsh -n`, parse `aliases.ps1` with `pwsh` (and Windows PowerShell 5.1 on Windows), `jq empty` the JSON (each where installed) |

Skipped by the JSON check: JSON containing `{{` (rendered by `templates`) and `dot_config/Code/User/settings.json` (JSONC). No `shfmt`: existing scripts mix tab and space indents.

## CI (`.github/workflows/ci.yml`)

On every PR and push to `main`:

1. `gitleaks` scans the git history for secrets.
2. `check` (the required check) runs `make tools`, then `make check`.

### Pinned tools

`make tools` installs exactly the pinned versions, so CI never follows `latest`:

- `agents` and `chezmoi-dots`: version, URL and linux/amd64 sha256 from `[release.*]` in `home/.chezmoidata/packages.toml` (the same rows machines install; `chezmoi dots pkg bump` moves them). The download is checked against the sha256 and fails on a mismatch.
- `chezmoi`: machines get it from their bootstrap, so the registry has no row. Its CI version is `[tools] chezmoi` in `.github/tools.toml`, installed with the official installer (which verifies its own checksum). Bump it by hand.

`chezmoi dots` is run as `chezmoi-dots` with `CHEZMOI_SOURCE_DIR` set (the `DOTS` variable in the `Makefile`): chezmoi's plugin dispatch ignores `--source` placed before `dots` and always uses the default source directory, so it cannot target a CI checkout or a worktree.

A red `check`: open the run, read the failing step, reproduce with `make check`, fix, push. A gitleaks hit: rotate the secret first, then remove it from the change (encrypted file or `bws`, see [OS.md](./OS.md)).

### Weekly pin bump

`.github/workflows/bump-releases.yml`: every Monday (and on `workflow_dispatch`) it runs `chezmoi dots pkg bump` on `home/` and opens or updates a PR on `chore/bump-releases` titled `chore(deps): bump pinned releases` with the moved release versions, archive commits and checksums ([OS.md](./OS.md#packages)). The job is only the checkout (`persist-credentials: false`) and the [`pkg-bump` action](https://github.com/giocaizzi/chezmoi-dots/blob/main/docs/bump-action.md) of chezmoi-dots, which installs its own checksum-verified `chezmoi` and `chezmoi-dots`, runs the bump and opens the PR; only its last step gets the token. Inputs, token rules and troubleshooting are in the action's docs.

- **Token:** a `BUMP_TOKEN` secret (contents and pull-requests write on this repo) lets CI run on the PR; without it the `GITHUB_TOKEN` fallback opens the PR but its checks must be started by hand.
- **Versions:** the action does not use `make tools`. It runs the `chezmoi-dots` named by its `version` input and the `chezmoi` pinned inside the action release, which may differ from `[tools] chezmoi` above; `pkg bump` only reads the source with `chezmoi data`, so the difference does not matter there. `make tools` and its pins stay the CI story for `check`.
- **Upgrading the action:** the `uses:` ref is pinned by commit SHA with the release as a comment, and `version:` must equal that release. Dependabot (`.github/dependabot.yml`, `github-actions`) may move the ref and its comment in its weekly `ci(deps)` PR, but not `version:`: edit it in the same PR before merging. The `[release.chezmoi-dots]` pin that machines install moves separately, through this workflow's own PR.

A failed run: open it, read the failing step, and match its message in the action's troubleshooting table.
