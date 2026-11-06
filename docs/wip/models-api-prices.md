# WIP: first-party API prices and automatic tiers for `chezmoi agents models`

Status: A and B in implementation on `feat/models-api-prices`; C (automatic tiers) waits on the user's choice of rule; D partly waits on C. Delete this file when the work ships; its lasting content moves to `docs/OS.md` and `docs/MODELS.md`.

## A. Untiered models: nothing was lost

- The seed tiers (`docs/wip/models-sync.md` → Seed tiers, commit `877f3b8`) listed 26 ids; the first live sync (`1def0f5`) and today's `models.json` hold the same 26, unchanged. No tier entry was removed or renamed by the sync or by `Key` matching.
- Claude Opus 5.5, Opus 5, Opus 4.7, GPT-5.5, GPT-5.4, GPT-5.4 mini, GPT-5 mini, Gemini 3.5/3.6 Flash and Kimi K2.7 Code were never in the seed: the seed plan already lists them as "untiered Copilot models at seed time (sync must flag them)". Sync flagged them as designed.
- Root cause of the long untiered list: tiers are hand-picked, so every model the seed did not name stays untiered until someone edits `tiers`. C replaces that with a rule.

## B. Provider model: first-party API prices only

- Scope: `providers` (new manual key, the tracked first-party models.dev provider keys): `anthropic`, `google`, `moonshotai`, `openai`, `xai`. A model is tracked when the first-party provider lists it (provider key = prefix of its canonical id), it was released on or after `since`, outputs text, supports tool calls and is not `status: deprecated`, minus `ignore`. One loop over `providers`; no provider-specific branch in code or data.
- `github-copilot` is gone from sync, data, render, tests and docs. Copilot's per-token prices equal the first-party ones (checked on models.dev, 2026-11-06), so nothing is lost.
- Per model: `provider`, `id` (the id at the provider's API), `name`, `releaseDate`, `context`, `api {in, out, cacheRead, cacheWrite}` ($/Mtok). `api` is absent when models.dev lists no cost: rendered `n/a`. When a provider lists several ids for one canonical id (`gpt-5.6` and `gpt-5.6-sol`), the entry whose id equals the canonical id's part wins, else the first by id.
- Every table (tiers, untiered) has the same columns: Model, Provider, Id (canonical), $In, $Out, $Cache (read), Context, the arena ranks, the external benchmarks.
- Microsoft MAI-Code-1.1-Flash leaves the catalog: `microsoft` has no first-party provider on models.dev; it was Copilot-only.
- Schema 2: `models.json` is migrated by hand (schema, `providers`, MAI removed from `tiers`) and refilled by a live sync.

### Interface for the doctor advisory (later PR)

- The catalog is keyed by model and first-party provider, never by harness. `Model.Provider` uses the models.dev keys of `internal/provider` (`anthropic`, `openai`, `google`; `xai` and `moonshotai` when the usage tool learns them).
- Doctor input: `catalog.Catalog` (tiers, `Model.Provider`, `Model.ID`, `Model.API`) × the accounts the usage tool discovers per harness. Provider-tied harnesses (Claude Code, Codex, Antigravity, Copilot) map to their provider; agnostic ones (OpenCode, Pi) to every provider they hold a login for. Candidates for a role = models of the role's tier whose `Provider` is in that set; cost comparisons use `Model.API`.

## C. Automatic tiers (open question for the user)

See the driver report for the proposed rule and its table on today's data. Nothing is built until the user picks a rule.

## D. One command

`chezmoi agents models sync` refreshes prices, ranks, benchmarks and `docs/MODELS.md`; `render -check` stays in `make lint`. Open question: a scheduled weekly refresh PR (like the release bump workflow).
