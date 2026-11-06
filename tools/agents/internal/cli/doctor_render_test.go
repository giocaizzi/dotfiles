package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/doctor"
	"github.com/giocaizzi/dotfiles/tools/agents/internal/term"
)

func suggestFinding(harness string) doctor.Finding {
	return doctor.Finding{
		Level: doctor.LevelSuggest, Kind: doctor.KindBetterModel, Harness: harness, Role: doctor.RoleDriver,
		Text:      "driver: a better model is available",
		Current:   &doctor.ModelRef{ID: "acme/old", Name: "Old 4", In: 3, Out: 15, Category: "coding", Rank: 9},
		Suggested: &doctor.ModelRef{ID: "acme/new", Name: "New 5", In: 2.5, Out: 12.5, Category: "coding", Rank: 2},
		Reason:    "better-ranked", Billing: "included in your plan",
	}
}

func sampleFindings() []doctor.Finding {
	warn := func(h string) doctor.Finding {
		return doctor.Finding{Level: doctor.LevelWarn, Kind: doctor.KindUnknownModel, Harness: h, Text: "scout: model ghost is not in the catalog", Fix: "harnessPolicy.models.tiers.scout"}
	}
	return []doctor.Finding{
		{Level: doctor.LevelWarn, Kind: doctor.KindStaleCatalog, Text: "model catalog synced 40 days ago (2026-09-27)", Fix: "chezmoi agents models sync"},
		{Level: doctor.LevelInfo, Kind: doctor.KindNoUsage, Text: "no usage data: account checks skipped"},
		warn("claude"),
		suggestFinding("claude"),
		warn("codex"),
		suggestFinding("codex"),
		{Level: doctor.LevelSuggest, Kind: doctor.KindNoAccount, Harness: "pi", Text: "pi: no account for provider acme", Fix: "chezmoi agents usage -live"},
	}
}

const wantPlain = `chezmoi agents doctor — models, accounts and capacity   (live)

! model catalog synced 40 days ago (2026-09-27)  · chezmoi agents models sync
i no usage data: account checks skipped

claude · codex
  ! scout: model ghost is not in the catalog  · harnessPolicy.models.tiers.scout
  ↑ driver   Old 4 → New 5   better-ranked · coding #9 → #2 · $3/$15 → $2.5/$12.5 /Mtok · included in your plan

pi
  ↑ pi: no account for provider acme  · chezmoi agents usage -live
`

func TestRenderDoctorPlain(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	renderDoctor(&buf, term.Style{}, sampleFindings(), "(live)")
	if diff := cmp.Diff(wantPlain, buf.String()); diff != "" {
		t.Errorf("render mismatch (-want +got):\n%s", diff)
	}
}

func TestRenderDoctorStyled(t *testing.T) {
	t.Parallel()
	st := term.Style{B: "<B>", M: "<M>", R: "<R>", G: "<G>", Y: "<Y>", C: "<C>", D: "<D>", O: "<O>"}
	var buf bytes.Buffer
	renderDoctor(&buf, st, sampleFindings(), "(live)")
	want := `<B>chezmoi agents doctor — models, accounts and capacity<O>   <D>(live)<O>

<Y>!<O> model catalog synced 40 days ago (2026-09-27)  <D>· chezmoi agents models sync<O>
<D>i<O> no usage data: account checks skipped

<B>claude · codex<O>
  <Y>!<O> scout: model ghost is not in the catalog  <D>· harnessPolicy.models.tiers.scout<O>
  <G>↑<O> <M>driver   <O>Old 4 → <M>New 5<O>   <D>better-ranked · coding #9 → #2 · $3/$15 → $2.5/$12.5 /Mtok · included in your plan<O>

<B>pi<O>
  <G>↑<O> pi: no account for provider acme  <D>· chezmoi agents usage -live<O>
`
	if diff := cmp.Diff(want, buf.String()); diff != "" {
		t.Errorf("styled render mismatch (-want +got):\n%s", diff)
	}
}

func TestRenderDoctorNothingToReport(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	renderDoctor(&buf, term.Style{}, nil, "(no usage cache; -live to refresh)")
	want := "chezmoi agents doctor — models, accounts and capacity   (no usage cache; -live to refresh)\n\n  ✓ nothing to report\n"
	if diff := cmp.Diff(want, buf.String()); diff != "" {
		t.Errorf("render mismatch (-want +got):\n%s", diff)
	}
}

// Groups of different harnesses stay apart and keep harness order, whatever
// the order of the findings; a group sits where its first harness is.
func TestRenderDoctorGroupOrder(t *testing.T) {
	t.Parallel()
	a := func(h string) doctor.Finding {
		return doctor.Finding{Level: doctor.LevelInfo, Harness: h, Text: "same"}
	}
	findings := []doctor.Finding{a("pi"), {Level: doctor.LevelInfo, Harness: "codex", Text: "other"}, a("claude"), a("zed")}
	var buf bytes.Buffer
	renderDoctor(&buf, term.Style{}, findings, "")
	want := "claude · pi · zed\n  i same\n\ncodex\n  i other\n"
	if !strings.HasSuffix(buf.String(), want) {
		t.Errorf("groups = %q, want suffix %q", buf.String(), want)
	}
}

// A suggest finding without models falls back to its text line.
func TestRenderDoctorSuggestWithoutModels(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	renderDoctor(&buf, term.Style{}, []doctor.Finding{{Level: doctor.LevelSuggest, Text: "try acme", Fix: "do it"}}, "")
	if want := "↑ try acme  · do it\n"; !strings.HasSuffix(buf.String(), want) {
		t.Errorf("output = %q, want suffix %q", buf.String(), want)
	}
}

// Strings from findings may echo cache data: no control character reaches the terminal.
func TestRenderDoctorSanitizes(t *testing.T) {
	t.Parallel()
	f := suggestFinding("cl\x1b[2Jaude")
	f.Role = "dri\x1bver"
	f.Current.Name = "Old\x07"
	f.Reason = "bet\x1bter"
	findings := []doctor.Finding{
		{Level: doctor.LevelWarn, Text: "bad\x1b[31m text", Fix: "fi\x1bx"},
		f,
	}
	var buf bytes.Buffer
	renderDoctor(&buf, term.Style{}, findings, "(me\x1bta)")
	if strings.ContainsRune(buf.String(), '\x1b') || strings.ContainsRune(buf.String(), '\x07') {
		t.Errorf("control character in output: %q", buf.String())
	}
}
