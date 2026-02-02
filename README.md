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