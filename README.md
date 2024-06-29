# dotfiles

My configuration files backup and syncronization system.

Using [chezmoi](https://www.chezmoi.io/)

## Init chezmoi on a machine

Will be prompted for local data, like name and email.

```shell
chezmoi init gh:giocaizzi/dotfiles.git
```

## File structure

- chezmoi repository
- chezmoi local configuration

## Usage

Navigate to chezmoi repository for file ops.

```shell
chezmoi cd
```

Ensure that files are in the target state, applying changes if necessary.

```shell
chezmoi apply
```
