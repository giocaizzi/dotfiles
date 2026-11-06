package catalog

import (
	"strings"
	"testing"
)

func TestValidateRequiresProviders(t *testing.T) {
	t.Parallel()
	c := baseCatalog()
	c.Models["anthropic/claude-opus-5-5"] = Model{Name: "Claude Opus 5.5"}
	if err := c.Validate(); err != nil {
		t.Fatalf("valid catalog: %v", err)
	}
	c.Providers = nil
	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "providers") {
		t.Errorf("Validate() = %v, want a providers error", err)
	}
}
