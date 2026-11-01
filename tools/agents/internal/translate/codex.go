package translate

import (
	"github.com/giocaizzi/dotfiles/tools/agents/internal/policy"
)

// CodexConfig returns the managed subset of Codex configuration as a map.
func CodexConfig(p *policy.Policy) map[string]any {
	cfg := map[string]any{
		"tui": map[string]any{
			"notifications":       p.Notifications.TerminalBell,
			"notification_method": "bel",
		},
	}

	if p.Native.Codex.Model != nil {
		cfg["model"] = *p.Native.Codex.Model
	}
	if p.Native.Codex.ModelReasoningEffort != nil {
		cfg["model_reasoning_effort"] = *p.Native.Codex.ModelReasoningEffort
	}
	if p.Native.Codex.ApprovalPolicy != nil {
		cfg["approval_policy"] = *p.Native.Codex.ApprovalPolicy
	}
	if p.Native.Codex.SandboxMode != nil {
		cfg["sandbox_mode"] = *p.Native.Codex.SandboxMode
	}

	return cfg
}
