package policy_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/policy"
)

func loadTestPolicy(t *testing.T) (*policy.Policy, map[string]any) {
	t.Helper()
	f, err := os.Open("../../../../home/.chezmoidata/harness-policy.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	p, err := policy.ParseTOML(f)
	if err != nil {
		t.Fatalf("ParseTOML: %v", err)
	}

	b, err := os.ReadFile("testdata/harness-policy.golden.json")
	if err != nil {
		t.Fatal(err)
	}

	var golden map[string]any
	if err := json.Unmarshal(b, &golden); err != nil {
		t.Fatal(err)
	}
	return p, golden
}

func requireMap(t *testing.T, m map[string]any, key string) map[string]any {
	t.Helper()
	val, ok := m[key].(map[string]any)
	if !ok {
		t.Fatalf("missing or non-map key %q", key)
	}
	return val
}

func TestClaudeGolden(t *testing.T) {
	t.Parallel()
	p, golden := loadTestPolicy(t)
	linuxGolden := requireMap(t, golden, "linux")

	got := p.ClaudeSettings("linux")
	want := requireMap(t, linuxGolden, "claude")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ClaudeSettings(linux) mismatch (-want +got):\n%s", diff)
	}

	gotWin := p.ClaudePermissions("windows")
	wantRaw := requireMap(t, golden, "windowsClaudePermissions")
	bWant, err := json.Marshal(wantRaw)
	if err != nil {
		t.Fatalf("Marshal windowsClaudePermissions: %v", err)
	}
	var wantWin policy.ClaudePermissionsOutput
	if err := json.Unmarshal(bWant, &wantWin); err != nil {
		t.Fatalf("Unmarshal windowsClaudePermissions: %v", err)
	}

	if diff := cmp.Diff(wantWin, gotWin); diff != "" {
		t.Errorf("ClaudePermissions(windows) mismatch (-want +got):\n%s", diff)
	}
}

func TestAntigravityGolden(t *testing.T) {
	t.Parallel()
	p, golden := loadTestPolicy(t)
	linuxGolden := requireMap(t, golden, "linux")

	got := p.AntigravitySettings()
	want := requireMap(t, linuxGolden, "antigravity")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("AntigravitySettings() mismatch (-want +got):\n%s", diff)
	}
}

func TestCopilotGolden(t *testing.T) {
	t.Parallel()
	p, golden := loadTestPolicy(t)
	linuxGolden := requireMap(t, golden, "linux")

	got := p.CopilotSettings()
	want := requireMap(t, linuxGolden, "copilot")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CopilotSettings() mismatch (-want +got):\n%s", diff)
	}
}

func TestCodexGolden(t *testing.T) {
	t.Parallel()
	p, _ := loadTestPolicy(t)

	got := p.CodexConfig()
	want := map[string]any{
		"tui": map[string]any{
			"notifications":       true,
			"notification_method": "bel",
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CodexConfig() mismatch (-want +got):\n%s", diff)
	}
}

func TestPolicyValidation(t *testing.T) {
	t.Parallel()
	p, _ := loadTestPolicy(t)
	findings := p.Validate("test-policy")
	if len(findings) != 0 {
		t.Errorf("expected no findings for valid policy, got: %v", findings)
	}
}

func TestPolicyValidationErrors(t *testing.T) {
	t.Parallel()

	invalidTOML := `
[permissions]
denyRead = [""]
[native.claude]
defaultMode = "invalid_mode"
[native.antigravity]
agentMode = "super_auto"
[native.copilot.agents.scout]
model = ""
effortLevel = ""
`
	p, err := policy.ParseTOML(strings.NewReader(invalidTOML))
	if err != nil {
		t.Fatalf("ParseTOML failed: %v", err)
	}

	findings := p.Validate("bad-policy")
	if len(findings) < 4 {
		t.Errorf("expected at least 4 findings, got %d: %v", len(findings), findings)
	}
}
