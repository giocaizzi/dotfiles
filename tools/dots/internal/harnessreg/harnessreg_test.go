package harnessreg

import (
	"strings"
	"testing"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/finding"
)

func TestCheck(t *testing.T) {
	t.Parallel()
	choices := []string{"claude", "codex"}
	plugins := map[string][]string{"api": {"claude"}, "go": {"claude", "codex"}}
	tests := []struct {
		name     string
		plugins  map[string][]string
		personal []string
		upstream []string
		want     string
	}{
		{"in step", plugins, []string{"go"}, []string{"api", "go"}, ""},
		{"missing here", plugins, nil, []string{"api", "go", "hud"}, "hud: plugin of giocaizzi/harness missing here"},
		{"gone upstream", plugins, nil, []string{"api"}, "go: not a plugin of giocaizzi/harness"},
		{"unknown personal", plugins, []string{"nope"}, []string{"api", "go"}, "nope is not in [harnessPlugins]"},
		{"unknown harness", map[string][]string{"api": {"emacs"}}, nil, []string{"api"}, "unknown harness emacs"},
		{"nothing upstream", plugins, nil, nil, "no plugin directories found upstream"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := finding.Lines(Check(tt.plugins, tt.personal, choices, tt.upstream))
			if tt.want == "" && got != "" || !strings.Contains(got, tt.want) {
				t.Errorf("Check() = %q, want %q", got, tt.want)
			}
		})
	}
}
