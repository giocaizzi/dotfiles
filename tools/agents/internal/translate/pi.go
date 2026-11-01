package translate

import (
	"github.com/giocaizzi/dotfiles/tools/agents/internal/policy"
)

// PiSettings returns the managed subset of Pi settings.
func PiSettings(p *policy.Policy) map[string]any {
	provider := p.Native.Pi.DefaultProvider
	if provider == "" {
		provider = "openai"
	}
	model := p.Native.Pi.DefaultModel
	if model == "" {
		model = "gpt-5.5"
	}
	thinking := p.Native.Pi.DefaultThinkingLevel
	if thinking == "" {
		thinking = "medium"
	}

	pkgs := p.Native.Pi.Packages
	if len(pkgs) == 0 {
		pkgs = []string{"../../.local/share/giocaizzi-harness"}
	}

	return map[string]any{
		"defaultProvider":      provider,
		"defaultModel":         model,
		"defaultThinkingLevel": thinking,
		"packages":             pkgs,
	}
}
