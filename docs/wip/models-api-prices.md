# WIP: doctor advisory inputs

Status: the catalog work (first-party API prices, automatic tiers, one `models sync` command) shipped. This note feeds the later doctor PR; delete it when that ships.

## Doctor advisory: inputs

- The catalog is keyed by model and first-party provider, never by harness. `Model.Provider` uses the models.dev keys of `internal/provider` (`anthropic`, `openai`, `google`; `xai` and `moonshotai` when the usage tool learns them).
- Doctor input: `catalog.Catalog` (tiers, `Model.Provider`, `Model.ID`, `Model.API`) × the accounts the usage tool discovers per harness. Provider-tied harnesses (Claude Code, Codex, Antigravity, Copilot) map to their provider; agnostic ones (OpenCode, Pi) to every provider they hold a login for. Candidates for a role = models of the role's tier whose `Provider` is in that set; cost comparisons use `Model.API`.
