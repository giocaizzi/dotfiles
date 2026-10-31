// Package harnessreg checks .chezmoidata/harness.toml against the harness repo.
package harnessreg

import (
	"slices"
	"sort"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/finding"
)

// Check compares the plugins listed here with the plugin directories upstream.
// plugins maps plugin to the harnesses it is enabled for; choices are the
// harnesses a machine can select.
func Check(plugins map[string][]string, personal, choices, upstream []string) []finding.Finding {
	var out []finding.Finding
	listed := make([]string, 0, len(plugins))
	for p, hs := range plugins {
		listed = append(listed, p)
		for _, h := range hs {
			if !slices.Contains(choices, h) {
				out = append(out, finding.New("harnessPlugins."+p, "unknown harness %s", h))
			}
		}
	}
	sort.Strings(listed)
	for _, p := range personal {
		if _, ok := plugins[p]; !ok {
			out = append(out, finding.New("harnessPersonalPlugins", "%s is not in [harnessPlugins]", p))
		}
	}
	if len(upstream) == 0 {
		return append(out, finding.New("", "no plugin directories found upstream"))
	}
	for _, p := range listed {
		if !slices.Contains(upstream, p) {
			out = append(out, finding.New("harnessPlugins."+p, "not a plugin of giocaizzi/harness"))
		}
	}
	for _, p := range upstream {
		if _, ok := plugins[p]; !ok {
			out = append(out, finding.New("harnessPlugins."+p, "plugin of giocaizzi/harness missing here"))
		}
	}
	return out
}
