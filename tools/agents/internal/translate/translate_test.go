package translate_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/policy"
	"github.com/giocaizzi/dotfiles/tools/agents/internal/translate"
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

	got := translate.ClaudeSettings(p, "linux")
	want := requireMap(t, linuxGolden, "claude")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ClaudeSettings(linux) mismatch (-want +got):\n%s", diff)
	}
}

func TestAntigravityGolden(t *testing.T) {
	t.Parallel()
	p, golden := loadTestPolicy(t)
	linuxGolden := requireMap(t, golden, "linux")

	got := translate.AntigravitySettings(p)
	want := requireMap(t, linuxGolden, "antigravity")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("AntigravitySettings mismatch (-want +got):\n%s", diff)
	}
}

func TestCopilotGolden(t *testing.T) {
	t.Parallel()
	p, golden := loadTestPolicy(t)
	linuxGolden := requireMap(t, golden, "linux")

	got := translate.CopilotSettings(p)
	want := requireMap(t, linuxGolden, "copilot")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CopilotSettings mismatch (-want +got):\n%s", diff)
	}
}

func TestCodex(t *testing.T) {
	t.Parallel()
	p, _ := loadTestPolicy(t)
	got := translate.CodexConfig(p)
	tui, ok := got["tui"].(map[string]any)
	if !ok {
		t.Fatalf("missing tui section in codex config: %v", got)
	}
	if tui["notifications"] != true {
		t.Errorf("expected notifications=true, got %v", tui["notifications"])
	}
	if tui["notification_method"] != "bel" {
		t.Errorf("expected notification_method=bel, got %v", tui["notification_method"])
	}
}

func TestPiSettings(t *testing.T) {
	t.Parallel()
	p, _ := loadTestPolicy(t)
	got := translate.PiSettings(p)
	if got["defaultProvider"] != "openai" {
		t.Errorf("expected defaultProvider=openai, got %v", got["defaultProvider"])
	}
	if got["defaultModel"] != "gpt-5.5" {
		t.Errorf("expected defaultModel=gpt-5.5, got %v", got["defaultModel"])
	}
}

func TestVSCodeSettings(t *testing.T) {
	t.Parallel()
	p, _ := loadTestPolicy(t)
	got := translate.VSCodeSettings(p)
	terminal, ok := got["chat.tools.terminal.autoApprove"].(map[string]bool)
	if !ok {
		t.Fatalf("expected map[string]bool for terminalAutoApprove, got %T", got["chat.tools.terminal.autoApprove"])
	}
	if !terminal["git *"] {
		t.Errorf("expected git * to be true")
	}
	if terminal["env"] {
		t.Errorf("expected env to be false")
	}
	urls, ok := got["chat.tools.urls.autoApprove"].(map[string]bool)
	if !ok {
		t.Fatalf("expected map[string]bool for urlsAutoApprove, got %T", got["chat.tools.urls.autoApprove"])
	}
	if !urls["https://github.com/**"] {
		t.Errorf("expected https://github.com/** to be true")
	}
}
