package policy

// CopilotSettings returns the managed subset of Copilot CLI settings.
func (p Policy) CopilotSettings() map[string]any {
	agents := make(map[string]any, len(p.Native.Copilot.Agents))
	for name, a := range p.Native.Copilot.Agents {
		agents[name] = map[string]any{
			"model":       a.Model,
			"effortLevel": a.EffortLevel,
		}
	}

	return map[string]any{
		"beep": p.Notifications.TerminalBell,
		"subagents": map[string]any{
			"agents": agents,
		},
	}
}
