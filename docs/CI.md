# CI and `make check`

[GitHub standard](https://github.com/giocaizzi/.github/blob/main/docs/standard.md): one `ci.yml`, required status check `check`, CI runs `make check`.

## Locally

```bash
make help    # list targets
make check   # lint + test; run before pushing
```

Needs `go`, `chezmoi`, `shellcheck`, `jq`, `bash`, `zsh` (all in the chezmoi install list on macOS). `pwsh` is optional locally (the PowerShell alias parse is skipped without it) and present on CI runners. `make` builds the `chezmoi-dots` plugin ([OS.md](./OS.md#the-chezmoi-dots-plugin)) into `tools/dots/bin/` and runs it against this checkout; the plugin's own linters and tools are pinned Go modules run with `go tool`, so nothing else is installed.
On Windows, `make` uses Git for Windows' POSIX shell; zsh syntax checks run when
zsh is installed and are otherwise skipped because zsh is not a Windows dependency.
ShellCheck also ignores working-tree CRLF diagnostics on Windows; tracked shell
files remain LF through `.gitattributes`.

| Target | Does |
|---|---|
| `lint` | `shellcheck` on the bash rc files (`dot_bashrc`, `dot_bash_profile`, `dot_profile`, as bash); `jq empty` on tracked JSON; `chezmoi dots lint` (below); `golangci-lint` on the plugin for linux, darwin and windows |
| `dots` | builds the plugin into `tools/dots/bin/` |
| `test-syntax` | `bash -n` on the same shell files, `zsh -n` on `dot_zshrc`, `dot_zprofile` |
| `test` | `test-syntax`, then the plugin's tests: unit tests with fakes, and an end-to-end test of the edit flow against a real chezmoi, a throwaway source with dummy encrypted files, a throwaway age key and a temp destination (`go test -race -shuffle=on`) |
| `aliases-audit` | not in `check`: lists aliases live in this machine's zsh and bash that the registry does not account for |
| `check` | `lint` + `test` |
| `check-fast` | `lint` + `test-syntax`: the agent stop hook (guard plugin) runs it instead of `check`; run `make check` before pushing |

`chezmoi dots lint` runs these checks (name them to run a subset):

| Check | Verifies |
|---|---|
| `registry` | the edit registry: every entry has a `desc`, an existing `source` or a home-relative `local` folder, a `check` that is a command list, and only known keys; every `.chezmoidata` file and every encrypted file has an entry |
| `aliases` | the alias registry (the encrypted part too where the age key exists): known keys, every alias names its shells, a one-shell alias says why, known OS and role values |
| `packages` | the package and release tables, including a sha256 for every release target |
| `plugins` | `.chezmoidata/harness.toml` lists exactly the `plugins/` directories of `giocaizzi/harness`: main through `gh api` (CI passes `GH_TOKEN`), or a local checkout with `--harness-dir <path>`; every `harnessPersonalPlugins` entry is one of them |
| `matrix` | the configuration matrix built from `.chezmoi.toml.tmpl` lists the harnesses it offers |
| `templates` | renders every script template, `aliases.sh`, `omz-aliases.zsh`, `aliases.ps1` and the settings templates per OS × role × harness selection; `shellcheck` the shell output, `zsh -n`, parse `aliases.ps1` with `pwsh` (and Windows PowerShell 5.1 on Windows), `jq empty` the JSON (each where installed) |

Skipped by the JSON check: JSON containing `{{` (rendered by `templates`) and `dot_config/Code/User/settings.json` (JSONC). No `shfmt`: existing scripts mix tab and space indents.

## CI (`.github/workflows/ci.yml`)

On every PR and push to `main`:

1. `gitleaks` scans the git history for secrets.
2. `dots` (ubuntu, macOS, Windows) installs chezmoi and runs the plugin's `make check` (Windows: `go test -race -shuffle=on ./...`, since the runner has no make).
3. `dots-vuln` runs `govulncheck` on the plugin's dependencies.
4. `check` (the required check) fails unless the jobs above succeeded, installs `zsh` and chezmoi, runs `make check`.

A red `check`: open the run, read the failing step, reproduce with `make check`, fix, push. A red `dots` on one OS: read its log; the tests are cross-platform Go, so `go test -race ./...` in `tools/dots` on that OS reproduces it. A red `dots-vuln`: bump the dependency named in the report (`go get <module>@<fixed>` in `tools/dots`). A gitleaks hit: rotate the secret first, then remove it from the change (encrypted file or `bws`, see [OS.md](./OS.md)).

Other workflows: `release-dots` publishes the plugin on a `dots-v*` tag and `bump-releases` opens the weekly PR that pins newer release binaries ([OS.md](./OS.md#packages)); the latter needs a `BUMP_TOKEN` secret for CI to run on its PR.

Dependabot (`.github/dependabot.yml`) bumps action versions and the plugin's Go modules weekly. `golangci-lint` has its own modfile that Dependabot does not read: upgrade it with `go get -tool -modfile=golangci-lint.mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest` in `tools/dots`, in its own PR.
