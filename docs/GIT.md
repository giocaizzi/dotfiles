# Git Commit Conventions

This repository uses [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/).

## Required Header Format

```text
<type>[optional scope][!]: <description>
```

## Optional Sections

```text
<type>[optional scope][!]: <description>

[optional body]

[optional footer(s)]
```

## Rules

- Use `feat` for new features.
- Use `fix` for bug fixes.
- Scope is optional and should be a noun in parentheses, for example `feat(shell): ...`.
- Mark breaking changes with `!` before `:` and/or with a `BREAKING CHANGE: <description>` footer.
- Keep commits small and focused.

## Common Types In This Repository

- `feat`
- `fix`
- `docs`
- `style`
- `refactor`
- `perf`
- `test`
- `build`
- `ci`
- `chore`
- `revert`

## Examples

```text
feat(shell): add fzf preview to history search
fix(env)!: remove deprecated PATH fallback

BREAKING CHANGE: legacy PATH fallback is no longer supported
```
