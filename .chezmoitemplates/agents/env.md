{{- /* Environment layer of the global agent instructions: this machine's role and OS.
       Runbook: docs/OS.md#agent-instructions. */ -}}
{{- $os := .chezmoi.os -}}
{{- $work := get . "work" -}}
{{- $pm := dict "darwin" "`brew`" "linux" "`apt` (or release binaries in `~/.local/bin`)" "windows" "`winget`" | dig $os "`brew`" -}}
{{- $install := dict "darwin" "brew install" "linux" "apt install" "windows" "winget install" | dig $os "brew install" -}}
{{- $tools := list "jq" "fd" "fzf" "gh" "uv" "gitleaks" "actionlint" "hadolint" -}}
{{- if ne $os "windows" }}{{ $tools = append $tools "tmux" }}{{ else }}{{ $tools = concat $tools (list "rg" "make" "pwsh") }}{{ end -}}
{{- if $work }}{{ $tools = concat $tools (list "glab" "az") }}{{ end -}}
# Environment

- `chezmoi` governs the CLI environment: configs and the tool install list (`.chezmoiscripts/run_onchange_install-pkgs.*`). The source is `~/.local/share/chezmoi`; machine roles (`light`, `work`), per-OS layout and Bitwarden are documented in its `docs/OS.md`.
- Tools: packages come from {{ $pm }}, through that list only; never an ad-hoc `{{ $install }}`. One-off tools run without installing (`uvx`, `npx`, `docker run`). Propose, never add on your own, a tool for the list once it was needed in two separate sessions or a repo's standard requires it.
- Assume {{ range $i, $t := $tools }}{{ if $i }}, {{ end }}`{{ $t }}`{{ end }} are available; prefer them (e.g. `jq` for JSON) over hand-rolled parsing.
- LSP (Claude Code): chezmoi installs `pyright` and `typescript-language-server` for its LSP plugins. For symbols (definitions, references, renames) and type errors, use the LSP tool before grep; fall back to grep for text, config and unsupported languages.
{{- if not $work }}

# Own org (`giocaizzi`)

- Skills, agents and mods come from `github.com/giocaizzi/harness`. If a rule is missing or stale, add it there, generic and reusable; personal or machine-specific facts stay in these instructions.
- Parity: every feature of agents, skills and mods covers all supported harnesses (Claude Code, Copilot CLI, Codex, opencode, pi; matrix in `giocaizzi/harness` docs/compatibility.md) and hosts (macOS, Linux, Windows). Build every side in the same change, deployed through chezmoi; a side that cannot be built gets a GitHub issue with a priority label and the reason, before shipping.
- `giocaizzi/*` repos follow the standard in `github.com/giocaizzi/.github` (`docs/standard.md`: labels, rulesets, settings, templates, CI, exceptions, plan limits; PRs stacked and auto-merged behind required CI checks on `main`); `make audit` there reports drift. Change the standard there first, then the repos. New repos start from `giocaizzi/template` (`gh repo create giocaizzi/<name> --template giocaizzi/template`), then `make onboard REPO=<name> TYPE=<type>` in `giocaizzi/.github`.
{{- end }}
