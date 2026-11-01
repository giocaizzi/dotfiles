// Package policy parses, validates and holds canonical AI harness policies
// across Claude Code, Antigravity, Copilot CLI, Codex, Pi, and opencode.
package policy

// Policy is the canonical cross-harness security, notification and runtime configuration.
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
	Pi          PiNative          `json:"pi"          toml:"pi"`
	OpenCode    OpenCodeNative    `json:"opencode"    toml:"opencode"`
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
	Agents map[string]AgentBinding `json:"agents,omitempty" toml:"agents,omitempty"`
}

// AgentBinding specifies a model and reasoning effort for an agent.
type AgentBinding struct {
	Model       string `json:"model"       toml:"model"`
	EffortLevel string `json:"effortLevel" toml:"effortLevel"`
}

// PiNative defines native settings for the Pi coding agent.
type PiNative struct {
	DefaultProvider      string   `json:"defaultProvider,omitempty"      toml:"defaultProvider,omitempty"`
	DefaultModel         string   `json:"defaultModel,omitempty"         toml:"defaultModel,omitempty"`
	DefaultThinkingLevel string   `json:"defaultThinkingLevel,omitempty" toml:"defaultThinkingLevel,omitempty"`
	Packages             []string `json:"packages,omitempty"             toml:"packages,omitempty"`
}

// OpenCodeNative defines native settings for opencode.
type OpenCodeNative struct {
	DefaultModel string `json:"defaultModel,omitempty" toml:"defaultModel,omitempty"`
	Sandbox      bool   `json:"sandbox,omitempty"      toml:"sandbox,omitempty"`
}
