package catalog

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func pm(provider, id, name, canonical, date string, in, out float64) ProviderModel {
	return ProviderModel{
		Provider: provider, ID: id, Name: name, CanonicalID: canonical, ReleaseDate: date,
		In: in, Out: out, ToolCall: true, TextOutput: true,
	}
}

func baseCatalog() Catalog {
	return Catalog{
		Schema: Schema,
		Since:  "2026-01-01",
		Tiers: map[string][]string{
			"powerful": {"anthropic/claude-opus-5-5"},
		},
		Ignore: []string{"openai/gpt-realtime-2.1"},
		Models: map[string]Model{},
	}
}

func TestApplyKeepRules(t *testing.T) {
	t.Parallel()
	c := baseCatalog()
	c.Tiers["lightweight"] = []string{"anthropic/claude-haiku-3"}
	c.Tiers["extra"] = []string{"openai/gpt-realtime-2.1"}
	src := Sources{Today: "2026-11-06", ModelsDev: []ProviderModel{
		pm("anthropic", "claude-opus-5-5", "Claude Opus 5.5", "anthropic/claude-opus-5-5", "2026-09-22", 4, 20),
		pm("github-copilot", "claude-opus-5.5", "Claude Opus 5.5 (copilot)", "anthropic/claude-opus-5-5", "2026-09-23", 4, 20),
		pm("github-copilot", "gemini-3.6-flash", "Gemini 3.6 Flash", "google/gemini-3.6-flash", "2026-07-21", 0.75, 3.75),
		pm("anthropic", "claude-3-opus", "Claude 3 Opus", "anthropic/claude-3-opus", "2024-02-29", 15, 75),
		pm("anthropic", "claude-haiku-3", "Claude Haiku 3", "anthropic/claude-haiku-3", "2024-03-07", 0.25, 1.25),
		pm("openai", "gpt-6-luna", "GPT-6 Luna", "openai/gpt-6-luna", "2026-09-22", 0.1, 0.5),
		pm("openai", "gpt-realtime-2.1", "GPT-Realtime-2.1", "openai/gpt-realtime-2.1", "2026-07-06", 4, 24),
		pm("openai", "gpt-image", "GPT Image", "openai/gpt-image", "2026-07-06", 4, 24),
		pm("openai", "gpt-ignored", "Ignored", "openai/gpt-ignored", "2026-07-06", 1, 1),
	}}
	src.ModelsDev[len(src.ModelsDev)-2].TextOutput = false
	c.Ignore = append(c.Ignore, "openai/gpt-ignored")

	got, rep := Apply(c, src)

	wantIDs := []string{
		"anthropic/claude-haiku-3", "anthropic/claude-opus-5-5", "google/gemini-3.6-flash",
		"openai/gpt-6-luna", "openai/gpt-realtime-2.1",
	}
	if diff := cmp.Diff(wantIDs, sortedKeys(got.Models)); diff != "" {
		t.Errorf("model ids (-want +got):\n%s", diff)
	}
	opus := got.Models["anthropic/claude-opus-5-5"]
	wantOpus := Model{
		Name: "Claude Opus 5.5", ReleaseDate: "2026-09-22",
		Providers: map[string]Offer{
			"anthropic":      {ID: "claude-opus-5-5", In: 4, Out: 20},
			"github-copilot": {ID: "claude-opus-5.5", In: 4, Out: 20},
		},
	}
	if diff := cmp.Diff(wantOpus, opus); diff != "" {
		t.Errorf("opus (-want +got):\n%s", diff)
	}
	if got.Models["google/gemini-3.6-flash"].Name != "Gemini 3.6 Flash" {
		t.Errorf("copilot-only model missing its name: %+v", got.Models["google/gemini-3.6-flash"])
	}
	if got.SyncedAt != "2026-11-06" || got.Schema != Schema || got.Since != c.Since {
		t.Errorf("header = %q %d %q", got.SyncedAt, got.Schema, got.Since)
	}
	wantUntiered := []string{"google/gemini-3.6-flash", "openai/gpt-6-luna"}
	if diff := cmp.Diff(wantUntiered, rep.Untiered); diff != "" {
		t.Errorf("untiered (-want +got):\n%s", diff)
	}
}

func TestApplyFirstPartyNameWins(t *testing.T) {
	t.Parallel()
	src := Sources{ModelsDev: []ProviderModel{
		pm("anthropic", "claude-opus-5-5", "First Party", "anthropic/claude-opus-5-5", "2026-09-22", 4, 20),
		pm("github-copilot", "claude-opus-5.5", "Copilot Name", "anthropic/claude-opus-5-5", "2026-09-30", 4, 20),
		pm("github-copilot", "gpt-x", "Copilot GPT", "openai/gpt-x", "2026-02-02", 1, 2),
	}}
	got, _ := Apply(baseCatalog(), src)
	if m := got.Models["anthropic/claude-opus-5-5"]; m.Name != "First Party" || m.ReleaseDate != "2026-09-22" {
		t.Errorf("opus = %q %q", m.Name, m.ReleaseDate)
	}
	if m := got.Models["openai/gpt-x"]; m.Name != "Copilot GPT" || m.ReleaseDate != "2026-02-02" {
		t.Errorf("gpt-x = %q %q", m.Name, m.ReleaseDate)
	}
}

func TestApplyStaleAndMissing(t *testing.T) {
	t.Parallel()
	c := baseCatalog()
	c.Tiers["versatile"] = []string{"openai/gone", "openai/never-seen"}
	c.Models["openai/gone"] = Model{Name: "Gone", Providers: map[string]Offer{"openai": {ID: "gone", In: 1, Out: 2}}}
	got, rep := Apply(c, Sources{Today: "2026-11-06"})
	if diff := cmp.Diff([]string{"anthropic/claude-opus-5-5", "openai/never-seen"}, rep.Missing); diff != "" {
		t.Errorf("missing (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"openai/gone"}, rep.Stale); diff != "" {
		t.Errorf("stale (-want +got):\n%s", diff)
	}
	if got.Models["openai/gone"].Name != "Gone" {
		t.Errorf("stale entry not kept: %+v", got.Models)
	}
	if _, ok := got.Models["openai/never-seen"]; ok {
		t.Error("missing id must not appear in models")
	}
}

func arenaSrc(rows ...ArenaRow) Sources {
	return Sources{
		Today: "2026-11-06",
		ModelsDev: []ProviderModel{
			pm("anthropic", "claude-opus-5-5", "Claude Opus 5.5", "anthropic/claude-opus-5-5", "2026-09-22", 4, 20),
			pm("openai", "gpt-6-1-sol", "GPT-6.1 Sol", "openai/gpt-6.1-sol", "2026-09-29", 5, 30),
		},
		Arena: rows,
	}
}

func TestApplyArena(t *testing.T) {
	t.Parallel()
	text := func(name, cat string, rank int, rating float64, date string) ArenaRow {
		return ArenaRow{
			Board: ArenaBoard, Category: cat, ModelName: name, Org: "anthropic",
			Rank: rank, Rating: rating, Votes: 100, PublishDate: date,
		}
	}
	src := arenaSrc(
		text("Claude Opus 5.5 (Max)", "overall", 5, 1490, "2026-11-03"),
		text("Claude Opus 5.5 (High)", "overall", 3, 1508, "2026-11-04"),
		text("Claude Opus 5.5", "coding", 7, 1400, "2026-11-04"),
		text("Claude Opus 5.5 (High)", "hard_prompts", 2, 1450, "2026-11-04"),
		text("Claude Opus 5.5 (High)", "languages", 1, 1, "2026-11-04"),
		ArenaRow{Board: "webdev", Category: "overall", ModelName: "Claude Opus 5.5 (Max)", Org: "anthropic", Rank: 4, Rating: 1300, PublishDate: "2026-12-01"},
		ArenaRow{Board: "agent", Category: "overall", ModelName: "Claude Opus 5.5 (Max)", Org: "anthropic", Rank: 1, Rating: 0.2, Votes: 14169, PublishDate: "2026-11-04"},
	)
	c := baseCatalog()
	c.Sources.Arena.PublishDate = "old"
	got, rep := Apply(c, src)
	want := map[string]Rank{
		"text":        {Rank: 3, Rating: 1508, Votes: 100, Variant: "High"},
		"coding":      {Rank: 7, Rating: 1400, Votes: 100},
		"hardPrompts": {Rank: 2, Rating: 1450, Votes: 100, Variant: "High"},
		"webdev":      {Rank: 4, Rating: 1300, Variant: "Max"},
		"agent":       {Rank: 1, Rating: 0.2, Votes: 14169, Variant: "Max"},
	}
	if diff := cmp.Diff(want, got.Models["anthropic/claude-opus-5-5"].Arena); diff != "" {
		t.Errorf("arena (-want +got):\n%s", diff)
	}
	if got.Sources.Arena != (ArenaInfo{Board: ArenaBoard, PublishDate: "2026-11-04"}) {
		t.Errorf("sources = %+v", got.Sources.Arena)
	}
	if len(rep.Unmatched) != 0 {
		t.Errorf("unmatched = %v", rep.Unmatched)
	}
}

func TestApplyArenaPublishDateKeptWithoutTextRows(t *testing.T) {
	t.Parallel()
	c := baseCatalog()
	c.Sources.Arena.PublishDate = "2026-10-01"
	got, _ := Apply(c, arenaSrc())
	if got.Sources.Arena != (ArenaInfo{Board: ArenaBoard, PublishDate: "2026-10-01"}) {
		t.Errorf("sources = %+v", got.Sources.Arena)
	}
}

func TestApplyArenaAliasWinsOverKey(t *testing.T) {
	t.Parallel()
	c := baseCatalog()
	c.Aliases.Arena = map[string]string{"Claude Opus 5.5 (High)": "openai/gpt-6.1-sol"}
	src := arenaSrc(ArenaRow{
		Board: ArenaBoard, Category: "overall", ModelName: "Claude Opus 5.5 (High)", Org: "anthropic",
		Rank: 1, Rating: 1500, PublishDate: "2026-11-04",
	})
	got, _ := Apply(c, src)
	if len(got.Models["anthropic/claude-opus-5-5"].Arena) != 0 {
		t.Errorf("key match must lose to the alias: %+v", got.Models["anthropic/claude-opus-5-5"].Arena)
	}
	if r := got.Models["openai/gpt-6.1-sol"].Arena["text"]; r.Rank != 1 || r.Variant != "High" {
		t.Errorf("alias target rank = %+v", r)
	}
}

func TestApplyArenaMatchesByNameAndOfferID(t *testing.T) {
	t.Parallel()
	src := arenaSrc(
		ArenaRow{Board: ArenaBoard, Category: "overall", ModelName: "GPT-6.1 Sol (Max)", Org: "openai", Rank: 2, Rating: 1, PublishDate: "d"},
		ArenaRow{Board: ArenaBoard, Category: "math", ModelName: "gpt-6-1-sol", Org: "openai", Rank: 4, Rating: 1, PublishDate: "d"},
	)
	got, _ := Apply(baseCatalog(), src)
	arena := got.Models["openai/gpt-6.1-sol"].Arena
	if arena["text"].Rank != 2 || arena["math"].Rank != 4 {
		t.Errorf("arena = %+v", arena)
	}
}

func TestApplyUnmatchedLimitedToTrackedOrgs(t *testing.T) {
	t.Parallel()
	row := func(name, org string) ArenaRow {
		return ArenaRow{Board: ArenaBoard, Category: "overall", ModelName: name, Org: org, Rank: 1, PublishDate: "d"}
	}
	src := arenaSrc(
		row("Claude Mystery (High)", "Anthropic"),
		row("Claude Mystery (Max)", "anthropic"),
		row("Other Lab Model", "otherlab"),
		ArenaRow{Board: ArenaBoard, Category: "languages", ModelName: "Claude Skipped", Org: "anthropic"},
	)
	src.Epoch = []EpochRow{
		{Benchmark: "eci", Model: "Claude Mystery", Score: 1},
		{Benchmark: "terminalBench", Model: "claude-mystery_max", Score: 1},
		{Benchmark: "eci", Model: "Qwen 9", Score: 1},
		{Benchmark: "eci", Model: "GPT-9 Zeta", Score: 1},
	}
	_, rep := Apply(baseCatalog(), src)
	want := []string{"arena: Claude Mystery", "epoch: claude-mystery", "epoch: gpt-9-zeta"}
	if diff := cmp.Diff(want, rep.Unmatched); diff != "" {
		t.Errorf("unmatched (-want +got):\n%s", diff)
	}
}

func TestApplyExternal(t *testing.T) {
	t.Parallel()
	old := External{ECI: 100, OSWorld: 0.1}
	c := baseCatalog()
	c.Models["anthropic/claude-opus-5-5"] = Model{Name: "Claude Opus 5.5", External: old}
	c.Aliases.Epoch = map[string]string{"weird-name": "openai/gpt-6.1-sol"}

	src := arenaSrc()
	got, _ := Apply(c, src)
	if got.Models["anthropic/claude-opus-5-5"].External != old {
		t.Errorf("nil epoch must keep external, got %+v", got.Models["anthropic/claude-opus-5-5"].External)
	}

	src.Epoch = []EpochRow{
		{Benchmark: "eci", Model: "Claude Opus 5.5", Score: 168},
		{Benchmark: "osworld", Model: "claude-opus-5-5_max", Score: 0.31},
		{Benchmark: "osworld", Model: "claude-opus-5-5_high", Score: 0.29},
		{Benchmark: "terminalBench", Model: "weird-name", Score: 0.8},
	}
	got, _ = Apply(c, src)
	want := External{ECI: 168, OSWorld: 0.31}
	if diff := cmp.Diff(want, got.Models["anthropic/claude-opus-5-5"].External); diff != "" {
		t.Errorf("opus external (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(External{TerminalBench: 0.8}, got.Models["openai/gpt-6.1-sol"].External); diff != "" {
		t.Errorf("alias external (-want +got):\n%s", diff)
	}
}

func TestApplyDoesNotMutateInput(t *testing.T) {
	t.Parallel()
	c := baseCatalog()
	c.Models["anthropic/claude-opus-5-5"] = Model{
		Name:      "Old",
		Providers: map[string]Offer{"anthropic": {ID: "old"}},
		Arena:     map[string]Rank{"text": {Rank: 9}},
		External:  External{ECI: 1},
	}
	c.Models["openai/gone"] = Model{Name: "Gone", Arena: map[string]Rank{"text": {Rank: 8}}}
	c.Tiers["versatile"] = []string{"openai/gone"}
	want := Catalog{
		Schema: c.Schema, Since: c.Since,
		Tiers:  map[string][]string{"powerful": {"anthropic/claude-opus-5-5"}, "versatile": {"openai/gone"}},
		Ignore: []string{"openai/gpt-realtime-2.1"},
		Models: map[string]Model{
			"anthropic/claude-opus-5-5": {
				Name:      "Old",
				Providers: map[string]Offer{"anthropic": {ID: "old"}},
				Arena:     map[string]Rank{"text": {Rank: 9}},
				External:  External{ECI: 1},
			},
			"openai/gone": {Name: "Gone", Arena: map[string]Rank{"text": {Rank: 8}}},
		},
	}
	src := arenaSrc(ArenaRow{Board: ArenaBoard, Category: "overall", ModelName: "Claude Opus 5.5", Org: "anthropic", Rank: 1, PublishDate: "d"})
	src.Epoch = []EpochRow{{Benchmark: "eci", Model: "Claude Opus 5.5", Score: 5}}
	Apply(c, src)
	if diff := cmp.Diff(want, c); diff != "" {
		t.Errorf("input changed (-want +got):\n%s", diff)
	}
}

func TestApplyKeepsManualKeys(t *testing.T) {
	t.Parallel()
	c := baseCatalog()
	c.Aliases = Aliases{Arena: map[string]string{"a": "b"}, Epoch: map[string]string{"c": "d"}}
	got, _ := Apply(c, arenaSrc())
	if diff := cmp.Diff(c.Tiers, got.Tiers); diff != "" {
		t.Errorf("tiers: %s", diff)
	}
	if diff := cmp.Diff(c.Ignore, got.Ignore); diff != "" {
		t.Errorf("ignore: %s", diff)
	}
	if diff := cmp.Diff(c.Aliases, got.Aliases); diff != "" {
		t.Errorf("aliases: %s", diff)
	}
	if got.Since != c.Since {
		t.Errorf("since = %q", got.Since)
	}
}
