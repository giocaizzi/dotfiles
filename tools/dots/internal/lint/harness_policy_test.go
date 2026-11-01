package lint_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/matrix"
)

// The golden captures the managed policy before it was extracted into shared data.
// It deliberately retains native matching differences instead of assuming that two
// clients enforce equivalent rules.
func TestHarnessPolicyPreservesNativeSettings(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile("testdata/harness-policy.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	var golden map[string]map[string]any
	if err := json.Unmarshal(b, &golden); err != nil {
		t.Fatal(err)
	}
	for _, config := range matrix.All() {
		t.Run(config.Label(), func(t *testing.T) {
			t.Parallel()
			cz, src := policyClient(t)
			for _, harness := range []string{"claude", "copilot", "antigravity"} {
				for _, selection := range [][]string{{harness}, {"claude", "copilot", "vscode", "codex", "opencode", "pi", "antigravity"}} {
					o := policyOverrides(config, selection)
					got := renderPolicyJSON(t, cz, src, ".chezmoitemplates/"+harness+"-settings.json", o)
					want, ok := golden["linux"][harness].(map[string]any)
					if !ok {
						t.Fatalf("missing golden for %s", harness)
					}
					assertGoldenMatch(t, got, want, golden["windowsClaudePermissions"], harness, config.OS, selection)
				}
			}
		})
	}
}

func assertGoldenMatch(t *testing.T, got, want map[string]any, winClaude any, harness, goos string, sel []string) {
	t.Helper()
	for key, value := range want {
		if harness == "claude" && key == "permissions" && goos == "windows" {
			value = winClaude
		}
		if diff := cmp.Diff(value, got[key]); diff != "" {
			t.Errorf("%s %s %v (-want +got):\n%s", harness, key, sel, diff)
		}
	}
}

func TestHarnessPolicyChangesReachNativeAdapters(t *testing.T) {
	t.Parallel()
	cz, src := policyClient(t)
	o := policyOverrides(matrix.Config{OS: "linux", Role: "personal"}, []string{"claude", "copilot", "codex", "antigravity"})
	o["harnessPolicy"] = map[string]any{
		"notifications": map[string]any{"terminalBell": false},
		"permissions":   map[string]any{"denyRead": []string{"**/example-private/**"}},
	}
	for _, tt := range []struct {
		harness, channel, prefix string
		wantChannel              any
	}{
		{harness: "claude", channel: "preferredNotifChannel", prefix: "Read", wantChannel: "notifications_disabled"},
		{harness: "antigravity", prefix: "read_file"},
		{harness: "copilot", channel: "beep", wantChannel: false},
	} {
		got := renderPolicyJSON(t, cz, src, ".chezmoitemplates/"+tt.harness+"-settings.json", o)
		if tt.channel != "" && got[tt.channel] != tt.wantChannel {
			t.Errorf("%s notification = %v, want %v", tt.harness, got[tt.channel], tt.wantChannel)
		}
		if tt.prefix != "" {
			assertPathDenied(t, got, tt.harness, tt.prefix, "**/example-private/**")
		}
	}
	got := renderPolicyTOML(t, cz, src, o)
	tui, ok := got["tui"].(map[string]any)
	if !ok || tui["notifications"] != false {
		t.Errorf("Codex tui = %v, want notifications=false", got["tui"])
	}
}

func assertPathDenied(t *testing.T, got map[string]any, harness, prefix, path string) {
	t.Helper()
	permissions, ok := got["permissions"].(map[string]any)
	if !ok {
		t.Fatalf("%s permissions missing", harness)
	}
	deny, ok := permissions["deny"].([]any)
	if !ok {
		t.Fatalf("%s deny rules missing", harness)
	}
	target := prefix + "(" + path + ")"
	for _, rule := range deny {
		if rule == target {
			return
		}
	}
	t.Errorf("%s does not protect %s", harness, target)
}

func TestHarnessPolicyMergesPreserveRuntimeState(t *testing.T) {
	t.Parallel()
	cz, src := policyClient(t)
	for _, harness := range []string{"claude", "copilot", "antigravity"} {
		o := policyOverrides(matrix.Config{OS: "linux", Role: "personal"}, []string{harness})
		o["chezmoi"] = map[string]any{"os": "linux", "stdin": `{"model":"local-choice","customState":{"keep":true},"allowedUrls":["https://example.com"]}`}
		path := "dot_" + harness + "/modify_private_settings.json"
		if harness == "claude" {
			path = "dot_claude/modify_settings.json"
		}
		if harness == "antigravity" {
			path = "dot_gemini/antigravity-cli/modify_private_settings.json"
		}
		got := renderPolicyJSON(t, cz, src, path, o)
		if got["model"] != "local-choice" || cmp.Diff(map[string]any{"keep": true}, got["customState"]) != "" {
			t.Errorf("%s lost local state: model=%v customState=%v", harness, got["model"], got["customState"])
		}
	}
}

func TestCodexPolicyOptionalOverridesPreserveLocalState(t *testing.T) {
	t.Parallel()
	cz, src := policyClient(t)
	for _, config := range matrix.All() {
		t.Run(config.Label(), func(t *testing.T) {
			t.Parallel()
			o := policyOverrides(config, []string{"codex"})
			o["chezmoi"] = map[string]any{"os": config.OS, "stdin": `model = "local-choice"
model_reasoning_effort = "low"
approval_policy = "on-request"
sandbox_mode = "read-only"
[plugins."example@marketplace"]
enabled = true
[hooks.state.example]
trusted_hash = "example-hash"
[customState]
keep = true
`}
			got := renderPolicyTOML(t, cz, src, o)
			if got["model"] != "local-choice" || got["sandbox_mode"] != "read-only" || got["approval_policy"] != "on-request" {
				t.Errorf("absent policy replaced Codex local preferences")
			}
			o["harnessPolicy"] = map[string]any{"native": map[string]any{"codex": map[string]any{
				"model": "managed-choice", "model_reasoning_effort": "medium", "approval_policy": "never", "sandbox_mode": "workspace-write",
			}}}
			changed := renderPolicyTOML(t, cz, src, o)
			for key, value := range map[string]any{"model": "managed-choice", "model_reasoning_effort": "medium", "approval_policy": "never", "sandbox_mode": "workspace-write"} {
				if changed[key] != value {
					t.Errorf("Codex %s = %v, want %v", key, changed[key], value)
				}
			}
			for _, key := range []string{"plugins", "hooks", "customState"} {
				if diff := cmp.Diff(got[key], changed[key]); diff != "" {
					t.Errorf("Codex %s changed (-want +got):\n%s", key, diff)
				}
			}
		})
	}
}

func policyClient(t *testing.T) (*chez.Client, string) {
	t.Helper()
	if _, err := exec.LookPath("chezmoi"); err != nil {
		t.Skip("chezmoi is not installed")
	}
	src, err := filepath.Abs("../../../../home")
	if err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(t.TempDir(), "chezmoi.toml")
	if err := os.WriteFile(config, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	return chez.New(chez.ExecRunner{}, "chezmoi", "--source", src, "--config", config), src
}

func policyOverrides(c matrix.Config, harnesses []string) chez.Overrides {
	return chez.Overrides{
		"chezmoi": map[string]any{"os": c.OS}, "work": c.Role == "work", "light": c.Role == "light",
		"harnesses": harnesses, "otelEndpoint": "",
	}
}

func renderPolicyJSON(t *testing.T, cz *chez.Client, src, name string, overrides chez.Overrides) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(src, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	out, err := cz.Execute(t.Context(), overrides, b, false)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	return got
}

func renderPolicyTOML(t *testing.T, cz *chez.Client, src string, overrides chez.Overrides) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(src, "dot_codex", "modify_private_config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := cz.Execute(t.Context(), overrides, b, false)
	if err != nil {
		t.Fatal(err)
	}
	converted, err := cz.Execute(t.Context(), chez.Overrides{"renderedPolicy": string(out)}, []byte(`{{ .renderedPolicy | fromToml | toJson }}`), false)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(converted, &got); err != nil {
		t.Fatal(err)
	}
	return got
}
