package cli

import (
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/doctor"
	"github.com/giocaizzi/dotfiles/tools/agents/internal/term"
)

const roleCol = 9

// renderDoctor writes the report. meta is the dim text after the title, e.g.
// "(live)". Global findings come first, then one group per set of harnesses
// with identical findings.
func renderDoctor(w io.Writer, st term.Style, findings []doctor.Finding, meta string) {
	fmt.Fprintf(w, "%schezmoi agents doctor — models, accounts and capacity%s   %s%s%s\n\n", st.B, st.O, st.D, sanitize(meta), st.O)
	if len(findings) == 0 {
		fmt.Fprintf(w, "  %s✓%s nothing to report\n", st.G, st.O)
		return
	}

	var global []doctor.Finding
	byHarness := map[string][]doctor.Finding{}
	var names []string
	for _, f := range findings {
		if f.Harness == "" {
			global = append(global, f)
			continue
		}
		if _, ok := byHarness[f.Harness]; !ok {
			names = append(names, f.Harness)
		}
		byHarness[f.Harness] = append(byHarness[f.Harness], f)
	}

	for _, f := range global {
		fmt.Fprintln(w, findingLine(st, f, ""))
	}
	groups := harnessGroups(orderHarnesses(names), byHarness)
	if len(global) > 0 && len(groups) > 0 {
		fmt.Fprintln(w)
	}
	for i, g := range groups {
		if i > 0 {
			fmt.Fprintln(w)
		}
		heads := make([]string, len(g))
		for j, h := range g {
			heads[j] = sanitize(h)
		}
		fmt.Fprintf(w, "%s%s%s\n", st.B, strings.Join(heads, " · "), st.O)
		for _, f := range byHarness[g[0]] {
			fmt.Fprintln(w, findingLine(st, f, "  "))
		}
	}
}

// harnessGroups merges the harnesses (already in display order) whose
// findings are identical but for the Harness field; a group sits where its
// first harness is.
func harnessGroups(ordered []string, by map[string][]doctor.Finding) [][]string {
	var groups [][]string
	for _, h := range ordered {
		i := slices.IndexFunc(groups, func(g []string) bool { return sameFindings(by[g[0]], by[h]) })
		if i < 0 {
			groups = append(groups, []string{h})
			continue
		}
		groups[i] = append(groups[i], h)
	}
	return groups
}

func sameFindings(a, b []doctor.Finding) bool {
	return slices.EqualFunc(a, b, func(x, y doctor.Finding) bool {
		x.Harness, y.Harness = "", ""
		return reflect.DeepEqual(x, y)
	})
}

func marker(st term.Style, l doctor.Level) string {
	switch l {
	case doctor.LevelWarn:
		return st.Y + "!" + st.O
	case doctor.LevelSuggest:
		return st.G + "↑" + st.O
	case doctor.LevelInfo:
	}
	return st.D + "i" + st.O
}

// findingLine is one finding as a line, indented by indent.
func findingLine(st term.Style, f doctor.Finding, indent string) string {
	if f.Level == doctor.LevelSuggest && f.Current != nil && f.Suggested != nil {
		return indent + suggestLine(st, f)
	}
	line := indent + marker(st, f.Level) + " " + sanitize(f.Text)
	if f.Fix != "" {
		line += "  " + st.D + "· " + sanitize(f.Fix) + st.O
	}
	return line
}

func suggestLine(st term.Style, f doctor.Finding) string {
	cur, sug := f.Current, f.Suggested
	var parts []string
	if f.Reason != "" {
		parts = append(parts, sanitize(f.Reason))
	}
	if cur.Category != "" {
		parts = append(parts, fmt.Sprintf("%s #%d → #%d", sanitize(cur.Category), cur.Rank, sug.Rank))
	}
	parts = append(parts, fmt.Sprintf("$%s/$%s → $%s/$%s /Mtok", num(cur.In), num(cur.Out), num(sug.In), num(sug.Out)))
	if f.Billing != "" {
		parts = append(parts, sanitize(f.Billing))
	}
	return fmt.Sprintf("%s↑%s %s%s%s%s → %s%s%s   %s%s%s",
		st.G, st.O,
		st.M, term.Pad(sanitize(f.Role), roleCol), st.O,
		sanitize(cur.Name), st.M, sanitize(sug.Name), st.O,
		st.D, strings.Join(parts, " · "), st.O)
}
