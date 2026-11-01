package policy

import (
	"fmt"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/finding"
)

const (
	defaultAuto = "auto"
)

// Validate checks that the policy conforms to house rules and valid enum values.
func (p Policy) Validate(where string) []finding.Finding {
	var fs []finding.Finding

	checkPatterns := func(list []string, name string) {
		for i, pat := range list {
			if strings.TrimSpace(pat) == "" {
				fs = append(fs, finding.New(where, "%s[%d] is empty", name, i))
			}
		}
	}

	checkPatterns(p.Permissions.DenyRead, "permissions.denyRead")
	checkPatterns(p.Permissions.DenyWrite, "permissions.denyWrite")
	checkPatterns(p.Permissions.DenyCommandPatterns, "permissions.denyCommandPatterns")
	checkPatterns(p.Permissions.DenyCommandExact, "permissions.denyCommandExact")
	checkPatterns(p.Permissions.DenyCommandPrefixes, "permissions.denyCommandPrefixes")
	checkPatterns(p.Permissions.AskCommandPrefixes, "permissions.askCommandPrefixes")

	if p.Native.Claude.DefaultMode != "" {
		switch p.Native.Claude.DefaultMode {
		case defaultAuto, "manual", "prompt":
		default:
			fs = append(fs, finding.New(where, "native.claude.defaultMode %q must be auto, manual or prompt", p.Native.Claude.DefaultMode))
		}
	}

	if p.Native.Antigravity.AgentMode != "" {
		switch p.Native.Antigravity.AgentMode {
		case defaultAuto, "manual":
		default:
			fs = append(fs, finding.New(where, "native.antigravity.agentMode %q must be auto or manual", p.Native.Antigravity.AgentMode))
		}
	}

	for agent, binding := range p.Native.Copilot.Agents {
		if strings.TrimSpace(binding.Model) == "" {
			fs = append(fs, finding.New(where, "native.copilot.agents.%s missing model", agent))
		}
		if strings.TrimSpace(binding.EffortLevel) == "" {
			fs = append(fs, finding.New(where, "native.copilot.agents.%s missing effortLevel", agent))
		}
	}

	if p.Native.Pi.DefaultThinkingLevel != "" {
		switch p.Native.Pi.DefaultThinkingLevel {
		case "off", "minimal", "low", "medium", "high", "xhigh", "max":
		default:
			fs = append(fs, finding.New(where, "native.pi.defaultThinkingLevel %q must be off, minimal, low, medium, high, xhigh or max", p.Native.Pi.DefaultThinkingLevel))
		}
	}

	return fs
}

func (p Policy) String() string {
	return fmt.Sprintf("Policy(notifications=%v, denyRead=%d, denyCommands=%d)",
		p.Notifications.TerminalBell,
		len(p.Permissions.DenyRead),
		len(p.Permissions.DenyCommandPatterns)+len(p.Permissions.DenyCommandExact)+len(p.Permissions.DenyCommandPrefixes))
}
