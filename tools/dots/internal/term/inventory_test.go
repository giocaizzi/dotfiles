package term

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/inventory"
)

var update = flag.Bool("update", false, "rewrite the golden files")

var now = time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

func richReport() inventory.Report {
	return inventory.Report{
		OS: "darwin", Role: "personal",
		Groups: []inventory.Group{
			{Source: inventory.Release, Tools: []inventory.Tool{
				{Name: "age", Installed: "1.2.1", Pinned: "1.2.1", Latest: "1.2.1"},
				{Name: "chezmoi-agents", Installed: "0.7.0", Pinned: "0.7.1", Latest: "0.7.1"},
				{Name: "gitleaks", Installed: "8.28.0", Pinned: "8.28.0", Latest: "8.30.0"},
				{Name: "shfmt", Missing: true, Pinned: "3.12.0", Latest: "3.12.0"},
			}},
			{Source: inventory.Brew, Tools: []inventory.Tool{
				{Name: "fd", ID: "fd-find", Installed: "10.2.0", Latest: "10.3.0"},
				{Name: "jq", Installed: "1.8.1", Latest: "1.8.1"},
			}, Note: "latest versions come from brew outdated"},
			{Source: inventory.UV, Tools: []inventory.Tool{{Name: "ruff", Installed: "0.14.0"}}, Note: "offline: latest versions not looked up"},
			{Source: inventory.Npm, Tools: []inventory.Tool{
				{Name: "typescript", Installed: "6.0.1", Pinned: "5", Latest: "5.9.2"},
				{Name: "prettier", Installed: "3.6.0", Latest: "3.6.0"},
			}},
			{Source: inventory.Apt},
			{Source: inventory.Script, Tools: []inventory.Tool{{Name: "bun", Missing: true}}},
		},
		Tree: inventory.Tree{Known: true, Branch: "main", Upstream: "origin/main", Behind: 2, Dirty: 3, Fetched: now.Add(-3 * time.Hour)},
	}
}

func tokens() Style {
	return Style{B: "<b>", M: "<m>", R: "<r>", G: "<g>", Y: "<y>", C: "<c>", D: "<d>", O: "</>"}
}

func TestInventoryGolden(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		file  string
		style Style
	}{
		{"inventory.golden", Style{}},
		{"inventory_tokens.golden", tokens()},
	} {
		t.Run(tt.file, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			tt.style.Inventory(&buf, richReport(), now)
			path := filepath.Join("testdata", tt.file)
			if *update {
				if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(path) //nolint:gosec // a fixed testdata path
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(string(want), buf.String()); diff != "" {
				t.Errorf("Inventory() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestInventoryClean(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	Style{}.Inventory(&buf, inventory.Report{OS: "linux", Role: "light", Offline: true,
		Groups: []inventory.Group{{Source: inventory.Brew, Tools: []inventory.Tool{{Name: "jq", Installed: "1", Latest: "1"}}}},
		Tree:   inventory.Tree{Branch: "main"},
	}, now)
	want := "chezmoi dots outdated — every tool on this machine   linux · light · offline\n\n" +
		"Homebrew (1)\n" +
		"  TOOL  INSTALLED  LATEST  STATE\n" +
		"  jq    1          1       ✓ up to date\n\n" +
		"source  not a git checkout\n" +
		"summary 1 tools · 1 up to date\n"
	if diff := cmp.Diff(want, buf.String()); diff != "" {
		t.Errorf("Inventory() mismatch (-want +got):\n%s", diff)
	}
}

func TestTreeText(t *testing.T) {
	t.Parallel()
	fetched := now.Add(-12 * time.Minute)
	tests := []struct {
		name string
		tree inventory.Tree
		want string
	}{
		{"no git", inventory.Tree{}, "not a git checkout"},
		{"no upstream", inventory.Tree{Known: true, Branch: "dev"}, "dev · no upstream · not fetched"},
		{"even", inventory.Tree{Known: true, Branch: "main", Upstream: "origin/main", Fetched: fetched}, "main · up to date with upstream · fetched 12m ago"},
		{"ahead and behind", inventory.Tree{Known: true, Branch: "main", Upstream: "origin/main", Ahead: 1, Behind: 4, Dirty: 2, Fetched: now}, "main · 4 behind upstream · 1 ahead · 2 local change(s) · fetched just now"},
		{"detached", inventory.Tree{Known: true, Upstream: "u", Fetched: now.Add(-50 * time.Hour)}, "(detached) · up to date with upstream · fetched 2d ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := (Style{}).treeText(tt.tree, now); got != tt.want {
				t.Errorf("treeText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAge(t *testing.T) {
	t.Parallel()
	tests := []struct {
		d    time.Duration
		want string
	}{
		{-time.Hour, "just now"},
		{59 * time.Second, "just now"},
		{time.Minute, "1m"},
		{59 * time.Minute, "59m"},
		{3*time.Hour + 20*time.Minute, "3h"},
		{23 * time.Hour, "23h"},
		{96 * time.Hour, "4d"},
	}
	for _, tt := range tests {
		if got := age(tt.d); got != tt.want {
			t.Errorf("age(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
