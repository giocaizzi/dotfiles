# dotfiles

My configuration files backup and syncronization system.

Using [chezmoi](https://www.chezmoi.io/)

## Init chezmoi on a machine

Will be prompted for local data, like name and email.

```shell
chezmoi init gh:giocaizzi/dotfiles.git
```

## File structure

- chezmoi repository: `~/.local/share/chezmoi`
- chezmoi local configuration and state: `~/.config/chezmoi`

## Usage

- Navigate to chezmoi repository for file ops.

```shell
chezmoi cd
```

- Ensure that files are in the target state, applying changes if necessary.

```shell
chezmoi update
chezmoi apply
```

- Add new files to be managed by chezmoi.

```shell
chezmoi add <file>
```

- Edit files directly in the chezmoi repository.

```shell
chezmoi edit <file>
```

- View the status of files managed by chezmoi.

```shell
chezmoi status
```

- View differences between source and target states.

```shell
chezmoi diff
```

- Merge changes from the source state into the target state.

```shell
chezmoi merge
chezmoi merge --all
```

## Scripts

`chezmoi` offers the ability to run custom scripts during multiple phases of its operation. 

The scripts are saved in `.chezmoiscripts` directory, and they are executed at specific points in the process of applying changes to the target system.
The filenames indicate when they are executed:
- `run_` prefix: executed everytime `chezmoi apply` is run.
- `run_onchange_` prefix: if their content changed since last time they were successfully run.

## Applications

Using symlinks, the [apps](./apps/) folder contains configuration files for specific applications, which are linked to their expected locations in the filesystem.

## Ignore files

The [.chezmoiignore](./.chezmoiignore) file is used to exclude certain files or directories from being managed by `chezmoi`, ensuring that only the intended configuration files are tracked and applied. This will consider OS and other variables.

## Documentation

See the official [chezmoi documentation](https://www.chezmoi.io/docs/) for more details and advanced usage.