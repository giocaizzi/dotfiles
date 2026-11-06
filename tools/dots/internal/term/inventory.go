package term

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/inventory"
)

func groupTitle(src inventory.Source) string {
	switch src {
	case inventory.Release:
		return "Pinned releases"
	case inventory.Brew:
		return "Homebrew"
	case inventory.UV:
		return "uv tools"
	case inventory.Npm:
		return "npm tools"
	case inventory.Script:
		return "Official installers"
	case inventory.Apt, inventory.Winget:
		return string(src)
	}
	return string(src)
}

// Inventory prints the tooling overview: a table per source, the dotfiles checkout
// against its upstream, a summary and the next steps.
func (s Style) Inventory(w io.Writer, r inventory.Report, now time.Time) {
	link := "online"
	if r.Offline {
		link = "offline"
	}
	fmt.Fprintf(w, "%schezmoi dots outdated — every tool on this machine%s   %s%s · %s · %s%s\n\n", s.B, s.O, s.D, r.OS, r.Role, link, s.O)

	for _, g := range r.Groups {
		if len(g.Tools) == 0 {
			continue
		}
		fmt.Fprintf(w, "%s%s%s %s(%d)%s\n", s.B, groupTitle(g.Source), s.O, s.D, len(g.Tools), s.O)
		s.inventoryTable(w, g)
		if g.Note != "" {
			fmt.Fprintf(w, "  %s%s%s\n", s.D, g.Note, s.O)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "%ssource%s  %s\n", s.B, s.O, s.treeText(r.Tree, now))

	total := 0
	var parts []string
	counts := r.Counts()
	for _, c := range []struct {
		state inventory.State
		color string
		label string
	}{
		{inventory.UpToDate, s.G, "up to date"},
		{inventory.Outdated, s.Y, "outdated"},
		{inventory.Drifted, s.C, "drifted"},
		{inventory.Missing, s.R, "not installed"},
		{inventory.Unknown, s.D, "unknown"},
	} {
		n := counts[c.state]
		total += n
		if n > 0 {
			parts = append(parts, fmt.Sprintf("%s%d %s%s", c.color, n, c.label, s.O))
		}
	}
	parts = slices.Insert(parts, 0, fmt.Sprintf("%d tools", total))
	fmt.Fprintf(w, "%ssummary%s %s\n", s.B, s.O, strings.Join(parts, " "+s.D+"·"+s.O+" "))

	hints := r.Hints()
	if len(hints) == 0 {
		return
	}
	cmdW := 0
	for _, h := range hints {
		cmdW = max(cmdW, Width(h.Cmd))
	}
	fmt.Fprintf(w, "\n%snext%s\n", s.B, s.O)
	for _, h := range hints {
		if h.Cmd == "" {
			fmt.Fprintf(w, "  %s%s%s\n", s.D, h.Why, s.O)
			continue
		}
		fmt.Fprintf(w, "  %s  %s%s%s\n", Pad(s.M+h.Cmd+s.O, cmdW), s.D, h.Why, s.O)
	}
}

func (s Style) inventoryTable(w io.Writer, g inventory.Group) {
	pinned := g.Source == inventory.Release ||
		(g.Source == inventory.Npm && slices.ContainsFunc(g.Tools, func(t inventory.Tool) bool { return t.Pinned != "" }))
	header := []string{"TOOL", "INSTALLED", "LATEST", "STATE"}
	if pinned {
		header = []string{"TOOL", "INSTALLED", "PINNED", "LATEST", "STATE"}
	}
	dash := s.D + "—" + s.O
	unknown := s.D + "unknown" + s.O
	rows := make([][]string, 0, len(g.Tools))
	for _, t := range g.Tools {
		name := s.M + t.Name + s.O
		if t.ID != "" {
			name += " " + s.D + "(" + t.ID + ")" + s.O
		}
		installed := orElse(t.Installed, unknown)
		if t.Missing {
			installed = dash
		}
		row := []string{name, installed}
		if pinned {
			row = append(row, orElse(t.Pinned, dash))
		}
		row = append(row, orElse(t.Latest, unknown), s.stateText(t.State(g.Source)))
		rows = append(rows, row)
	}
	s.Table(w, header, rows)
}

func orElse(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func (s Style) stateText(st inventory.State) string {
	switch st {
	case inventory.UpToDate:
		return s.G + "✓ up to date" + s.O
	case inventory.Outdated:
		return s.Y + "↑ outdated" + s.O
	case inventory.Drifted:
		return s.C + "≠ drifted" + s.O
	case inventory.Missing:
		return s.R + "✗ not installed" + s.O
	case inventory.Unknown:
		return s.D + "? unknown" + s.O
	}
	return ""
}

// treeText describes the dotfiles checkout against its upstream.
func (s Style) treeText(t inventory.Tree, now time.Time) string {
	if !t.Known {
		return s.D + "not a git checkout" + s.O
	}
	branch := orElse(t.Branch, "(detached)")
	parts := []string{s.M + branch + s.O}
	switch {
	case t.Upstream == "":
		parts = append(parts, "no upstream")
	case t.Behind == 0 && t.Ahead == 0:
		parts = append(parts, s.G+"up to date with upstream"+s.O)
	default:
		if t.Behind > 0 {
			parts = append(parts, fmt.Sprintf("%s%d behind upstream%s", s.Y, t.Behind, s.O))
		}
		if t.Ahead > 0 {
			parts = append(parts, fmt.Sprintf("%s%d ahead%s", s.Y, t.Ahead, s.O))
		}
	}
	if t.Dirty > 0 {
		parts = append(parts, fmt.Sprintf("%s%d local change(s)%s", s.Y, t.Dirty, s.O))
	}
	if t.Fetched.IsZero() {
		parts = append(parts, s.D+"not fetched"+s.O)
	} else if a := age(now.Sub(t.Fetched)); a == "just now" {
		parts = append(parts, s.D+"fetched just now"+s.O)
	} else {
		parts = append(parts, s.D+"fetched "+a+" ago"+s.O)
	}
	return strings.Join(parts, " "+s.D+"·"+s.O+" ")
}

// age shortens d to one unit: "just now", "12m", "3h" or "4d".
func age(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return strconv.Itoa(int(d/time.Minute)) + "m"
	case d < 24*time.Hour:
		return strconv.Itoa(int(d/time.Hour)) + "h"
	default:
		return strconv.Itoa(int(d/(24*time.Hour))) + "d"
	}
}
