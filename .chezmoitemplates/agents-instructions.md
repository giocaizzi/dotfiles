{{- /* Global agent instructions, the same text for every harness (~/.config/agents/AGENTS.md,
       Claude, Copilot, Codex, opencode, pi), composed from layers: core (public, every machine),
       env (public, per role and OS: the CLI environment) and the role's private layer, inlined
       because Copilot, Codex, opencode and pi do not expand `@file` imports: personal machines
       AGENTS.private.md, work machines AGENTS.work.md (both age-encrypted). Light machines render
       none (.chezmoiignore). Runbook: docs/OS.md#agent-instructions. */ -}}
{{- $private := ternary "AGENTS.work.md" "AGENTS.private.md" (get . "work" | default false) -}}
{{ include ".chezmoitemplates/agents/core.md" }}
{{ template "agents/env.md" . }}
{{ include (printf "dot_config/agents/encrypted_%s.age" $private) | decrypt -}}
