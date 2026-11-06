package parity_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/parity"
	"github.com/giocaizzi/dotfiles/tools/agents/internal/policy"
)

func loadPolicy(t *testing.T) *policy.Policy {
	t.Helper()
	path := filepath.Join("..", "..", "..", "..", "home", ".chezmoidata", "harness-policy.toml")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	p, err := policy.ParseTOML(f)
	if err != nil {
		t.Fatalf("ParseTOML: %v", err)
	}
	return p
}

func TestCheckParity_AllHarnesses(t *testing.T) {
	t.Parallel()
	p := loadPolicy(t)

	findings := parity.CheckParity(p)
	if len(findings) > 0 {
		t.Errorf("unexpected parity findings: %v", findings)
	}

	summary := parity.Summary(p)
	for _, h := range parity.SupportedHarnesses() {
		if !strings.Contains(summary, h) {
			t.Errorf("summary missing harness %s", h)
		}
	}
}

func TestCheckParity_VoiceTracksPolicy(t *testing.T) {
	t.Parallel()
	for _, enabled := range []bool{true, false} {
		p := loadPolicy(t)
		p.Voice.Enabled = enabled
		if findings := parity.CheckParity(p); len(findings) > 0 {
			t.Errorf("voice.enabled=%v: unexpected findings: %v", enabled, findings)
		}
	}
}
