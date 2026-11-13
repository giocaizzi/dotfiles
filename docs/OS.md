# OS-Specific Configuration

How each managed file maps to its destination per OS. Canonical sources live under `dot_config/` in the source state (`home/`, [Source layout](#source-layout-chezmoiroot)); OS-specific destinations symlink to them via `{{ .chezmoi.sourceDir }}/...`.

## Destination matrix

| Canonical source (in repo)                          | macOS                                                       | Linux                                          | Windows                                                                            |
| ---------------------------------------------------- | ----------------------------------------------------------- | ---------------------------------------------- | ---------------------------------------------------------------------------------- |
| `dot_config/Code/User/{settings,keybindings}.json` | `~/Library/Application Support/Code/User/` (symlinks)      | `~/.config/Code/User/` (direct)                | `%APPDATA%\Code\User\` (symlinks)                                                |
| `dot_config/agents/instructions/` (layers) and `.chezmoitemplates/agents-config.toml` (the config) | `~/.config/agents/instructions/`, `~/.config/agents/config.toml` (rendered, owner-only); [`agents apply`](#agents-apply) then writes every harness file from them (`~/.claude`, `~/.copilot`, `~/.codex`, `~/.config/opencode`, `~/.pi`, `~/.gemini`, `~/ops`; [AI agent configuration](#ai-agent-configuration)) | same | same |
| `dot_config/ghostty/config`                          | `~/Library/.../com.mitchellh.ghostty/` (symlink)            | `~/.config/ghostty/config` (direct)            | — (no Windows build)                                                               |
| `dot_config/windows-terminal/settings.json` *        | —                                                           | —                                              | `%LOCALAPPDATA%\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\` (symlink) |
| `dot_config/powershell/profile.ps1` *                | —                                                           | —                                              | `<Documents>/{PowerShell,WindowsPowerShell}/profile.ps1` (symlinks, created by `run_after_link-pwsh-profile.ps1`) |
| `.chezmoidata/aliases.toml` (+ `.aliases.private.toml.age`) | `~/.config/shell/aliases.sh`, `~/.config/shell/omz-aliases.zsh` | `~/.config/shell/aliases.sh`                   | `~/.config/powershell/aliases.ps1` (+ `~/.config/shell/aliases.sh` for Git Bash)   |
| `dot_config/git/config.tmpl`                         | `~/.config/git/config`                                      | `~/.config/git/config`                         | `~/.config/git/config`                                                             |
| `dot_config/oh-my-posh/theme.omp.json`               | `~/.config/oh-my-posh/theme.omp.json`                       | same                                           | same                                                                               |
| `dot_config/vim/vimrc`                               | `~/.config/vim/vimrc`                                       | same                                           | `~/vimfiles/vimrc` (symlink)                                                       |
| Vim native packages (`catppuccin`, `previm`, `mermaid`) | `~/.config/vim/pack/`                                  | same                                           | `~/vimfiles/pack/`                                                                 |
| `dot_config/git/{ignore,commit-template}`            | `~/.config/git/`                                            | same                                           | same                                                                               |
| `dot_profile`, `dot_bashrc`, `dot_bash_profile`      | `~/.profile`, `~/.bashrc`, `~/.bash_profile`                | same                                           | same (used by Git Bash / WSL)                                                      |
| `dot_zshrc`, `dot_zprofile`                          | `~/.zshrc`, `~/.zprofile`                                   | — (ignored)                                    | — (ignored)                                                                        |
| `dot_config/git/personal.tmpl` (work machines only)  | `~/.config/git/personal`                                    | same                                           | same                                                                               |
| `dot_config/readline/inputrc`                        | `~/.config/readline/inputrc` (via `$INPUTRC`)               | same                                           | same (Git Bash)                                                                    |
| `dot_config/tmux/tmux.conf.tmpl`                     | `~/.config/tmux/tmux.conf`                                  | same                                           | `~/.config/tmux/tmux.conf` (deployed); `~/.psmux.conf` symlinks to that deployed file for PSMux |
| `private_dot_ssh/encrypted_private_config.age`       | `~/.ssh/config` (decrypted, owner-only)                     | same                                           | same                                                                               |
| `dot_config/git/hooks/` (`pre-push` secret scan + one dispatcher per other client hook) | `~/.config/git/hooks/`                 | same                                           | same (run by Git for Windows' sh)                                                  |
| `core.hooksPath` in `dot_config/git/config.tmpl`     | `~/.config/git/hooks`                                       | same                                           | same                                                                               |
| `dot_config/gitleaks/encrypted_private-names.toml.age` (personal machines only) | `~/.config/gitleaks/private-names.toml` (decrypted; private-name patterns for the push scan) | same | same |

\* Symlink-target-only sources — never deployed as regular files anywhere (listed in `.chezmoiignore` unconditionally); only referenced via `{{ .chezmoi.sourceDir }}/...` from the Windows-specific symlinks.

## Global git hooks

`core.hooksPath` points every repo at `~/.config/git/hooks` on all machine roles. `pre-push` runs `gitleaks` over the pushed commits and blocks on a finding (and when `gitleaks` is missing); where `GITLEAKS_PRIVATE_CONFIG` names a readable file (personal machines, set in `.profile` and the PowerShell profile) and the remote is public (unknown counts as public), a second pass with that config runs. Every other client hook is a dispatcher that runs the repo's own `.git/hooks/<name>`, so existing hooks keep working. A repo that sets its own `core.hooksPath` (husky, lefthook) skips these global hooks. `make test-git-hooks` (part of `make check`) tests them.

## Bootstrap scripts (`.chezmoiscripts/`)

| Script | macOS | Linux | Windows |
| --- | :---: | :---: | :---: |
| `run_once_init-untracked-env.sh.tmpl` → `~/.config/shell/secrets.sh` | ✓ | ✓ | — |
| `run_once_init-untracked-env.ps1.tmpl` → `~/.config/powershell/secrets.ps1` | — | — | ✓ |
| `run_once_after_create-xdg-dirs.sh.tmpl` → `$XDG_STATE_HOME/{zsh,bash,vim}`, `$XDG_CACHE_HOME/zsh` | ✓ | ✓ | — |
| `run_onchange_after_install-pkgs.sh.tmpl` → the packages of [Packages](#packages) | ✓ | ✓ | — |
| `run_onchange_after_install-pkgs.ps1.tmpl` → the same, with winget | — | — | ✓ |
| `run_onchange_after_zz-agents-apply.*` → [`agents apply`](#agents-apply); `run_after_update-agent-plugins.*` and `run_after_agents-doctor.*` → [plugins](#agent-plugins-giocaizzi-harness), [doctor](#doctor-agents-doctor); `run_onchange_after_sync-vscode-extensions.*` | per [AI harnesses](#ai-harnesses) | | |
| `run_after_link-pwsh-profile.ps1.tmpl` → resolves the real Documents folder (OneDrive/locale) and links the profile | — | — | ✓ |

Each script states where it runs: its body is wrapped in a condition (OS; harness for the harness scripts), and elsewhere it renders empty, which chezmoi skips. No script is listed in `.chezmoiignore`. The `run_onchange_` scripts re-run when their rendered text changes, so each embeds what it depends on (the package lists, the harness checkout's commit, the extension list).

## Externals (`.chezmoiexternal.toml.tmpl`)

Declarative upstream sources cloned/fetched by chezmoi on `apply`, refreshed per `refreshPeriod`. Prefer over an install-script `git clone` whenever the upstream is "drop this repo at this path" (no build, no PATH wiring).

| Entry (target path)                                    | Upstream                                  | macOS | Linux | Windows |
| ------------------------------------------------------ | ----------------------------------------- | :---: | :---: | :-----: |
| `~/.config/vim/pack/catppuccin/start/catppuccin/`      | `github.com/catppuccin/vim`               |  ✓    |  ✓    |   —     |
| `~/vimfiles/pack/catppuccin/start/catppuccin/`         | `github.com/catppuccin/vim`               |  —    |  —    |   ✓     |
| `~/.config/vim/pack/lsp/start/lsp/`                    | `github.com/yegappan/lsp` (pinned commit archive) |  ✓ (not light) |  ✓ (not light) |   —     |
| `~/vimfiles/pack/lsp/start/lsp/`                       | `github.com/yegappan/lsp` (pinned commit archive) |  —    |  —    |   ✓ (not light) |
| `~/.config/vim/pack/markdown/start/{previm,mermaid}/` | `github.com/previm/previm`, `github.com/mracos/mermaid.vim` | ✓ | ✓ | — |
| `~/vimfiles/pack/markdown/start/{previm,mermaid}/`    | `github.com/previm/previm`, `github.com/mracos/mermaid.vim` | — | — | ✓ |
| `~/.local/share/fonts/JetBrainsMonoNerdFont/`          | Nerd Fonts `JetBrainsMono.tar.xz` release |  —    |  ✓ (not light) |   —     |
| `~/.oh-my-zsh/`                                        | `github.com/ohmyzsh/ohmyzsh`              |  ✓    |  —    |   —     |

Release binaries (`bws`, `gitleaks`, `actionlint`, `hadolint`, the Go toolchain, `goreleaser`, `syft`, `glab`, `golangci-lint` (personal), `terraform`, `terraform-ls` and `tflint` (`terraform` answer), the `chezmoi-dots` plugin) are externals too, generated from the `[release.*]` tables of the [package registry](#packages) for this OS, architecture and role, each pinned and checksum-verified. `dotsRender` (an override-data flag set by `chezmoi dots` for previews) makes the file render empty, so previewing another OS or role never downloads its binaries.

The Vim LSP client is pinned by commit SHA and checksum in the URL: bump it by hand (new SHA and checksum), since `chezmoi dots outdated` does not see externals (tracked in [issue #185](https://github.com/giocaizzi/dotfiles/issues/185)).

**When NOT to use externals:** anything needing `chmod +x` on a downloaded binary with arch detection (oh-my-posh), official installer scripts (uv), or package-manager registration (brew/apt/winget). Those stay in `.chezmoiscripts/`.

### Vim Markdown previews

Open a Markdown preview in the default browser with `:PrevimOpen`; it refreshes after each buffer update. Previm renders fenced `mermaid` blocks without Node or another converter, while `mermaid.vim` provides syntax and indentation for Markdown fences and `.mmd` / `.mermaid` files. `:PrevimRefresh` forces an update and `:PrevimWipeCache` removes generated previews.

Previews are created under the user's cache directory (`%LOCALAPPDATA%\vim\previm` on Windows, `~/Library/Caches/vim/previm` on macOS, and `$XDG_CACHE_HOME/vim/previm` or `~/.cache/vim/previm` on Linux). They require `start`, `open`, or `xdg-open` respectively; headless systems retain syntax support but cannot open a browser preview.

## Asymmetries (by design)

- **Ghostty** has no Windows build → Windows Terminal fills the gap with matching theme/font.
- **zsh** is ignored on Linux and Windows (bash is the Unix default; PowerShell is the Windows default).
- **PowerShell profile** is Windows-only; trivially extendable to Unix by mirroring the `dot_config/powershell/` snippet pattern.

## Machine roles

Three independent yes/no answers (`light`, `work`, `terraform`), asked once per machine by `chezmoi init` as a choice (arrow keys to pick `yes` or `no`, Enter to confirm; default `no`) and stored as booleans only in that machine's `~/.config/chezmoi/chezmoi.toml` (`data.light`, `data.work`, `data.terraform`). A config that already holds the boolean is not asked again. Templates read them with `get . "light"` / `get . "work"`, which treats a missing key as false.

| Flag | Prompt | When true |
|---|---|---|
| `light` | Light install (yes/no) — shell, git, vim, prompt only | `.chezmoiignore` skips VS Code, Ghostty and the agents config and instruction layers ([no agents on light machines](#ai-agent-configuration)) and the encrypted SSH config (so light machines need **no age key**); no Nerd Font download; the install script installs only `fzf git vim fd tmux (Unix) / PSMux (Windows) jq shellcheck shfmt` + oh-my-posh (no uv/Python, Node, language servers, linters or forge CLIs). |
| `work` | Work computer (yes/no) — work git identity, no personal telemetry | Asks the work email, which becomes the default `user.email` (`gitEmail` stays the personal one), the **personal GitHub username** (`githubUser`) and optional **extra personal remotes** (`personalGitRemote`: github.com org names or URL globs, comma-separated; blank = none). The username and each org name (an entry of letters, digits and `-`) expand to the four github.com URL forms (`git@github.com:<owner>/**`, `ssh://git@github.com/<owner>/**`, `https://github.com/<owner>/**`, `https://*@github.com/<owner>/**`); each of those and each extra glob becomes `includeIf "hasconfig:remote.*.url:<glob>"` (git ≥ 2.36) → `~/.config/git/personal`, so repos in any folder with a matching remote commit with the personal email. Everything else, including repos with no remote, uses the work email. In an extra glob `*` stays within one path segment: write `https://git.example.com/**`, `https://*@git.example.com/**` (URL with user), `git@git.example.com:*/**` (scp-like) and `git@<ssh-alias>:*/**` for an SSH `Host` alias. To add more later see [Adding personal remotes](#adding-personal-remotes). Keep the SSH key consistent with the identity through `~/.ssh/config.d/*` (e.g. `Host github.com` → personal key, a work alias → work key). No agent telemetry goes to the OTel endpoint (the personal profile alone enables it). No Go stack (`go`, `gopls`, `goreleaser`, `syft`, `golangci-lint`, the `gopls` server) and the personal-only harness plugins are disabled (`[profiles.work]`, [Agent plugins](#agent-plugins-giocaizzi-harness)). The install scripts add the GitLab and Azure DevOps CLIs: `glab` and Azure CLI (`az`, brew `azure-cli`, winget `Microsoft.AzureCLI` machine-scope MSI) plus its `azure-devops` extension. |
| `terraform` | Install Terraform tooling (yes/no; not asked on light machines, where it is no) | Installs `terraform`, `terraform-ls` and `tflint` (their packages and linux releases carry `requires = "terraform"`), makes the `terraform` and `tflint` language servers available ([Language servers](#language-servers)) and the `hashicorp.terraform` VS Code extension. The configuration matrix seeds it true on personal and work, false on light. |

Full personal machines are also asked for the Claude Code OTLP endpoint (`otelEndpoint`, local config only); blank omits the telemetry keys.

Private agent instructions and personal shell aliases are not in this public repo: they are age-encrypted layers of the global agent instructions (`personal` on personal machines, `work` on work ones, see [Agent instructions](#agent-instructions)) and personal aliases in `.aliases.private.toml.age`, merged into the alias registry at render time (see [SHELL.md → Personal aliases](./SHELL.md#personal-aliases)). Light machines skip both. VS Code trusted domains for private hosts are set per machine in the local user settings, not in the shared file.

Current machines: Mac = full/personal, Raspberry Pi = light/personal, Windows PC = full/work.

Change an answer later: `chezmoi edit-config` and set `light` / `work` under `[data]` to `true` or `false`, or re-answer every prompt with `chezmoi init --prompt`; then `chezmoi apply`.

Re-running `chezmoi init`:

- Plain `chezmoi init` asks only for keys missing from the local config (e.g. `harnesses` on a machine set up before that prompt existed) and keeps every stored answer.
- `chezmoi init --prompt` asks every prompt again, each prefilled with the stored answer (text as the default, yes/no and the harness list preselected), so Enter keeps it. `chezmoi init --prompt --promptDefaults` re-renders the config with every stored answer and no questions.

### AI harnesses

Full machines are also asked which AI harnesses to configure: a multiple choice of `claude`, `copilot`, `vscode`, `codex`, `opencode`, `pi`, `antigravity` (arrow keys to move, space to toggle, Enter to confirm; none preselected), stored as `data.harnesses`. chezmoi never installs the CLIs. Light machines are not asked and get none (even with the key set); a config without the key gets none. `.chezmoitemplates/harnesses` is the one reader of the selection.

The selection is a machine fact: it becomes `harnesses` in the agents config, and everything else about the harnesses (settings, MCP servers, plugins, instructions, LSP) is written by `agents apply`, not by chezmoi. With no harness selected chezmoi deploys neither `~/.config/agents` nor the apply, doctor and plugin-update scripts. See [AI agent configuration](#ai-agent-configuration).

The one harness-related set of files chezmoi still owns is VS Code's: the static user settings and keybindings (`dot_config/Code/User/`, [Destination matrix](#destination-matrix)). `chezmoi apply` also installs the missing extensions listed in `.chezmoitemplates/vscode-extensions` through the `code` CLI (a shared list plus a `work` block, PowerShell only on Windows, or a personal block). Names only, so VS Code auto-updates them; add-only, so uninstall a dropped extension by hand. It warns and skips when the `code` CLI is unavailable. VS Code's agent chat runs on the Copilot harness, so `agents apply` gives `vscode` the `~/.copilot` files too.

Add or remove a harness later: `chezmoi edit-config` and edit the list under `[data]`, e.g. `harnesses = ["claude", "copilot", "vscode"]` (or re-answer every prompt with `chezmoi init --prompt`), then `chezmoi apply`. Removing a harness only stops `agents` managing its keys; files already in the home directory stay.

Simulate the prompts without touching the real config (prompt flags are keyed by the prompt text, as printed by `chezmoi init`):

```sh
chezmoi execute-template --init \
  --promptChoice "[ INSTALL MODE ] Light setup? (shell, git, vim, and prompt only)=no" \
  --promptMultichoice "[ AI HARNESSES ] Which harness configs to manage? (space toggles; Enter confirms; chezmoi never installs the CLIs)=claude/copilot" \
  < .chezmoi.toml.tmpl
```

`execute-template` keeps the answer already in your config for a key it finds there (and, unlike `chezmoi init`, re-asks a stored `harnesses` list); point `--config` at an empty `.toml` file to simulate a new machine.

### Adding personal remotes

On a work machine, add another personal org or host by editing the local config (never the repo): in `~/.config/chezmoi/chezmoi.toml` under `[data]`, append to `personalGitRemote` (comma-separated) a github.com org name, which expands to its four URL forms, or a URL glob for any other host, then `chezmoi apply`:

```toml
personalGitRemote = "other-org,https://ghe.example.com/**,git@ghe.example.com:*/**"
```

Git matches the URL case-sensitively: write the org as it appears in the remote URLs.

`chezmoi init --prompt` also works but re-asks every prompt. Check a repo with `git config --show-origin user.email`. If the host is not github.com, give it a `Host` entry and a personal key in `~/.ssh/config.d/` so the SSH key matches the identity.

The corporate-PC workarounds below are tied to Windows, not to `work`: they're harmless on any Windows machine.

## Bitwarden

Two Bitwarden CLIs, used per machine role:

| Role | `bw` (password vault) | `bws` (Secrets Manager) |
|---|---|---|
| full + personal (Mac) | Homebrew `bitwarden-cli`; fetches the age key at setup | own machine-account token → **personal** project |
| light (Pi) | — | — (`secrets.sh` by hand) |
| work (Windows PC) | — (personal vault stays off the work PC; age key copied by hand) | own machine-account token → **work** project |

- **Env secrets** (`~/.config/shell/secrets.sh`, `~/.config/powershell/secrets.ps1`) are rendered by chezmoi from the machine's project (`bws secret list <project> --output json`, via `.chezmoitemplates/bws-secrets.json`): each secret's name becomes the variable name. Edit secrets in Bitwarden, then `chezmoi apply`. Names that aren't valid variable names are skipped.
- **One token per machine**, created as a Bitwarden machine account with read access to the machine's project (the personal Mac's token can also read `rp5-homeserver`, for the home-server functions below), stored only in that machine's OS keyring (Windows Credential Manager, macOS Keychain, Linux Secret Service), never in `chezmoi.toml`: `chezmoi secret keyring set --service=bws --user=access-token`. `chezmoi.toml` keeps only the project ID, which switches the feature on. Give the token an expiry date, and revoke a single machine's token in Bitwarden if it's lost. The token never goes on a command line (chezmoi echoes failed commands, which would print it): the template runs `bws` through `powershell -Command` / `sh -c`, which reads it with `chezmoi secret keyring get` and passes it as `BWS_ACCESS_TOKEN`. A project ID without a stored token fails with `secret not found in keyring`; headless Linux without a Secret Service keyring can't use this, so leave the project ID blank and edit the secrets file by hand.
- **`bws` is fetched** by `.chezmoiexternal.toml.tmpl` into `~/.local/bin` (`[release.bws]` in `.chezmoidata/packages.toml`, pinned and checksummed; `chezmoi dots bump` updates it). The first apply on a new machine fetches it; the next apply renders the secrets.
- **Home-server functions (personal Mac):** `rp5-secrets <stack>` (Swarm secrets, `create_secrets.sh`) and `rp5-sync-infra` (infra deploy with its file secrets, `sync_infra.sh`; needs `DOMAIN`) read that token from the keyring when they run and pass it only to the script, so it never sits in the shell environment. They are why `BWS_ACCESS_TOKEN` is empty in an interactive shell even though the machine can read the project.
- **Online needed:** on machines with a project ID, every `chezmoi status`/`diff`/`apply` calls `bws` once and fails offline.
- No Bitwarden MCP server: it has no read-only mode and would give an AI agent full vault access.
- **Tab alert:** every harness rings the terminal bell when it waits on the person, so the terminal marks that tab. Claude Code `preferredNotifChannel: "terminal_bell"`, Copilot `beep`, Codex `[tui] notification_method = "bel"`; opencode and pi through the harness `sounds` plugin. Ghostty `bell-features` shows 🔔 in the title, a border, a Dock bounce and the system sound; Windows Terminal `bellStyle: "all"` flashes the window and taskbar and plays the sound. tmux forwards bells to the outer terminal and marks belling windows with a red `!`; it does not monitor activity (see [tmux](SHELL.md#tmux)).
- **Claude Code TUI:** `tui` is `fullscreen` everywhere (needed for `viewMode: focus`, clicks and the side-by-side `/diff`, which also needs Claude Code ≥ 2.1.287 and a terminal ≥ 110 columns). macOS needs Ghostty mouse reporting on. On Windows Terminal, if stale text fragments appear, set `CLAUDE_CODE_ALT_SCREEN_FULL_REPAINT=1`; `CLAUDE_CODE_SCROLL_SPEED=3` is already set there.
- **Claude Code and `bw`:** every `bw` command Claude runs needs your approval (an `ask` rule in the managed Claude settings; ask rules beat the blanket Bash allow, also inside pipes and in auto mode). To let Claude act, unlock in your own terminal into a private file, `(umask 077; bw unlock --raw > <file>)`; Claude runs `bw --session "$(cat <file>)" …` and prints only non-secret fields. Finish with `bw lock` and delete the file. The rule matches the command as written, so it's an approval gate, not a security boundary.
- **Claude Code and auto-merge:** `gh pr merge … --auto` is allowed by name, so agents can arm auto-merge in auto mode too, where the blanket Bash allow does not count; the merge still waits for required CI and the repo ruleset. `gh pr merge … --admin` is denied. Copilot CLI, Codex, opencode and pi have no rule for it here and ask as for any command.

## Editing your dotfiles

Many targets are composed: `~/.config/agents/config.toml` is a one-line stub over a shared template, an alias file is rendered from the alias registry, VS Code settings are symlinks into `dot_config/`. `chezmoi edit <target>` opens the stub; `chezmoi dots` opens the part you mean and shows what the edit changes on every machine type. `dots` is an alias for `chezmoi dots` on full machines; `chezmoi dots help` prints the full usage. It is [a chezmoi plugin](#the-chezmoi-dots-plugin) (Go, [giocaizzi/chezmoi-dots](https://github.com/giocaizzi/chezmoi-dots)), the same on macOS, Linux and Windows. On a terminal the output uses the basic ANSI colours, which the terminal's Catppuccin theme maps (headings blue, configurations and entries mauve, failures red, success green, prompts yellow), and the composition tree is drawn with `├─`/`└─`; a pipe, `NO_COLOR` or `TERM=dumb` gives plain text.

### Edit

| Command | Edits |
|---|---|
| `dots` | a registry entry, picked with fzf (a numbered prompt without fzf) |
| `dots <entry>` (also `dots edit <entry>`) | that entry, e.g. `dots agents/config`, `dots agents/work`, `dots aliases/private`, `dots ssh/config`; `dots list` shows them all |
| `dots <target>` | what a target is made of: a picker over the nodes of its composition tree (`dots ~/.config/agents/config.toml` → the stub, the template); a target made of one file opens it directly |
| `dots <source file>` | a source file by its source-relative path, e.g. `dots dot_config/git/config.tmpl` |
| `dots <home>/<local>/<name>` | a file of a `local` entry's folder, e.g. `dots ~/.ssh/config.d/pi` (`ssh/local`: this machine only, unmanaged; a new name creates it): the entry's check parses a private temp copy, which replaces the live file only when it passes (else edit again, or keep the live file); no render, diff, apply or PR. `dots ssh/local` picks an existing file |

Every source edit, in order:

1. **Render before:** the target state (`chezmoi dump`) of this machine and of 9 configurations — darwin, linux, windows × personal, work, light — each with every harness selected (light: none), rendered in parallel. The configurations come from the real `.chezmoi.toml.tmpl`, rendered per role with seeded answers, so there is no second copy of the init prompts to keep in step.
2. **Edit** in `$EDITOR` (then `$VISUAL`, then `vi`; on Windows `notepad`). An encrypted file is decrypted into a private temp file (owner-only), re-encrypted (armored) only when its content changed, and the temp file is always removed. Inside tmux or, on Windows, Windows Terminal, a side pane (`chezmoi dots live`) re-renders this machine from a private copy of the source on every save and shows the diff of each target the edit changes so far; it closes with the editor.
3. **Check:** the entry's check, when it has one (`chezmoi dots lint aliases templates`, `ssh -G`, ...).
4. **Render after** and report the changed targets per configuration: a list, `same as <configuration>` for an identical list, or `nothing`. A configuration that no longer renders fails the edit; one that already failed before is only reported.
5. **This machine only:** `chezmoi diff` of its changed targets (one diff when the entry says they render the same text), `apply now? [y/N]` (yes applies only those targets; skipped when they are already up to date), and the entry's note (e.g. open a new shell after an alias change). The current `dots` release prints source status with a PR reminder. The dotfiles repository supports direct pushes to `main`. A changed script is listed and runs on the next full `chezmoi apply`.

`chezmoi init` configures native chezmoi edits to stage, commit, and push source changes automatically. Commit and push changes made with `dots` separately.

A failed check or render applies nothing; the source keeps the edit, so run the command again to fix it. The diff and the live pane show decrypted text: do not run them where the screen is shared or recorded.

### View

| Command | Shows |
|---|---|
| `dots view <target>` | the composition tree (each node with its registry entry, encrypted nodes marked) and the rendered content on this machine |
| `dots view <entry>` | the entry's tree and the targets it reaches on this machine |
| `... --as <os>[,<role>]` | the same for one configuration: os `darwin`, `linux` or `windows`; role `personal` (default), `work` or `light`. A preview, never applied |
| `... --matrix` | every configuration: 3×3 Catppuccin matrix grid (`darwin`, `linux`, `windows` × `personal`, `work`, `light`) showing reachability and distinct variant paths/renders, followed by a color-coded legend |
| `dots view [--matrix] [--all]` | whole-setup matrix overview across all registry entries and configurations, with a compact visual combo strip |

The composition tree is read from the source: `template`/`includeTemplate` names, `include` paths (a `printf` pattern is globbed, so both private layers show), the file a `symlink_` template points at, and the `.chezmoidata` files whose top-level keys a template reads. It lists every branch, whichever role takes it; `--matrix` shows which configuration takes which.

Limits of the simulation: identity answers this machine lacks render as placeholders (`<workEmail>`), the others blank, as at init. Paths come from this machine (`.chezmoi.homeDir`, `.chezmoi.sourceDir`), so a windows render shows Mac paths. Previews render with `dotsRender` set, so no external is downloaded for them. The 9 configurations render without a Bitwarden project, so secrets stay this machine's and parallel renders stay under the Secrets Manager rate limit; this machine's render still calls `bws` and needs to be online.

Short aliases: `cz` (`chezmoi`), `czd` (`chezmoi diff`), `cza` (`chezmoi apply`); edit through `dots`, not `chezmoi edit`.

### Edit registry

`.chezmoidata/edit.toml` names the components: each `[edit."<name>"]` has a `desc` (shown by the picker and `dots list`) and a `source` (source-relative) or a `local` folder (home-relative, unmanaged files edited in place), optionally a `check` (a command as a list, run without a shell in the repo root, chezmoi's working tree, with `{file}` replaced by the edited file; non-zero applies nothing; the same on every OS), `note` (printed after an apply) and `diff = "first"`. Any other source file or target is editable without an entry.

**Add an editable component**

**When:** a new file under `.chezmoidata/` or a new encrypted file (`chezmoi dots lint registry` fails without an entry), or any component worth a name or a check.

1. `dots registry` and add the entry (or, for a folder of unmanaged files, `local = "<home-relative folder>"` instead of `source`):
   ```toml
   [edit."<name>"]
   desc = "<what it is>"
   source = "<source-relative path>"
   check = ["<command>", "{file}"]   # optional
   ```
2. Apply when asked; ship the PR.

**Verify:** `chezmoi dots lint registry` passes; `dots list` lists the entry; `dots view <name>` shows the targets it reaches.

**Rollback:** delete the entry (`dots registry`).

## Packages

Every tool a machine installs is declared once, in `.chezmoidata/packages.toml`; the install scripts and the release externals render from it, so there is no second list. The file header documents the schema.

- `[package.<name>]`: installed by the install scripts with the platform's native manager (brew on macOS, apt on Linux, winget on Windows), plus uv, npm, an Azure CLI extension or a PowerShell module, per role (`light`, `work`, `personal`) and, with `requires = "<answer>"`, only where that boolean init answer is true (e.g. `terraform`). `agents = true` also names the tool to AI agents in the environment layer of the [agent instructions](#agent-instructions).
- `[release.<name>]`: a binary from a release page, pinned by version and checksum and installed by chezmoi as an external into `~/.local/bin` (no sudo, no package manager). It takes `roles` and `requires` like a package; `url` and `path` expand `{version}` and `{target}`. Upstreams `bump` and `outdated` can read: GitHub and GitLab tags, go.dev and `releases.hashicorp.com` (resolved to the tags of the matching `hashicorp/<name>` repository). Tap-qualified brew ids (`hashicorp/tap/terraform`) match by their short name.

`chezmoi dots lint packages` (part of `make check`) validates both tables, including that every release target has a 64-hex sha256 and that an optional `versionCmd` is a non-empty list of strings.

**What `chezmoi dots outdated` shows**

`chezmoi dots outdated [--offline] [--fetch] [--check]` prints every tool the registry gives this machine (its OS and role), one table per source: pinned releases, then the OS's manager (Homebrew formulae and casks, apt, winget), uv tools, npm tools and official installers (entries with `script` for this OS and no native-manager key). Columns: `TOOL`, `INSTALLED`, `PINNED` (releases, and npm specs that pin a version), `LATEST`, `STATE`.

- States: `up to date`; `outdated` (upstream has a newer release than the pin, or the manager knows a newer package); `drifted` (installed differs from the pin, e.g. after a pull before `chezmoi apply`); `not installed`; `unknown`.
- Footer: a `source` line (this checkout against its upstream from the last fetched refs: behind/ahead, local changes, when last fetched), a `summary` line and `next` hints per state (`chezmoi dots upgrade dotfiles`, `chezmoi dots bump <release>`, `chezmoi apply`, `chezmoi dots upgrade <brew|apt|winget|uv|npm>`).
- Installed: the manager's local query, or, for releases and official installers, the tool's version command (`versionCmd`, run without a shell, with a timeout). Latest: releases from upstream tags; brew and apt from their local index as of the last `brew update` / `apt update`; uv and npm from their registries; winget and official installers are not looked up. Lookups run in parallel with timeouts and never print command output.
- `--offline` skips every network lookup (LATEST shows `unknown`, never an error); `--fetch` runs `git fetch` on the dotfiles first; `--check` exits 1 when anything is outdated, drifted or not installed (the overview alone always exits 0).

| Command | Moves | Reads |
|---|---|---|
| `chezmoi dots outdated` | nothing | remote (upstream releases, `origin/main`), source (`packages.toml` pins), installed |
| `chezmoi dots bump` | remote → source (pins in `packages.toml`) | upstream releases |
| `chezmoi apply` | source → installed (the pins) | this checkout |
| `chezmoi dots upgrade` | remote → source → installed (`chezmoi update`), plus system packages, uv and npm tools | |

Releases move only with a bump. `upgrade` still ends with the release-only check.

**Add a version command**

**When:** a tool's installed version shows `unknown` because `--version` does not print it.

1. Add `versionCmd = [...]` to its entry in `packages.toml` (argv; `{dest}` expands to a release's installed path). The first dotted number in the output is the version.
2. `chezmoi dots lint packages`.

**Verify:** `chezmoi dots outdated --offline` shows the tool's installed version.

**Rollback:** remove the key; the default applies.

**Update everything on this machine**

**When:** whenever, e.g. weekly.

1. `chezmoi dots upgrade`: updates the dotfiles (`chezmoi update --refresh-externals`), then the system packages (brew on macOS, apt on Linux, winget on Windows), uv tools, global npm tools and, on Windows, PowerShell modules. A step whose tool is missing is skipped; a failing step does not stop the others. `--dry-run` prints the plan, and naming steps (`chezmoi dots upgrade brew npm`) runs only those.
2. Pinned release binaries do not move with it: they change when the pin changes (below), then `chezmoi apply` fetches the new version.

**Verify:** `chezmoi dots outdated` shows every tool up to date; `chezmoi status` prints nothing.

**Bump the pinned release binaries**

**When:** `chezmoi dots outdated` marks a release outdated (the weekly `bump-releases` workflow opens a PR for this too).

1. `chezmoi dots bump` (or `chezmoi dots bump <release>...`; `--dry-run` previews): finds the newest tag of each release (git tags on GitHub/GitLab, go.dev for Go), downloads the assets for every target, and rewrites the version and sha256 lines in `packages.toml`, leaving comments and order alone.
2. Review the diff, `make check`, ship the PR; `chezmoi apply` installs the new versions.

**Rollback:** revert the PR; the old pin and checksums come back together.

## The chezmoi-dots plugin

`chezmoi dots` is a [chezmoi plugin](https://www.chezmoi.io/user-guide/command-overview/#plugins): chezmoi runs the `chezmoi-dots` executable found on `PATH` for `chezmoi dots ...`, with the chezmoi environment set (`CHEZMOI_SOURCE_DIR`, `CHEZMOI_EXECUTABLE`, ...). The plugin lives independently at [github.com/giocaizzi/chezmoi-dots](https://github.com/giocaizzi/chezmoi-dots).

| Command | Does |
|---|---|
| `chezmoi dots [edit] [<what>]`, `view`, `list`, `live` | [Editing your dotfiles](#editing-your-dotfiles) |
| `chezmoi dots lint [check]...` | `registry`, `aliases` (the encrypted registry too where the age key exists), `packages`, `matrix`, `templates` (every script, alias file and settings template, rendered per configuration and shellchecked, zsh-checked or pwsh-parsed where those tools are installed) |
| `chezmoi dots audit aliases` | lists aliases live in this machine's zsh and bash that the registry does not account for |
| `chezmoi dots outdated`, `bump`, `upgrade` | [Packages](#packages) |

**Install it**

1. `chezmoi apply` installs the pinned release (`[release.chezmoi-dots]` in `packages.toml`) into `~/.local/bin`.
2. Standalone install: `go install github.com/giocaizzi/chezmoi-dots/cmd/chezmoi-dots@latest`.

**Verify:** `chezmoi dots version`.

**Rollback:** delete the release and tag, and revert the pin.

## Encrypted files (age)

Secrets such as `~/.ssh/config` are committed age-encrypted (`encrypted_` in the source name). Each personal and work machine needs the private key at `~/.config/chezmoi/key.txt` (owner-only) **before** `chezmoi apply` — light machines don't, and must not hold it; copy it from another machine over a secure channel, never through the repo. The public key in `.chezmoi.toml.tmpl` is safe to publish. Add or update an encrypted file with `chezmoi add --encrypt <file>`, or edit one in place with `chezmoi edit --apply <file>` (decrypts to an editor, re-encrypts on save); `~/.ssh/config` with `dots ssh/config`, which also checks it ([Editing your dotfiles](#editing-your-dotfiles)).

**SSH key names (same on every machine):** `~/.ssh/id_<type>_<service>_<account>` (+ `.pub`) — type first, then the service, then the account or organisation: `id_ed25519_github_<github-user>`, `id_ed25519_github_<work-account>`, `id_rsa_devops_<org>`. The shared `~/.ssh/config` points `Host github.com` at the personal GitHub key, so don't redefine `github.com` in `config.d/` (it would shadow the shared entry); rename the key instead. Work keys and their `Host` aliases (e.g. `github-<work>`, `azure-<org>`) live in `config.d/<name>`.

**SSH hosts per machine:** the shared `~/.ssh/config` starts with `Include config.d/*` and holds only hosts every full machine uses (GitHub). Hosts that belong to one machine only — e.g. work hosts on the corporate PC, the home-server host on the personal Mac — go in `~/.ssh/config.d/<name>` there; chezmoi doesn't manage that folder, so they never enter this repo, not even encrypted. Edit them with `dots ~/.ssh/config.d/<name>` ([Editing your dotfiles](#editing-your-dotfiles)). Work keys are generated on the work machine itself. If a work Git host is the same as a personal one, give the work account its own `Host` alias in `config.d/<name>` (own key, `IdentitiesOnly yes`) and clone work repos through that alias.

## Hardcoded by design

chezmoi target paths are literal (they can't be templated), so a few values stay fixed on purpose:

- **`~/.config`, `~/.local/{share,state}`, `~/.cache`** — chezmoi deploys to these exact paths, so `.profile` pins `XDG_*` to the spec defaults. Pointing `XDG_CONFIG_HOME` elsewhere would split config from where chezmoi writes it.
- **Homebrew prefixes** `/opt/homebrew` (Apple Silicon) and `/usr/local` (Intel) — fixed by Homebrew.
- **Vendor paths** — `Library/Application Support/…`, `AppData/…`, the Windows Terminal package family name.
- **`{{ .chezmoi.homeDir }}`** in templates — rendered per machine.
- **Personal data** — the OTel endpoint (personal machines only), project aliases. Git identity is prompted at `chezmoi init`.

Anything that genuinely differs per machine and can't be templated (e.g. a redirected Documents folder) is resolved at runtime by a script.

## Managed (corporate) Windows PCs

Endpoint policies on the work PC block several defaults; the workarounds apply to every Windows machine:

| Blocked | Workaround |
|---|---|
| Machine-wide installs (winget font package) | JetBrainsMono installed per user by `oh-my-posh font install` |
| `Set-Acl` (needs SeSecurityPrivilege) | `icacls` restricts `secrets.ps1` |
| uv's `python.exe` launchers in `~\.local\bin` (unsigned exe named python.exe) | `uv python install --no-bin`; `profile.ps1` puts uv's `cpython-3.NN-*` interpreter folder on PATH |
| `New-Item -ItemType SymbolicLink` in Windows PowerShell 5.1 without admin | `mklink` (honours Developer Mode) in `run_after_link-pwsh-profile.ps1` |
| Documents redirected to OneDrive and localised | profile link target resolved at runtime with `GetFolderPath('MyDocuments')` |

## AI agent configuration

Agent harness configuration lives in the standalone [`agents`](https://github.com/giocaizzi/agents) CLI; chezmoi supplies its inputs and runs it. The split:

| Piece | Owner | Where |
|---|---|---|
| The `agents` binary | chezmoi installs the pinned release | `[release.agents]` in `.chezmoidata/packages.toml` ([Packages](#packages)) |
| The config (`~/.config/agents/config.toml`) | chezmoi renders it from one template | `.chezmoitemplates/agents-config.toml`, deployed by `dot_config/agents/private_config.toml.tmpl`; edit with `dots agents/config` |
| The instruction layers (`~/.config/agents/instructions/`) | chezmoi deploys the files | `dot_config/agents/instructions/{core.md, env.md, personal, work}`; edit with `dots agents/<layer>` |
| Every harness file (settings, MCP, rules, plugins, LSP, instructions) | `agents apply` | `~/.claude`, `~/.copilot`, `~/.codex`, `~/.config/opencode`, `~/.pi`, `~/.gemini`, workspace folders |

Reference, not repeated here: [config.md](https://github.com/giocaizzi/agents/blob/main/docs/config.md) (every key), [cli.md](https://github.com/giocaizzi/agents/blob/main/docs/cli.md) (commands and runbooks), [features.md](https://github.com/giocaizzi/agents/blob/main/docs/features.md) (what each harness supports and how each key is mapped). The procedures for this repo are in [Runbook: agent configuration](#runbook-agent-configuration).

Light machines, and machines with no harness selected, get none of it: no config, no layers, no scripts, and no age key is needed for them.

### The agents config

`.chezmoitemplates/agents-config.toml` is plain agents config except for a few machine facts that are template output: `profile` (`personal` or `work`, from the `work` answer), `harnesses` (the selection, [AI harnesses](#ai-harnesses)), the telemetry endpoint (`otelEndpoint`, personal machines only; blank omits it), and the instruction variables (`instructions.vars`: the tools of the package registry marked `agents = true` that this OS and role install, plus the package manager and install command). Everything else is edited in that file as written: autonomy and permissions, model tiers, context, UI, voice, telemetry, language servers, instruction layers, plugins, MCP servers, per-harness settings (`[harness.<name>]`, with raw native keys under `[harness.<name>.settings]`), Windows-only overrides (`[os.windows...]`) and the per-profile differences (`[profiles.personal]`, `[profiles.work]`).

Nothing private is in it: account names and client ids appear as `${VAR}` references that `agents apply` expands from the environment (Bitwarden secrets exported by `secrets.sh`, see [SHELL.md](./SHELL.md)). A client secret is never in the config; it is set once per machine and harness.

`agents config show -origin` shows the resolved config on this machine and which layer set each key; `chezmoi cat ~/.config/agents/config.toml` shows the rendered file.

### `agents apply`

`run_onchange_after_zz-agents-apply.{sh,ps1}.tmpl` runs `agents apply` after a `chezmoi apply` in which its inputs changed. It runs last (`zz`), after the package install has put the `agents` binary in place. The script embeds a hash of the rendered config, of each instruction layer file and of the pinned `agents` release, so it re-runs on a change to any of them and on an `agents` bump. Machines with no harness selected render it empty, as does the other OS's variant.

`agents apply` writes the native files of every selected harness. Files the apps also write (Claude Code and Copilot CLI settings, Codex `config.toml`, opencode, pi) are merged, not replaced: only the keys `agents` manages are written, in the live key order, nothing unchanged is rewritten, and a key that left the config is removed. Keys the app writes itself (a chosen model, hook trust, local plugins, authentication) stay local and never reach this public repo. The VS Code `settings.json` is check-only (`agents` reports drift and never writes it). Each target is `create`, `update`, `unchanged`, `drift`, `refused` (a symlink: it never writes through one) or `error`; the full semantics are in [cli.md](https://github.com/giocaizzi/agents/blob/main/docs/cli.md#apply).

The script never fails the chezmoi apply: a missing binary is skipped with a note (run `chezmoi apply` again once the package install has run), and a failed `agents apply` is a warning.

- **Preview:** `agents apply -n` prints the diff of each target and writes nothing; `agents render -harness claude` lists the managed keys of one harness.
- **Run by hand:** `agents apply`. The script re-runs on its own only when an input changes.
- **Check:** `agents check` validates the config; `agents check -live` also reports native files that differ from what apply would write.
- **Skip for one run:** `chezmoi apply --exclude=scripts` (skips every script; the apply script runs at the next `chezmoi apply`).
- **Roll back:** there is no automatic undo. Revert the config change in git and apply again; a key that leaves the config is removed from the live file.

### Agent instructions

Every harness reads the same global instructions, composed by `agents` from layers and written whole to `~/.claude/CLAUDE.md`, `~/.copilot/copilot-instructions.md`, `~/.codex/AGENTS.md`, `~/.config/opencode/AGENTS.md`, `~/.pi/agent/AGENTS.md` and `~/.gemini/config/AGENTS.md`. Copilot CLI and VS Code (Copilot harness) both read `~/.copilot/copilot-instructions.md` natively, so VS Code needs no `chat.instructionsFilesLocations` entry: with `~/.copilot` listed there, VS Code offers to migrate the file into `~/.copilot/instructions/*.instructions.md`, and Copilot CLI then loads the instructions twice. Decline that migration.

The layer files are deployed by chezmoi to `~/.config/agents/instructions/`; the config declares them (`[instructions.layers.*]`) and each profile lists the ones it uses, in order (`[profiles.<name>] instructions`):

| Layer | Source (under `dot_config/agents/instructions/`) | Public | Profiles |
|---|---|---|---|
| core | `core.md` | yes | personal, work |
| env | `env.md` (an agents template: `.profile`, `.os`, `.vars.*`) | yes | personal, work |
| personal | `encrypted_private_personal.md.age` | no (age) | personal |
| work | `encrypted_private_work.md.age` | no (age) | work |

`.chezmoiignore` skips the other role's private layer, so a work machine never receives the personal one.

Core holds how to work on any machine, in any harness: it names no harness-only tool (skills carry the per-harness detail), except that the plan-first routing line uses Claude Code's plugin-namespaced agent names, which the other harnesses know by the plain names, kept verbatim because evals validated that wording. Env holds what depends on the machine: how its CLI environment is managed, the tool-install rule (brew, apt or winget, from `instructions.vars`), the tools it has and, on personal machines only, the own-org rules. The private layers hold what must not be public: on personal machines the side projects and the boundary list; on work machines the employer context and the boundary rule seen from the work side. They are inlined into one text rather than imported with `@file`, which only Claude Code expands.

- **Edit a layer:** `dots agents/<core|env|personal|work>` ([Editing your dotfiles](#editing-your-dotfiles)); a private layer goes through a private temp file and is checked before saving.
- **Check:** `agents prompt matrix` (layers per profile and harness), `agents prompt check` (renders every cell; the other role's missing layer is "skipped (not on this machine)") and `agents prompt show -profile <personal|work> -harness <name>` (the composed text of one cell).
- **Edit a private layer by hand:** `chezmoi decrypt` it into a `umask 077` temp file, edit, `chezmoi encrypt` it back, delete the temp file, `chezmoi apply`.

### Model tiers

Models are grouped in free-named tiers in the config, and agents are assigned to a tier; nothing else joins them.

- `[models.tiers.<name>]` sets, per harness, a native model id (`claude`, `copilot`, `codex`, ...) and a reasoning `effort`. A harness with no id keeps its plugin or session default. This repo defines `light`, `standard`, `strong` and `frontier`.
- `[models.assign]` maps an agent id (`<plugin>:<agent>`, globs allowed, an exact key wins) to a tier. `models.default` would set the tier of the main session model.
- `[models.catalog]` holds the preferences of the benchmark catalog (`providers`, `since`, `pins`, `ignore`, `aliases`). The catalog itself is cached locally (`models.json` in the agents cache directory), refreshed by `agents models sync`, shown by `agents models`, and never committed.

`agents models tiers` shows what each tier gets per harness and each assigned agent's effective model. How a tier reaches each harness, and where a harness cannot set a per-agent model: [features.md](https://github.com/giocaizzi/agents/blob/main/docs/features.md). To change a tier: [Runbook](#change-a-model-tier-or-an-assignment).

### Agent plugins (`giocaizzi-harness`)

Plugins come from the `giocaizzi/harness` registry. `[plugins] registry` and `[plugins.enabled]` in the config list, per plugin directory of that repo, the harnesses that enable it; `[profiles.work] plugins.disabled` keeps the personal-only plugins off work machines. `agents apply` writes the enablement into each harness's settings, and `agents plugins` installs and updates the plugins through each harness's own CLI (the Claude Code and Copilot CLI marketplaces, Codex, Antigravity) from the registry checkout `agents` keeps; opencode gets copies made by the checkout's installer, pi gets the checkout as a package. chezmoi no longer clones the registry or wires plugins itself. Hooks (guards, `make check` on stop, post-edit lint, keep-awake) ship in the `guard` plugin: [guard.md](https://github.com/giocaizzi/harness/blob/main/docs/guard.md); install detail: [install.md](https://github.com/giocaizzi/harness/blob/main/docs/install.md). Codex skips plugin hooks until trusted: run `/hooks` once per `guard` version.

`run_after_update-agent-plugins.{sh,ps1}.tmpl` runs `agents plugins update` after every chezmoi apply and update (it pulls the registry at most once every 24 hours; a failure only warns). Inspect by hand: `agents plugins list` (installed against upstream per harness), `agents plugins update -n` then `agents plugins update`, `agents plugins prune`; [cli.md → plugins](https://github.com/giocaizzi/agents/blob/main/docs/cli.md#plugins).

- **New plugin upstream:** add `<name> = [harnesses]` under `[plugins.enabled]` (`dots agents/config`), `chezmoi apply`.
- **Plugin removed upstream:** delete its key, apply; disable or uninstall it by hand where the harness keeps it.
- The Claude Code status line (`hud`) is set under `[harness.claude.settings.statusLine]`; it needs `node` 22+ and a Nerd Font.

### MCP servers

Rule: harness `repo-standard`, "MCP servers, least privilege". Two layers, both in the config:

| Layer | Config | Reaches |
|---|---|---|
| User | `[mcp.servers]` (every machine; none today) and `[profiles.personal.mcp.servers]` (personal machines) | the user-scope MCP file of every selected harness |
| Workspace | `[profiles.personal.mcp.workspaces.ops]` (folder `~/ops`, personal machines) | the native project files of that folder (`.mcp.json` and each harness's equivalent) |

VS Code follows the Copilot CLI files: its Copilot harness reads `~/.copilot/mcp-config.json` and the workspace-root `.mcp.json`, and runs OAuth sign-in itself.

- **Connectors off:** `disable_claude_ai_connectors = true` under `[harness.claude]`.
- **Private URL parts:** `${VAR}` in a server `url` or `oauth.client_id` (e.g. `${MCP_CONNECTOR_ACCOUNT}`), a Bitwarden secret of the same name exported by `secrets.sh`; `agents apply` expands it from the environment, so the written files (never committed) hold the value and this public repo does not. Unset (first apply on a new machine, before a new shell), the server is skipped with a warning: open a new shell and apply again.
- **`~/ops` is personal:** the workspace is declared in the personal profile only. Never put a `.mcp.json` in `~` (Claude Code merges it into every repo below).
- **Repo servers** live in each repo, not here; `agents mcp add` and `agents mcp sync <dir>` write a repo's native files from its `.mcp.json`.
- **Sign in:** OAuth servers ask once per machine and harness: Claude Code `/mcp`, Copilot CLI `/mcp`, VS Code on first connect, `codex mcp login <name>`, `opencode mcp auth <name>`, `pi mcp login <name>`. First run in a workspace folder: Claude Code asks to approve the project servers; Copilot CLI and pi ask to trust the folder.
- **Check:** `claude mcp list`, `copilot mcp list`, `codex mcp list`, `opencode mcp list`, `pi mcp list` from `~`, `~/ops` and a repo: only that layer's servers.

A server whose authorization server has no dynamic client registration takes a pre-registered OAuth client, `oauth = { client_id = "${MCP_<NAME>_CLIENT_ID}", callback_port = <port> }`. Which redirect URI each harness sends (register both the `localhost` and the `127.0.0.1` form for a client shared across harnesses) and which harnesses cannot take one: [cli.md → OAuth redirect hosts](https://github.com/giocaizzi/agents/blob/main/docs/cli.md#oauth-redirect-hosts). The client secret is set once per machine, harness and server, never in the config or Secrets Manager; keep it in the password vault. Procedure: [Runbook](#add-a-plugin-an-mcp-server-or-a-language-server).

### Language servers

`[lsp.servers.<name>]` entries in the config give each server's `command` (argv), `extensions` (`".ext"` to language id), optional `filetypes` (Vim filetypes, where they differ) and `format` (the editor formats on save with it). A server is configured only where its command is on `PATH`, so it follows the packages the machine installs (Go servers on personal machines, Terraform ones with the `terraform` answer, none on light machines); `[lsp] enabled = false` turns LSP off everywhere. Servers today: `pyright`, `ruff`, `typescript` (TypeScript 7's own `tsc --lsp --stdio`), `gopls`, `bash` (`bash-language-server`, which runs shellcheck and shfmt when present), `yaml-ls`, `json`, `terraform` and `tflint`.

| Consumer | How it gets the servers |
|---|---|
| Claude Code, Copilot CLI, opencode | `agents apply` (Claude Code through a plugin it generates, the other two in their LSP config); how a file extension claimed by two servers is resolved is in [features.md](https://github.com/giocaizzi/agents/blob/main/docs/features.md) |
| Vim | `.chezmoitemplates/vim-lsp.vim` reads the same rendered `~/.config/agents/config.toml` through `fromToml` → `~/.config/vim/plugin/lsp.vim` (Unix), `~/vimfiles/plugin/lsp.vim` (Windows). Client `yegappan/lsp`, a pinned external ([Externals](#externals-chezmoiexternaltomltmpl)); `features.documentFormatting` is on only for `format` servers. `vimrc` sets `g:filetype_tf = 'terraform'` (Vim sniffs `*.tf` as TinyFugue) |
| VS Code | Static settings, set by hand: formatters on save for Python (Ruff), TypeScript/JavaScript (`TypeScriptTeam.native-preview`), Go, shell (`bash-ide`) and Terraform; `js/ts.experimental.useTsgo`; `go.lintTool` = `golangci-lint-v2`; `terraform.languageServer.path` = `terraform-ls`. Extensions come from `.chezmoitemplates/vscode-extensions` |
| Codex, pi, Antigravity | None: no native LSP config |

To add a server: a `[lsp.servers.<name>]` entry plus, when chezmoi installs the binary, its row in `packages.toml`.

### Permissions, autonomy, voice and telemetry

These are the shared preferences in the config (`[autonomy]`, `[permissions]`, `[voice]`, `[telemetry]`, `[context]`, `[ui]`, `[notifications]`), written by `agents apply` in each client's own syntax. A shared pattern is not a guarantee of equivalent enforcement: harnesses differ in what they can express (wildcards, ask rules, execution modes), and launch, project and session settings can override user defaults. The per-harness mapping and the gaps are in [features.md](https://github.com/giocaizzi/agents/blob/main/docs/features.md); voice input per harness (where audio goes, which speech model): [voice.md](https://github.com/giocaizzi/agents/blob/main/docs/voice.md); parity: [parity.md](https://github.com/giocaizzi/agents/blob/main/docs/parity.md).

- Lease pushes (`git push --force-with-lease origin *`) are allowed only to feature branches: deny patterns take `main` and `master` back. The `guard` hook and the server ruleset stay the real protection: no pattern sees `git -C <dir> push`, a refspec such as `feat/x:main` or reordered arguments.
- Telemetry is on in the personal profile only when an `otelEndpoint` was given at `chezmoi init`; work machines send nothing.
- The VS Code voice keys in `dot_config/Code/User/settings.json` (`dictation.enabled`, `agents.voice.language`) are static: keep them in step with `[voice]` by hand; `agents check -live` reports drift.
- OS dictation fallback for harnesses without their own (types into the focused terminal): macOS Dictation (Microphone key), Windows voice typing (Win+H); Linux has none built in.

## Models and subscriptions

### Subscription usage (`agents usage`)

`agents usage [-provider <id>]` queries every provider account the machine is logged into (found from the credentials each harness already stores) and shows the plan, rolling windows, monthly allowances, spend and credits; one account seen through several harnesses is one row. Credentials are read into memory only, never printed, logged or cached; expired tokens are not refreshed (open the harness once to refresh). The usage cache (`usage.json` in the agents cache directory) holds numbers, plan names and labels, no credentials; `agents doctor` only reads it. How accounts are found per harness and profile directory: [cli.md → usage and doctor](https://github.com/giocaizzi/agents/blob/main/docs/cli.md#usage-and-doctor).

Runbook: check capacity before fanning out agents.

- **When:** before starting several parallel agents or a long run.
- **Steps:** `agents usage`; read the bars (green below 70%, yellow from 70%, red from 90%) and the reset times; back off or pick another provider near 90%.
- **Verify:** every account you use shows a bar and a reset time, with a fresh age.
- **Troubleshooting:** "token expired" or "unauthorized": open the harness named on that row once, then run `agents usage` again. "managed by org" or "not visible" is expected for a team seat or a cloud or gateway mode. "unavailable: timeout" on every HTTP account while `curl` reaches the same hosts: a per-app network filter (e.g. Little Snitch) blocks `agents`; allow it. One account twice or "same account? unknown": an unproven identity (expired or rejected token); open the harness named on the row, then run it again.
- **Rollback:** nothing to undo; delete the cache file to forget old snapshots.

### Doctor (`agents doctor`)

`agents doctor [-q] [-json] [-harness <list>]` warns, offline and without changing anything, when a newer or better model or config is available for what you have. It reads the config, the model catalog cache and the usage cache; no network, no credentials. It checks the models of the assigned tiers against the catalog (retired, renamed or untracked models, tier drift), suggests a better, newer or cheaper model of the same tier that the harness can use (following the billing of the account that would pay for it), and warns on a stale catalog (older than 30 days), on capacity at 90% or more, and on a configured model whose provider has no logged-in account. Without a usage cache the account, capacity and suggestion checks are skipped. Exit 0 when it ran, 1 on an unexpected error, 2 on a flag error. Output markers: `!` warning, `↑` suggestion, `i` info.

**Hook:** `run_after_agents-doctor.{sh,ps1}.tmpl` run `agents doctor -q` after every `chezmoi apply` and `chezmoi update` on machines with a harness selected (never light machines). `-q` prints nothing when there is nothing to report, otherwise at most a short head; run the plain command for everything. It never prompts, never fails or blocks the apply, and a binary older than the doctor is skipped silently. Skip it for one run with `chezmoi apply --exclude=scripts`.

Runbook: a doctor warning.

- **When:** `agents doctor` (or the hook after an apply) prints a `!` warning or a `↑` suggestion.
- **Steps**, by kind:
    - *Unknown or drifted model:* edit the tier the line names in `[models.tiers.<name>]` (`dots agents/config`), picking a model of the right tier from `agents models`; `chezmoi apply`.
    - *Suggestion:* advice only. If the model is worth the change, edit the tier's model for that harness and apply; if not, leave it.
    - *Stale catalog:* `agents models sync` (network), then `agents models` to review the tiers.
    - *Capacity:* back off, or use the account the info hint names; check with `agents usage`.
    - *No account:* log in to the named harness, then `agents usage`.
- **Verify:** after `agents usage`, `agents doctor` no longer prints the warning.
- **Rollback:** the doctor changes nothing; revert the config edit in git and apply again.

## Runbook: agent configuration

Every change to the agent setup is an edit of the config or of a layer, then a chezmoi apply, which runs [`agents apply`](#agents-apply). Preconditions: a machine with a harness selected, the `agents` binary installed (the first `chezmoi apply` installs it) and, for a private layer, the age key.

### Change the agents config

1. `dots agents/config` opens `.chezmoitemplates/agents-config.toml`, shows what the edit changes on every OS and role, and offers to apply. The entry has no check (a template cannot be piped to `agents check` without a shell on every OS): `make lint` renders it and runs `agents check`, and the apply script checks the rendered file before `agents apply`. By hand: edit the file, then `make check`.
2. Preview on this machine: `chezmoi apply --exclude=scripts` deploys the config only; `agents apply -n` prints the diff each harness file would get.
3. `chezmoi apply`: deploys the config, then the script runs `agents apply` because the config hash changed.

**Verify:** `agents check -live` prints `config ok`; open a new session of the harness and inspect its effective settings (an open session keeps its launch settings).
**Rollback:** `git revert` the change and apply again; keys that leave the config are removed from the live files.

### Add or change an instruction layer

1. Change a layer: `dots agents/<core|env|personal|work>`.
2. Add a layer: create `home/dot_config/agents/instructions/<name>.md`, declare it with `[instructions.layers.<name>]` and list it in the `instructions` of the profiles that use it (`dots agents/config`); add the file to the hash list of both `run_onchange_after_zz-agents-apply.*` scripts and an entry for it in `.chezmoidata/edit.toml` (`chezmoi dots lint registry` fails without one). A private layer is committed age-encrypted (`chezmoi add --encrypt`), and `.chezmoiignore` must skip it on the other role.
3. `agents prompt matrix` and `agents prompt check`, then `agents prompt show -profile <p> -harness <h>` for the composed text.
4. `chezmoi apply`.

**Verify:** the head of `~/.claude/CLAUDE.md` (and the other instruction files) shows the change.
**Rollback:** `git revert` and apply.

### Change a model tier or an assignment

1. `dots agents/config`: edit `[models.tiers.<name>]` (a harness's model id, or `effort`) or `[models.assign]` (agent id or glob to tier name; the tier must exist).
2. `chezmoi apply`.
3. `agents models tiers` shows the effective model per harness.

**Verify:** `agents models tiers` as expected and `agents doctor` has no model warning.
**Rollback:** `git revert` and apply.

### Add a plugin, an MCP server or a language server

- *Plugin:* a key under `[plugins.enabled]` listing its harnesses (it must be a directory of the `giocaizzi/harness` registry); `chezmoi apply`; `agents plugins list` shows it installed.
- *MCP server, every machine:* an entry under `[mcp.servers]`; *personal machines:* `[profiles.personal.mcp.servers]`; *the `~/ops` folder:* `[profiles.personal.mcp.workspaces.ops.servers]`. A private part of a URL is a `${VAR}` reference to a secret exported by `secrets.sh`, never the value. `chezmoi apply`, then sign in once per machine and harness ([MCP servers](#mcp-servers)).
- *MCP server with a pre-registered OAuth client:* register the redirect URIs on the OAuth client ([cli.md](https://github.com/giocaizzi/agents/blob/main/docs/cli.md#oauth-redirect-hosts)); put the client id in Bitwarden Secrets Manager as `MCP_<NAME>_CLIENT_ID` and open a new shell; add `oauth = { client_id = "${MCP_<NAME>_CLIENT_ID}", callback_port = <port> }` to the server; `chezmoi apply`; set the client secret once per machine, harness and server from the password vault (never in the config); sign in to one server at a time (they share the callback port).
- *Remove a server:* delete the entry and apply; `agents` removes the key it wrote. Check `claude mcp list` and the other `mcp list` commands.
- *Language server:* a `[lsp.servers.<name>]` entry (command, extensions, optional `format`) and, if chezmoi installs the binary, its `packages.toml` row; it is configured only where the command is on `PATH`.

**Verify:** `agents check -live` and the harness's own list command.
**Rollback:** remove the entry and apply.

### Troubleshoot

| Symptom | Do |
|---|---|
| A harness does not show a change | `agents check -live`, then `agents apply -n`; an open session keeps its launch settings: start a new one |
| `refused <path>` | the target is a symlink; replace it with a regular file, then `agents apply` ([cli.md](https://github.com/giocaizzi/agents/blob/main/docs/cli.md#recover-from-drift-or-a-refused-symlink)) |
| `agents apply: ... not installed yet, skipped` | the package install has not run; `chezmoi apply` again |
| `agents apply failed` warning | run `agents apply` by hand to see the cause; `agents check` for config errors |
| A `${VAR}` or server skipped with a warning | the secret is not exported yet: open a new shell and apply again |
| A language server missing | its command is not on `PATH`; install it or open a new shell |
| Doctor warning or suggestion | [Doctor](#doctor-agents-doctor) |
| Plugins stale | `agents plugins list`, then `agents plugins update` |

## Source layout (`.chezmoiroot`)

`.chezmoiroot` makes `home/` the source state: every source path in these docs (`dot_config/…`, `.chezmoidata/…`) is relative to it, and `chezmoi source-path` prints `~/.local/share/chezmoi/home`. The repo root holds what is never deployed (docs, `scripts/`, `tools/`, `Makefile`, agent config), so `.chezmoiignore` has no repo-meta block, and a git worktree anywhere in the repo (e.g. `.claude/worktrees/<slug>`) sits outside the source state: chezmoi never reads its `.chezmoidata`.

**After pulling the move:** run `chezmoi init` once (it asks nothing new): it drops the old `[hooks.apply.pre]` worktree guard from this machine's config.

## Testing the ignore matrix

```sh
chezmoi execute-template \
  --override-data '{"light":false,"work":false,"harnesses":["claude"],"chezmoi":{"os":"darwin"}}' \
  < home/.chezmoiignore   # vary os (darwin/linux/windows), light, work, harnesses
```

For the live OS, `chezmoi managed` and `chezmoi ignored` show the actual split.
