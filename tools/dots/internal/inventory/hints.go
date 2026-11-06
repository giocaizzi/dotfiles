package inventory

import (
	"fmt"
	"slices"
	"strings"
)

// Counts tallies the tools of every group by state.
func (r Report) Counts() map[State]int {
	n := map[State]int{}
	for _, g := range r.Groups {
		for _, t := range g.Tools {
			n[t.State(g.Source)]++
		}
	}
	return n
}

// Failing reports whether any tool is outdated, drifted or not installed.
func (r Report) Failing() bool {
	n := r.Counts()
	return n[Outdated]+n[Drifted]+n[Missing] > 0
}

// Hint is one next step: the command to run (empty when there is none) and why.
type Hint struct{ Cmd, Why string }

// names lists the sorted names of the tools of source s in one of states.
func (r Report) names(s Source, states ...State) []string {
	var out []string
	for _, g := range r.Groups {
		if g.Source != s {
			continue
		}
		for _, t := range g.Tools {
			if slices.Contains(states, t.State(s)) {
				out = append(out, t.Name)
			}
		}
	}
	slices.Sort(out)
	return out
}

// Hints lists the next steps the report calls for, in the order to take them.
func (r Report) Hints() []Hint {
	var hs []Hint
	add := func(cmd, why string) { hs = append(hs, Hint{cmd, why}) }
	t := r.Tree
	if t.Known && t.Behind > 0 {
		add("chezmoi dots upgrade dotfiles", fmt.Sprintf("pull %d commit(s) from %s, then re-run", t.Behind, t.Upstream))
	}
	if t.Known && t.Upstream != "" && t.Fetched.IsZero() {
		add("chezmoi dots outdated --fetch", t.Upstream+" was never fetched here")
	}
	if t.Ahead > 0 {
		add("git push", fmt.Sprintf("%d local commit(s) not on %s", t.Ahead, t.Upstream))
	}
	if n := r.names(Release, Outdated); len(n) > 0 {
		add("chezmoi dots bump "+strings.Join(n, " "), "pin the newer release(s)")
	}
	if n := r.names(Release, Drifted, Missing); len(n) > 0 {
		add("chezmoi apply", "install the pinned "+strings.Join(n, ", "))
	}
	for _, s := range []Source{Brew, Apt, Winget, UV, Npm} {
		if n := r.names(s, Outdated); len(n) > 0 {
			add("chezmoi dots upgrade "+string(s), fmt.Sprintf("upgrade %d %s package(s)", len(n), s))
		}
	}
	var specs []string
	for _, g := range r.Groups {
		if g.Source != Npm {
			continue
		}
		for _, tl := range g.Tools {
			if tl.State(Npm) == Drifted {
				id := tl.ID
				if id == "" {
					id = tl.Name
				}
				specs = append(specs, id+"@"+tl.Pinned)
			}
		}
	}
	if len(specs) > 0 {
		slices.Sort(specs)
		add("npm install -g "+strings.Join(specs, " "), "back to the registry pin")
	}
	var missing []string
	for _, s := range Sources() {
		if s != Release {
			missing = append(missing, r.names(s, Missing)...)
		}
	}
	if len(missing) > 0 {
		slices.Sort(missing)
		add("", "not installed: "+strings.Join(missing, ", ")+" (the install script runs when packages.toml changes; docs/OS.md#packages)")
	}
	return hs
}
