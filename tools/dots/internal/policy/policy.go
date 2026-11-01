// Package policy parses, validates and renders shared AI harness policies
// into client-native configurations for Claude Code, Antigravity, Copilot CLI,
// and Codex.
package policy

const defaultAuto = "auto"

// Policy is the canonical cross-harness security and runtime configuration.
type Policy struct {
	Notifications Notifications `json:"notifications" toml:"notifications"`
	Permissions   Permissions   `json:"permissions"   toml:"permissions"`
	Native        Native        `json:"native"        toml:"native"`
}

// Notifications configures shared notification preferences across harnesses.
type Notifications struct {
	TerminalBell bool `json:"terminalBell" toml:"terminalBell"`
}

// Permissions defines forbidden and sensitive resources across all harnesses.
type Permissions struct {
	DenyRead            []string `json:"denyRead"            toml:"denyRead"`
	DenyWrite           []string `json:"denyWrite"           toml:"denyWrite"`
	DenyCommandPatterns []string `json:"denyCommandPatterns" toml:"denyCommandPatterns"`
	DenyCommandExact    []string `json:"denyCommandExact"    toml:"denyCommandExact"`
	DenyCommandPrefixes []string `json:"denyCommandPrefixes" toml:"denyCommandPrefixes"`
	AskCommandPrefixes  []string `json:"askCommandPrefixes"  toml:"askCommandPrefixes"`
}

// Native holds harness-specific configuration settings and overrides.
type Native struct {
	Claude      ClaudeNative      `json:"claude"      toml:"claude"`
	Antigravity AntigravityNative `json:"antigravity" toml:"antigravity"`
	Codex       CodexNative       `json:"codex"       toml:"codex"`
	Copilot     CopilotNative     `json:"copilot"     toml:"copilot"`
}

// ClaudeNative defines Claude Code native settings.
type ClaudeNative struct {
	DefaultMode              string `json:"defaultMode"              toml:"defaultMode"`
	SkipAutoPermissionPrompt bool   `json:"skipAutoPermissionPrompt" toml:"skipAutoPermissionPrompt"`
}

// AntigravityNative defines Antigravity native settings.
type AntigravityNative struct {
	AgentMode               string `json:"agentMode"               toml:"agentMode"`
	ToolPermission          string `json:"toolPermission"          toml:"toolPermission"`
	EnableTerminalSandbox   bool   `json:"enableTerminalSandbox"   toml:"enableTerminalSandbox"`
	AllowNonWorkspaceAccess bool   `json:"allowNonWorkspaceAccess" toml:"allowNonWorkspaceAccess"`
}

// CodexNative defines optional Codex preferences.
type CodexNative struct {
	Model                *string `json:"model,omitempty"                  toml:"model,omitempty"`
	ModelReasoningEffort *string `json:"model_reasoning_effort,omitempty" toml:"model_reasoning_effort,omitempty"`
	ApprovalPolicy       *string `json:"approval_policy,omitempty"        toml:"approval_policy,omitempty"`
	SandboxMode          *string `json:"sandbox_mode,omitempty"           toml:"sandbox_mode,omitempty"`
}

// CopilotNative defines Copilot CLI native agent bindings.
type CopilotNative struct {
	Agents map[string]CopilotAgent `json:"agents" toml:"agents"`
}

// CopilotAgent specifies the model and effort level for a Copilot subagent.
type CopilotAgent struct {
	Model       string `json:"model"       toml:"model"`
	EffortLevel string `json:"effortLevel" toml:"effortLevel"`
}
