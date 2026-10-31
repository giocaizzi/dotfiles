# chezmoi-dots — design

`chezmoi-dots` is a [chezmoi plugin](https://www.chezmoi.io/reference/plugins/): `chezmoi dots …`
runs it (the registry alias `dots` is `chezmoi dots`). It replaces the twin POSIX/PowerShell
`dots` functions, the `.chezmoitemplates/dots-edit/` helpers and the `scripts/*-lint.sh`
checks with one tested Go binary.

## As built

Where the code differs from the first design below, the code wins:

- Flags use the standard library `flag` (no cobra); a hand-written dispatcher in `internal/cli`.
- `live` polls the edited file's mtime and hash (no fsnotify); the side pane opens in tmux, or in Windows Terminal on Windows.
- The picker runs fzf when installed, else a numbered prompt (no fuzzy-finder library).
- The composition tree reads template edges with patterns (`internal/tree`), not the template AST: exact for how this repository writes templates, and a branch is listed whichever role takes it.
- TOML is read through `chezmoi data` (no TOML library); the registry rewrite in `pkgreg.SetRelease` is a line-preserving text edit so comments and order survive.
- End-to-end tests call `cli.Run` against a real chezmoi and a fixture source (`internal/cli/e2e_test.go`), not `testscript`.
- `dotsRender`: an override-data flag set on every read-only render; `.chezmoiexternal.toml.tmpl` renders empty under it, because chezmoi fetches every external while reading the source state, even for a `dump` that excludes them.
- `audit aliases` (port of the old `aliases-audit.sh`) and the release version (`-ldflags -X main.buildVersion`, set by the release workflow) were added.

## Goals and rules

1. **Never on the apply path.** Templates, `modify_` files, hooks and scripts never call the
   plugin: it is installed by an external *during* apply, so a fresh machine has none.
   It is a tool for editing and linting the source, nothing else.
2. **chezmoi is the engine.** The plugin never re-implements chezmoi: it talks to it through
   its machine-readable commands (`data`, `dump`, `managed`, `cat`, `source-path`,
   `target-path`, `status`, `diff`, `apply`, `encrypt`, `decrypt`, `execute-template`). chezmoi's
   Go packages are `internal/` and cannot be imported; the CLI is the stable contract.
3. **One source of truth per fact.** Nothing the plugin needs is copied by hand from another
   file; what used to be (the configuration matrix) is derived (see *Matrix*).
4. **Pure core, thin edges.** Decisions live in packages with no I/O; processes, files,
   terminals and editors sit behind small interfaces, so every decision is unit-tested
   without chezmoi installed.
5. **One behaviour on every OS.** No shell-specific twins: checks are argv lists, the live
   preview is the same binary, paths go through `path/filepath`.

## Environment

chezmoi sets up its environment before it runs a plugin (`internal/cmd/cmd.go`), so the
plugin reads, never guesses:

| Variable | Use |
|---|---|
| `CHEZMOI_SOURCE_DIR` | source-state root (`<repo>/home` with `.chezmoiroot`) |
| `CHEZMOI_WORKING_TREE` | repo root (git, `tools/`, `Makefile`) |
| `CHEZMOI_CONFIG_FILE` | passed back as `--config` so every child call sees the same config |
| `CHEZMOI_EXECUTABLE` | the chezmoi binary to call back (same version as the caller) |
| `CHEZMOI_OS`, `CHEZMOI_ARCH` | this machine |

Run directly (tests, `go run`), the plugin falls back to `chezmoi` on `PATH` and asks it
(`chezmoi source-path`, `chezmoi dump-config`).

## Layout

```
tools/dots/
  go.mod                       module github.com/giocaizzi/dotfiles/tools/dots
  cmd/chezmoi-dots/main.go     composition root: build adapters, hand them to cli, exit code
  internal/
    cli/       cobra commands; flag parsing, usage, error → exit code. No logic.
    app/       use cases (Edit, View, List, Lint, Audit) and the ports they need (ports.go)
    registry/  edit registry: model, decoding, validation                       (pure)
    aliases/   alias registry: model, validation, rendering checks             (pure)
    harness/   harness plugin list: model, validation                          (pure)
    matrix/    Config{OS, Role}, the matrix, override data                     (pure)
    tree/      composition tree from template ASTs                             (pure, fs.FS)
    state/     target state, change sets, report grouping                      (pure)
    finding/   lint findings: File, Line, Rule, Message; text and GitHub output (pure)
    chezmoi/   adapter: typed client over a Runner (os/exec)
    term/      styles, tree/diff/report printers, picker, prompts              (io.Writer)
    editor/    $EDITOR / $VISUAL launcher, file watcher (fsnotify)
    secure/    private temp dirs and files, atomic replace
```

Packages are named after what they hold, not after layers; there is no `utils`.

**Dependency rule** (enforced by `depguard` in `.golangci.yml`):

```
cmd → cli → app → {registry, aliases, harness, matrix, tree, state, finding}
cmd → {chezmoi, term, editor, secure}   (adapters, wired into app ports)
pure packages import only the standard library and each other, acyclically
```

`app` never imports an adapter; adapters never import `app` (they satisfy its interfaces
structurally).

## Ports (defined in `app`, by the consumer)

```go
type Chezmoi interface {
	Data(ctx context.Context, o matrix.Overrides) (map[string]any, error)
	Dump(ctx context.Context, o matrix.Overrides) (state.TargetState, error)
	Managed(ctx context.Context, o matrix.Overrides) (map[string]string, error) // target → source
	Cat(ctx context.Context, target string, o matrix.Overrides) ([]byte, error)
	SourcePath(ctx context.Context, target string, o matrix.Overrides) (string, error)
	Status(ctx context.Context, targets []string) ([]state.Status, error)
	Diff(ctx context.Context, targets []string, w io.Writer) error
	Apply(ctx context.Context, targets []string) error
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
	InitData(ctx context.Context, answers map[string]any) (map[string]any, error)
	WithSource(dir string) Chezmoi // same client over another source dir (live preview)
}

type Editor   interface { Edit(ctx context.Context, path string) error }
type Picker   interface { Pick(label string, items []Item) (Item, error) }
type Prompter interface { Confirm(question string, def bool) (bool, error) }
type Runner   interface { Run(ctx context.Context, argv []string, dir string) error } // checks
type Git      interface { Status(ctx context.Context) ([]string, error) }
```

The `chezmoi` adapter is itself built over `exec.Runner` (`argv → stdout, stderr, err`),
so its argument building and JSON decoding are unit-tested against recorded output.

## Domain

### Edit registry (`registry`)

`.chezmoidata/edit.toml`, read through `chezmoi data` (so encrypted and private data merge
exactly as chezmoi merges them). **Changed:** `check` becomes an argv list with a `{file}`
placeholder, run without a shell; the `powershell` twin field goes away.

```toml
[edit."ssh/config"]
desc   = "SSH client config (encrypted)"
source = "private_dot_ssh/encrypted_private_config.age"
check  = ["ssh", "-G", "-F", "{file}", "localhost"]
```

Validation returns `[]finding.Finding`; unknown keys, missing `desc`, `source` xor `local`,
missing files, and *every `.chezmoidata` file and encrypted file has an entry*.

### Matrix (`matrix`)

A configuration is `{OS ∈ darwin|linux|windows, Role ∈ personal|work|light}`. **Changed:**
its data is no longer a hand-copied template (`dots-edit/matrix`) kept in sync by a lint.
It is produced by rendering the real `.chezmoi.toml.tmpl` (`chezmoi execute-template --init`)
under a throwaway config seeded with the role's answers: every `prompt*Once` returns the
seeded value, so the result is exactly the `[data]` that `chezmoi init` writes for that role.
Seeds: `light`/`work` from the role; `harnesses` = every choice the template offers (read
from its `promptMultichoiceOnce "harnesses"` call in the parsed template); identity answers
this machine lacks become `<name>` placeholders; `bwsProjectId` empty (no secrets, no API
rate limits). The OS is applied as `--override-data {"chezmoi":{"os":…}}`.

### Composition tree (`tree`)

**Changed:** built from the template AST (`text/template/parse` with `SkipFuncCheck`), not
regexes, and honouring `chezmoi:template:left-delimiter`/`right-delimiter` directives.
Edges: `template`/`includeTemplate` → `.chezmoitemplates/<name>`; `include "<path>"`;
`include (printf "<pattern>" …)` → globbed; `symlink_` → its `.chezmoi.sourceDir` target;
`data` → the `.chezmoidata` file defining a top-level key read as `.key`, `$.key` or
`get . "key"`. Input is an `fs.FS` rooted at the source dir, so tests use `fstest.MapFS`.

### Target state and reports (`state`)

`TargetState` is decoded from `chezmoi dump --format json --exclude dirs,externals`.
`Changes(before, after)` returns the targets whose entry differs; `Report` groups the
configurations whose change sets are equal ("same as linux/personal") and classifies render
failures as *new* (fails the edit) or *pre-existing* (reported only).

## Use cases (`app`)

**Edit** is an explicit sequence of steps, each a method returning an error; cleanup is
`defer`red; `signal.NotifyContext` turns Ctrl-C into context cancellation:

1. resolve what to edit (entry → source; local folder file; source path; target → pick a
   node of its tree)
2. snapshot: dump this machine and every configuration (bounded `errgroup`)
3. open: plain file in place, or decrypt into a `secure` temp file
4. edit (with the live preview when available)
5. re-encrypt only when changed, atomic replace (armored, as chezmoi writes it)
6. run the entry's check (argv, cwd = working tree); failure → keep the edit, apply nothing
7. snapshot again, report per configuration; a newly broken configuration → apply nothing
8. this machine: `status` → `diff` (first target only for `diff = "first"`) → confirm → `apply`
9. print `note`, list changed scripts (they run on the next full apply), `git status` hint

**View**, **List**, **Lint** (`registry`, `aliases`, `harness`, `templates`) and **Audit**
(live aliases not in the registry) are separate use cases over the same ports.

### Live preview

**Changed:** no `live.sh`/`live.ps1`, no `sleep 1` polling, no full-source `tar` copy per
edit. The edit opens a pane (tmux `split-window`, or Windows Terminal `wt sp`) running the
same binary's hidden `chezmoi dots live --session <dir>`; it watches the edited file with
fsnotify (debounced), writes it into a shadow source (a copy made once, hard links where
the OS allows) and prints this machine's diff. Without tmux/WT the edit simply runs
without the pane.

## CLI

```
chezmoi dots [edit] [<what>]                 edit (picker when <what> is empty)
chezmoi dots view <what> [--as os[,role] | --matrix]
chezmoi dots list
chezmoi dots lint [registry|aliases|harness|templates]...   (all when none)
chezmoi dots audit aliases
chezmoi dots version
```

Exit codes: 0 ok, 1 failure, 2 usage, 130 interrupted. Colour: the basic ANSI colours
(the terminal's Catppuccin theme maps them), off for non-TTY, `NO_COLOR`, `TERM=dumb`.
Lint output is `file:line: rule: message`, and GitHub `::error file=…,line=…::` annotations
under `GITHUB_ACTIONS`.

## Testing

| Level | What | How |
|---|---|---|
| Unit | pure packages | table-driven tests, `testdata/` fixtures, golden files (`-update`) |
| Use case | `app` | hand-written fakes for each port (no mocking framework) |
| Adapter | `chezmoi` | argv/JSON tests over a fake `exec.Runner`; a **contract suite** run against both the fake and the real chezmoi, so the fake cannot drift |
| End to end | the binary | [`testscript`](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript) (what chezmoi itself uses): `.txtar` scripts with a mini source tree, a throwaway age key and a scripted editor; CI on Linux, macOS and Windows |

`go test ./...` needs no chezmoi; the contract and testscript suites skip without it
(`CHEZMOI_BIN` or `chezmoi` on `PATH`), and CI always provides it.
Tooling: `gofmt`, `go vet`, `golangci-lint` (with `depguard`), `go test -race`.

## Dependencies

Standard library first. Then: `spf13/cobra` (chezmoi's own CLI library),
`fsnotify/fsnotify`, `pelletier/go-toml/v2` (line numbers for findings),
`ktr0731/go-fuzzyfinder` (picker), `golang.org/x/sync` (errgroup), `golang.org/x/term`,
`rogpeppe/go-internal` (tests only).

## Distribution

- A `release-dots` workflow on `dots-v*` tags cross-compiles darwin/linux/windows ×
  amd64/arm64 (`CGO_ENABLED=0`, `-trimpath`), publishes archives and `checksums.txt`.
- `.chezmoidata/dots.toml` pins `version` and the per-asset SHA-256; an `archive-file`
  external in `.chezmoiexternal.toml.tmpl` installs `~/.local/bin/chezmoi-dots[.exe]` on
  full machines with `checksum.sha256` (chezmoi verifies it).
- Development: `go run ./cmd/chezmoi-dots …` from `tools/dots`, or `make dots` to build
  into `tools/dots/bin/` and put it first on `PATH` for the session.
- `.chezmoiversion` pins the minimum chezmoi version the plugin's commands need.

## Migration

Each slice deletes what it replaces, with its tests, in the same PR:

1. Module, CI (`go test`, lint, the three-OS testscript job), release workflow, external.
2. `lint` (registry, aliases, harness, templates) → delete `scripts/*-lint.*`,
   `edit-lint.sh`, the Makefile shell loops; Makefile targets call the plugin.
3. `list`, `view` → delete `dots-edit/{entries,field,tree,tree-of,feeds,matrix,usage}`.
4. `edit` (+ local entries, encrypted files) → delete the sh/ps1 `dots` functions and
   `scripts/dots-edit-test.sh`; `dots` becomes a registry alias for `chezmoi dots`.
5. live preview → delete `dots-edit/live.{sh,ps1}`.
6. `audit aliases` → delete `scripts/aliases-audit.sh`.
