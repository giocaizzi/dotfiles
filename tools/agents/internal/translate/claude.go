// Package translate converts canonical policies into client-native configurations.
package translate

import (
	"fmt"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/policy"
)

// ClaudePermissionsOutput is the structure for Claude Code permissions.
type ClaudePermissionsOutput struct {
	Allow       []string `json:"allow"`
	Deny        []string `json:"deny"`
	Ask         []string `json:"ask"`
	DefaultMode string   `json:"defaultMode"`
}

// ToMap converts ClaudePermissionsOutput into an untyped map for JSON/template consumers.
func (c ClaudePermissionsOutput) ToMap() map[string]any {
	allow := make([]any, len(c.Allow))
	for i, v := range c.Allow {
		allow[i] = v
	}
	deny := make([]any, len(c.Deny))
	for i, v := range c.Deny {
		deny[i] = v
	}
	ask := make([]any, len(c.Ask))
	for i, v := range c.Ask {
		ask[i] = v
	}
	return map[string]any{
		"allow":       allow,
		"deny":        deny,
		"ask":         ask,
		"defaultMode": c.DefaultMode,
	}
}

// ClaudePermissions translates policy permissions into Claude's native schema.
func ClaudePermissions(p *policy.Policy, goos string) ClaudePermissionsOutput {
	allow := []string{
		"WebSearch",
		"WebFetch",
		"Bash",
		"Bash(gh pr merge * --auto*)",
	}

	deny := make([]string, 0, len(p.Permissions.DenyCommandPatterns)+
		len(p.Permissions.DenyRead)+
		len(p.Permissions.DenyWrite)+
		len(p.Permissions.DenyCommandExact)+
		len(p.Permissions.DenyCommandPrefixes)*2)

	for _, pat := range p.Permissions.DenyCommandPatterns {
		deny = append(deny, fmt.Sprintf("Bash(%s)", pat))
	}
	for _, path := range p.Permissions.DenyRead {
		deny = append(deny, fmt.Sprintf("Read(%s)", path))
	}
	for _, path := range p.Permissions.DenyWrite {
		deny = append(deny, fmt.Sprintf("Edit(%s)", path))
	}
	for _, cmd := range p.Permissions.DenyCommandExact {
		deny = append(deny, fmt.Sprintf("Bash(%s)", cmd))
	}
	for _, pfx := range p.Permissions.DenyCommandPrefixes {
		deny = append(deny, fmt.Sprintf("Bash(%s:*)", pfx))
	}

	if goos == "windows" {
		for _, pfx := range p.Permissions.DenyCommandPrefixes {
			if pfx != "printenv" {
				deny = append(deny, fmt.Sprintf("PowerShell(%s *)", pfx))
			}
		}
	}

	ask := make([]string, 0, len(p.Permissions.AskCommandPrefixes)*2)
	for _, pfx := range p.Permissions.AskCommandPrefixes {
		ask = append(ask, fmt.Sprintf("Bash(%s)", pfx))
		ask = append(ask, fmt.Sprintf("Bash(%s *)", pfx))
	}

	mode := p.Native.Claude.DefaultMode
	if mode == "" {
		mode = "auto"
	}

	return ClaudePermissionsOutput{
		Allow:       allow,
		Deny:        deny,
		Ask:         ask,
		DefaultMode: mode,
	}
}

// ClaudeSettings returns the managed subset of Claude Code settings.
func ClaudeSettings(p *policy.Policy, goos string) map[string]any {
	notif := "notifications_disabled"
	if p.Notifications.TerminalBell {
		notif = "terminal_bell"
	}

	return map[string]any{
		"permissions":              ClaudePermissions(p, goos).ToMap(),
		"preferredNotifChannel":    notif,
		"skipAutoPermissionPrompt": p.Native.Claude.SkipAutoPermissionPrompt,
	}
}
