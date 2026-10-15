# OS-Specific Configuration

How each managed file maps to its destination per OS. Canonical sources live under `dot_config/` in the repo; OS-specific destinations symlink to them via `{{ .chezmoi.sourceDir }}/...`.

## Destination matrix

| Canonical source (in repo)                          | macOS                                                       | Linux                                          | Windows                                                                            |
| ---------------------------------------------------- | ----------------------------------------------------------- | ---------------------------------------------- | ---------------------------------------------------------------------------------- |
| `dot_config/Code/User/{settings,keybindings,mcp}.json` | `~/Library/Application Support/Code/User/` (symlinks)      | `~/.config/Code/User/` (direct)                | `%APPDATA%\Code\User\` (symlinks)                                                |
| same + Insiders                                      | `~/Library/.../Code - Insiders/User/` (symlinks → stable)   | `~/.config/Code - Insiders/User/` (symlinks → stable) | `%APPDATA%\Code - Insiders\User\` (symlinks → stable)                       |
| `dot_config/agents/AGENTS.md`                        | `~/.config/agents/AGENTS.md`                                | `~/.config/agents/AGENTS.md`                   | `~/.config/agents/AGENTS.md`                                                       |
| `dot_config/ghostty/config`                          | `~/Library/.../com.mitchellh.ghostty/` (symlink)            | `~/.config/ghostty/config` (direct)            | — (no Windows build)                                                               |
| `dot_config/windows-terminal/settings.json` *        | —                                                           | —                                              | `%LOCALAPPDATA%\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\` (symlink) |
| `dot_config/powershell/profile.ps1` *                | —                                                           | —                                              | `<Documents>/{PowerShell,WindowsPowerShell}/profile.ps1` (symlinks, created by `run_after_link-pwsh-profile.ps1`) |
| `dot_copilot/modify_private_settings.json` (merge)   | `~/.copilot/settings.json`                                  | `~/.copilot/settings.json`                     | `~/.copilot/settings.json`                                                         |
| `dot_claude/symlink_CLAUDE.md.tmpl`                  | `~/.claude/CLAUDE.md` → `~/.config/agents/AGENTS.md`        | same                                           | same                                                                               |
| `dot_copilot/symlink_copilot-instructions.md.tmpl`   | `~/.copilot/copilot-instructions.md` → `~/.config/agents/AGENTS.md` | same                                    | same                                                                               |
| `dot_codex/AGENTS.md.tmpl`, `dot_config/opencode/AGENTS.md.tmpl`, `dot_pi/agent/AGENTS.md.tmpl` (`.chezmoitemplates/agents-instructions.md`) | `~/.codex/AGENTS.md`, `~/.config/opencode/AGENTS.md`, `~/.pi/agent/AGENTS.md`: `AGENTS.md` with `AGENTS.private.md` inlined | same | same |
| `.chezmoitemplates/{claude,copilot}-settings.json` → `giocaizzi-harness` plugins | — | — | enabled only (`enabledPlugins` + `extraKnownMarketplaces`, repo `giocaizzi/harness`); each harness installs and updates them from the marketplace. `guard` holds the agent hooks (pip deny, `make check` on stop, post-edit lint, skill/title/host/risky guards, keep-awake) for Claude Code and Copilot CLI on every OS; `hud` the status line (both CLIs) and the Claude band and pane; `sounds` the Claude toasts and sounds; full machines only |
| `.chezmoidata/shortcuts.toml`                        | `~/.config/shell/aliases.sh`                                | `~/.config/shell/aliases.sh`                   | `~/.config/powershell/aliases.ps1` (+ `~/.config/shell/aliases.sh` for Git Bash)   |
| `dot_config/git/config.tmpl`                         | `~/.config/git/config`                                      | `~/.config/git/config`                         | `~/.config/git/config`                                                             |
| `dot_config/oh-my-posh/theme.omp.json`               | `~/.config/oh-my-posh/theme.omp.json`                       | same                                           | same                                                                               |
| `dot_config/vim/vimrc`                               | `~/.config/vim/vimrc`                                       | same                                           | `~/vimfiles/vimrc` (symlink)                                                       |
| `dot_config/git/{ignore,commit-template}`            | `~/.config/git/`                                            | same                                           | same                                                                               |
| `dot_profile`, `dot_bashrc`, `dot_bash_profile`      | `~/.profile`, `~/.bashrc`, `~/.bash_profile`                | same                                           | same (used by Git Bash / WSL)                                                      |
| `dot_zshrc`, `dot_zprofile`                          | `~/.zshrc`, `~/.zprofile`                                   | — (ignored)                                    | — (ignored)                                                                        |
| `dot_claude/modify_settings.json` (merge)            | `~/.claude/settings.json`                                   | same                                           | same                                                                               |
| `dot_config/git/personal.tmpl` (work machines only)  | `~/.config/git/personal`                                    | same                                           | same                                                                               |
| `dot_config/readline/inputrc`                        | `~/.config/readline/inputrc` (via `$INPUTRC`)               | same                                           | same (Git Bash)                                                                    |
| `dot_config/tmux/tmux.conf`                          | `~/.config/tmux/tmux.conf`                                  | same                                           | — (no tmux)                                                                        |
| `private_dot_ssh/encrypted_private_config.age`       | `~/.ssh/config` (decrypted, owner-only)                     | same                                           | same                                                                               |

\* Symlink-target-only sources — never deployed as regular files anywhere (listed in `.chezmoiignore` unconditionally); only referenced via `{{ .chezmoi.sourceDir }}/...` from the Windows-specific symlinks.

## Bootstrap scripts (`.chezmoiscripts/`)

| Script                                       | macOS | Linux | Windows |
| -------------------------------------------- | :---: | :---: | :-----: |
| `run_once_init-untracked-env.sh` → `~/.config/shell/secrets.sh`             |  ✓    |  ✓    |   —     |
| `run_once_init-untracked-env.ps1` → `~/.config/powershell/secrets.ps1`      |  —    |  —    |   ✓     |
| `run_once_after_create-xdg-dirs.sh` → creates `$XDG_STATE_HOME/{zsh,bash,vim}`, `$XDG_CACHE_HOME/zsh` |  ✓    |  ✓    |   —     |
| `run_onchange_install-pkgs.sh.tmpl`          |  ✓    |  ✓    |   —     |
| `run_after_link-pwsh-profile.ps1` → resolves the real Documents folder (OneDrive/locale) and links the profile |  —    |  —    |   ✓     |
| `run_onchange_install-pkgs.ps1.tmpl`         |  —    |  —    |   ✓     |

The two install scripts keep the same tool list (brew on macOS, apt or official release binaries in `~/.local/bin` on Linux, winget on Windows). Full machines get the core CLIs (`fzf jq git vim fd gh tmux shellcheck shfmt`), uv + Python, oh-my-posh, Node, the language servers and the linters `gitleaks`, `actionlint`, `hadolint`; [Machine roles](#machine-roles) lists what `light` and `work` change.

## Externals (`.chezmoiexternal.toml.tmpl`)

Declarative upstream sources cloned/fetched by chezmoi on `apply`, refreshed per `refreshPeriod`. Prefer over an install-script `git clone` whenever the upstream is "drop this repo at this path" (no build, no PATH wiring).

| Entry (target path)                                    | Upstream                                  | macOS | Linux | Windows |
| ------------------------------------------------------ | ----------------------------------------- | :---: | :---: | :-----: |
| `~/.config/vim/pack/catppuccin/start/catppuccin/`      | `github.com/catppuccin/vim`               |  ✓    |  ✓    |   —     |
| `~/vimfiles/pack/catppuccin/start/catppuccin/`         | `github.com/catppuccin/vim`               |  —    |  —    |   ✓     |
| `~/.local/share/fonts/JetBrainsMonoNerdFont/`          | Nerd Fonts `JetBrainsMono.tar.xz` release |  —    |  ✓ (not light) |   —     |
| `~/.oh-my-zsh/`                                        | `github.com/ohmyzsh/ohmyzsh`              |  ✓    |  —    |   —     |

**When NOT to use externals:** anything needing `chmod +x` on a downloaded binary with arch detection (oh-my-posh), official installer scripts (uv), or package-manager registration (brew/apt/winget). Those stay in `.chezmoiscripts/`.

## Asymmetries (by design)

- **Ghostty** has no Windows build → Windows Terminal fills the gap with matching theme/font.
- **zsh** is ignored on Linux and Windows (bash is the Unix default; PowerShell is the Windows default).
- **PowerShell profile** is Windows-only; trivially extendable to Unix by mirroring the `dot_config/powershell/` snippet pattern.

## Machine roles

Two independent yes/no answers, asked once per machine by `chezmoi init` as a choice (arrow keys to pick `yes` or `no`, Enter to confirm; default `no`) and stored as booleans only in that machine's `~/.config/chezmoi/chezmoi.toml` (`data.light`, `data.work`). A config that already holds the boolean is not asked again. Templates read them with `get . "light"` / `get . "work"`, which treats a missing key as false.

| Flag | Prompt | When true |
|---|---|---|
| `light` | Light install (yes/no) — shell, git, vim, prompt only | `.chezmoiignore` skips VS Code, Ghostty, Claude Code, Copilot, Codex, opencode, pi, AGENTS.md and the encrypted SSH config (so light machines need **no age key**); no Nerd Font download; the install script installs only `fzf git vim fd tmux jq shellcheck shfmt` + oh-my-posh (no uv/Python, Node, language servers, linters or forge CLIs). |
| `work` | Work computer (yes/no) — work git identity, no personal telemetry | Asks the work email, which becomes the default `user.email` (`gitEmail` stays the personal one), the **personal GitHub username** (`githubUser`) and optional **extra personal remote URL globs** (`personalGitRemote`, for other orgs or hosts; blank = none). The username expands to the four github.com URL forms (`git@github.com:<user>/**`, `ssh://git@github.com/<user>/**`, `https://github.com/<user>/**`, `https://*@github.com/<user>/**`); each of those and each extra glob becomes `includeIf "hasconfig:remote.*.url:<glob>"` (git ≥ 2.36) → `~/.config/git/personal`, so repos in any folder with a matching remote commit with the personal email. Everything else, including repos with no remote, uses the work email. In an extra glob `*` stays within one path segment: write `https://git.example.com/**`, `https://*@git.example.com/**` (URL with user), `git@git.example.com:*/**` (scp-like) and `git@<ssh-alias>:*/**` for an SSH `Host` alias. To add more later see [Adding personal remotes](#adding-personal-remotes). Keep the SSH key consistent with the identity through `~/.ssh/config.d/*` (e.g. `Host github.com` → personal key, a work alias → work key). Claude Code sends no telemetry to the OTel endpoint (the keys are left out and stripped from the live file). The install scripts add the GitLab and Azure DevOps CLIs: `glab` and Azure CLI (`az`, brew `azure-cli`, winget `Microsoft.AzureCLI` machine-scope MSI) plus its `azure-devops` extension. |

Full personal machines are also asked for the Claude Code OTLP endpoint (`otelEndpoint`, local config only); blank omits the telemetry keys.

Personal agent instructions and shell aliases are not in this public repo: they are age-encrypted sources that render to `~/.config/agents/AGENTS.private.md` (referenced from `AGENTS.md`, imported by Claude Code) and `~/.config/shell/aliases.private.sh` (sourced by `aliases.sh` when present). Light machines skip both. VS Code trusted domains for private hosts are set per machine in the local user settings, not in the shared file.

Current machines: Mac = full/personal, Raspberry Pi = light/personal, Windows PC = full/work.

Change an answer later: `chezmoi edit-config` and set `light` / `work` under `[data]` to `true` or `false`, or re-answer every prompt with `chezmoi init --prompt`; then `chezmoi apply`.

Re-running `chezmoi init`:

- Plain `chezmoi init` asks only for keys missing from the local config (e.g. `harnesses` on a machine set up before that prompt existed) and keeps every stored answer.
- `chezmoi init --prompt` asks every prompt again, each prefilled with the stored answer (text as the default, yes/no and the harness list preselected), so Enter keeps it. `chezmoi init --prompt --promptDefaults` re-renders the config with every stored answer and no questions.

### AI harnesses

Full machines are also asked which AI harnesses chezmoi writes configs for: a multiple choice of `claude`, `copilot`, `vscode`, `codex`, `opencode`, `pi` (arrow keys to move, space to toggle, Enter to confirm; all six preselected), stored as `data.harnesses`. chezmoi never installs the CLIs; it only writes their configs. Light machines are not asked and get none. A config without the key (created before the prompt existed) gets all six; the selection lives in `.chezmoitemplates/harnesses`, which every consumer reads.

| Harness | Managed only when selected |
|---|---|
| `claude` | `~/.claude` (`settings.json` merge, `CLAUDE.md` symlink) |
| `copilot` | `~/.copilot` (`settings.json` merge, `copilot-instructions.md` symlink) |
| `vscode` | VS Code and VS Code Insiders user settings, keybindings and `mcp.json` (`~/.config/Code*` on Linux, `~/Library/Application Support/Code*` on macOS, `%APPDATA%\Code*` on Windows). The whole files, not only the Copilot keys. `chezmoi apply` also installs the missing extensions listed in `.chezmoitemplates/vscode-extensions` through the `code` CLI: a shared list plus a `work` block (PowerShell only on Windows) or a personal block. Names only, so VS Code auto-updates them; add-only, so uninstall a dropped extension by hand. It warns and skips when the `code` CLI is unavailable. |
| `codex` | `~/.codex/AGENTS.md`; Codex wiring in `run_after_install-harness.*` |
| `opencode` | `~/.config/opencode/AGENTS.md`; opencode wiring in `run_after_install-harness.*` |
| `pi` | `~/.pi/agent/AGENTS.md`; pi wiring in `run_after_install-harness.*` |

Shared: `~/.config/agents/` (`AGENTS.md`, `AGENTS.private.md`) stays while any of `claude`, `copilot`, `codex`, `opencode`, `pi` is selected; the `giocaizzi-harness` external and the `run_after_install-harness.*` scripts while any of `codex`, `opencode`, `pi` is.

Add or remove a harness later: `chezmoi edit-config` and edit the list under `[data]`, e.g. `harnesses = ["claude", "copilot", "vscode"]` (or re-answer every prompt with `chezmoi init --prompt`), then `chezmoi apply`. Removing a harness only stops chezmoi managing its files: the files already in the home directory stay; delete them by hand if they are unwanted.

Simulate the prompts without touching the real config (prompt flags are keyed by the prompt text, as printed by `chezmoi init`):

```sh
chezmoi execute-template --init \
  --promptChoice "[ INSTALL MODE ] Light setup? (shell, git, vim, and prompt only)=no" \
  --promptMultichoice "[ AI HARNESSES ] Which harness configs to manage? (space toggles; Enter confirms; chezmoi never installs the CLIs)=claude/copilot" \
  < .chezmoi.toml.tmpl
```

`execute-template` keeps the answer already in your config for a key it finds there (and, unlike `chezmoi init`, re-asks a stored `harnesses` list); point `--config` at an empty `.toml` file to simulate a new machine.

### Adding personal remotes

On a work machine, add another personal org or host by editing the local config (never the repo): in `~/.config/chezmoi/chezmoi.toml` under `[data]`, append a glob to `personalGitRemote` (comma-separated), then `chezmoi apply`:

```toml
personalGitRemote = "https://ghe.example.com/**,git@ghe.example.com:*/**,git@github.com:other-org/**,https://github.com/other-org/**"
```

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
- **One token per machine**, created as a Bitwarden machine account with read access to one project, stored only in that machine's OS keyring (Windows Credential Manager, macOS Keychain, Linux Secret Service), never in `chezmoi.toml`: `chezmoi secret keyring set --service=bws --user=access-token`. `chezmoi.toml` keeps only the project ID, which switches the feature on. Give the token an expiry date, and revoke a single machine's token in Bitwarden if it's lost. The token never goes on a command line (chezmoi echoes failed commands, which would print it): the template runs `bws` through `powershell -Command` / `sh -c`, which reads it with `chezmoi secret keyring get` and passes it as `BWS_ACCESS_TOKEN`. A project ID without a stored token fails with `secret not found in keyring`; headless Linux without a Secret Service keyring can't use this, so leave the project ID blank and edit the secrets file by hand.
- **`bws` is fetched** by `.chezmoiexternal.toml.tmpl` into `~/.local/bin` (version pinned in `.chezmoidata/bws.toml`; bump to update). The first apply on a new machine fetches it; the next apply renders the secrets.
- **Online needed:** on machines with a project ID, every `chezmoi status`/`diff`/`apply` calls `bws` once and fails offline.
- No Bitwarden MCP server: it has no read-only mode and would give an AI agent full vault access.
- **Claude Code TUI:** `tui` is `fullscreen` everywhere (needed for `viewMode: focus`, clicks and the side-by-side `/diff`, which also needs Claude Code ≥ 2.1.287 and a terminal ≥ 110 columns). macOS needs Ghostty mouse reporting on. On Windows Terminal, if stale text fragments appear, set `CLAUDE_CODE_ALT_SCREEN_FULL_REPAINT=1`; `CLAUDE_CODE_SCROLL_SPEED=3` is already set there.
- **Claude Code and `bw`:** every `bw` command Claude runs needs your approval (an `ask` rule in the managed Claude settings; ask rules beat the blanket Bash allow, also inside pipes and in auto mode). To let Claude act, unlock in your own terminal into a private file, `(umask 077; bw unlock --raw > <file>)`; Claude runs `bw --session "$(cat <file>)" …` and prints only non-secret fields. Finish with `bw lock` and delete the file. The rule matches the command as written, so it's an approval gate, not a security boundary.

## Encrypted files (age)

Secrets such as `~/.ssh/config` are committed age-encrypted (`encrypted_` in the source name). Each personal and work machine needs the private key at `~/.config/chezmoi/key.txt` (owner-only) **before** `chezmoi apply` — light machines don't, and must not hold it; copy it from another machine over a secure channel, never through the repo. The public key in `.chezmoi.toml.tmpl` is safe to publish. Add or update an encrypted file with `chezmoi add --encrypt <file>`, or edit one in place with `chezmoi edit --apply <file>` (decrypts to an editor, re-encrypts on save).

**SSH key names (same on every machine):** `~/.ssh/id_<type>_<service>_<account>` (+ `.pub`) — type first, then the service, then the account or organisation: `id_ed25519_github_<github-user>`, `id_ed25519_github_<work-account>`, `id_rsa_devops_<org>`. The shared `~/.ssh/config` points `Host github.com` at the personal GitHub key, so don't redefine `github.com` in `config.d/` (it would shadow the shared entry); rename the key instead. Work keys and their `Host` aliases (e.g. `github-<work>`, `azure-<org>`) live in `config.d/<name>`.

**SSH hosts per machine:** the shared `~/.ssh/config` starts with `Include config.d/*`. Hosts that belong to one machine only — e.g. work hosts on the corporate PC — go in `~/.ssh/config.d/<name>` there; chezmoi doesn't manage that folder, so they never enter this repo, not even encrypted. Work keys are generated on the work machine itself. If a work Git host is the same as a personal one, give the work account its own `Host` alias in `config.d/<name>` (own key, `IdentitiesOnly yes`) and clone work repos through that alias.

## Hardcoded by design

chezmoi target paths are literal (they can't be templated), so a few values stay fixed on purpose:

- **`~/.config`, `~/.local/{share,state}`, `~/.cache`** — chezmoi deploys to these exact paths, so `.profile` pins `XDG_*` to the spec defaults. Pointing `XDG_CONFIG_HOME` elsewhere would split config from where chezmoi writes it.
- **Homebrew prefixes** `/opt/homebrew` (Apple Silicon) and `/usr/local` (Intel) — fixed by Homebrew.
- **Vendor paths** — `Library/Application Support/…`, `AppData/…`, the Windows Terminal package family name.
- **`{{ .chezmoi.homeDir }}`** in templates — rendered per machine.
- **Personal data** — plugin marketplaces, the OTel endpoint (personal machines only), project aliases. Git identity is prompted at `chezmoi init`.

Anything that genuinely differs per machine and can't be templated (e.g. a redirected Documents folder) is resolved at runtime by a script.

## Files merged, not replaced

`~/.claude/settings.json` and `~/.copilot/settings.json` are also written by the apps themselves, so chezmoi doesn't own them whole. `modify_` templates overlay `.chezmoitemplates/{claude,copilot}-settings.json` onto the live file: managed keys win (Claude: merged with jq when installed, so the live file's key order survives and `chezmoi diff` stays quiet), keys the app writes stay local and never reach this public repo (Claude `autoMode`, Copilot `model`; Copilot `allowedUrls` is the union of both). Removing a managed key from a template doesn't delete it from the live file; delete it there once (except the telemetry keys on work machines, which are stripped automatically).

## Managed (corporate) Windows PCs

Endpoint policies on the work PC block several defaults; the workarounds apply to every Windows machine:

| Blocked | Workaround |
|---|---|
| Machine-wide installs (winget font package) | JetBrainsMono installed per user by `oh-my-posh font install` |
| `Set-Acl` (needs SeSecurityPrivilege) | `icacls` restricts `secrets.ps1` |
| uv's `python.exe` launchers in `~\.local\bin` (unsigned exe named python.exe) | `uv python install --no-bin`; `profile.ps1` puts uv's `cpython-3.NN-*` interpreter folder on PATH |
| `New-Item -ItemType SymbolicLink` in Windows PowerShell 5.1 without admin | `mklink` (honours Developer Mode) in `run_after_link-pwsh-profile.ps1` |
| Documents redirected to OneDrive and localised | profile link target resolved at runtime with `GetFolderPath('MyDocuments')` |

## Agent plugins (`giocaizzi-harness`)

Plugins come from the `giocaizzi/harness` marketplace, not from chezmoi files: the settings templates only declare the marketplace and enable the plugins, and Claude Code / Copilot CLI fetch them. Codex, opencode and pi have no settings that do this, so `run_after_install-harness.{sh,ps1}.tmpl` (full machines with any of them selected, see [AI harnesses](#ai-harnesses); every apply, idempotent) wires them from a checkout kept by the `.local/share/giocaizzi-harness` external: Codex adds the marketplace and plugins (`codex plugin add`) and gets the agents from `scripts/install.mjs codex`; opencode gets skills, agents and the guard plugin stub from `scripts/install.mjs opencode`; pi installs the checkout as a package, so the external's daily pull updates it. chezmoi does not install the three CLIs: the script wires up only those already on PATH. Codex skips plugin hooks until trusted: run `/hooks` once per `guard` version. Runbook: [install.md](https://github.com/giocaizzi/harness/blob/main/docs/install.md). Hooks (guards, `make check` on stop, post-edit lint, keep-awake) ship in its `guard` plugin; runbook: [guard.md](https://github.com/giocaizzi/harness/blob/main/docs/guard.md).

```sh
claude plugin marketplace add giocaizzi/harness
claude plugin install guard@giocaizzi-harness   # likewise hud, sounds, python, api, javascript, dev-tools
copilot plugin marketplace add giocaizzi/harness
copilot plugin install guard@giocaizzi-harness  # likewise hud, python, api, javascript, dev-tools
```

Managed settings per harness:

- Claude Code: marketplace and `enabledPlugins` (`.chezmoitemplates/claude-settings.json`).
- Claude Code status line: `hud`'s `hooks/statusline.mjs claude`, run from the marketplace clone (`~/.claude/plugins/marketplaces/giocaizzi-harness`, a path that does not change with the plugin version); `node` 22+ and a Nerd Font are required. It replaces the old `statusline-command.sh`; delete `~/.claude/statusline-command.sh` once on each machine.
- Copilot CLI: marketplace, `enabledPlugins` and `statusLine` (`.chezmoitemplates/copilot-settings.json`). The status line runs `hud`'s `hooks/statusline.mjs` (`ctx N% · cache N%`) from `~/.copilot/installed-plugins/giocaizzi-harness/hud`; Copilot expands `~` itself on every OS, and `node` 22+ must be on PATH. Copilot CLI has no prompt-cache TTL or auto-compact setting. `subagents.agents` pins the harness agents to their tiers, keyed `<plugin>:<agent>`: `dev-tools:scout` GPT-6 Luna at `low`, `dev-tools:escalate` GPT-6 Astra at `high` (as Claude's `effort`). A model the account lacks falls back to an available one, so Astra runs as Luna until the account gets it.
- Codex, opencode, pi: global instructions are `AGENTS.md` rendered with `AGENTS.private.md` inlined where Claude's `@` import sits, since none of them expands `@file` (`.chezmoitemplates/agents-instructions.md`). Rendered, not symlinked: an edit to `AGENTS.md` reaches them on the next `chezmoi apply`.
- VS Code Copilot (`dot_config/Code/User/settings.json`, all OSes): `chat.plugins.marketplaces` lists `giocaizzi/harness` and `chat.useHooks` enables the `guard` hooks. Plugins are installed from `@agentPlugins` in the Extensions view (or picked up from Copilot CLI's installs); VS Code has no status-line hook.

## Testing the ignore matrix

```sh
chezmoi execute-template \
  --override-data '{"light":false,"work":false,"harnesses":["claude"],"chezmoi":{"os":"darwin"}}' \
  < .chezmoiignore   # vary os (darwin/linux/windows), light, work, harnesses
```

For the live OS, `chezmoi managed` and `chezmoi ignored` show the actual split.
