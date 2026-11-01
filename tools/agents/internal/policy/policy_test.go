package policy_test

import (
	"strings"
	"testing"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/policy"
)

func TestParseTOML_Valid(t *testing.T) {
	t.Parallel()
	input := `
[harnessPolicy.notifications]
terminalBell = true

[harnessPolicy.permissions]
denyRead = ["**/.env"]
denyWrite = ["**/.env"]
denyCommandPatterns = ["gh pr merge * --admin*"]
denyCommandExact = ["env"]
denyCommandPrefixes = ["printenv"]
askCommandPrefixes = ["bw"]

[harnessPolicy.native.claude]
defaultMode = "auto"

[harnessPolicy.native.pi]
defaultProvider = "openai"
defaultModel = "gpt-5.5"
defaultThinkingLevel = "medium"
packages = ["local/pkg"]
`
	p, err := policy.ParseTOML(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseTOML failed: %v", err)
	}

	if !p.Notifications.TerminalBell {
		t.Errorf("expected TerminalBell=true")
	}
	if len(p.Permissions.DenyRead) != 1 || p.Permissions.DenyRead[0] != "**/.env" {
		t.Errorf("unexpected DenyRead: %v", p.Permissions.DenyRead)
	}
	if p.Native.Pi.DefaultModel != "gpt-5.5" {
		t.Errorf("unexpected Pi default model: %s", p.Native.Pi.DefaultModel)
	}

	findings := p.Validate("test")
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %v", findings)
	}
}

func TestValidate_Invalid(t *testing.T) {
	t.Parallel()
	p := policy.Policy{
		Permissions: policy.Permissions{
			DenyRead: []string{""},
		},
		Native: policy.Native{
			Claude: policy.ClaudeNative{
				DefaultMode: "invalid-mode",
			},
			Pi: policy.PiNative{
				DefaultThinkingLevel: "invalid-level",
			},
		},
	}

	fs := p.Validate("test")
	if len(fs) < 3 {
		t.Errorf("expected at least 3 validation errors, got %d", len(fs))
	}
}
