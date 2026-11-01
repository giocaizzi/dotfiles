package policy

import (
	"fmt"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/finding"
)

const (
	defaultAuto   = "auto"
	defaultManual = "manual"
)

// Validate checks that the policy conforms to house rules and valid enum values.
func (p Policy) Validate(where string) []finding.Finding {
	fs := make([]finding.Finding, 0, 8)
	fs = append(fs, p.validateAutonomy(where)...)
	fs = append(fs, p.validatePatterns(where)...)
	fs = append(fs, p.validateModels(where)...)
	fs = append(fs, p.validateNative(where)...)
	return fs
}

func (p Policy) validateAutonomy(where string) []finding.Finding {
	var fs []finding.Finding
	mode := p.Autonomy.ExecutionMode
	if mode == "" {
		mode = p.Autonomy.Mode
	}
	if mode != "" {
		switch mode {
		case defaultAuto, defaultManual, "prompt", "on-request":
		default:
			fs = append(fs, finding.New(where, "autonomy mode %q must be auto, manual, on-request or prompt", mode))
		}
	}
	return fs
}

func (p Policy) validatePatterns(where string) []finding.Finding {
	var fs []finding.Finding
	check := func(list []string, name string) {
		for i, pat := range list {
			if strings.TrimSpace(pat) == "" {
				fs = append(fs, finding.New(where, "%s[%d] is empty", name, i))
			}
		}
	}
	check(p.Permissions.AllowCommands, "permissions.allowCommands")
	check(p.Permissions.AllowedUrls, "permissions.allowedUrls")
	check(p.Permissions.DenyRead, "permissions.denyRead")
	check(p.Permissions.DenyWrite, "permissions.denyWrite")
	check(p.Permissions.DenyCommandPatterns, "permissions.denyCommandPatterns")
	check(p.Permissions.DenyCommandExact, "permissions.denyCommandExact")
	check(p.Permissions.DenyCommandPrefixes, "permissions.denyCommandPrefixes")
	check(p.Permissions.AskCommandPrefixes, "permissions.askCommandPrefixes")
	return fs
}

func (p Policy) validateModels(where string) []finding.Finding {
	var fs []finding.Finding
	for tierName, tier := range p.Models.Tiers {
		if strings.TrimSpace(tier.Model) == "" {
			fs = append(fs, finding.New(where, "models.tiers.%s missing model", tierName))
		}
		if strings.TrimSpace(tier.Effort) == "" {
			fs = append(fs, finding.New(where, "models.tiers.%s missing effort", tierName))
		}
	}
	return fs
}

func (p Policy) validateNative(where string) []finding.Finding {
	var fs []finding.Finding
	if p.Native.Claude.DefaultMode != "" {
		switch p.Native.Claude.DefaultMode {
		case defaultAuto, defaultManual, "prompt":
		default:
			fs = append(fs, finding.New(where, "native.claude.defaultMode %q must be auto, manual or prompt", p.Native.Claude.DefaultMode))
		}
	}

	if p.Native.Antigravity.AgentMode != "" {
		switch p.Native.Antigravity.AgentMode {
		case defaultAuto, defaultManual:
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
	bell := p.UI.TerminalBell || p.Notifications.TerminalBell
	return fmt.Sprintf("Policy(bell=%v, allowCommands=%d, denyRead=%d, denyCommands=%d)",
		bell,
		len(p.Permissions.AllowCommands),
		len(p.Permissions.DenyRead),
		len(p.Permissions.DenyCommandPatterns)+len(p.Permissions.DenyCommandExact)+len(p.Permissions.DenyCommandPrefixes))
}
