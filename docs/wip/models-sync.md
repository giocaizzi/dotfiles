# WIP: `chezmoi agents models` (model tier list sync)

Status: `models`, `models sync`, `models render` in implementation (PR #162); the apply-time advisory (`doctor`) is planned below and waits on the user's answers. Data sources checked on 2026-11-06. Delete this file when the feature ships; its lasting content moves to `docs/OS.md` and `docs/MODELS.md`.

## Goal

A `chezmoi agents models` subcommand family in `tools/agents` (chezmoi-agents) that keeps a model tier list in sync for three providers: Anthropic, OpenAI and GitHub Copilot.

## Decisions made

1. Tiers are manual, data is automatic. Four tiers: Lightweight, Versatile, Powerful, Extra. Sync updates ranks, scores and costs, and flags models from the three providers that have no tier yet.
2. Scope: Anthropic and OpenAI models, plus any model GitHub Copilot serves (Copilot also serves Gemini, Grok, Kimi and MAI models; all of them in the seed table are in Copilot's list).
3. Storage: a structured JSON file `home/.chezmoidata/models.json` (data + manual tiers) and a generated quicklook `docs/MODELS.md`. `chezmoi agents models` prints the table in the terminal; `chezmoi agents models sync` refreshes the data.
4. The quicklook always shows cost: $/Mtok input/output per provider.
5. Rankings ("best when") come from arena.ai official data. Columns: text overall, text coding, agent, webdev overall, text hard prompts, math, instruction following, multi-turn.
6. External benchmarks are optional, behind `-external`: Epoch Capabilities Index (ECI), Terminal-Bench, SWE-bench Verified, METR time horizons, OSWorld.
7. The sync commands are not linked to `harnessPolicy.models.tiers` (scout/worker/driver/escalate). The apply-time advisory (below) reads both, in a later PR.
8. Simplest correct solution: no speculative features.

## Seed tiers

Canonical ids are models.dev `canonical_model_id` (see Matching).

| Tier | Models |
|---|---|
| lightweight | `openai/gpt-6-luna`, `openai/gpt-5.6-luna`, `openai/gpt-5.4-nano`, `anthropic/claude-haiku-5-5`, `anthropic/claude-haiku-4-5`, `google/gemini-3.8-flash`, `google/gemini-3.7-flash`, `microsoft/mai-code-1.1-flash` |
| versatile | `openai/gpt-6.1-terra`, `openai/gpt-5.6-terra`, `openai/gpt-5.3-codex`, `anthropic/claude-sonnet-5-5`, `anthropic/claude-sonnet-5`, `anthropic/claude-sonnet-4-6`, `moonshotai/kimi-k3`, `xai/grok-4.5`, `xai/grok-4.6` |
| powerful | `openai/gpt-6.1-sol`, `openai/gpt-6-sol`, `openai/gpt-5.6-sol`, `anthropic/claude-opus-4-8`, `google/gemini-4-argon`, `xai/grok-4.7` |
| extra | `openai/gpt-6-astra`, `anthropic/claude-fable-5-1`, `anthropic/claude-fable-5` |

Untiered Copilot models at seed time (sync must flag them): `claude-opus-5.5`, `claude-opus-5`, `claude-opus-4.7`, `gpt-5.5`, `gpt-5.4`, `gpt-5.4-mini`, `gpt-5-mini`, `gemini-3.5-flash`, `gemini-3.6-flash`, `kimi-k2.7-code`. Claude Opus 5.5 and Opus 5 are also on arena and missing from the table.

## Data sources (all without auth)

### Costs and model list: models.dev

- URL: `https://models.dev/api.json`
- Provider keys: `anthropic`, `openai`, `github-copilot`; each `.models.<id>` has `name`, `cost.input`, `cost.output` ($/Mtok), `release_date`, `canonical_model_id`, `family`, `modalities.{input,output}`, `tool_call`, `reasoning_options`.
- `canonical_model_id` is `<org>/<id>` (e.g. `anthropic/claude-opus-5-5`, `google/gemini-4-argon`). Copilot ids use dots (`claude-opus-5.5`) but their `canonical_model_id` is the same as the Anthropic entry's: it is the model key.
- Copilot bills by usage (AI Credits) since June 2026, so token prices are the cost metric; legacy multipliers are ignored.
- Avoid openrouter.ai as a source (TLS failures seen on some hosts).

### Rankings: arena.ai via Hugging Face

- arena.ai has no public API and answers 403 to scripts. The official data is the dataset `lmarena-ai/leaderboard-dataset`, updated daily.
- Configs used: `text` or `text_style_control` (see open questions), `webdev`, `agent`. Each has splits `latest` and `full`; only `latest` is needed.
- Columns: `model_name`, `organization`, `license`, `rating`, `rating_lower`, `rating_upper`, `variance`, `vote_count`, `rank`, `category`, `leaderboard_publish_date`.
- Display names carry a reasoning-effort suffix: `Claude Opus 5.5 (High)`, `GPT-6 Astra (Max)`.
- Categories (read from the datasets-server `/statistics` endpoint on 2026-11-06): text boards have `overall`, `coding`, `hard_prompts`, `math`, `instruction_following`, `multi_turn` (plus languages, industries, `hard_prompts_english`, `exclude_ties`, ...); webdev has `overall` (plus `webdev-html`, `webdev-react`, `image_to_webdev`); agent has only `overall`. Mapping: text `overall`→`text`, `coding`→`coding`, `hard_prompts`→`hardPrompts`, `math`→`math`, `instruction_following`→`instructionFollowing`, `multi_turn`→`multiTurn`; webdev `overall`→`webdev`; agent `overall`→`agent`.
- The agent board has different columns: `score` (0..1) instead of `rating`, `session_count` instead of `vote_count`; they fill `rating` and `votes`.
- Sizes (latest): text about 11.1k rows, text_style_control the same, webdev about 650, agent about 58.
- Access options:
  - JSON: `https://datasets-server.huggingface.co/rows?dataset=lmarena-ai/leaderboard-dataset&config=<config>&split=latest&offset=<n>&length=100` (max 100 per page; rows sorted by category, then rank). About 112 pages for text. Worked reliably on 2026-11-06.
  - `https://datasets-server.huggingface.co/filter?...&where="category"='coding'` exists but failed repeatedly on 2026-11-06 ("Authentication check on the Hugging Face Hub failed or timed out").
  - Parquet: `https://huggingface.co/datasets/lmarena-ai/leaderboard-dataset/resolve/main/<config>/latest-00000-of-00001.parquet` (text: 624 KB). Reliable, but needs a Go parquet library.

### External benchmarks: Epoch AI (with `-external`)

- URL: `https://epoch.ai/data/benchmark_data.zip` (about 2.3 MB). License CC-BY: credit "Epoch AI" wherever the data is shown.
- Files and score columns:

| Benchmark | File | Key column | Score column |
|---|---|---|---|
| ECI | `epoch_capabilities_index/eci_scores.csv` | `Model` (display name, e.g. `GPT-6 Astra`) | `eci` |
| Terminal-Bench | `terminalbench_external.csv` | `Model version` | `Accuracy mean` |
| SWE-bench Verified | `swe_bench_verified.csv` | `Model version` | `mean_score` |
| METR time horizon | `metr_time_horizons_external.csv` | `Model version` | `Time horizon` (minutes) |
| OSWorld | `osworld_2_external.csv` only (the current benchmark; mixing it with `os_world_external.csv` would compare two benchmarks) | `Model version` | `Score` |

- `Model version` forms: `claude-sonnet-5-5_max`, `gpt-6.1-sol_max`, `gpt-5-2025-08-07_high`, `claude-sonnet-4-5-20250929`, sometimes empty (skip those rows).
- Coverage of the newest models is thin: Terminal-Bench, SWE-bench and METR stop around GPT-5.5; only `osworld_2` has `claude-opus-5_*` and `gpt-5.6-sol_max`. ECI is current. This is why external data stays optional.
- Reading needs only Go's standard library (`archive/zip`, `encoding/csv`).

## Matching across sources

One function `Key(s)` normalizes any name or id:

1. Lowercase.
2. Drop a trailing `(...)` (arena effort variants, models.dev `(latest)`).
3. Drop a trailing `_<suffix>` (Epoch effort).
4. Drop date suffixes `-YYYY-MM-DD` and `-YYYYMMDD`.
5. Collapse runs of spaces, dots and underscores into `-`.

A source row matches a model when, after the alias map is checked, `Key(row)` equals `Key` of the id part of the model's canonical id, or `Key` of its models.dev `name`. Examples: `Claude Opus 5.5 (High)` and `anthropic/claude-opus-5-5` both give `claude-opus-5-5`; `GPT-6.1 Sol` and `openai/gpt-6.1-sol` both give `gpt-6-1-sol`; `gpt-5-2025-08-07_high` gives `gpt-5`.

Manual aliases in `models.json` (`aliases.arena`, `aliases.epoch`: source name → canonical id) win over the rule. Sync lists rows it could not match, limited to the organizations of the providers' models so other labs do not flood the report.

Arena variants: when a model has several variants (High, Max, ...), the best-ranked one wins per category and its variant is recorded (pending open question 2).

## JSON schema (`home/.chezmoidata/models.json`)

```json
{"modelCatalog": {
  "schema": 1,
  "syncedAt": "2026-11-06",
  "since": "2026-01-01",
  "sources": {
    "arena": {"publishDate": "2026-10-28", "board": "text_style_control"},
    "modelsDev": {},
    "epoch": {}
  },
  "tiers": {
    "lightweight": ["openai/gpt-6-luna"],
    "versatile": [],
    "powerful": [],
    "extra": []
  },
  "ignore": ["openai/gpt-realtime-2.1"],
  "aliases": {"arena": {"<display name>": "<canonical id>"}, "epoch": {}},
  "models": {
    "openai/gpt-6-luna": {
      "name": "GPT-6 Luna",
      "releaseDate": "2026-09-22",
      "providers": {
        "openai": {"id": "gpt-6-luna", "in": 0.1, "out": 0.5},
        "github-copilot": {"id": "gpt-6-luna", "in": 0.1, "out": 0.5}
      },
      "arena": {
        "text": {"rank": 12, "rating": 1480, "votes": 0, "variant": "High"},
        "coding": {}, "agent": {}, "webdev": {},
        "hardPrompts": {}, "math": {}, "instructionFollowing": {}, "multiTurn": {}
      },
      "external": {"eci": 0, "terminalBench": 0, "sweBenchVerified": 0, "metrHorizonMin": 0, "osworld": 0}
    }
  }
}}
```

- Manual: `tiers`, `ignore`, `aliases`, `since`. Sync rewrites `models`, `sources`, `syncedAt` and never touches the manual keys.
- Deterministic output (sorted keys, 2-space indent) for clean diffs.
- Without `-external`, sync keeps existing `external` values.
- Untiered is computed at display time, not stored: every Copilot model, plus Anthropic and OpenAI models with `release_date >= since`, text output and `tool_call: true`, minus tiered and ignored ones.
- Validation: tier names are the four above; every tier id exists in `models`; no id in two tiers.
- `.chezmoidata/*.json` is loaded as chezmoi template data on every apply and cannot be excluded by `.chezmoiignore`; the single top-level key `modelCatalog` avoids collisions (open question 4). `chezmoi dots lint registry` requires an `edit.toml` entry; root `make lint` already JSON-checks it.

## Commands

- `chezmoi agents models [-file <path>]`: validate the JSON and print a colored terminal table (tiers, then untiered).
- `chezmoi agents models sync [-file <path>] [-doc <path>] [-external]`: fetch all sources, update `models.json` and `docs/MODELS.md`. All-or-nothing: any source failure writes nothing. Reports unmatched rows, untiered models, and tier entries no provider lists any more (kept, with a warning).
- `chezmoi agents models render [-check] [-file <path>] [-doc <path>]`: regenerate `docs/MODELS.md` offline (after a manual tier edit); `-check` compares instead of writing and runs in `make lint`.
- `models.json` is located like `loadPolicy` finds `harness-policy.toml` (`CHEZMOI_SOURCE_DIR`, then relative candidates). The doc is `docs/MODELS.md` in the repo root, the parent of `home/`.

## Quicklook layout (`docs/MODELS.md`, generated)

1. Header: "Generated by `make models-sync`; do not edit", sync date, arena publish date and board.
2. One table per tier:

   | Model | Providers ($in/$out per Mtok) | Text | Code | Agent | WebDev | Hard | Math | IF | Multi |
   |---|---|---|---|---|---|---|---|---|---|
   | GPT-6 Luna | API $0.1/$0.5 · Copilot $0.1/$0.5 | #12 | ... | | | | | | |

   Rank cells are arena ranks (`#3`); empty when the model is not ranked. External columns (ECI, TB, SWE-V, METR, OSWorld) appear only when data is present.
3. Untiered: list with providers and cost.
4. Best when: per arena category, the top-ranked model among all models the three providers serve, tiered or not.
5. Footer when external data is shown: "Benchmark data: Epoch AI, CC-BY". Rankings: "LMArena leaderboard dataset".

The terminal table uses the same columns and the `term` Catppuccin styles.

## Package layout

Matching the `tools/agents` layering (pure domain packages, `cli` wiring):

- `internal/catalog` (pure: no network; renamed from `internal/models` because the Go house rules ban a package named `models`):
  - `catalog.go`: types; `file.go`: `Parse`, `Marshal`, `Validate`, `TierOf`, `Untiered`.
  - `key.go`: `Key` normalization and matching.
  - `modelsdev.go`, `arena.go`, `epoch.go`: parse raw bytes into rows.
  - `merge.go`: `Apply(catalog, sources) (Catalog, Report)`.
  - `render.go`: `Markdown(catalog)` and `WriteTerminal(w, style, catalog)`, built on `term.Table`/`term.Width`/`term.Pad`/`term.Warn` (`internal/term/table.go`).
- `internal/cli/models.go`: flags, file discovery, HTTP fetching (arena pages at most 4 in parallel, each retried 3 times with backoff); base URLs overridable for tests. Dispatch like `mcp` in `cli.go`, plus usage lines.
- No new Go dependencies with the JSON arena endpoint (open question 1); `golang.org/x/sync/errgroup` moves from indirect to direct (already in `go.sum`).
- Sync keeps in `models` every Copilot model, Anthropic and OpenAI models with `release_date >= since`, text output and `tool_call`, and every tiered model (a tiered model no provider lists any more keeps its last data and is reported as stale).

## Tests (offline)

- Fixtures in `internal/models/testdata/`: trimmed real models.dev JSON (a few models per provider), arena `/rows` pages, Epoch CSVs; the zip is built in the test with `archive/zip`.
- Table tests for `Key` with every id form above.
- `Apply`: tiers, ignore and aliases preserved; untiered and unmatched reported; `external` kept without `-external`.
- Golden `MODELS.md`.
- CLI tests against an `httptest` server. No network in tests.
- Standard `testing` plus `go-cmp`, `t.Parallel()`, `t.TempDir()`, `-race -shuffle=on` (per `tools/agents/AGENTS.md`).

## Make, docs, registry

- Root `Makefile`:
  - `models-sync: agents ## Refresh model data and docs/MODELS.md (network)` → `tools/agents/bin/chezmoi-agents models sync`. External data: run `chezmoi-agents models sync -external` directly.
  - `lint`: add `tools/agents/bin/chezmoi-agents models render -check`.
- `home/.chezmoidata/edit.toml`: `[edit.models]`, desc "Model tiers (manual) and synced model data", note "Regenerate docs/MODELS.md: chezmoi agents models render". Its `check = ["chezmoi-agents", "models", "-file", "{file}"]` is added only after the release is pinned (the installed binary has no `models` before that).
- `docs/OS.md`: new section "Model tiers (`chezmoi-agents models`)" with two runbooks (When / Steps / Verify / Rollback): refresh the data (`make models-sync`, review the diff, ship); re-tier a model (`dots models`, `chezmoi agents models render`, ship).
- Root `AGENTS.md`: one bullet (tiers in `.chezmoidata/models.json`, generated `docs/MODELS.md`, link to the OS.md section) and `docs/MODELS.md` in the Docs list.
- `tools/agents/AGENTS.md`: add `models` to the pure package list.
- `README.md`: `docs/MODELS.md` in the docs list.

## Release

New feature on a pre-1.0 tool: `agents-v0.5.0` (current 0.4.0).

1. PR 1 off `main`: `feat(agents): add models subcommand for model tier sync` — code, seeded `models.json` (seed tiers above, then one live sync), `docs/MODELS.md`, docs, make targets, `edit.toml` entry without its check. This WIP doc is deleted in it.
2. Tag `agents-v0.5.0` on the merge commit; `release-agents` publishes the binaries.
3. PR 2 off `main` (independent): `chore(packages): bump chezmoi-agents to 0.5.0` (`chezmoi dots bump chezmoi-agents`) and add the `edit.toml` check.

Implementation split: commit the `internal/models` types and signatures first; then in parallel the domain package with tests, and the CLI wiring with tests; integrate and run the first sync; then docs, Makefile, `edit.toml` and AGENTS. Done means `make check` passes at the root and in `tools/agents`.

## Shared setup with the usage tool

Agreed on 2026-11-06 with the driver of `chezmoi agents usage` (plan: `docs/wip/agents-usage.md` on branch `docs/agents-usage-plan`). One of each, no duplicated configuration:

| Piece | Owner | Where |
|---|---|---|
| Provider ids | shared | models.dev provider keys (`anthropic`, `openai`, `github-copilot`, `google`, ...); no second list |
| Provider/account/harness registry | usage tool | Go code in `internal/provider` (`Discover` reads the plan from each harness's credential at run time; no init prompt, no new `.chezmoidata` file, no `edit.toml` entry) |
| Usage snapshots | usage tool | `internal/usage`, cache file `os.UserCacheDir()/chezmoi-agents/usage.json` written only by `chezmoi agents usage`; `internal/cache.Dir(getenv)` owned by the usage tool |
| Model catalog | this plan | `internal/catalog`, `home/.chezmoidata/models.json` (the only `edit.toml` entry of the two), `docs/MODELS.md` |
| Terminal helpers | split | this plan: `term.Width`, `term.Pad`, `term.Table`, `term.Warn` in `internal/term/table.go`; usage tool: `term.Bar` in `internal/term/bar.go`; `term.go` untouched, no new colours |
| Apply hook and warning surface | this plan | one `chezmoi agents doctor` subcommand and one `run_after_agents-doctor.{sh,ps1}.tmpl`, later PR; capacity facts come from the usage tool (below), printed with `term.Warn` |
| Shared files | both | `cli.go`: one `case` and usage lines each; `docs/OS.md`: parent section "Models and subscriptions" (created here) with one subsection each; one bullet each in the AGENTS.md files; whoever merges second rebases |

Subscription and billing model (owned by the usage tool, read here; covers individual, team, enterprise seat, enterprise usage-based, overage, API key, cloud-hosted, org gateway, free and unknown future plans):

- `usage.Plan{Provider, Kind, Billing, Payer, Raw}`: Kind free | individual | team | enterprise | api | cloud | gateway | unknown; Billing flat | metered | mixed | unknown; Payer self | org | cloud | unknown; Raw is the provider's own plan string (shown as `unknown (<raw>)` when unmapped).
- `usage.Limit{Name, Kind (window | allowance | spend | credits), Scope (seat | pool | account), UsedPct, Used, Cap, Remaining, ResetsAt, Overage, Visible}`; `Visible=false` means managed by the org and not visible.
- `usage.Snapshot{Plan, Account, Limits, Status, FetchedAt}`, cached in `<cache dir>/chezmoi-agents/usage.json`.
- Pure functions the doctor calls: `usage.Summaries(snaps, now) []Summary{Provider, Plan, Headroom}` (one per provider; Headroom is the fraction left on the tightest visible limit with a future reset, nil when unknown; several accounts take the best headroom) and `usage.Advisories(snaps, now) []Advisory{Provider, Level, Text}` (capacity facts only).

## Apply-time advisory (`chezmoi agents doctor`, later PR)

Requirement (2026-11-06): `chezmoi update` / `chezmoi apply` warn, never change anything, when a newer or better model or tier config is available for what this machine can actually use. Not built until the questions below are answered.

What it checks, per harness selected on the machine (`includeTemplate "harnesses"`):

1. Reachable models: the harness's providers (fixed for Claude Code, Codex, Antigravity and Copilot; discovered from `auth.json` for opencode and pi) that have a `usage.Summary`; a model is reachable when one of those providers serves it in `models.json`. No account, no recommendation.
2. Better config: for each configured model (`harnessPolicy.models.tiers.*.model`, the `native.*` default models and Copilot agent models), the catalog tier it sits in; warn when a reachable model in the same catalog tier is newer (`releaseDate`) and ranks better on arena `text` and `coding`, or when the configured model is untiered or no reachable provider serves it any more.
   Billing shapes the suggestion (the judgement lives in the doctor's glue, since it needs the catalog): flat billing with healthy headroom may suggest a stronger model inside the plan; metered billing (api, cloud, enterprise usage-based) or an org payer with a spend cap prefers the cheaper model of the same tier by catalog $/Mtok and never suggests moving up a tier for rank alone; headroom below 10% on the provider of a default model suggests the same-tier model from a provider with headroom (info level); Kind unknown or a limit not visible means no capacity-based suggestion, rank only. Nothing is ever guessed or changed.
3. Stale data: `syncedAt` older than 30 days ("run `make models-sync`").
4. Capacity: the usage tool's advisories from its cached snapshot.

Output: nothing when there is nothing to say; otherwise a blue `chezmoi agents doctor` heading and one `term.Warn` line per advisory, on stderr. Always exit 0.

Integration point options:

| Option | How | For | Against |
|---|---|---|---|
| (a) `run_after_agents-doctor.{sh,ps1}.tmpl` | runs after every apply/update; body rendered only when harnesses are selected and the binary is on PATH; `chezmoi-agents doctor` offline, 2 s budget | matches the repo's `.chezmoiscripts` convention; sees account changes made outside chezmoi | runs on every apply (cheap: local files only) |
| (b) `run_onchange_after_` keyed on hashes of `models.json` and `harness-policy.toml` | runs only when the data changes | quietest | misses account/plan changes and ageing data; never repeats a warning |
| (c) a harness SessionStart hook (harness repo `hud`/`guard`) | warns inside the agent session | seen where models are chosen | outside this repo; needs every harness side built (parity rule) |
| (d) no hook, manual `chezmoi agents doctor` | none | zero apply cost | nobody runs it |

Recommended: (a), offline. Freshness then comes from `models.json` itself: `chezmoi update` pulls a catalog synced elsewhere, so open question 5 (a weekly sync PR) is what makes "newer model available" timely without network on apply.

## Open questions

Questions 1 to 6 were answered by approving this plan with its recommendations (`pickup and implement`, 2026-11-06); the implementation follows (a) for 1 to 5 and `2026-01-01` for 6.

### For the apply-time advisory (block the `doctor` PR only)

7. Integration point: (a) `run_after_` script, (b) `run_onchange_after_`, (c) harness SessionStart hook, (d) manual only. Recommended (a).
8. Network on apply: (a) offline only (local `models.json`, cached usage); (b) refresh models.dev once a day with a 2 s timeout, quiet on failure. Recommended (a), with question 5 answered (b) (weekly sync PR).
9. Catalog tier ↔ policy tier mapping, to judge "better in the same tier": proposed scout→lightweight, worker→lightweight, driver→powerful, escalate→extra (matches today's config: gpt-6-luna, gpt-6-luna, gpt-6.1-sol, gpt-6-astra). Alternatively an explicit `tierMap` key in `models.json`. Recommended: the fixed mapping in code.
10. Plan-level availability: models.dev lists what a provider serves, not what each subscription plan includes (e.g. Copilot Free vs Pro+). (a) provider-level reachability only; (b) a hand-kept plan → model allow-list. Recommended (a): (b) is a second, stale source.
11. Subcommand name: `doctor` (agreed with the usage tool) or `notice`.
12. Registry data: harness → provider facts as Go code in `internal/provider` (both drivers recommend) or a `[harnessProviders]` table in `harness.toml`.

### For the sync commands

1. Arena download method:
   - (a) HF `/rows` JSON, about 112 requests per sync with retries, no new dependency; the HF JSON service was unreliable on 2026-11-06 (`/filter`).
   - (b) Parquet files plus `github.com/parquet-go/parquet-go`: one reliable download per board, large dependency tree.
   - Recommended: (a), since sync is manual and all-or-nothing.
2. Arena variant rule: (a) best-ranked variant per model, variant shown; (b) a fixed variant. Recommended: (a).
3. Text board: (a) `text_style_control`; (b) `text`. They rank differently (Claude Fable 5.1 is #4 on one, #5 on the other). Recommended: (a), assuming arena.ai shows style-controlled ranks by default; confirm against the site.
4. Location of `models.json`: (a) `home/.chezmoidata/models.json` under one key `modelCatalog`, loaded as template data on every apply (small cost), editable with `dots`; (b) outside `home/` (e.g. `tools/agents/models.json`), not loaded, no registry entry, not editable with `dots`. Recommended: (a).
5. Scheduled refresh: (a) manual `make models-sync` only; (b) weekly workflow opening a PR, like `bump-releases`. Recommended: (a) for now.
6. `since` cutoff for flagging Anthropic/OpenAI models: `2026-01-01` proposed (keeps out legacy models such as GPT-5 nano and Claude 4.5).

## Out of scope

- `tools/agents/AGENTS.md` points to a release runbook in `docs/OS.md` → AI harnesses that does not exist (only the dots release runbook is there). Worth a docs-sync pass, possibly in PR 1.
