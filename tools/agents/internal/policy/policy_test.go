package policy_test

import (
	"strings"
	"testing"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/policy"
)

func TestParseTOML_Valid(t *testing.T) {
	t.Parallel()
	input := `
[harnessPolicy.autonomy]
executionMode = "auto"

[harnessPolicy.notifications]
terminalBell = true

[harnessPolicy.permissions]
allowCommands = ["git *"]
allowedUrls = ["https://github.com/**"]
denyRead = ["**/.env"]
denyWrite = ["**/.env"]
denyCommandPatterns = ["gh pr merge * --admin*"]
denyCommandExact = ["env"]
denyCommandPrefixes = ["printenv"]
askCommandPrefixes = ["bw"]

[harnessPolicy.models.tiers.scout]
model = "gpt-6-luna"
effort = "low"

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
		t.Errorf("expected terminalBell true")
	}
	if len(p.Permissions.DenyRead) != 1 {
		t.Errorf("expected 1 denyRead entry, got %d", len(p.Permissions.DenyRead))
	}
	if len(p.Permissions.AllowCommands) != 1 {
		t.Errorf("expected 1 allowCommands entry, got %d", len(p.Permissions.AllowCommands))
	}
	if p.Native.Claude.DefaultMode != "auto" {
		t.Errorf("expected defaultMode auto, got %q", p.Native.Claude.DefaultMode)
	}

	findings := p.Validate("test")
	if len(findings) > 0 {
		t.Errorf("unexpected validation findings: %v", findings)
	}
}

func TestValidate_Voice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		voice policy.Voice
		want  int
	}{
		{"empty", policy.Voice{}, 0},
		{"hold", policy.Voice{Enabled: true, Mode: "hold"}, 0},
		{"tap", policy.Voice{Mode: "tap"}, 0},
		{"bad mode", policy.Voice{Mode: "toggle"}, 1},
		{"language short", policy.Voice{Language: "it"}, 0},
		{"language region", policy.Voice{Language: "it-IT"}, 0},
		{"bad language", policy.Voice{Language: "english"}, 1},
		{"bad language underscore", policy.Voice{Language: "en_US"}, 1},
		{"both bad", policy.Voice{Mode: "x", Language: "1"}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := policy.Policy{Voice: tt.voice}
			if got := p.Validate("test"); len(got) != tt.want {
				t.Errorf("Validate() = %v, want %d findings", got, tt.want)
			}
		})
	}
}
