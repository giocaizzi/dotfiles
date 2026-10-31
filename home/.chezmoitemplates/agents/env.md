{{- /* Environment layer of the global agent instructions: this machine's role and OS.
       Runbook: docs/OS.md#agent-instructions. */ -}}
{{- $os := .chezmoi.os -}}
{{- $work := get . "work" -}}
{{- $pm := dict "darwin" "`brew`" "linux" "`apt` (or pinned release binaries in `~/.local/bin`, chezmoi externals)" "windows" "`winget`" | dig $os "`brew`" -}}
{{- $install := dict "darwin" "brew install" "linux" "apt install" "windows" "winget install" | dig $os "brew install" -}}
{{- /* The tools marked agents = true in the package registry that this machine's OS
       and role install (.chezmoidata/packages.toml), by command name. */ -}}
{{- $role := includeTemplate "role" . -}}
{{- $manager := dict "darwin" "brew" "linux" "apt" "windows" "winget" | dig $os "brew" -}}
{{- $releases := includeTemplate "releases" . | fromJson -}}
{{- $tools := list -}}
{{- range $name, $p := .package -}}
{{-   if and (get $p "agents") (has $role (get $p "roles" | default (list "personal" "work"))) -}}
{{-     if or (get $p $manager) (has $os (get $p "script" | default list)) (hasKey $releases $name) -}}
{{-       $tools = append $tools (get $p "cmd" | default $name) -}}
{{-     end -}}
{{-   end -}}
{{- end -}}
# Environment

- `chezmoi` governs the CLI environment: configs and the tool install list (the package registry, `.chezmoidata/packages.toml` in its source state `home/`). The source is `~/.local/share/chezmoi`; machine roles (`light`, `work`), per-OS layout and Bitwarden are documented in its `docs/OS.md`.
- Tools: packages come from {{ $pm }}, through that list only; never an ad-hoc `{{ $install }}`. One-off tools run without installing (`uvx`, `npx`, `docker run`). Propose, never add on your own, a tool for the list once it was needed in two separate sessions or a repo's standard requires it.
- Assume {{ range $i, $t := $tools }}{{ if $i }}, {{ end }}`{{ $t }}`{{ end }} are available; prefer them (e.g. `jq` for JSON) over hand-rolled parsing.
{{- if not $work }}

# Own org (`giocaizzi`)

- Skills, agents and mods come from `github.com/giocaizzi/harness`. If a rule is missing or stale, add it there, generic and reusable; personal or machine-specific facts stay in these instructions.
- Parity: every feature of agents, skills and mods covers all supported harnesses (Claude Code, Copilot CLI, Codex, opencode, pi; matrix in `giocaizzi/harness` docs/compatibility.md) and hosts (macOS, Linux, Windows). Build every side in the same change, deployed through chezmoi; a side that cannot be built gets a GitHub issue with a priority label and the reason, before shipping.
- `giocaizzi/*` repos follow the standard in `github.com/giocaizzi/.github` (`docs/standard.md`: labels, rulesets, settings, templates, CI, exceptions, plan limits; PRs stacked and auto-merged behind required CI checks on `main`); `make audit` there reports drift. Change the standard there first, then the repos. New repos start from `giocaizzi/template` (`gh repo create giocaizzi/<name> --template giocaizzi/template`), then `make onboard REPO=<name> TYPE=<type>` in `giocaizzi/.github`.
{{- end }}
