package catalog

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/term"
)

var update = flag.Bool("update", false, "rewrite golden files")

func sample() Catalog {
	offer := func(id string, in, out float64) Offer { return Offer{ID: id, In: in, Out: out} }
	rank := func(r int, v string) Rank { return Rank{Rank: r, Variant: v} }
	return Catalog{
		Schema:   Schema,
		SyncedAt: "2026-11-06",
		Sources:  SourceInfo{Arena: ArenaInfo{Board: ArenaBoard, PublishDate: "2026-10-28"}},
		Tiers: map[string][]string{
			"lightweight": {"openai/gpt-6-luna", "copilot/only-mini"},
			"versatile":   {"anthropic/claude-sonnet-5-5", "openai/ghost"},
			"powerful":    {"anthropic/claude-opus-5-5"},
		},
		Models: map[string]Model{
			"openai/gpt-6-luna": {
				Name: "GPT-6 Luna",
				Providers: map[string]Offer{
					"openai":         offer("gpt-6-luna", 0.1, 0.5),
					"github-copilot": offer("gpt-6-luna", 0.1, 0.5),
				},
				Arena: map[string]Rank{"text": rank(12, "Low"), "coding": rank(60, "")},
			},
			"copilot/only-mini": {
				Name:      "Only Mini",
				Providers: map[string]Offer{"github-copilot": offer("only-mini", 0.25, 2)},
				Arena:     map[string]Rank{"text": rank(30, "")},
			},
			"anthropic/claude-sonnet-5-5": {
				Name:      "Claude Sonnet 5.5",
				Providers: map[string]Offer{"anthropic": offer("claude-sonnet-5-5", 3, 15)},
				Arena:     map[string]Rank{"text": rank(4, "High"), "coding": rank(2, ""), "math": rank(9, "")},
				External:  External{ECI: 151.4, METRHorizonMin: 95},
			},
			"anthropic/claude-opus-5-5": {
				Name: "Claude Opus 5.5",
				Providers: map[string]Offer{
					"anthropic":      offer("claude-opus-5-5", 5, 25),
					"github-copilot": offer("claude-opus-5-5", 5, 25),
				},
				Arena: map[string]Rank{"text": rank(1, "High"), "coding": rank(2, "High"), "webdev": rank(3, "")},
			},
			"openai/untiered-pipe": {
				Name:      "Untiered | Pipe",
				Providers: map[string]Offer{"openai": offer("x", 1, 4)},
				Arena:     map[string]Rank{"agent": rank(7, "")},
			},
			"openai/gpt-6-sol": {
				Name:      "GPT-6 Sol",
				Providers: map[string]Offer{"openai": offer("gpt-6-sol", 2.5, 10)},
			},
		},
	}
}

var colorStyle = term.Style{B: "<b>", M: "<m>", R: "<r>", G: "<g>", Y: "<y>", C: "<c>", D: "<d>", O: "</>"}

func golden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.MkdirAll("testdata", 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(string(want), string(got)); diff != "" {
		t.Errorf("%s mismatch (-want +got):\n%s", name, diff)
	}
}

func terminal(st term.Style, c Catalog) string {
	var b bytes.Buffer
	WriteTerminal(&b, st, c)
	return b.String()
}

func TestRenderGolden(t *testing.T) {
	t.Parallel()
	c := sample()
	t.Run("markdown", func(t *testing.T) {
		t.Parallel()
		golden(t, "render_markdown.golden", Markdown(c))
	})
	t.Run("plain", func(t *testing.T) {
		t.Parallel()
		golden(t, "render_terminal_plain.golden", []byte(terminal(term.Style{}, c)))
	})
	t.Run("color", func(t *testing.T) {
		t.Parallel()
		golden(t, "render_terminal_color.golden", []byte(terminal(colorStyle, c)))
	})
}

func TestExternalColumnsHiddenWhenZero(t *testing.T) {
	t.Parallel()
	c := sample()
	m := c.Models["anthropic/claude-sonnet-5-5"]
	m.External = External{}
	c.Models["anthropic/claude-sonnet-5-5"] = m
	md := string(Markdown(c))
	for _, h := range []string{"ECI", "TB", "SWE-V", "METR", "OSWorld", "Epoch"} {
		if strings.Contains(md, h) {
			t.Errorf("markdown shows %q with no external data", h)
		}
	}
	if out := terminal(term.Style{}, c); strings.Contains(out, "ECI") || strings.Contains(out, "Epoch") {
		t.Errorf("terminal shows external data:\n%s", out)
	}
}

func TestExternalFormats(t *testing.T) {
	t.Parallel()
	c := Catalog{
		Tiers: map[string][]string{"extra": {"a/x"}},
		Models: map[string]Model{"a/x": {Name: "X", External: External{
			ECI: 150.6, TerminalBench: 0.512, SWEBenchVerified: 80.4, METRHorizonMin: 120, OSWorld: 0.5,
		}}},
	}
	want := "| X |  |  |  |  |  |  |  |  |  | 151 | 51% | 80% | 120m | 50% |"
	if got := string(Markdown(c)); !strings.Contains(got, want) {
		t.Errorf("row %q not in:\n%s", want, got)
	}
}

func TestUntieredEmpty(t *testing.T) {
	t.Parallel()
	c := Catalog{Tiers: map[string][]string{}, Models: map[string]Model{}}
	out := terminal(term.Style{}, c)
	if !strings.Contains(out, "  ✓ every tracked model has a tier\n") {
		t.Errorf("missing check line:\n%s", out)
	}
	if md := string(Markdown(c)); !strings.Contains(md, "## Untiered\n\n_None._\n") {
		t.Errorf("markdown untiered not empty:\n%s", md)
	}
}

func TestMarkdownPipeEscapingAndEnding(t *testing.T) {
	t.Parallel()
	md := string(Markdown(sample()))
	if !strings.Contains(md, `| Untiered \| Pipe |`) {
		t.Errorf("pipe not escaped:\n%s", md)
	}
	if !strings.HasSuffix(md, ".\n") || strings.HasSuffix(md, "\n\n") {
		t.Errorf("want exactly one trailing newline, got %q", md[len(md)-10:])
	}
}

func TestBestWhenTieBreak(t *testing.T) {
	t.Parallel()
	c := Catalog{Models: map[string]Model{
		"b/two": {Name: "Two", Arena: map[string]Rank{"text": {Rank: 1}}},
		"a/one": {Name: "One", Arena: map[string]Rank{"text": {Rank: 1, Variant: "High"}}},
	}}
	got := bestWhen(c)
	want := []best{{"Text", "One", "High", 1}}
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(best{})); diff != "" {
		t.Errorf("bestWhen mismatch (-want +got):\n%s", diff)
	}
}
