# dotfiles

My configuration for macOS, Linux and Windows from a single source, managed with [chezmoi](https://www.chezmoi.io/). Configs follow the XDG Base Directory spec (`~/.config`, `~/.local/{bin,share,state}`, `~/.cache`).

## Set up a new machine

1. **Install chezmoi** into `~/.local/bin`:

   ```shell
   sh -c "$(curl -fsLS get.chezmoi.io)" -- -b ~/.local/bin   # macOS / Linux
   winget install twpayne.chezmoi                              # Windows
   ```

2. **Copy the age key** to `~/.config/chezmoi/key.txt` (owner-only) from another machine, over a secure channel such as a password manager — never through this repo or chat. Without it, `apply` stops at the first encrypted file (see [docs/OS.md](./docs/OS.md#encrypted-files-age)).

3. **Windows only:** turn on Developer Mode (Settings → System → For developers) so chezmoi can create symlinks without admin.

4. **Init and apply:**

   ```shell
   chezmoi init --apply giocaizzi/dotfiles
   ```

   `chezmoi init` asks once per machine and stores the answers only in that machine's `~/.config/chezmoi/chezmoi.toml`:

   | Prompt | Answer |
   |---|---|
   | User name / Email address | git identity |
   | Light install (shell, git, vim, prompt only) | yes for headless boxes over SSH (e.g. the Raspberry Pi) |
   | Work computer (work git identity, no personal telemetry) | yes on the corporate PC; then asks the work repos folder and work email |

   What each answer changes: [docs/OS.md → Machine roles](./docs/OS.md#machine-roles). To change one later, e.g. `chezmoi init --promptBool "Light install (shell, git, vim, prompt only)=true"` (the flag matches the prompt text).

## Everyday use

| Command | Does |
|---|---|
| `chezmoi update` | pull this repo and apply it |
| `chezmoi edit <file>` | edit a managed file in the source (auto-committed and pushed) |
| `chezmoi add <file>` | start managing a file (`--encrypt` for secrets) |
| `chezmoi status` / `chezmoi diff` | what an apply would change |
| `chezmoi merge <file>` | reconcile a file changed on both sides |
| `chezmoi cd` | open a shell in the source repo |

`autoCommit`/`autoPush` are on: `chezmoi edit`/`add` commit and push immediately. If a push fails, the next `chezmoi update` on another machine stops on diverged history — push from the machine that has the commits first.

## Secrets

- **Environment variables** (API keys, tokens) live in `~/.config/shell/secrets.sh` (Windows: `~/.config/powershell/secrets.ps1`), never tracked. Created empty on first apply with owner-only permissions; edit it directly on each machine. See [docs/SHELL.md](./docs/SHELL.md#secrets).
- **Secret files** (e.g. `~/.ssh/config`) are committed age-encrypted: `chezmoi add --encrypt <file>`.

## Documentation

- [docs/OS.md](./docs/OS.md) — what deploys where per OS, machine roles (light/work), encryption, Windows notes.
- [docs/SHELL.md](./docs/SHELL.md) — shell architecture, aliases and shortcuts, tools, secrets.
- [AGENTS.md](./AGENTS.md) — repo conventions and the Catppuccin style system.
- [chezmoi documentation](https://www.chezmoi.io/).
