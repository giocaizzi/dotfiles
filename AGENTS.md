This is user's [chezmoi](https://chezmoi.io/) configuration for a user. 

You have deep knowledge of shell configurations, dotfiles management, and best practices for setting up development environments.

## Principles
- Always consider OS and separate configuration for MacOS, Linux, and Windows when necessary.
- Always check current [chezmoi documentation](https://chezmoi.io/docs/) for the latest features and best practices.
- If required, use symbolic links to avoid duplication and keep centralized configuration management.

## Chezmoi 

- Use main [.chezmoiignore](./.chezmoiignore) to exclude files and folders that should not be managed by chezmoi, depending on OS or other criteria.
- Use Chezmoi template language with the functions and variables provided.
- Use [chezmoi scripts](./.chezmoiscripts) with correct naming conventions to manage scripts and when to use them. (Filenames indicate when they are executed.)

## Docs

Keep documentation easy to read yet comprehensive.

Keep a simple, essential documentation in [./docs](./docs) folder, with markdown files for different topics. Ensure to keep them updated with the latest changes, and to cover all the necessary information for the user to understand how to use and manage their chezmoi configuration.

Keep also the main [README.md](./README.md) file updated with everyday usage instructions. Ensure always up to date with the latest changes.

