package inventory_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/inventory"
)

func TestCountsAndFailing(t *testing.T) {
	t.Parallel()
	ok := inventory.Tool{Name: "a", Installed: "1", Latest: "1"}
	old := inventory.Tool{Name: "b", Installed: "1", Latest: "2"}
	tests := []struct {
		name   string
		rep    inventory.Report
		counts map[inventory.State]int
		fail   bool
	}{
		{"empty", inventory.Report{}, map[inventory.State]int{}, false},
		{"fine", inventory.Report{Groups: []inventory.Group{{Source: inventory.Brew, Tools: []inventory.Tool{ok}}}},
			map[inventory.State]int{inventory.UpToDate: 1}, false},
		{"outdated", inventory.Report{Groups: []inventory.Group{{Source: inventory.Brew, Tools: []inventory.Tool{ok, old}}}},
			map[inventory.State]int{inventory.UpToDate: 1, inventory.Outdated: 1}, true},
		{"missing", inventory.Report{Groups: []inventory.Group{{Source: inventory.Apt, Tools: []inventory.Tool{{Name: "c", Missing: true}}}}},
			map[inventory.State]int{inventory.Missing: 1}, true},
		{"unknown only", inventory.Report{Groups: []inventory.Group{{Source: inventory.UV, Tools: []inventory.Tool{{Name: "d", Installed: "1"}}}}},
			map[inventory.State]int{inventory.Unknown: 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.counts, tt.rep.Counts()); diff != "" {
				t.Errorf("Counts() mismatch (-want +got):\n%s", diff)
			}
			if got := tt.rep.Failing(); got != tt.fail {
				t.Errorf("Failing() = %v, want %v", got, tt.fail)
			}
		})
	}
}

func TestHints(t *testing.T) {
	t.Parallel()
	fetched := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rel := inventory.Group{Source: inventory.Release, Tools: []inventory.Tool{
		{Name: "zeta", Installed: "1.0", Pinned: "1.0", Latest: "1.1"},
		{Name: "alpha", Installed: "1.0", Pinned: "1.0", Latest: "2.0"},
		{Name: "drift", Installed: "0.9", Pinned: "1.0", Latest: "1.0"},
		{Name: "gone", Missing: true, Pinned: "1.0"},
	}}
	npm := inventory.Group{Source: inventory.Npm, Tools: []inventory.Tool{
		{Name: "tsc", ID: "typescript", Installed: "6.0.1", Pinned: "5", Latest: "5.9.2"},
		{Name: "old", Installed: "1", Latest: "2"},
	}}
	brew := inventory.Group{Source: inventory.Brew, Tools: []inventory.Tool{
		{Name: "b", Installed: "1", Latest: "2"},
		{Name: "a", Installed: "1", Latest: "2"},
		{Name: "z", Missing: true},
	}}
	script := inventory.Group{Source: inventory.Script, Tools: []inventory.Tool{{Name: "bun", Missing: true}}}
	tests := []struct {
		name string
		rep  inventory.Report
		want []inventory.Hint
	}{
		{"nothing", inventory.Report{Tree: inventory.Tree{Known: true, Fetched: fetched}}, nil},
		{"unknown tree", inventory.Report{Tree: inventory.Tree{Behind: 2, Upstream: "origin/main"}}, nil},
		{"tree", inventory.Report{Tree: inventory.Tree{Known: true, Upstream: "origin/main", Behind: 3, Ahead: 1}}, []inventory.Hint{
			{"chezmoi dots upgrade dotfiles", "pull 3 commit(s) from origin/main, then re-run"},
			{"chezmoi dots outdated --fetch", "origin/main was never fetched here"},
			{"git push", "1 local commit(s) not on origin/main"},
		}},
		{"no upstream", inventory.Report{Tree: inventory.Tree{Known: true}}, nil},
		{"all", inventory.Report{Groups: []inventory.Group{rel, brew, npm, script}, Tree: inventory.Tree{Known: true, Upstream: "origin/main", Fetched: fetched}}, []inventory.Hint{
			{"chezmoi dots bump alpha zeta", "pin the newer release(s)"},
			{"chezmoi apply", "install the pinned drift, gone"},
			{"chezmoi dots upgrade brew", "upgrade 2 brew package(s)"},
			{"chezmoi dots upgrade npm", "upgrade 1 npm package(s)"},
			{"npm install -g typescript@5", "back to the registry pin"},
			{"", "not installed: bun, z (the install script runs when packages.toml changes; docs/OS.md#packages)"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.rep.Hints()); diff != "" {
				t.Errorf("Hints() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
