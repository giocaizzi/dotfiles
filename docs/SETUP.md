# Set up a new machine

Runbook for a fresh machine; the [README](../README.md) has the short version.

First decide the machine's role ([details](./OS.md#machine-roles)): **personal** (full, your vault available), **work** (full, corporate; no personal vault) or **light** (headless box over SSH, e.g. the Raspberry Pi). Then:

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

3. **Put the age key** at `~/.config/chezmoi/key.txt` (owner-only) — personal and work only, or `apply` stops at the first encrypted file ([why](./OS.md#encrypted-files-age)). **Light machines skip this step:** they get no encrypted files, so a compromised box can't decrypt your secrets.
   - **Personal:** from the vault (secure note `chezmoi age key`):

     ```shell
     bw login
     mkdir -p ~/.config/chezmoi && (umask 077; bw get notes "chezmoi age key" --session "$(bw unlock --raw)" > ~/.config/chezmoi/key.txt)
     bw lock
     ```

   - **Work:** copy it by hand over a secure channel (your personal vault doesn't go on a work PC) — never through this repo or chat.

4. **Create this machine's Secrets Manager token** (personal and work only; [details](./OS.md#bitwarden)): in the Bitwarden **web app** → Secrets Manager → *Machine accounts* → new account named after the machine → *Projects*: **Can read** on `personal` (or `work`) → *Access tokens*: create one (set an expiry date) and copy it (shown once). Also copy the project's ID. Then store the token in the OS keyring, never in a file: `chezmoi secret keyring set --service=bws --user=access-token` (hidden prompt).

5. **Init and apply:**

   ```shell
   chezmoi init --apply giocaizzi/dotfiles
   ```

   `chezmoi init` asks once per machine and stores the answers only in that machine's `~/.config/chezmoi/chezmoi.toml`:

   | Prompt | personal | work | light |
   |---|---|---|---|
   | Git user name (text) / PERSONAL git email (text) | your name / personal email | your name / personal email (used only for personal repos) | your name / personal email |
   | Light setup (choice yes/no) — shell, git, vim, prompt only | no | no | **yes** |
   | Work computer (choice yes/no) — work git identity, no personal telemetry | no | **yes** → WORK git email (default for every repo) + PERSONAL GitHub username + extra personal remote globs ([add later](./OS.md#adding-personal-remotes)) | no |
   | AI harnesses (multiple choice: claude, copilot, vscode, codex, opencode, pi; [details](./OS.md#ai-harnesses)) | the ones you use | the ones you use | not asked |
   | Bitwarden Secrets Manager project ID (text) | `personal` ID | `work` ID | not asked |

   The token is not an init answer: it sits in the OS keyring (step 4). To change answers later, `chezmoi init --prompt` (asks everything again).

6. **Apply once more** — the first apply downloads `bws`, the second renders the env secrets from Bitwarden:

   ```shell
   chezmoi apply && chezmoi status   # status prints nothing when done
   ```

   Open a new shell (Windows: a new PowerShell window). Light machines are done after step 5.

## Secrets

- **Environment variables** (API keys, tokens) live in `~/.config/shell/secrets.sh` (Windows: `~/.config/powershell/secrets.ps1`), owner-only, never committed. Machines with a Bitwarden Secrets Manager project ID (token in the OS keyring) get them **generated** from Bitwarden on `chezmoi apply` (edit them in Bitwarden); others edit the file by hand. See [docs/OS.md → Bitwarden](./OS.md#bitwarden).
- **Secret files** (e.g. `~/.ssh/config`) are committed age-encrypted: `chezmoi add --encrypt <file>`.
