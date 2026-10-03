# dotfiles

My configuration for macOS, Linux and Windows from a single source, managed with [chezmoi](https://www.chezmoi.io/). Configs follow the XDG Base Directory spec (`~/.config`, `~/.local/{bin,share,state}`, `~/.cache`).

## Set up a new machine

First decide the machine's role ([details](./docs/OS.md#machine-roles)): **personal** (full, your vault available), **work** (full, corporate; no personal vault) or **light** (headless box over SSH, e.g. the Raspberry Pi). Then:

1. **Prerequisites**
   - macOS: [Homebrew](https://brew.sh) (the install script stops without it).
   - Windows: Developer Mode on (Settings → System → For developers), so chezmoi can create symlinks without admin; winget (App Installer).
   - Linux: `sudo` for apt.

2. **Install chezmoi** (and `bw` on personal machines; chezmoi's install script adds it later too, but you need it now for step 3):

   ```shell
   brew install chezmoi bitwarden-cli                          # macOS
   sh -c "$(curl -fsLS get.chezmoi.io)" -- -b ~/.local/bin   # Linux (bw: see bitwarden.com/help/cli)
   winget install twpayne.chezmoi                              # Windows
   ```

3. **Put the age key** at `~/.config/chezmoi/key.txt` (owner-only) — personal and work only, or `apply` stops at the first encrypted file ([why](./docs/OS.md#encrypted-files-age)). **Light machines skip this step:** they get no encrypted files, so a compromised box can't decrypt your secrets.
   - **Personal:** from the vault (secure note `chezmoi age key`):

     ```shell
     bw login
     mkdir -p ~/.config/chezmoi && (umask 077; bw get notes "chezmoi age key" --session "$(bw unlock --raw)" > ~/.config/chezmoi/key.txt)
     bw lock
     ```

   - **Work:** copy it by hand over a secure channel (your personal vault doesn't go on a work PC) — never through this repo or chat.

4. **Create this machine's Secrets Manager token** (personal and work only; [details](./docs/OS.md#bitwarden)): in the Bitwarden **web app** → Secrets Manager → *Machine accounts* → new account named after the machine → *Projects*: **Can read** on `personal` (or `work`) → *Access tokens*: create one and copy it (shown once). Also copy the project's ID.

5. **Init and apply:**

   ```shell
   chezmoi init --apply giocaizzi/dotfiles
   ```

   `chezmoi init` asks once per machine and stores the answers only in that machine's `~/.config/chezmoi/chezmoi.toml`:

   | Prompt | personal | work | light |
   |---|---|---|---|
   | Git user name (text) / PERSONAL git email (text) | your name / personal email | your name / personal email (used only for personal repos) | your name / personal email |
   | Light install (yes/no) — shell, git, vim, prompt only | no | no | **yes** |
   | Work computer (yes/no) — work git identity, no personal telemetry | no | **yes** → WORK git email (default for every repo) + PERSONAL GitHub username + extra personal remote globs ([add later](./docs/OS.md#adding-personal-remotes)) | no |
   | Bitwarden Secrets Manager access token (text) / project ID (text) | token + `personal` ID | token + `work` ID | not asked |

   The token is visible while typed. To change answers later, `chezmoi init --prompt` (asks everything again; keeps the token out of shell history).

6. **Apply once more** — the first apply downloads `bws`, the second renders the env secrets from Bitwarden:

   ```shell
   chezmoi apply && chezmoi status   # status prints nothing when done
   ```

   Open a new shell (Windows: a new PowerShell window). Light machines are done after step 5.

## Everyday use

| Command | Does |
|---|---|
| `chezmoi update` | pull this repo and apply it |
| `chezmoi edit <file>` | edit a managed file in the source (auto-committed and pushed) |
| `chezmoi add <file>` | start managing a file (`--encrypt` for secrets) |
| `chezmoi status` / `chezmoi diff` | what an apply would change |
| `chezmoi merge <file>` | reconcile a file changed on both sides |
| `chezmoi cd` | open a shell in the source repo |

`autoCommit`/`autoPush` are on: `chezmoi edit`/`add` commit and push immediately, with Conventional Commit messages (`chore: add .ssh/config`, from `.commit-message.tmpl`). If a push fails, the next `chezmoi update` on another machine stops on diverged history — push from the machine that has the commits first.

## Secrets

- **Environment variables** (API keys, tokens) live in `~/.config/shell/secrets.sh` (Windows: `~/.config/powershell/secrets.ps1`), owner-only, never committed. Machines with a Bitwarden Secrets Manager token get them **generated** from Bitwarden on `chezmoi apply` (edit them in Bitwarden); others edit the file by hand. See [docs/OS.md → Bitwarden](./docs/OS.md#bitwarden).
- **Secret files** (e.g. `~/.ssh/config`) are committed age-encrypted: `chezmoi add --encrypt <file>`.

## Documentation

- [docs/OS.md](./docs/OS.md) — what deploys where per OS, machine roles (light/work), Bitwarden, encryption, Windows notes.
- [docs/SHELL.md](./docs/SHELL.md) — shell architecture, aliases and shortcuts, tools, secrets.
- [AGENTS.md](./AGENTS.md) — repo conventions and the Catppuccin style system.
- [chezmoi documentation](https://www.chezmoi.io/).
