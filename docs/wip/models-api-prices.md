# WIP: doctor advisory inputs

Status: the catalog work (first-party API prices, automatic tiers, one `models sync` command) shipped. This note feeds the later doctor PR; delete it when that ships.

## Doctor advisory: inputs

- The catalog is keyed by model and first-party provider, never by harness. `Model.Provider` uses the models.dev keys of `internal/provider` (`anthropic`, `openai`, `google`; `xai` and `moonshotai` when the usage tool learns them).
- Doctor input: `catalog.Catalog` (tiers, `Model.Provider`, `Model.ID`, `Model.API`) × the accounts the usage tool discovers per harness. A harness maps to the first-party providers whose models it can run with the accounts the usage tool finds. Provider-tied harnesses (Claude Code → `anthropic`, Codex → `openai`, Antigravity → `google`) map to their provider; harnesses that serve several vendors' models through one account (Copilot) or any provider (OpenCode, Pi) map to every first-party provider whose models they expose. `github-copilot` is an account/billing provider, never a catalog provider. Costs always use the model's first-party API price. Candidates for a role = models of the role's tier whose `Provider` is in that set; cost comparisons use `Model.API`.
