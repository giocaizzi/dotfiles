{{- /* Global agent instructions for harnesses without Claude's `@file` import (Codex, opencode, pi):
       AGENTS.md with the private file inlined where Claude imports it. Claude and Copilot keep
       symlinks to the source. */ -}}
{{- include "dot_config/agents/AGENTS.md" | replace "@~/.config/agents/AGENTS.private.md" (include "dot_config/agents/encrypted_AGENTS.private.md.age" | decrypt) -}}
