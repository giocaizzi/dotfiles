{{- /* Global agent instructions, the same text for every harness (Claude, Copilot, Codex,
       opencode, pi), composed from layers in .chezmoitemplates/agents/: core (public, every
       machine), env (public, per role and OS: the CLI environment) and the role's private layer,
       inlined because Copilot, Codex, opencode and pi do not expand `@file` imports: personal
       machines personal.md.age, work machines work.md.age (both age-encrypted). Light machines
       render none (.chezmoiignore). Edit with `dots agents/<layer>`. Runbook: docs/OS.md#agent-instructions. */ -}}
{{- $private := ternary "work" "personal" (get . "work" | default false) -}}
{{ include ".chezmoitemplates/agents/core.md" }}
{{ template "agents/env.md" . }}
{{ include (printf ".chezmoitemplates/agents/%s.md.age" $private) | decrypt -}}
