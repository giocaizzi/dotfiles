{{- /* Environment layer of the agent instructions: this machine's role and OS, as an agents template
       (vars from the agents config, rendered by chezmoi). Runbook: docs/OS.md#agent-instructions. */ -}}
# Environment

- `chezmoi` governs the CLI environment: configs and the tool install list (the package registry, `.chezmoidata/packages.toml` in its source state `home/`). The source is `~/.local/share/chezmoi`; machine roles (`light`, `work`), per-OS layout and Bitwarden are documented in its `docs/OS.md`.
- Tools: packages come from {{ .vars.package_source }}, through that list only; never an ad-hoc `{{ .vars.install_command }}`. One-off tools run without installing (`uvx`, `npx`, `docker run`). Propose, never add on your own, a tool for the list once it was needed in two separate sessions or a repo's standard requires it.
- Assume {{ range $i, $t := .vars.tools }}{{ if $i }}, {{ end }}`{{ $t }}`{{ end }} are available; prefer them (e.g. `jq` for JSON) over hand-rolled parsing.
{{- if ne .profile "work" }}

# Own org (`giocaizzi`)

- Skills, agents and mods come from `github.com/giocaizzi/harness`. If a rule is missing or stale, add it there, generic and reusable; personal or machine-specific facts stay in these instructions.
- Parity: every feature of agents, skills and mods covers all supported harnesses (Claude Code, Copilot CLI, Codex, opencode, pi; matrix in `giocaizzi/harness` docs/compatibility.md) and hosts (macOS, Linux, Windows). Build every side in the same change, deployed through chezmoi; a side that cannot be built gets a GitHub issue with a priority label and the reason, before shipping.
- `giocaizzi/*` repos follow the standard in `github.com/giocaizzi/.github` (`docs/standard.md`: labels, rulesets, settings, templates, CI, exceptions, plan limits; PRs stacked and auto-merged behind required CI checks on `main`); `make audit` there reports drift. Change the standard there first, then the repos. New repos start from `giocaizzi/template` (`gh repo create giocaizzi/<name> --template giocaizzi/template`), then `make onboard REPO=<name> TYPE=<type>` in `giocaizzi/.github`.
{{- end }}
