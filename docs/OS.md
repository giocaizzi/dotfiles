# OS-Specific Configuration

How each managed file maps to its destination per OS. Canonical sources live under `dot_config/` in the source state (`home/`, [Source layout](#source-layout-chezmoiroot)); OS-specific destinations symlink to them via `{{ .chezmoi.sourceDir }}/...`.

## Destination matrix

| Canonical source (in repo)                          | macOS                                                       | Linux                                          | Windows                                                                            |
| ---------------------------------------------------- | ----------------------------------------------------------- | ---------------------------------------------- | ---------------------------------------------------------------------------------- |
| `dot_config/Code/User/{settings,keybindings}.json` | `~/Library/Application Support/Code/User/` (symlinks)      | `~/.config/Code/User/` (direct)                | `%APPDATA%\Code\User\` (symlinks)                                                |
| `.chezmoitemplates/vscode-mcp.json` (`mcp.json.tmpl` in each Code / Code - Insiders `User/` folder) | `~/Library/Application Support/Code*/User/mcp.json` (rendered) | `~/.config/Code*/User/mcp.json` (rendered) | `%APPDATA%\Code*\User\mcp.json` (rendered) |
| same + Insiders                                      | `~/Library/.../Code - Insiders/User/` (symlinks → stable)   | `~/.config/Code - Insiders/User/` (symlinks → stable) | `%APPDATA%\Code - Insiders\User\` (symlinks → stable)                       |
| `.chezmoitemplates/agents-instructions.md` (layers, [Agent instructions](#agent-instructions)) | `~/.claude/CLAUDE.md`, `~/.copilot/copilot-instructions.md`, `~/.codex/AGENTS.md`, `~/.config/opencode/AGENTS.md`, `~/.pi/agent/AGENTS.md`, `~/.gemini/config/AGENTS.md` (rendered) | same | same |
| `dot_config/ghostty/config`                          | `~/Library/.../com.mitchellh.ghostty/` (symlink)            | `~/.config/ghostty/config` (direct)            | — (no Windows build)                                                               |
| `dot_config/windows-terminal/settings.json` *        | —                                                           | —                                              | `%LOCALAPPDATA%\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\` (symlink) |
| `dot_config/powershell/profile.ps1` *                | —                                                           | —                                              | `<Documents>/{PowerShell,WindowsPowerShell}/profile.ps1` (symlinks, created by `run_after_link-pwsh-profile.ps1`) |
| `dot_copilot/modify_private_settings.json` (merge)   | `~/.copilot/settings.json`                                  | `~/.copilot/settings.json`                     | `~/.copilot/settings.json`                                                         |
| `.chezmoitemplates/{claude,copilot}-settings.json` → `giocaizzi-harness` plugins | — | — | enabled only (`enabledPlugins` + `extraKnownMarketplaces`, repo `giocaizzi/harness`); each harness installs and updates them from the marketplace. `guard` holds the agent hooks (pip deny, `make check` on stop, post-edit lint, skill/title/host/risky guards, keep-awake) for Claude Code and Copilot CLI on every OS; `hud` the status line (both CLIs) and the Claude band and pane; `sounds` the Claude toasts and sounds; full machines only |
| `.chezmoidata/aliases.toml` (+ `.aliases.private.toml.age`) | `~/.config/shell/aliases.sh`, `~/.config/shell/omz-aliases.zsh` | `~/.config/shell/aliases.sh`                   | `~/.config/powershell/aliases.ps1` (+ `~/.config/shell/aliases.sh` for Git Bash)   |
| `dot_config/git/config.tmpl`                         | `~/.config/git/config`                                      | `~/.config/git/config`                         | `~/.config/git/config`                                                             |
| `dot_config/oh-my-posh/theme.omp.json`               | `~/.config/oh-my-posh/theme.omp.json`                       | same                                           | same                                                                               |
| `dot_config/vim/vimrc`                               | `~/.config/vim/vimrc`                                       | same                                           | `~/vimfiles/vimrc` (symlink)                                                       |
| `dot_config/git/{ignore,commit-template}`            | `~/.config/git/`                                            | same                                           | same                                                                               |
| `dot_profile`, `dot_bashrc`, `dot_bash_profile`      | `~/.profile`, `~/.bashrc`, `~/.bash_profile`                | same                                           | same (used by Git Bash / WSL)                                                      |
| `dot_zshrc`, `dot_zprofile`                          | `~/.zshrc`, `~/.zprofile`                                   | — (ignored)                                    | — (ignored)                                                                        |
| `dot_claude/modify_settings.json` (merge)            | `~/.claude/settings.json`                                   | same                                           | same                                                                               |
| `dot_gemini/antigravity-cli/modify_private_settings.json` (merge) | `~/.gemini/antigravity-cli/settings.json`      | same                                           | same                                                                               |
| `dot_config/git/personal.tmpl` (work machines only)  | `~/.config/git/personal`                                    | same                                           | same                                                                               |
| `dot_config/readline/inputrc`                        | `~/.config/readline/inputrc` (via `$INPUTRC`)               | same                                           | same (Git Bash)                                                                    |
| `dot_config/tmux/tmux.conf`                          | `~/.config/tmux/tmux.conf`                                  | same                                           | — (no tmux)                                                                        |
| `private_dot_ssh/encrypted_private_config.age`       | `~/.ssh/config` (decrypted, owner-only)                     | same                                           | same                                                                               |

\* Symlink-target-only sources — never deployed as regular files anywhere (listed in `.chezmoiignore` unconditionally); only referenced via `{{ .chezmoi.sourceDir }}/...` from the Windows-specific symlinks.

## Bootstrap scripts (`.chezmoiscripts/`)

| Script | macOS | Linux | Windows |
| --- | :---: | :---: | :---: |
| `run_once_init-untracked-env.sh.tmpl` → `~/.config/shell/secrets.sh` | ✓ | ✓ | — |
| `run_once_init-untracked-env.ps1.tmpl` → `~/.config/powershell/secrets.ps1` | — | — | ✓ |
| `run_once_after_create-xdg-dirs.sh.tmpl` → `$XDG_STATE_HOME/{zsh,bash,vim}`, `$XDG_CACHE_HOME/zsh` | ✓ | ✓ | — |
| `run_onchange_after_install-pkgs.sh.tmpl` → the packages of [Packages](#packages) | ✓ | ✓ | — |
| `run_onchange_after_install-pkgs.ps1.tmpl` → the same, with winget | — | — | ✓ |
| `run_onchange_after_install-harness.*`, `run_onchange_after_claude-mcp-user.*`, `run_onchange_after_sync-vscode-extensions.*` | per [AI harnesses](#ai-harnesses) | | |
| `run_after_link-pwsh-profile.ps1.tmpl` → resolves the real Documents folder (OneDrive/locale) and links the profile | — | — | ✓ |

Each script states where it runs: its body is wrapped in a condition (OS; harness for the harness scripts), and elsewhere it renders empty, which chezmoi skips. No script is listed in `.chezmoiignore`. The `run_onchange_` scripts re-run when their rendered text changes, so each embeds what it depends on (the package lists, the harness checkout's commit, the extension list).

## Externals (`.chezmoiexternal.toml.tmpl`)

Declarative upstream sources cloned/fetched by chezmoi on `apply`, refreshed per `refreshPeriod`. Prefer over an install-script `git clone` whenever the upstream is "drop this repo at this path" (no build, no PATH wiring).

| Entry (target path)                                    | Upstream                                  | macOS | Linux | Windows |
| ------------------------------------------------------ | ----------------------------------------- | :---: | :---: | :-----: |
| `~/.config/vim/pack/catppuccin/start/catppuccin/`      | `github.com/catppuccin/vim`               |  ✓    |  ✓    |   —     |
| `~/vimfiles/pack/catppuccin/start/catppuccin/`         | `github.com/catppuccin/vim`               |  —    |  —    |   ✓     |
| `~/.local/share/fonts/JetBrainsMonoNerdFont/`          | Nerd Fonts `JetBrainsMono.tar.xz` release |  —    |  ✓ (not light) |   —     |
| `~/.oh-my-zsh/`                                        | `github.com/ohmyzsh/ohmyzsh`              |  ✓    |  —    |   —     |

Release binaries (`bws`, `gitleaks`, `actionlint`, `hadolint`, the Go toolchain, `goreleaser`, `syft`, `glab`, the `chezmoi-dots` plugin) are externals too, generated from the `[release.*]` tables of the [package registry](#packages) for this OS, architecture and role, each pinned and checksum-verified. `dotsRender` (an override-data flag set by `chezmoi dots` for previews) makes the file render empty, so previewing another OS or role never downloads its binaries.

**When NOT to use externals:** anything needing `chmod +x` on a downloaded binary with arch detection (oh-my-posh), official installer scripts (uv), or package-manager registration (brew/apt/winget). Those stay in `.chezmoiscripts/`.

## Asymmetries (by design)

- **Ghostty** has no Windows build → Windows Terminal fills the gap with matching theme/font.
- **zsh** is ignored on Linux and Windows (bash is the Unix default; PowerShell is the Windows default).
- **PowerShell profile** is Windows-only; trivially extendable to Unix by mirroring the `dot_config/powershell/` snippet pattern.

## Machine roles

Two independent yes/no answers, asked once per machine by `chezmoi init` as a choice (arrow keys to pick `yes` or `no`, Enter to confirm; default `no`) and stored as booleans only in that machine's `~/.config/chezmoi/chezmoi.toml` (`data.light`, `data.work`). A config that already holds the boolean is not asked again. Templates read them with `get . "light"` / `get . "work"`, which treats a missing key as false.

| Flag | Prompt | When true |
|---|---|---|
| `light` | Light install (yes/no) — shell, git, vim, prompt only | `.chezmoiignore` skips VS Code, Ghostty, Claude Code, Copilot, Codex, opencode, pi, the agent instructions ([no agents on light machines](#agent-instructions)) and the encrypted SSH config (so light machines need **no age key**); no Nerd Font download; the install script installs only `fzf git vim fd tmux jq shellcheck shfmt` + oh-my-posh (no uv/Python, Node, language servers, linters or forge CLIs). |
| `work` | Work computer (yes/no) — work git identity, no personal telemetry | Asks the work email, which becomes the default `user.email` (`gitEmail` stays the personal one), the **personal GitHub username** (`githubUser`) and optional **extra personal remotes** (`personalGitRemote`: github.com org names or URL globs, comma-separated; blank = none). The username and each org name (an entry of letters, digits and `-`) expand to the four github.com URL forms (`git@github.com:<owner>/**`, `ssh://git@github.com/<owner>/**`, `https://github.com/<owner>/**`, `https://*@github.com/<owner>/**`); each of those and each extra glob becomes `includeIf "hasconfig:remote.*.url:<glob>"` (git ≥ 2.36) → `~/.config/git/personal`, so repos in any folder with a matching remote commit with the personal email. Everything else, including repos with no remote, uses the work email. In an extra glob `*` stays within one path segment: write `https://git.example.com/**`, `https://*@git.example.com/**` (URL with user), `git@git.example.com:*/**` (scp-like) and `git@<ssh-alias>:*/**` for an SSH `Host` alias. To add more later see [Adding personal remotes](#adding-personal-remotes). Keep the SSH key consistent with the identity through `~/.ssh/config.d/*` (e.g. `Host github.com` → personal key, a work alias → work key). Claude Code sends no telemetry to the OTel endpoint (the keys are left out and stripped from the live file). No Go stack (`go`, `gopls`, `goreleaser`, `syft`, the Claude `gopls-lsp` plugin) and the personal-only harness plugins are disabled ([Agent plugins](#agent-plugins-giocaizzi-harness)). The install scripts add the GitLab and Azure DevOps CLIs: `glab` and Azure CLI (`az`, brew `azure-cli`, winget `Microsoft.AzureCLI` machine-scope MSI) plus its `azure-devops` extension. |

Full personal machines are also asked for the Claude Code OTLP endpoint (`otelEndpoint`, local config only); blank omits the telemetry keys.

Private agent instructions and personal shell aliases are not in this public repo: they are age-encrypted layers of the global agent instructions (`personal.md.age` on personal machines, `work.md.age` on work ones, see [Agent instructions](#agent-instructions)) and personal aliases in `.aliases.private.toml.age`, merged into the alias registry at render time (see [SHELL.md → Personal aliases](./SHELL.md#personal-aliases)). Light machines skip both. VS Code trusted domains for private hosts are set per machine in the local user settings, not in the shared file.

Current machines: Mac = full/personal, Raspberry Pi = light/personal, Windows PC = full/work.

Change an answer later: `chezmoi edit-config` and set `light` / `work` under `[data]` to `true` or `false`, or re-answer every prompt with `chezmoi init --prompt`; then `chezmoi apply`.

Re-running `chezmoi init`:

- Plain `chezmoi init` asks only for keys missing from the local config (e.g. `harnesses` on a machine set up before that prompt existed) and keeps every stored answer.
- `chezmoi init --prompt` asks every prompt again, each prefilled with the stored answer (text as the default, yes/no and the harness list preselected), so Enter keeps it. `chezmoi init --prompt --promptDefaults` re-renders the config with every stored answer and no questions.

### AI harnesses

Full machines are also asked which AI harnesses chezmoi writes configs for: a multiple choice of `claude`, `copilot`, `vscode`, `codex`, `opencode`, `pi`, `antigravity` (arrow keys to move, space to toggle, Enter to confirm; none preselected), stored as `data.harnesses`. chezmoi never installs the CLIs; it only writes their configs. Light machines are not asked and get none (even with the key set). A config without the key gets none: select a harness to get its config; the selection lives in `.chezmoitemplates/harnesses`, which every consumer reads. Each harness's targets are listed once, in `[harnessTargets]` (`.chezmoidata/harness.toml`): `.chezmoiignore` skips the targets of every harness not selected.

| Harness | Managed only when selected |
|---|---|
| `claude` | `~/.claude` (`settings.json` merge, `CLAUDE.md`); user-scope MCP servers (`run_onchange_after_claude-mcp-user.*`, [MCP servers](#mcp-servers)) |
| `copilot` | `~/.copilot` (`settings.json` and `mcp-config.json` merge, `copilot-instructions.md`) |
| `vscode` | VS Code and VS Code Insiders user settings, keybindings and `mcp.json` (`~/.config/Code*` on Linux, `~/Library/Application Support/Code*` on macOS, `%APPDATA%\Code*` on Windows). The whole files, not only the Copilot keys. Subagent nesting is on (`chat.subagents.allowInvocationsFromSubagents`), so a driver agent can spawn its own workers. Also `~/.copilot/copilot-instructions.md` (the global instructions, which VS Code reads from there), even without the `copilot` harness. `chezmoi apply` also installs the missing extensions listed in `.chezmoitemplates/vscode-extensions` through the `code` CLI: a shared list plus a `work` block (PowerShell only on Windows) or a personal block. Names only, so VS Code auto-updates them; add-only, so uninstall a dropped extension by hand. It warns and skips when the `code` CLI is unavailable. |
| `codex` | `~/.codex/AGENTS.md`, `~/.codex/config.toml` (merge: `[tui]` bell keys, user MCP servers, `~/ops` trust); Codex wiring in `run_onchange_after_install-harness.*` |
| `opencode` | `~/.config/opencode/AGENTS.md`, `opencode.jsonc` (merge: user MCP servers); opencode wiring in `run_onchange_after_install-harness.*` |
| `pi` | `~/.pi/agent/AGENTS.md`, `settings.json` (merge: `defaultProvider`, `defaultModel`, `defaultThinkingLevel`, `packages`), `mcp.json` (merge: user MCP servers); pi wiring in `run_onchange_after_install-harness.*` |
| `antigravity` | `~/.gemini/config/AGENTS.md`, `~/.gemini/config/mcp_config.json` (merge: user MCP servers), `~/.gemini/antigravity-cli/settings.json` (merge: `agentMode: "auto"`, `toolPermission: "always-proceed"`, `enableTerminalSandbox: false`, `allowNonWorkspaceAccess: true`, `permissions`); Antigravity wiring in `run_onchange_after_install-harness.*` |

Shared: the `giocaizzi-harness` external and the `run_onchange_after_install-harness.*` scripts while any of `codex`, `opencode`, `pi`, `antigravity` is.

Add or remove a harness later: `chezmoi edit-config` and edit the list under `[data]`, e.g. `harnesses = ["claude", "copilot", "vscode"]` (or re-answer every prompt with `chezmoi init --prompt`), then `chezmoi apply`. Removing a harness only stops chezmoi managing its files: the files already in the home directory stay; delete them by hand if they are unwanted.

Simulate the prompts without touching the real config (prompt flags are keyed by the prompt text, as printed by `chezmoi init`):

```sh
chezmoi execute-template --init \
  --promptChoice "[ INSTALL MODE ] Light setup? (shell, git, vim, and prompt only)=no" \
  --promptMultichoice "[ AI HARNESSES ] Which harness configs to manage? (space toggles; Enter confirms; chezmoi never installs the CLIs)=claude/copilot" \
  < .chezmoi.toml.tmpl
```

`execute-template` keeps the answer already in your config for a key it finds there (and, unlike `chezmoi init`, re-asks a stored `harnesses` list); point `--config` at an empty `.toml` file to simulate a new machine.

### Shared harness policy

`.chezmoidata/harness-policy.toml` holds common protected read/write paths, restricted command prefixes, terminal-bell preferences and native execution/model overrides. The existing registries remain the source for plugins (`harness.toml`), MCP servers (`mcp.toml`) and layered agent instructions (`.chezmoitemplates/agents/`). Telemetry already shares the machine's `otelEndpoint` and work-role exclusion; exporter formats and other client-specific settings stay in their adapters.

| Harness | Shared policy adapter | Native settings and limits |
|---|---|---|
| Claude Code | Protected paths and commands → `Read`/`Edit`/`Bash` rules; Windows adds its existing `PowerShell` restrictions; bell → `preferredNotifChannel` | `native.claude` keeps `defaultMode` and `skipAutoPermissionPrompt`; broad allows remain native. |
| Copilot CLI | Bell → `beep` | `native.copilot.agents` keeps per-agent model/effort bindings. Shared permission rules and execution mode are not managed here. |
| VS Code Copilot | Existing shared instructions and MCP registry | No shared permission, notification or execution mapping. Editor settings remain native. |
| Codex | Bell → `[tui] notifications` and `notification_method` | Optional `native.codex` keys: `model`, `model_reasoning_effort`, `approval_policy`, `sandbox_mode`. Empty by default, preserving local values. Shared path/command permission rules are not mapped. Installed agents retain their definitions. |
| opencode | Existing shared instructions, MCP and plugin wiring | Permission, execution and model settings remain unmanaged. Notifications come from the `sounds` plugin. |
| pi | Existing shared instructions, MCP and plugin wiring | Permission, execution and model settings remain unmanaged. Notifications come from the `sounds` plugin. |
| Antigravity | Protected paths and commands → `read_file`/`write_file`/`command` rules | `native.antigravity` keeps `agentMode`, `toolPermission`, `enableTerminalSandbox`, `allowNonWorkspaceAccess`; broad allows remain native. No shared bell mapping. |

These adapters preserve each client's matching rules; a shared pattern is not a guarantee of equivalent enforcement. Native execution modes are separate preferences, not interchangeable values of a universal auto mode. Launch, project and session settings can override user defaults, including in Codex's VS Code integration. Runtime state (model choices without a managed override, hook trust, plugins, authentication and trusted-workspace history) stays local under the existing merge rules.

To change the shared setup:

1. Run `dots agents/policy` and edit the common preference or the named native override. Use the existing `dots mcp`, `dots plugins` and `dots agents/<layer>` entries for those registries.
2. The editor's `lint templates` check renders the policy adapters across OSes and roles before applying. Review the configuration diffs; unsupported mappings in the table above require a native setting, not an assumed translation.
3. Start a new client session and inspect its effective settings. An existing session can retain launch/session overrides. Do not use a successful file render as evidence that every native permission is enforced.

Verify a source change with `make check`: policy regression tests preserve the pre-refactor native settings, exercise shared preference changes, optional Codex overrides and runtime-state merges. The template matrix validates the containing settings templates, which include the permission partials. PowerShell execution remains a Windows CI check.

Rollback: restore the previous policy file through git and run `dots agents/policy` to review and apply it. Removing an optional Codex override stops managing that key; it does not delete a value already written to the live config. Restore that local preference explicitly if needed.

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
- **Tab alert:** every harness rings the terminal bell when it waits on the person, so the terminal marks that tab. Claude Code `preferredNotifChannel: "terminal_bell"`, Copilot `beep`, Codex `[tui] notification_method = "bel"`; opencode and pi through the harness `sounds` plugin. Ghostty `bell-features` shows 🔔 in the title, a border, a Dock bounce and the system sound; Windows Terminal `bellStyle: "all"` flashes the window and taskbar and plays the sound. tmux forwards the bell to the outer terminal.
- **Claude Code TUI:** `tui` is `fullscreen` everywhere (needed for `viewMode: focus`, clicks and the side-by-side `/diff`, which also needs Claude Code ≥ 2.1.287 and a terminal ≥ 110 columns). macOS needs Ghostty mouse reporting on. On Windows Terminal, if stale text fragments appear, set `CLAUDE_CODE_ALT_SCREEN_FULL_REPAINT=1`; `CLAUDE_CODE_SCROLL_SPEED=3` is already set there.
- **Claude Code and `bw`:** every `bw` command Claude runs needs your approval (an `ask` rule in the managed Claude settings; ask rules beat the blanket Bash allow, also inside pipes and in auto mode). To let Claude act, unlock in your own terminal into a private file, `(umask 077; bw unlock --raw > <file>)`; Claude runs `bw --session "$(cat <file>)" …` and prints only non-secret fields. Finish with `bw lock` and delete the file. The rule matches the command as written, so it's an approval gate, not a security boundary.
- **Claude Code and auto-merge:** `gh pr merge … --auto` is allowed by name, so agents can arm auto-merge in auto mode too, where the blanket Bash allow does not count; the merge still waits for required CI and the repo ruleset. `gh pr merge … --admin` is denied. Copilot CLI, Codex, opencode and pi have no rule for it here and ask as for any command.

## Editing your dotfiles

Many targets are composed: `~/.claude/CLAUDE.md` is a one-line stub over a partial over four layers, an `mcp.json` is rendered from `.chezmoidata/mcp.toml` through two partials, VS Code settings are symlinks into `dot_config/`. `chezmoi edit <target>` opens the stub; `chezmoi dots` opens the part you mean and shows what the edit changes on every machine type. `dots` is an alias for `chezmoi dots` on full machines; `chezmoi dots help` prints the full usage. It is [a chezmoi plugin](#the-chezmoi-dots-plugin) (Go, `tools/dots`), the same on macOS, Linux and Windows. On a terminal the output uses the basic ANSI colours, which the terminal's Catppuccin theme maps (headings blue, configurations and entries mauve, failures red, success green, prompts yellow), and the composition tree is drawn with `├─`/`└─`; a pipe, `NO_COLOR` or `TERM=dumb` gives plain text.

### Edit

| Command | Edits |
|---|---|
| `dots` | a registry entry, picked with fzf (a numbered prompt without fzf) |
| `dots <entry>` (also `dots edit <entry>`) | that entry, e.g. `dots mcp`, `dots agents/work`, `dots aliases/private`, `dots ssh/config`; `dots list` shows them all |
| `dots <target>` | what a target is made of: a picker over the nodes of its composition tree (`dots ~/.claude/CLAUDE.md` → the stub, the partial, each layer); a target made of one file opens it directly |
| `dots <source file>` | a source file by its source-relative path, e.g. `dots dot_config/git/config.tmpl` |
| `dots <home>/<local>/<name>` | a file of a `local` entry's folder, e.g. `dots ~/.ssh/config.d/pi` (`ssh/local`: this machine only, unmanaged; a new name creates it): the entry's check parses a private temp copy, which replaces the live file only when it passes (else edit again, or keep the live file); no render, diff, apply or PR. `dots ssh/local` picks an existing file |

Every source edit, in order:

1. **Render before:** the target state (`chezmoi dump`) of this machine and of 9 configurations — darwin, linux, windows × personal, work, light — each with every harness selected (light: none), rendered in parallel. The configurations come from the real `.chezmoi.toml.tmpl`, rendered per role with seeded answers, so there is no second copy of the init prompts to keep in step.
2. **Edit** in `$EDITOR` (then `$VISUAL`, then `vi`; on Windows `notepad`). An encrypted file is decrypted into a private temp file (owner-only), re-encrypted (armored) only when its content changed, and the temp file is always removed. Inside tmux or, on Windows, Windows Terminal, a side pane (`chezmoi dots live`) re-renders this machine from a private copy of the source on every save and shows the diff of each target the edit changes so far; it closes with the editor.
3. **Check:** the entry's check, when it has one (`chezmoi dots lint aliases templates`, `ssh -G`, ...).
4. **Render after** and report the changed targets per configuration: a list, `same as <configuration>` for an identical list, or `nothing`. A configuration that no longer renders fails the edit; one that already failed before is only reported.
5. **This machine only:** `chezmoi diff` of its changed targets (one diff when the entry says they render the same text), `apply now? [y/N]` (yes applies only those targets; skipped when they are already up to date), the entry's note (e.g. open a new shell after an alias change), and `git status` of the source as a reminder to ship a PR (e.g. `/ship`). A changed script is listed: it runs on the next full `chezmoi apply`.

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

Limits of the simulation: identity answers this machine lacks render as placeholders (`<workEmail>`), the others blank, as at init. Paths come from this machine (`.chezmoi.homeDir`, `.chezmoi.sourceDir`), so a windows render shows Mac paths. `modify_` targets (Claude, Copilot settings) merge into this machine's live file. Previews render with `dotsRender` set, so no external is downloaded for them. The 9 configurations render without a Bitwarden project, so secrets stay this machine's and parallel renders stay under the Secrets Manager rate limit; this machine's render still calls `bws` and needs to be online.

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

- `[package.<name>]`: installed by the install scripts with the platform's native manager (brew on macOS, apt on Linux, winget on Windows), plus uv, npm, an Azure CLI extension or a PowerShell module, per role (`light`, `work`, `personal`). `agents = true` also names the tool to AI agents in the environment layer of the [agent instructions](#agent-instructions).
- `[release.<name>]`: a binary from a release page, pinned by version and checksum and installed by chezmoi as an external into `~/.local/bin` (no sudo, no package manager).

`chezmoi dots lint packages` (part of `make check`) validates both tables, including that every release target has a 64-hex sha256.

**Update everything on this machine**

**When:** whenever, e.g. weekly.

1. `chezmoi dots upgrade`: updates the dotfiles (`chezmoi update --refresh-externals`), then the system packages (brew on macOS, apt on Linux, winget on Windows), uv tools, global npm tools and, on Windows, PowerShell modules. A step whose tool is missing is skipped; a failing step does not stop the others. `--dry-run` prints the plan, and naming steps (`chezmoi dots upgrade brew npm`) runs only those.
2. Pinned release binaries do not move with it: they change when the pin changes (below), then `chezmoi apply` fetches the new version.

**Verify:** `chezmoi dots outdated` lists nothing; `chezmoi status` prints nothing.

**Bump the pinned release binaries**

**When:** `chezmoi dots outdated` lists a newer version (the weekly `bump-releases` workflow opens a PR for this too).

1. `chezmoi dots bump` (or `chezmoi dots bump <release>...`; `--dry-run` previews): finds the newest tag of each release (git tags on GitHub/GitLab, go.dev for Go), downloads the assets for every target, and rewrites the version and sha256 lines in `packages.toml`, leaving comments and order alone.
2. Review the diff, `make check`, ship the PR; `chezmoi apply` installs the new versions.

**Rollback:** revert the PR; the old pin and checksums come back together.

## The chezmoi-dots plugin

`chezmoi dots` is a [chezmoi plugin](https://www.chezmoi.io/user-guide/command-overview/#plugins): chezmoi runs the `chezmoi-dots` executable found on `PATH` for `chezmoi dots ...`, with the chezmoi environment set (`CHEZMOI_SOURCE_DIR`, `CHEZMOI_EXECUTABLE`, ...). The code is the Go module `tools/dots`: pure packages for the registries, the configuration matrix, the composition tree and the package updates, behind small ports; adapters for chezmoi (`chezmoi data`, `dump`, `managed`, `cat`, ... as the stable interface), the editor, git and the terminal; tests use hand-written fakes plus an end-to-end test against a real chezmoi. [tools/dots/DESIGN.md](../tools/dots/DESIGN.md) has the architecture.

| Command | Does |
|---|---|
| `chezmoi dots [edit] [<what>]`, `view`, `list`, `live` | [Editing your dotfiles](#editing-your-dotfiles) |
| `chezmoi dots lint [check]...` | `registry`, `aliases` (the encrypted registry too where the age key exists), `packages`, `plugins` (against `giocaizzi/harness`; `--harness-dir` for a checkout), `matrix`, `templates` (every script, alias file and settings template, rendered per configuration and shellchecked, zsh-checked or pwsh-parsed where those tools are installed) |
| `chezmoi dots audit aliases` | lists aliases live in this machine's zsh and bash that the registry does not account for |
| `chezmoi dots outdated`, `bump`, `upgrade` | [Packages](#packages) |

**Install it**

**When:** a new machine, or while no release is pinned yet.

1. `chezmoi apply` installs the pinned release (`[release.chezmoi-dots]` in `packages.toml`) into `~/.local/bin`; nothing else to do.
2. To run a local build instead: `make -C tools/dots install` (builds and copies it to `~/.local/bin`).

**Verify:** `chezmoi dots version`.

**Release it**

**When:** the plugin changed in a way worth shipping.

1. Tag the merge commit on `main`: `git tag dots-vX.Y.Z && git push origin dots-vX.Y.Z`. The `release-dots` workflow builds static binaries for the six platforms, writes `checksums.txt`, attests them and creates the GitHub release.
2. Pin it: `chezmoi dots bump chezmoi-dots` rewrites the version and the six checksums of `[release.chezmoi-dots]` in `packages.toml` (the weekly `bump-releases` PR does it too); review, ship the PR, `chezmoi apply` installs it.

**Verify:** the release lists six binaries and `checksums.txt`; `chezmoi dots outdated` is clean after the bump.

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

## MCP servers

Rule: harness `repo-standard`, "MCP servers, least privilege". One list, `.chezmoidata/mcp.toml`, feeds every selected harness:

| Layer | Data | Claude Code | Copilot CLI | VS Code | Codex | opencode | pi |
|---|---|---|---|---|---|---|---|
| User | `[mcp.user]` (every machine) + `[mcp.personal]` (not work), merged by `.chezmoitemplates/mcp-user.json` | `~/.claude.json` via `run_onchange_after_claude-mcp-user.*` (add-only) | `~/.copilot/mcp-config.json` (merge) | user `mcp.json` (whole file rendered by `.chezmoitemplates/vscode-mcp.json`, Code and Insiders) | `~/.codex/config.toml` `[mcp_servers]` (merge) | `~/.config/opencode/opencode.jsonc` `mcp` (merge) | `~/.pi/agent/mcp.json` (merge) |
| Workspace `~/ops` | `[mcp.ops]` | `~/ops/.mcp.json` | same file | `~/ops/.vscode/mcp.json` | `~/ops/.codex/config.toml`, trusted in the user config | `~/ops/opencode.json` | `~/ops/.pi/mcp.json` |

- **Connectors off:** `disableClaudeAiConnectors: true` in `.chezmoitemplates/claude-settings.json`.
- **Private URL parts:** `${VAR}` in a `mcp.toml` URL or `client_id` (e.g. `${MCP_CONNECTOR_ACCOUNT}`), a Bitwarden secret of the same name exported by `secrets.sh`. `.chezmoitemplates/mcp-servers.json` fills it from the environment at `chezmoi apply`, so the rendered files (never committed) hold the value and this public repo does not. Unset (first apply on a new machine, before a new shell), it stays a reference: open a new shell and apply again.
- **`~/ops` is personal:** ignored on light and work machines. Never put a `.mcp.json` in `~` (Claude Code merges it into every repo below).
- **Repo servers** live in each repo, not here.

Procedures:

- **Add a server:** add `name = "url"` under `[mcp.user]` (every machine), `[mcp.personal]` (personal machines) or `[mcp.ops]`, `chezmoi apply`. VS Code's user file is rendered whole: a server VS Code adds itself is overwritten on the next apply, so add stdio servers to `.chezmoitemplates/vscode-mcp.json`.
- **Work PC set up before `[mcp.personal]`** (once): `chezmoi apply`, then remove the personal server from the clients whose merge keeps it, for each CLI installed there: `claude mcp remove -s user notion`, `copilot mcp remove notion`, `codex mcp remove notion`, delete `notion` from the `mcp` key of `~/.config/opencode/opencode.jsonc`, `pi mcp remove notion`. VS Code's rendered `mcp.json` drops it on apply. Check: `claude mcp list` (and the other `mcp list` commands) shows no `notion`.
- **Add a server with a pre-registered OAuth client** (`[mcp.ops]` only; for authorization servers without dynamic client registration, e.g. the Google Workspace MCP servers `google-calendar`, `google-docs`, `google-drive`, `google-gmail`, `google-sheets`, `google-slides`, which share one OAuth client and callback port 8765):
  1. Register the redirect URIs on the OAuth client: `http://localhost:<port>/callback` (Claude Code) and, for VS Code, `http://127.0.0.1:33418` and `https://vscode.dev/redirect`. Servers that share a client and port share these URIs.
  2. Put the client ID in Bitwarden Secrets Manager under a `MCP_<NAME>_CLIENT_ID` secret (Google: `MCP_GOOGLE_CLIENT_ID`), open a new shell (so `secrets.sh` exports it).
  3. Add `name = { url = "…", client_id = "${MCP_<NAME>_CLIENT_ID}", callback_port = <port> }` under `[mcp.ops]`, `chezmoi apply`. Claude Code gets `oauth.clientId` + `oauth.callbackPort` in `~/ops/.mcp.json`, VS Code `oauth.clientId` in `~/ops/.vscode/mcp.json`. Unset, the client ID stays a reference and sign-in fails: open a new shell and apply again.
  4. Client secret, once per machine, harness and server (never in `mcp.toml`, the rendered files or Secrets Manager; keep it in the password vault). Claude Code keeps it in the OS keychain per server, keyed by the server's name and URL, so each server needs its own `claude mcp add --client-secret`; a new client ID keeps it, a renamed server or new URL needs it again. In `~/ops`, for each server: `claude mcp remove -s project <name>`, then `MCP_CLIENT_SECRET="$(bw get password <item>)" claude mcp add -s project --transport http --client-id "$MCP_<NAME>_CLIENT_ID" --client-secret --callback-port <port> <name> <url>` (PowerShell: `$env:MCP_CLIENT_SECRET = bw get password <item>`, the same `claude mcp add`, then `Remove-Item Env:MCP_CLIENT_SECRET`). Then `chezmoi apply` to put the rendered `.mcp.json` back. VS Code: the *Set Client Secret* code lens above each server in `~/ops/.vscode/mcp.json`, or the prompt on first connect (OS secret store).
  5. Sign in to one server at a time (they share the callback port): `/mcp` in Claude Code, first connect in VS Code. Check: `claude mcp get <name>` shows `OAuth: client_id configured, callback_port <port>`; `claude mcp list` shows it connected.
  6. Changing the client ID's variable name (once, e.g. `MCP_GOOGLE_DOCS_CLIENT_ID` to `MCP_GOOGLE_CLIENT_ID`): create the new secret, new shell, `chezmoi apply`, check `claude mcp list`, then delete the old secret.

  Not yet supported, so their templates skip the entry: Copilot CLI ([#125](https://github.com/giocaizzi/dotfiles/issues/125), random callback port; it still sees the Claude Code entry in `.mcp.json` and cannot sign in), Codex ([#126](https://github.com/giocaizzi/dotfiles/issues/126), plaintext-only secret), pi ([#127](https://github.com/giocaizzi/dotfiles/issues/127)) and opencode ([#128](https://github.com/giocaizzi/dotfiles/issues/128)), no secret store. A table under `[mcp.user]` or `[mcp.personal]` fails the render.
- **Try a server first:** `claude mcp add -s local --transport http <name> <url>` in that folder (this machine only), then move it to the repo, `mcp.toml`, and `claude mcp remove -s local <name>`.
- **Remove a server:** delete it from `mcp.toml`, `chezmoi apply`, then remove it by hand where the merge keeps it: `claude mcp remove -s user <name>`, `copilot mcp remove <name>`, `codex mcp remove <name>`, the `mcp` key of opencode, `pi mcp remove <name>`.
- **First run in `~/ops`:** Claude Code asks to approve the project servers; Copilot CLI and pi ask to trust the folder. Codex is trusted by the config.
- **Sign in:** OAuth servers ask once per machine and harness: Claude Code `/mcp`, Copilot CLI `/mcp`, VS Code on first connect, `codex mcp login <name>`, `opencode mcp auth <name>`, `pi mcp login <name>`.
- **Check:** `claude mcp list`, `copilot mcp list`, `codex mcp list`, `opencode mcp list`, `pi mcp list` from `~`, `~/ops` and a repo: only that layer's servers.

## Agent instructions

Every harness reads the same global instructions, rendered (not symlinked) by `.chezmoitemplates/agents-instructions.md` into `~/.claude/CLAUDE.md`, `~/.copilot/copilot-instructions.md`, `~/.codex/AGENTS.md`, `~/.config/opencode/AGENTS.md` and `~/.pi/agent/AGENTS.md`. The text is composed of layers, in this order:

| Layer | Source | Public | Machines |
|---|---|---|---|
| core | `.chezmoitemplates/agents/core.md` | yes | every machine with a harness |
| env | `.chezmoitemplates/agents/env.md` (template: role and OS conditionals) | yes | every machine with a harness |
| private, personal | `.chezmoitemplates/agents/personal.md.age` | no (age) | personal (`work` false) |
| private, work | `.chezmoitemplates/agents/work.md.age` | no (age) | work |

Core holds how to work on any machine, in any harness: it names no harness-only tool (skills carry the per-harness detail, e.g. `subagent-orchestration`, `host-hygiene`), except that the plan-first routing line uses Claude Code's plugin-namespaced agent names (`dev-tools:driver`, `dev-tools:worker`, `dev-tools:scout`), which the other harnesses know by the plain names, kept verbatim because evals validated that wording. Env holds what depends on the machine: how its CLI environment is managed and the tool-install rule (brew, apt or winget by OS), which tools it has (by OS and role) and, on personal machines only, the own-org (`giocaizzi`) rules; work machines follow their org's standard. The private layer holds what must not be public: on personal machines the side projects, personal infrastructure and the boundary list; on work machines the employer context, its SSH hosts and the boundary rule seen from the work side. The private layer is inlined rather than imported with `@file`, which only Claude Code expands. The layers are templates only: no file is deployed for them (`~/.config/agents/` is gone, see below).

Light machines (the Raspberry Pi) render none of it, nor any harness config: agents never run there, by design, and they need no age key. Work agents run on the work PC, with the work layer.

- **Edit a layer:** `dots agents/<core|env|personal|work>` (or `dots ~/.claude/CLAUDE.md` and pick the layer), the flow in [Editing your dotfiles](#editing-your-dotfiles). A private layer goes through a private temp file. The diff shows only the first changed instruction file (they all render the same text); apply writes all of them. `dots view ~/.claude/CLAUDE.md --matrix` shows the composed text per OS and role.
- **Edit by hand:** public layers are plain files; for a private one, `chezmoi decrypt` it into a `umask 077` temp file, edit, `chezmoi encrypt` it back, delete the temp file, `chezmoi apply`.
- **Check a role:** `chezmoi cat ~/.claude/CLAUDE.md --override-data '{"work":true}'` renders it as the work PC would.
- **`~/.config/agents/` removal:** it once held a rendered copy and the private layer as the `chezmoi edit` handle; no harness read it. `.chezmoiremove` deletes it on the next `chezmoi apply` (chezmoi may ask to confirm). Elsewhere, once: `rm -r ~/.config/agents`.
- **First apply after the switch from symlinks:** `~/.claude/CLAUDE.md` and `~/.copilot/copilot-instructions.md` become regular files; chezmoi replaces the old symlinks.

## Agent plugins (`giocaizzi-harness`)

Plugins come from the `giocaizzi/harness` marketplace, not from chezmoi files: the settings templates only declare the marketplace and enable the plugins, and Claude Code / Copilot CLI fetch them. Codex, opencode and pi have no settings that do this, so `run_onchange_after_install-harness.{sh,ps1}.tmpl` (full machines with any of them selected, see [AI harnesses](#ai-harnesses); every apply, idempotent) wires them from a checkout kept by the `.local/share/giocaizzi-harness` external: Codex adds the marketplace and plugins (`codex plugin add`) and gets the agents from `scripts/install.mjs codex`; opencode gets skills, agents and the guard plugin stub from `scripts/install.mjs opencode`; pi installs the checkout as a package, so the external's daily pull updates it. chezmoi does not install the three CLIs: the script wires up only those already on PATH. Codex skips plugin hooks until trusted: run `/hooks` once per `guard` version. Runbook: [install.md](https://github.com/giocaizzi/harness/blob/main/docs/install.md). Hooks (guards, `make check` on stop, post-edit lint, keep-awake) ship in its `guard` plugin; runbook: [guard.md](https://github.com/giocaizzi/harness/blob/main/docs/guard.md).

```sh
claude plugin marketplace add giocaizzi/harness
claude plugin install guard@giocaizzi-harness   # likewise hud, sounds, python, api, javascript, dev-tools
copilot plugin marketplace add giocaizzi/harness
copilot plugin install guard@giocaizzi-harness  # likewise hud, python, api, javascript, dev-tools
```

Which harness enables which plugin is one list, `.chezmoidata/harness.toml` (`[harnessPlugins]`: plugin → harnesses); the Claude and Copilot settings and the Codex step of `run_onchange_after_install-harness.*` are generated from it. `harnessPersonalPlugins` names the plugins for personal work only: work machines get them disabled (`false` in the Claude and Copilot settings, left out of the Codex step); opencode and pi take the whole checkout, so they still get them. `chezmoi dots lint plugins` (part of `make check` and CI) fails when its keys differ from the `plugins/` directories of `giocaizzi/harness`, or when `harnessPersonalPlugins` names a plugin not in the list.

- **New plugin upstream:** add `<name> = ["claude", ...]` to `.chezmoidata/harness.toml`, `chezmoi apply`.
- **Plugin removed upstream:** delete its key, `chezmoi apply`; disable or uninstall it by hand where the harness keeps it.

Managed settings per harness:

- Claude Code: marketplace and `enabledPlugins` (`.chezmoitemplates/claude-settings.json`).
- Claude Code status line: `hud`'s `hooks/statusline.mjs claude`, run from the marketplace clone (`~/.claude/plugins/marketplaces/giocaizzi-harness`, a path that does not change with the plugin version); `node` 22+ and a Nerd Font are required. The agent-panel rows (`subagentStatusLine`) run the same script with `subagent`: each subagent's model and effort. It replaces the old `statusline-command.sh`; delete `~/.claude/statusline-command.sh` once on each machine.
- Copilot CLI: marketplace, `enabledPlugins` and `statusLine` (`.chezmoitemplates/copilot-settings.json`). The status line runs `hud`'s `hooks/statusline.mjs` (`ctx N% · cache N%`) from `~/.copilot/installed-plugins/giocaizzi-harness/hud`; Copilot expands `~` itself on every OS, and `node` 22+ must be on PATH. Copilot CLI has no prompt-cache TTL or auto-compact setting. `subagents.agents` pins the harness agents to their tiers, keyed `<plugin>:<agent>`: `dev-tools:scout` GPT-6 Luna at `low` (tier 1), `api:api-reviewer` GPT-6.1 Sol at `medium` (tier 3, driver), `dev-tools:escalate` GPT-6 Astra at `medium` (tier 4), `dev-tools:reviewer` and `dev-tools:worker` GPT-6 Luna at `medium` (routine tier; a safety-path reviewer cannot be raised per spawn on Copilot) (as Claude's `effort`). A model the account lacks falls back to an available one, so Astra runs as Luna until the account gets it.
- Global instructions: one composed text for every harness, see [Agent instructions](#agent-instructions).
- VS Code Copilot (`dot_config/Code/User/settings.json`, all OSes): `chat.plugins.marketplaces` lists `giocaizzi/harness` and `chat.useHooks` enables the `guard` hooks. Plugins are installed from `@agentPlugins` in the Extensions view (or picked up from Copilot CLI's installs); VS Code has no status-line hook.

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
