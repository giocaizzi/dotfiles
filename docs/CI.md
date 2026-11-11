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
| `lint` | `shellcheck` on the bash rc files; `jq empty` on tracked JSON; `chezmoi dots lint` (below); `agents check` on agents config |
| `test-syntax` | `bash -n` on the same shell files, `zsh -n` on `dot_zshrc`, `dot_zprofile` |
| `test` | `test-syntax` |
| `aliases-audit` | not in `check`: lists aliases live in this machine's zsh and bash that the registry does not account for |
| `check` | `lint` + `test` |
| `check-fast` | `lint` + `test-syntax`: the agent stop hook (guard plugin) runs it instead of `check`; run `make check` before pushing |

`chezmoi dots lint` runs these checks (name them to run a subset):

| Check | Verifies |
|---|---|
| `registry` | the edit registry: every entry has a `desc`, an existing `source` or a home-relative `local` folder, a `check` that is a command list, and only known keys; every `.chezmoidata` file and every encrypted file has an entry |
| `aliases` | the alias registry (the encrypted part too where the age key exists): known keys, every alias names its shells, a one-shell alias says why, known OS and role values |
| `packages` | the package and release tables, including a sha256 for every release target and a well-formed `versionCmd` |
| `plugins` | `.chezmoidata/harness.toml` lists exactly the `plugins/` directories of `giocaizzi/harness`: main through `gh api` (CI passes `GH_TOKEN`), or a local checkout with `--harness-dir <path>`; every `harnessPersonalPlugins` entry is one of them |
| `matrix` | the configuration matrix built from `.chezmoi.toml.tmpl` lists the harnesses it offers |
| `templates` | renders every script template, `aliases.sh`, `omz-aliases.zsh`, `aliases.ps1` and the settings templates per OS × role × harness selection; `shellcheck` the shell output, `zsh -n`, parse `aliases.ps1` with `pwsh` (and Windows PowerShell 5.1 on Windows), `jq empty` the JSON (each where installed) |

Skipped by the JSON check: JSON containing `{{` (rendered by `templates`) and `dot_config/Code/User/settings.json` (JSONC). No `shfmt`: existing scripts mix tab and space indents.

## CI (`.github/workflows/ci.yml`)

On every PR and push to `main`:

1. `gitleaks` scans the git history for secrets.
2. `check` (the required check) installs `chezmoi`, `chezmoi-dots` and `agents`, and runs `make check`.

A red `check`: open the run, read the failing step, reproduce with `make check`, fix, push. A gitleaks hit: rotate the secret first, then remove it from the change (encrypted file or `bws`, see [OS.md](./OS.md)).

Other workflows: `bump-releases` opens the weekly PR that pins newer release binaries ([OS.md](./OS.md#packages)); `models-sync` opens the weekly PR that refreshes the model catalog ([OS.md](./OS.md#model-tiers-agents-models)); both need a `BUMP_TOKEN` secret for CI to run on their PR. Their checkout keeps no credential (`persist-credentials: false`); only the final step that pushes and opens the PR gets the token.
