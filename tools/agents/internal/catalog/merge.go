package catalog

import (
	"cmp"
	"maps"
	"regexp"
	"slices"
	"strings"
)

const (
	boardWebdev = "webdev"
	boardAgent  = "agent"

	catOverall = "overall"
	catCoding  = "coding"
	catMath    = "math"

	benchECI        = "eci"
	colModelVersion = "Model version"
)

var variantGroup = regexp.MustCompile(`\s*\(([^()]*)\)\s*$`)

// arenaKey maps (board, dataset category) to the Model.Arena key.
func arenaKey(board, category string) (string, bool) {
	switch board {
	case ArenaBoard:
		k, ok := map[string]string{
			catOverall:              "text",
			catCoding:               catCoding,
			"hard_prompts":          CatHardPrompts,
			catMath:                 catMath,
			"instruction_following": CatInstructionFollowing,
			"multi_turn":            CatMultiTurn,
		}[category]
		return k, ok
	case boardWebdev:
		return boardWebdev, category == catOverall
	case boardAgent:
		return boardAgent, category == catOverall
	}
	return "", false
}

// Apply merges freshly fetched sources into c, keeping its manual keys. It
// returns a new catalog and never mutates c's maps.
func Apply(c Catalog, src Sources) (Catalog, Report) {
	var rep Report
	out := c
	tiered := map[string]bool{}
	for _, ids := range c.Tiers {
		for _, id := range ids {
			tiered[id] = true
		}
	}

	fresh := freshModels(c, src.ModelsDev, tiered)
	models := maps.Clone(fresh)
	for _, tier := range sortedKeys(c.Tiers) {
		for _, id := range c.Tiers[tier] {
			if _, ok := fresh[id]; ok {
				continue
			}
			if old, ok := c.Models[id]; ok {
				models[id] = old
				rep.Stale = append(rep.Stale, id)
			} else {
				rep.Missing = append(rep.Missing, id)
			}
		}
	}

	idx := newIndex(models)
	for id, m := range models {
		m.Arena = nil
		if src.Epoch != nil {
			m.External = External{}
		} else {
			m.External = c.Models[id].External
		}
		models[id] = m
	}

	out.SyncedAt = src.Today
	out.Schema = Schema
	out.Sources.Arena = ArenaInfo{Board: ArenaBoard, PublishDate: arenaPublishDate(c, src.Arena)}
	rep.Unmatched = append(rep.Unmatched, applyArena(c, idx, models, fresh, src.Arena)...)
	if src.Epoch != nil {
		applyEpoch(c, idx, models, src.Epoch)
	}
	out.Models = models

	rep.Untiered = out.Untiered()
	rep.Unmatched = append(rep.Unmatched, danglingAliases(c.Aliases, models)...)
	rep.Unmatched = sortedUnique(rep.Unmatched)
	rep.Stale = sortedUnique(rep.Stale)
	rep.Missing = sortedUnique(rep.Missing)
	return out, rep
}

// freshModels builds the models of the kept canonical ids from models.dev.
func freshModels(c Catalog, entries []ProviderModel, tiered map[string]bool) map[string]Model {
	groups := map[string][]ProviderModel{}
	for _, e := range entries {
		groups[e.CanonicalID] = append(groups[e.CanonicalID], e)
	}
	fresh := map[string]Model{}
	for id, offers := range groups {
		if slices.Contains(c.Ignore, id) && !tiered[id] {
			continue
		}
		if tiered[id] || keep(c.Since, offers) {
			fresh[id] = buildModel(id, offers)
		}
	}
	return fresh
}

func keep(since string, offers []ProviderModel) bool {
	for _, o := range offers {
		if o.Provider == ProviderCopilot {
			return true
		}
		if (o.Provider == ProviderAnthropic || o.Provider == ProviderOpenAI) &&
			o.ReleaseDate >= since && o.TextOutput && o.ToolCall {
			return true
		}
	}
	return false
}

// buildModel takes name and release date from the first-party offer, else
// from the Copilot one. The first offer (by id) of a provider wins.
func buildModel(id string, offers []ProviderModel) Model {
	m := Model{Providers: map[string]Offer{}}
	org, _, _ := strings.Cut(id, "/")
	var first, copilot *ProviderModel
	for i, o := range offers {
		if _, dup := m.Providers[o.Provider]; !dup {
			m.Providers[o.Provider] = Offer{ID: o.ID, In: o.In, Out: o.Out}
		}
		switch o.Provider {
		case org:
			if first == nil {
				first = &offers[i]
			}
		case ProviderCopilot:
			if copilot == nil {
				copilot = &offers[i]
			}
		}
	}
	src := first
	if src == nil {
		src = copilot
	}
	if src == nil {
		src = &offers[0]
	}
	m.Name, m.ReleaseDate = src.Name, src.ReleaseDate
	return m
}

// index maps a Key to the canonical ids of the models it identifies.
type index map[string][]string

func newIndex(models map[string]Model) index {
	idx := index{}
	add := func(key, id string) {
		if key != "" && !slices.Contains(idx[key], id) {
			idx[key] = append(idx[key], id)
		}
	}
	for _, id := range sortedKeys(models) {
		m := models[id]
		_, part, _ := strings.Cut(id, "/")
		add(Key(part), id)
		add(Key(m.Name), id)
		for _, o := range m.Providers {
			add(Key(o.ID), id)
		}
	}
	return idx
}

// match resolves a source name to canonical ids: the alias if there is one,
// else Key matching.
func (idx index) match(models map[string]Model, aliases map[string]string, name string) []string {
	if target, ok := aliases[name]; ok {
		if _, ok := models[target]; ok {
			return []string{target}
		}
		return nil
	}
	return idx[Key(name)]
}

func arenaPublishDate(c Catalog, rows []ArenaRow) string {
	date := ""
	for _, r := range rows {
		if r.Board == ArenaBoard {
			date = max(date, r.PublishDate)
		}
	}
	if date == "" {
		return c.Sources.Arena.PublishDate
	}
	return date
}

func splitVariant(name string) (base, variant string) {
	m := variantGroup.FindStringSubmatchIndex(name)
	if m == nil {
		return strings.TrimSpace(name), ""
	}
	return strings.TrimSpace(name[:m[0]]), strings.TrimSpace(name[m[2]:m[3]])
}

// better reports whether a beats b: lower rank, then higher rating, then name.
func better(a, b ArenaRow) bool {
	return cmp.Or(
		cmp.Compare(a.Rank, b.Rank),
		-cmp.Compare(a.Rating, b.Rating),
		strings.Compare(a.ModelName, b.ModelName),
	) < 0
}

// unmatchedTopRank limits the unmatched report to arena rows ranked this high or
// better in their category: the rest are legacy models nobody needs to alias.
const unmatchedTopRank = 20

// reportUnmatched reports whether an unmatched arena row is worth listing: a
// tracked org's model ranked in the top unmatchedTopRank.
func reportUnmatched(r ArenaRow, orgs map[string]bool) bool {
	return r.Rank <= unmatchedTopRank && orgs[strings.ToLower(r.Org)]
}

// applyArena sets Model.Arena and returns the unmatched report lines.
func applyArena(c Catalog, idx index, models, fresh map[string]Model, rows []ArenaRow) []string {
	best := map[string]map[string]ArenaRow{}
	orgs := orgSet(fresh)
	var unmatched []string
	for _, r := range rows {
		key, ok := arenaKey(r.Board, r.Category)
		if !ok {
			continue
		}
		ids := idx.match(models, c.Aliases.Arena, r.ModelName)
		if len(ids) == 0 {
			if reportUnmatched(r, orgs) {
				base, _ := splitVariant(r.ModelName)
				unmatched = append(unmatched, "arena: "+base)
			}
			continue
		}
		for _, id := range ids {
			if best[id] == nil {
				best[id] = map[string]ArenaRow{}
			}
			if cur, ok := best[id][key]; !ok || better(r, cur) {
				best[id][key] = r
			}
		}
	}
	for id, byKey := range best {
		m := models[id]
		m.Arena = map[string]Rank{}
		for key, r := range byKey {
			_, variant := splitVariant(r.ModelName)
			m.Arena[key] = Rank{Rank: r.Rank, Rating: r.Rating, Votes: r.Votes, Variant: variant}
		}
		models[id] = m
	}
	return unmatched
}

// applyEpoch rebuilds Model.External. Unmatched Epoch rows are not reported:
// Epoch's history is mostly legacy models, and aliases.epoch covers a gap.
func applyEpoch(c Catalog, idx index, models map[string]Model, rows []EpochRow) {
	for _, r := range rows {
		ids := idx.match(models, c.Aliases.Epoch, r.Model)
		for _, id := range ids {
			m := models[id]
			setExternal(&m.External, r)
			models[id] = m
		}
	}
}

func setExternal(e *External, r EpochRow) {
	var field *float64
	switch r.Benchmark {
	case benchECI:
		field = &e.ECI
	case "terminalBench":
		field = &e.TerminalBench
	case "sweBenchVerified":
		field = &e.SWEBenchVerified
	case "metrHorizonMin":
		field = &e.METRHorizonMin
	case "osworld":
		field = &e.OSWorld
	default:
		return
	}
	*field = max(*field, r.Score)
}

func orgSet(models map[string]Model) map[string]bool {
	orgs := map[string]bool{}
	for id := range models {
		org, _, _ := strings.Cut(id, "/")
		orgs[org] = true
	}
	return orgs
}

func sortedUnique(s []string) []string {
	slices.Sort(s)
	return slices.Compact(s)
}

// danglingAliases reports aliases whose target is not in models, so a typo in
// aliases.* does not go unnoticed.
func danglingAliases(a Aliases, models map[string]Model) []string {
	var out []string
	for source, m := range map[string]map[string]string{"arena": a.Arena, "epoch": a.Epoch} {
		for name, target := range m {
			if _, ok := models[target]; !ok {
				out = append(out, "alias "+source+": "+name+" -> "+target+" (not in models)")
			}
		}
	}
	return out
}
