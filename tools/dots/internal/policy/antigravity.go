package policy

import (
	"fmt"
	"strings"
)

// AntigravityPermissionsOutput is the structure for Antigravity permissions.
type AntigravityPermissionsOutput struct {
	Allow []string `json:"allow"`
	Deny  []string `json:"deny"`
	Ask   []string `json:"ask"`
}

// ToMap converts AntigravityPermissionsOutput into an untyped map for JSON/template consumers.
func (a AntigravityPermissionsOutput) ToMap() map[string]any {
	allow := make([]any, len(a.Allow))
	for i, v := range a.Allow {
		allow[i] = v
	}
	deny := make([]any, len(a.Deny))
	for i, v := range a.Deny {
		deny[i] = v
	}
	ask := make([]any, len(a.Ask))
	for i, v := range a.Ask {
		ask[i] = v
	}
	return map[string]any{
		"allow": allow,
		"deny":  deny,
		"ask":   ask,
	}
}

// AntigravityPermissions translates the policy permissions into Antigravity's native schema.
func (p Policy) AntigravityPermissions() AntigravityPermissionsOutput {
	allow := []string{
		"command(*)",
		"command(gh pr merge * --auto*)",
		"read_url(*)",
		"execute_url(*)",
		"mcp(*)",
	}

	deny := make([]string, 0, len(p.Permissions.DenyCommandPatterns)+
		len(p.Permissions.DenyCommandExact)+
		len(p.Permissions.DenyCommandPrefixes)*2+
		len(p.Permissions.DenyRead)+
		len(p.Permissions.DenyWrite))

	for _, pat := range p.Permissions.DenyCommandPatterns {
		deny = append(deny, fmt.Sprintf("command(%s)", pat))
	}
	for _, cmd := range p.Permissions.DenyCommandExact {
		deny = append(deny, fmt.Sprintf("command(%s)", cmd))
	}
	for _, pfx := range p.Permissions.DenyCommandPrefixes {
		if !strings.Contains(pfx, " ") {
			deny = append(deny, fmt.Sprintf("command(%s)", pfx))
		}
		deny = append(deny, fmt.Sprintf("command(%s *)", pfx))
	}
	for _, path := range p.Permissions.DenyRead {
		deny = append(deny, fmt.Sprintf("read_file(%s)", path))
	}
	for _, path := range p.Permissions.DenyWrite {
		deny = append(deny, fmt.Sprintf("write_file(%s)", path))
	}

	ask := make([]string, 0, len(p.Permissions.AskCommandPrefixes)*2)
	for _, pfx := range p.Permissions.AskCommandPrefixes {
		ask = append(ask, fmt.Sprintf("command(%s)", pfx))
		ask = append(ask, fmt.Sprintf("command(%s *)", pfx))
	}

	return AntigravityPermissionsOutput{
		Allow: allow,
		Deny:  deny,
		Ask:   ask,
	}
}

// AntigravitySettings returns the managed subset of Antigravity settings.
func (p Policy) AntigravitySettings() map[string]any {
	agentMode := p.Native.Antigravity.AgentMode
	if agentMode == "" {
		agentMode = defaultAuto
	}
	toolPerm := p.Native.Antigravity.ToolPermission
	if toolPerm == "" {
		toolPerm = "always-proceed"
	}

	return map[string]any{
		"permissions":             p.AntigravityPermissions().ToMap(),
		"agentMode":               agentMode,
		"toolPermission":          toolPerm,
		"enableTerminalSandbox":   p.Native.Antigravity.EnableTerminalSandbox,
		"allowNonWorkspaceAccess": p.Native.Antigravity.AllowNonWorkspaceAccess,
	}
}
