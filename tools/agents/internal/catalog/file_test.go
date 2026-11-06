package catalog

import (
	"strings"
	"testing"
)

func TestValidateRequiresProviders(t *testing.T) {
	t.Parallel()
	c := baseCatalog()
	c.Models["anthropic/claude-opus-5-5"] = Model{Name: "Claude Opus 5.5"}
	c.Tiers = map[string][]string{"powerful": {"anthropic/claude-opus-5-5"}}
	if err := c.Validate(); err != nil {
		t.Fatalf("valid catalog: %v", err)
	}
	c.Providers = nil
	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "providers") {
		t.Errorf("Validate() = %v, want a providers error", err)
	}
}

func TestValidateRejectsHandEditedTiers(t *testing.T) {
	t.Parallel()
	c := baseCatalog()
	c.Models["anthropic/claude-opus-5-5"] = Model{Name: "Claude Opus 5.5"}
	c.Tiers = map[string][]string{"powerful": {"anthropic/claude-opus-5-5"}, "extra": {}}
	if err := c.Validate(); err != nil {
		t.Fatalf("rule-consistent tiers (empty tier allowed): %v", err)
	}
	c.Tiers = map[string][]string{"extra": {"anthropic/claude-opus-5-5"}}
	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "tiers.extra: differs from the rule") ||
		!strings.Contains(err.Error(), "tiers.powerful: differs from the rule") {
		t.Errorf("Validate() = %v, want tiers.extra and tiers.powerful drift", err)
	}
}
