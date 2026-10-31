package registry

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/finding"
)

func data(entries map[string]any) map[string]any { return map[string]any{"edit": entries} }

func TestParseValid(t *testing.T) {
	t.Parallel()
	reg, fs := Parse(data(map[string]any{
		"a/b":   map[string]any{"desc": "A", "source": "x/y.md", "diff": "first", "check": []any{"make", "{file}"}, "note": "n"},
		"local": map[string]any{"desc": "L", "local": ".ssh/config.d", "check": []any{"ssh"}},
	}))
	if len(fs) != 0 {
		t.Fatalf("Parse() findings = %v, want none", fs)
	}
	e, _ := reg.Get("a/b")
	want := Entry{Name: "a/b", Desc: "A", Source: "x/y.md", Check: []string{"make", "{file}"}, Note: "n", First: true}
	if diff := cmp.Diff(want, e); diff != "" {
		t.Errorf("entry (-want +got):\n%s", diff)
	}
	if got, ok := reg.BySource("x/y.md"); !ok || got.Name != "a/b" {
		t.Errorf("BySource() = %v, %v", got, ok)
	}
	if got, ok := reg.ByLocal(".ssh/config.d/work"); !ok || got.Name != "local" {
		t.Errorf("ByLocal() = %v, %v", got, ok)
	}
	if diff := cmp.Diff([]string{"a/b", "local"}, reg.Names()); diff != "" {
		t.Errorf("Names() (-want +got):\n%s", diff)
	}
}

func TestParseFindings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		entry map[string]any
		want  string
	}{
		{"no desc", map[string]any{"source": "a"}, "desc is required"},
		{"no source", map[string]any{"desc": "d"}, "source is required"},
		{"unknown key", map[string]any{"desc": "d", "source": "a", "colour": 1}, "unknown key colour"},
		{"bad diff", map[string]any{"desc": "d", "source": "a", "diff": "all"}, `diff must be "first"`},
		{"string check", map[string]any{"desc": "d", "source": "a", "check": "make lint"}, "check must be a command as a list"},
		{"empty check", map[string]any{"desc": "d", "source": "a", "check": []any{}}, "check must be a command as a list"},
		{"absolute local", map[string]any{"desc": "d", "local": "/etc"}, "local must be a folder relative"},
		{"parent local", map[string]any{"desc": "d", "local": "a/../b"}, "local must be a folder relative"},
		{"local with source", map[string]any{"desc": "d", "local": "a", "source": "b"}, "local excludes source"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, fs := Parse(data(map[string]any{"e": tt.entry}))
			if got := finding.Lines(fs); !strings.Contains(got, tt.want) {
				t.Errorf("findings = %q, want one containing %q", got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()
	src := fstest.MapFS{
		"a.md":                {},
		".chezmoidata/x.toml": {},
		".chezmoidata/y.toml": {},
		"dir/secret.age":      {},
		"dir/other.age":       {},
	}
	reg, _ := Parse(data(map[string]any{
		"a":    map[string]any{"desc": "d", "source": "a.md"},
		"x":    map[string]any{"desc": "d", "source": ".chezmoidata/x.toml"},
		"sec":  map[string]any{"desc": "d", "source": "dir/secret.age"},
		"gone": map[string]any{"desc": "d", "source": "missing.md"},
	}))
	want := []string{
		"gone: source missing.md does not exist",
		".chezmoidata/y.toml: no edit registry entry",
		"dir/other.age: no edit registry entry",
	}
	got := strings.Split(finding.Lines(reg.Validate(src)), "\n")
	if diff := cmp.Diff([]string{".chezmoidata/y.toml: no edit registry entry", "dir/other.age: no edit registry entry", "gone: source missing.md does not exist"}, got); diff != "" {
		t.Errorf("Validate() (-want +got):\n%s\n(want set %v)", diff, want)
	}
}
