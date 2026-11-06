// Package catalog is the model catalog of chezmoi-agents: manual model tiers
// plus model data synced from models.dev (costs), the LMArena leaderboard
// dataset (ranks) and Epoch AI (external benchmarks). It is pure: callers
// fetch the raw bytes and write the results.
package catalog

// Schema is the catalog format version.
const Schema = 1

// ArenaBoard is the LMArena text board ranks come from.
const ArenaBoard = "text_style_control"

// Tiers are the manual tier names, in display order.
func Tiers() []string { return []string{"lightweight", "versatile", "powerful", "extra"} }

// Providers are the models.dev provider keys the catalog tracks.
func Providers() []string { return []string{"anthropic", "openai", "github-copilot"} }

// Categories are the arena rank keys of Model.Arena, in display order.
func Categories() []string {
	return []string{
		"text", "coding", "agent", "webdev",
		"hardPrompts", "math", "instructionFollowing", "multiTurn",
	}
}

// File is the top level of models.json: one key, so the chezmoi template data
// namespace stays clean.
type File struct {
	Catalog Catalog `json:"modelCatalog"`
}

// Catalog is the model catalog. Tiers, Ignore, Aliases and Since are manual;
// sync rewrites Models, Sources and SyncedAt.
type Catalog struct {
	Schema   int                 `json:"schema"`
	SyncedAt string              `json:"syncedAt,omitzero"`
	Since    string              `json:"since"`
	Sources  SourceInfo          `json:"sources"`
	Tiers    map[string][]string `json:"tiers"`
	Ignore   []string            `json:"ignore,omitzero"`
	Aliases  Aliases             `json:"aliases,omitzero"`
	Models   map[string]Model    `json:"models"`
}

// SourceInfo records where the synced data came from.
type SourceInfo struct {
	Arena ArenaInfo `json:"arena"`
}

// ArenaInfo is the arena board and its publish date at sync time.
type ArenaInfo struct {
	PublishDate string `json:"publishDate,omitzero"`
	Board       string `json:"board"`
}

// Aliases map a source's model name to a canonical id; they win over Key matching.
type Aliases struct {
	Arena map[string]string `json:"arena,omitzero"`
	Epoch map[string]string `json:"epoch,omitzero"`
}

// Model is one model, keyed in Catalog.Models by its models.dev canonical id
// (`<org>/<id>`, e.g. `anthropic/claude-opus-5-5`).
type Model struct {
	Name        string           `json:"name"`
	ReleaseDate string           `json:"releaseDate,omitzero"`
	Providers   map[string]Offer `json:"providers,omitzero"`
	Arena       map[string]Rank  `json:"arena,omitzero"`
	External    External         `json:"external,omitzero"`
}

// Offer is one provider serving a model: its id there and $/Mtok cost.
type Offer struct {
	ID  string  `json:"id"`
	In  float64 `json:"in"`
	Out float64 `json:"out"`
}

// Rank is a model's place in one arena category; Variant is the reasoning
// effort of its best-ranked arena entry (e.g. "High"), empty when there is none.
type Rank struct {
	Rank    int     `json:"rank"`
	Rating  float64 `json:"rating"`
	Votes   int     `json:"votes,omitzero"`
	Variant string  `json:"variant,omitzero"`
}

// External holds Epoch AI benchmark scores; zero means no data.
type External struct {
	ECI              float64 `json:"eci,omitzero"`
	TerminalBench    float64 `json:"terminalBench,omitzero"`
	SWEBenchVerified float64 `json:"sweBenchVerified,omitzero"`
	METRHorizonMin   float64 `json:"metrHorizonMin,omitzero"`
	OSWorld          float64 `json:"osworld,omitzero"`
}

// ProviderModel is one models.dev entry of a tracked provider.
type ProviderModel struct {
	Provider    string // models.dev provider key
	ID          string // id at that provider
	Name        string
	CanonicalID string // canonical_model_id, or Provider/ID when absent
	ReleaseDate string
	In, Out     float64
	ToolCall    bool
	TextOutput  bool
}

// ArenaRow is one row of the LMArena leaderboard dataset.
type ArenaRow struct {
	Board       string // dataset config: text_style_control, webdev, agent
	Category    string // dataset category, e.g. overall, coding, hard_prompts
	ModelName   string // display name, e.g. "Claude Opus 5.5 (High)"
	Org         string
	Rank        int
	Rating      float64 // rating, or score on the agent board
	Votes       int     // vote_count, or session_count on the agent board
	PublishDate string
}

// EpochRow is one benchmark score from Epoch AI's benchmark_data.zip.
type EpochRow struct {
	Benchmark string // eci, terminalBench, sweBenchVerified, metrHorizonMin, osworld
	Model     string // `Model` (ECI) or `Model version` column
	Score     float64
}

// Sources is everything one sync fetched. Epoch is nil when external data was
// not requested; Apply then keeps the existing External values.
type Sources struct {
	ModelsDev []ProviderModel
	Arena     []ArenaRow
	Epoch     []EpochRow
	Today     string // YYYY-MM-DD, becomes SyncedAt
}

// Report is what a sync found that needs a human.
type Report struct {
	Unmatched []string // arena/epoch names of tracked orgs that matched no model
	Untiered  []string // canonical ids in Models with no tier and not ignored
	Stale     []string // tiered ids no provider lists any more (kept)
	Missing   []string // tiered ids with no data at all (Validate fails on them)
}
