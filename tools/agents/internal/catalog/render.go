package catalog

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/term"
)

const (
	rankingsNote   = "Rankings: [LMArena leaderboard dataset](https://huggingface.co/datasets/lmarena-ai/leaderboard-dataset)."
	benchmarksNote = " Benchmarks: [Epoch AI](https://epoch.ai/data), CC-BY."
	topRank        = 10
	midRank        = 50
	noneMark       = "_None._"
	sep            = " · "
	mdModel        = "Model"
	termModel      = "MODEL"
)

func rankHeads() []string {
	return []string{"Text", "Code", "Agent", "WebDev", "Hard", "Math", "IF", "Multi"}
}

func rankTerm() []string {
	return []string{"TEXT", "CODE", "AGENT", "WEB", "HARD", "MATH", "IF", "MULTI"}
}

func categoryLabel(cat string) string {
	switch cat {
	case CatHardPrompts:
		return "Hard prompts"
	case CatInstructionFollowing:
		return "Instruction following"
	case CatMultiTurn:
		return "Multi-turn"
	case "webdev":
		return "WebDev"
	default:
		return title(cat)
	}
}

// extColumn is one external benchmark column, shown only when some model has data.
type extColumn struct {
	head string
	val  func(External) float64
	fmt  func(float64) string
}

func percent(v float64) string {
	if v <= 1 {
		v *= 100
	}
	return fmt.Sprintf("%.0f%%", v)
}

func externalColumns(c Catalog) []extColumn {
	all := []extColumn{
		{"ECI", func(e External) float64 { return e.ECI }, func(v float64) string { return fmt.Sprintf("%.0f", v) }},
		{"TB", func(e External) float64 { return e.TerminalBench }, percent},
		{"SWE-V", func(e External) float64 { return e.SWEBenchVerified }, percent},
		{"METR", func(e External) float64 { return e.METRHorizonMin }, func(v float64) string { return fmt.Sprintf("%.0fm", v) }},
		{"OSWorld", func(e External) float64 { return e.OSWorld }, percent},
	}
	var out []extColumn
	for _, col := range all {
		for _, m := range c.Models {
			if col.val(m.External) != 0 {
				out = append(out, col)
				break
			}
		}
	}
	return out
}

func cost(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// offer is one labelled provider price.
type offer struct{ label, price string }

// offers lists the first-party offers (API) first, then Copilot.
func offers(m Model) []offer {
	var api, cop []offer
	for _, p := range sortedKeys(m.Providers) {
		o := m.Providers[p]
		price := "$" + cost(o.In) + "/$" + cost(o.Out)
		switch p {
		case ProviderAnthropic, ProviderOpenAI:
			api = append(api, offer{"API", price})
		case ProviderCopilot:
			cop = append(cop, offer{"Copilot", price})
		}
	}
	return append(api, cop...)
}

func providersPlain(m Model) string {
	os := offers(m)
	parts := make([]string, 0, len(os))
	for _, o := range os {
		parts = append(parts, o.label+" "+o.price)
	}
	return strings.Join(parts, sep)
}

func providersStyled(st term.Style, m Model) string {
	os := offers(m)
	parts := make([]string, 0, len(os))
	for _, o := range os {
		parts = append(parts, st.D+o.label+st.O+" "+o.price)
	}
	return strings.Join(parts, st.D+sep+st.O)
}

func escape(s string) string { return strings.ReplaceAll(s, "|", `\|`) }

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// lookup returns the model of id and its display name: the id itself when the
// model is missing from Models.
func lookup(c Catalog, id string) (Model, string) {
	m, ok := c.Models[id]
	if !ok || m.Name == "" {
		return m, id
	}
	return m, m.Name
}

// best is the top-ranked model in one category.
type best struct {
	label   string
	name    string
	variant string
	rank    int
}

func bestWhen(c Catalog) []best {
	var out []best
	for _, cat := range Categories() {
		id, rank := "", 0
		for _, k := range sortedKeys(c.Models) {
			r, ok := c.Models[k].Arena[cat]
			if !ok || r.Rank <= 0 {
				continue
			}
			if id == "" || r.Rank < rank {
				id, rank = k, r.Rank
			}
		}
		if id == "" {
			continue
		}
		m := c.Models[id]
		out = append(out, best{categoryLabel(cat), m.Name, m.Arena[cat].Variant, rank})
	}
	return out
}

func rankOf(m Model, cat string) (int, bool) {
	r, ok := m.Arena[cat]
	return r.Rank, ok && r.Rank > 0
}

func externalCell(m Model, e extColumn) string {
	if v := e.val(m.External); v != 0 {
		return e.fmt(v)
	}
	return ""
}

func mdTable(b *strings.Builder, header []string, rows [][]string) {
	b.WriteString("| " + strings.Join(header, " | ") + " |\n")
	b.WriteString("|" + strings.Repeat("---|", len(header)) + "\n")
	for _, r := range rows {
		cells := make([]string, len(r))
		for i, cell := range r {
			cells[i] = escape(cell)
		}
		b.WriteString("| " + strings.Join(cells, " | ") + " |\n")
	}
}

func mdTier(c Catalog, tier string, header []string, ext []extColumn) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## %s\n\n", title(tier))
	ids := c.Tiers[tier]
	if len(ids) == 0 {
		return b.String() + noneMark + "\n\n"
	}
	rows := make([][]string, 0, len(ids))
	for _, id := range ids {
		m, name := lookup(c, id)
		row := make([]string, 0, len(header))
		row = append(row, name, providersPlain(m))
		for _, cat := range Categories() {
			cell := ""
			if r, ok := rankOf(m, cat); ok {
				cell = "#" + strconv.Itoa(r)
			}
			row = append(row, cell)
		}
		for _, e := range ext {
			row = append(row, externalCell(m, e))
		}
		rows = append(rows, row)
	}
	mdTable(&b, header, rows)
	return b.String() + "\n"
}

func mdUntiered(c Catalog) string {
	un := c.Untiered()
	if len(un) == 0 {
		return "## Untiered\n\n" + noneMark + "\n\n"
	}
	rows := make([][]string, 0, len(un))
	for _, id := range un {
		m, name := lookup(c, id)
		rows = append(rows, []string{name, id, providersPlain(m)})
	}
	var b strings.Builder
	b.WriteString("## Untiered\n\n")
	mdTable(&b, []string{mdModel, "Id", "Providers"}, rows)
	return b.String() + "\n"
}

func mdBest(c Catalog) string {
	bw := bestWhen(c)
	if len(bw) == 0 {
		return "## Best when\n\n" + noneMark + "\n\n"
	}
	rows := make([][]string, 0, len(bw))
	for _, x := range bw {
		name := x.name
		if x.variant != "" {
			name += " (" + x.variant + ")"
		}
		rows = append(rows, []string{x.label, name, "#" + strconv.Itoa(x.rank)})
	}
	var b strings.Builder
	b.WriteString("## Best when\n\n")
	mdTable(&b, []string{"Category", mdModel, "Rank"}, rows)
	return b.String() + "\n"
}

// Markdown renders docs/MODELS.md.
func Markdown(c Catalog) []byte {
	ext := externalColumns(c)
	var b strings.Builder
	b.WriteString("# Models\n\n")
	b.WriteString("<!-- Generated by chezmoi-agents (make models-sync / chezmoi agents models render). Do not edit; edit home/.chezmoidata/models.json instead. -->\n\n")
	fmt.Fprintf(&b, "Synced %s · arena board `%s` published %s. Tiers are manual: `dots models`, then `chezmoi agents models render` ([runbook](OS.md#model-tiers-chezmoi-agents-models)).\n\n",
		c.SyncedAt, c.Sources.Arena.Board, c.Sources.Arena.PublishDate)

	header := append([]string{mdModel, "Providers ($in/$out per Mtok)"}, rankHeads()...)
	for _, e := range ext {
		header = append(header, e.head)
	}
	for _, tier := range Tiers() {
		b.WriteString(mdTier(c, tier, header, ext))
	}
	b.WriteString(mdUntiered(c))
	b.WriteString(mdBest(c))
	b.WriteString(rankingsNote)
	if len(ext) > 0 {
		b.WriteString(benchmarksNote)
	}
	b.WriteString("\n")
	return []byte(b.String())
}

func termRank(st term.Style, rank int, ok bool) string {
	switch {
	case !ok:
		return st.D + "·" + st.O
	case rank <= topRank:
		return st.G + "#" + strconv.Itoa(rank) + st.O
	case rank <= midRank:
		return "#" + strconv.Itoa(rank)
	default:
		return st.D + "#" + strconv.Itoa(rank) + st.O
	}
}

func termTier(w io.Writer, st term.Style, c Catalog, tier string, header []string, ext []extColumn) {
	ids := c.Tiers[tier]
	fmt.Fprintf(w, "\n%s%s%s %s(%d)%s\n", st.B, title(tier), st.O, st.D, len(ids), st.O)
	if len(ids) == 0 {
		fmt.Fprintf(w, "  %snone%s\n", st.D, st.O)
		return
	}
	rows := make([][]string, 0, len(ids))
	for _, id := range ids {
		m, name := lookup(c, id)
		row := make([]string, 0, len(header))
		row = append(row, st.M+name+st.O, providersStyled(st, m))
		for _, cat := range Categories() {
			r, ok := rankOf(m, cat)
			row = append(row, termRank(st, r, ok))
		}
		for _, e := range ext {
			row = append(row, externalCell(m, e))
		}
		rows = append(rows, row)
	}
	st.Table(w, header, rows)
}

func termUntiered(w io.Writer, st term.Style, c Catalog) {
	un := c.Untiered()
	fmt.Fprintf(w, "\n%sUntiered%s %s(%d)%s\n", st.B, st.O, st.D, len(un), st.O)
	if len(un) == 0 {
		fmt.Fprintf(w, "  %s✓%s every tracked model has a tier\n", st.G, st.O)
		return
	}
	rows := make([][]string, 0, len(un))
	for _, id := range un {
		m, name := lookup(c, id)
		rows = append(rows, []string{st.M + name + st.O, st.D + id + st.O, providersStyled(st, m)})
	}
	st.Table(w, []string{termModel, "ID", "PROVIDERS $/MTOK"}, rows)
}

func termBest(w io.Writer, st term.Style, c Catalog) {
	bw := bestWhen(c)
	if len(bw) == 0 {
		return
	}
	fmt.Fprintf(w, "\n%sBest when%s\n", st.B, st.O)
	rows := make([][]string, 0, len(bw))
	for _, x := range bw {
		name := st.M + x.name + st.O
		if x.variant != "" {
			name += " " + st.D + "(" + x.variant + ")" + st.O
		}
		rows = append(rows, []string{x.label, name, st.G + "#" + strconv.Itoa(x.rank) + st.O})
	}
	st.Table(w, []string{"CATEGORY", termModel, "RANK"}, rows)
}

// WriteTerminal prints the catalog as coloured tables: one per tier, then
// untiered models, then the best model per arena category.
func WriteTerminal(w io.Writer, st term.Style, c Catalog) {
	ext := externalColumns(c)
	fmt.Fprintf(w, "%sModel tiers%s %s· synced %s · arena %s %s%s\n",
		st.B, st.O, st.D, c.SyncedAt, c.Sources.Arena.Board, c.Sources.Arena.PublishDate, st.O)

	header := append([]string{termModel, "PROVIDERS $/MTOK"}, rankTerm()...)
	for _, e := range ext {
		header = append(header, strings.ToUpper(e.head))
	}
	for _, tier := range Tiers() {
		termTier(w, st, c, tier, header, ext)
	}
	termUntiered(w, st, c)
	termBest(w, st, c)

	note := "Rankings: LMArena leaderboard dataset"
	if len(ext) > 0 {
		note += " · Benchmarks: Epoch AI, CC-BY"
	}
	fmt.Fprintf(w, "\n%s%s%s\n", st.D, note, st.O)
}
