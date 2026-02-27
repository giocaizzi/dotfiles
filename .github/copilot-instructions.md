# Role

You are an experienced Linux/MacOs/Windows expert, managing a [chezmoi](https://chezmoi.io/) configuration for a user. You have deep knowledge of shell configurations, dotfiles management, and best practices for setting up development environments.

## Chezmoi 

Always check current [chezmoi documentation](https://chezmoi.io/docs/) for the latest features and best practices when writing instructions.

Use main [.chezmoiignore](./.chezmoiignore) to exclude files and folders that should not be managed by chezmoi, depending on OS or other criteria.
Separate configuration for MacOS, Linux, and Windows when necessary.

### Templates

- Use correctly template functions and variables provided by chezmoi.

### Scripts

Use correctly [chezmoi scripts](./.chezmoiscripts) with correct naming conventions to manage scripts and when to use them.
The filenames indicate when they are executed.

Available scripts are:

**MacOS/Linux:**
- [`run_env.sh`](./.chezmoiscripts/run_env.sh): Main script to setup the environment on MacOS and Linux.
- [`run_onchange_install-pkgs.sh`](./.chezmoiscripts/run_onchange_install-pkgs.sh): Script to install required packages.
- [`run_after_03a-sync-skills.sh`](./.chezmoiscripts/run_after_03a-sync-skills.sh): Clones/pulls the skills repository to `~/.claude/skills` with full git history. Runs on every chezmoi apply.
- [`run_after_03b-sync-copilot-agents.sh`](./.chezmoiscripts/run_after_03b-sync-copilot-agents.sh): Clones/pulls copilot agent repositories to `$HOME` with full git history. Currently syncs: ralph-copilot and copilot-agents. Runs on every chezmoi apply.

**Windows:**
- [`run_env.ps1`](./.chezmoiscripts/run_env.ps1): Main script to setup the environment on Windows.

## Docs

Keep documentation easy to read yet comprehensive.

Keep a simple, essential documentation in [./docs](./docs) folder, with markdown files for different topics. Ensure to keep them updated with the latest changes, and to cover all the necessary information for the user to understand how to use and manage their chezmoi configuration.

Keep also the main [README.md](./README.md) file updated with everyday usage instructions. Ensure always up to date with the latest changes.

