package cli

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/term"
	usg "github.com/giocaizzi/dotfiles/tools/agents/internal/usage"
)

const (
	providerCol = 16
	minNameCol  = 14
	barWidth    = 20
)

// sanitize drops control characters (C0, DEL, C1, ESC and so tab and newline
// too) from a string that came from a provider or the cache, so it cannot
// drive the terminal.
func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

// sanitized returns a copy of the snapshot with every printable string cleaned.
func sanitized(s usg.Snapshot) usg.Snapshot {
	s.Provider = sanitize(s.Provider)
	s.Account = sanitize(s.Account)
	s.Note = sanitize(s.Note)
	s.Source = sanitize(s.Source)
	s.Plan.Name = sanitize(s.Plan.Name)
	s.Plan.Raw = sanitize(s.Plan.Raw)
	s.Plan.Kind = usg.Kind(sanitize(string(s.Plan.Kind)))
	s.Status = usg.Status(sanitize(string(s.Status)))
	hs := make([]string, len(s.Harnesses))
	for i, h := range s.Harnesses {
		hs[i] = sanitize(h)
	}
	s.Harnesses = hs
	limits := make([]usg.Limit, len(s.Limits))
	for i, l := range s.Limits {
		l.Name = sanitize(l.Name)
		l.Note = sanitize(l.Note)
		l.Used, l.Remaining, l.Cap = cleanAmount(l.Used), cleanAmount(l.Remaining), cleanAmount(l.Cap)
		limits[i] = l
	}
	s.Limits = limits
	return s
}

func cleanAmount(a *usg.Amount) *usg.Amount {
	if a == nil {
		return nil
	}
	c := *a
	c.Unit = sanitize(c.Unit)
	return &c
}

// planText is the plan column: its name, or unknown (<raw>), then the account.
func planText(s usg.Snapshot) string {
	var text string
	switch {
	case s.Plan.Name != "":
		text = s.Plan.Name
	case s.Plan.Kind == usg.KindUnknown || s.Plan.Kind == "":
		text = "unknown"
		if s.Plan.Raw != "" {
			text += " (" + s.Plan.Raw + ")"
		}
	default:
		text = string(s.Plan.Kind)
	}
	if s.Account != "" {
		text += " · " + s.Account
	}
	return text
}

func sortSnapshots(snaps []usg.Snapshot) []usg.Snapshot {
	out := append([]usg.Snapshot(nil), snaps...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Provider != out[j].Provider {
			return out[i].Provider < out[j].Provider
		}
		if out[i].Account != out[j].Account {
			return out[i].Account < out[j].Account
		}
		return strings.Join(out[i].Harnesses, ",") < strings.Join(out[j].Harnesses, ",")
	})
	return out
}

// renderUsage writes the report. meta is the dim text after the title, e.g.
// "(live)" or "(cached 12m ago; -live to refresh)".
func renderUsage(w io.Writer, st term.Style, snaps []usg.Snapshot, meta string, now time.Time) {
	fmt.Fprintf(w, "%schezmoi agents usage — subscription usage per provider account%s   %s%s%s\n\n", st.B, st.O, st.D, meta, st.O)

	clean := make([]usg.Snapshot, len(snaps))
	for i, s := range snaps {
		clean[i] = sanitized(s)
	}
	snaps = sortSnapshots(clean)
	nameW := minNameCol
	for _, s := range snaps {
		for _, l := range s.Limits {
			nameW = max(nameW, term.Width(l.Name))
		}
	}
	planWidth := 0
	for _, s := range snaps {
		planWidth = max(planWidth, term.Width(planText(s)))
	}

	for _, s := range snaps {
		line := term.Pad(st.B+s.Provider+st.O, providerCol) +
			term.Pad(st.M+planText(s)+st.O, planWidth+2) +
			st.D + strings.Join(s.Harnesses, ", ") + st.O
		if s.Status == usg.StatusStale {
			line += "  " + st.D + staleText(s, now) + st.O
		}
		fmt.Fprintln(w, strings.TrimRight(line, " "))

		for _, l := range s.Limits {
			fmt.Fprintln(w, limitLine(st, l, nameW, now))
		}
		if msg := warning(s); msg != "" {
			fmt.Fprintf(w, "  %s!%s %s\n", st.Y, st.O, msg)
		}
	}
}

// warning is the text of the "!" line under a snapshot; empty when there is none.
func warning(s usg.Snapshot) string {
	switch {
	case s.Status == usg.StatusOK:
		return ""
	case s.Status == usg.StatusStale:
		return s.Note
	case s.Note != "":
		return s.Note
	}
	return string(s.Status)
}

func staleText(s usg.Snapshot, now time.Time) string {
	parts := []string{"stale"}
	if s.Source != "" {
		parts = append(parts, s.Source)
	}
	if !s.FetchedAt.IsZero() {
		parts = append(parts, usg.HumanDuration(now.Sub(s.FetchedAt))+" ago")
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func limitLine(st term.Style, l usg.Limit, nameW int, now time.Time) string {
	head := "  " + term.Pad(l.Name, nameW) + " "
	dim := func(s string) string { return st.D + s + st.O }
	switch {
	case l.Hidden:
		if l.Note != "" {
			return head + dim(l.Note)
		}
		return head + dim("managed by org")
	case l.Disabled:
		return head + dim("not applied now")
	}

	var parts []string
	if l.Scope == usg.ScopePool {
		parts = append(parts, "pool")
	}
	if t := amountText(l); t != "" {
		parts = append(parts, t)
	}
	if l.Unlimited {
		parts = append(parts, "unlimited")
	}
	if l.ResetsAt != nil {
		parts = append(parts, resetText(*l.ResetsAt, now))
	}
	if l.Note != "" {
		parts = append(parts, l.Note)
	}

	var line string
	if l.UsedPct != nil {
		line = head + st.Bar(*l.UsedPct, barWidth) + fmt.Sprintf("  %3.0f%%", *l.UsedPct)
		if len(parts) > 0 {
			line += "   " + dim(strings.Join(parts, " · "))
		}
		return line
	}
	if len(parts) == 0 {
		parts = []string{"no data"}
	}
	return head + dim(strings.Join(parts, " · "))
}

func resetText(at, now time.Time) string {
	until := at.Sub(now)
	switch {
	case until <= 0:
		return "reset"
	case until < 24*time.Hour:
		return "resets in " + usg.HumanDuration(until)
	case until < 7*24*time.Hour:
		return "resets " + at.In(now.Location()).Format("Mon 15:04")
	}
	return "resets " + at.In(now.Location()).Format("Jan 2")
}

func num(v float64) string {
	return strings.TrimSuffix(strings.TrimRight(strconv.FormatFloat(v, 'f', 2, 64), "0"), ".")
}

// amountText is "90 of 300 left", "12.5 USD left" or "3 of 10 used".
func amountText(l usg.Limit) string {
	unit := func(a *usg.Amount) string {
		if a == nil || a.Unit == "" || a.Unit == "requests" {
			return ""
		}
		return " " + a.Unit
	}
	switch {
	case l.Remaining != nil && l.Cap != nil:
		return num(l.Remaining.Value) + " of " + num(l.Cap.Value) + unit(l.Cap) + " left"
	case l.Remaining != nil:
		return num(l.Remaining.Value) + unit(l.Remaining) + " left"
	case l.Used != nil && l.Cap != nil:
		return num(l.Used.Value) + " of " + num(l.Cap.Value) + unit(l.Cap) + " used"
	case l.Used != nil:
		return num(l.Used.Value) + unit(l.Used) + " used"
	}
	return ""
}
