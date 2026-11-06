package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/term"
	usg "github.com/giocaizzi/dotfiles/tools/agents/internal/usage"
)

func TestSanitize(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"plain text":                "plain text",
		"\x1b[31mX\x1b[0m":          "[31mX[0m",
		"a\u009bb\u0085c\x7fd":      "abcd",
		"tab\there\nnew\rline\x00z": "tabherenewlinez",
		"héllo · ünï":               "héllo · ünï",
	}
	for in, want := range tests {
		if got := sanitize(in); got != want {
			t.Errorf("sanitize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderStripsControlCharacters(t *testing.T) {
	t.Parallel()
	snap := usg.Snapshot{
		Provider: "any\x1b]0;x\aprov", Account: "me\u009b@example.com", Harnesses: []string{"h\x1b[2J"},
		Plan:   usg.Plan{Name: "Pl\x1bus", Raw: "raw\x1b"},
		Status: usg.StatusStale, Note: "note\x1b[1m", Source: "src\x1b",
		FetchedAt: renderNow,
		Limits: []usg.Limit{
			{Name: "\x1b[31mX\x1b[0m and \u009b", Kind: usg.LimitWindow},
			{Name: "amount", Kind: usg.LimitCredits, Remaining: &usg.Amount{Value: 1, Unit: "US\x1bD"}},
		},
	}
	var out bytes.Buffer

	renderUsage(&out, term.Style{}, []usg.Snapshot{snap}, "(live)", renderNow)

	for _, r := range out.String() {
		if r == '\x1b' || (r >= 0x80 && r <= 0x9f) {
			t.Fatalf("output holds control character %q:\n%q", r, out.String())
		}
	}
	if !strings.Contains(out.String(), "[31mX[0m and ") {
		t.Errorf("limit name lost:\n%s", out.String())
	}
	if snap.Limits[0].Name != "\x1b[31mX\x1b[0m and \u009b" {
		t.Error("render mutated its input")
	}
}
