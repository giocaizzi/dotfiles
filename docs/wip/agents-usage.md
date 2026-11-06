# WIP: `chezmoi agents usage` (subscription usage per provider account)

Status: plan, not implemented. Sources checked on 2026-11-06. Delete this file when the feature ships; its lasting content moves to `docs/OS.md`.

## Goal

One command that shows, for every provider account the machine is logged into, the plan and how much of the subscription's capacity is used: rate-limit windows (percent used, reset time), monthly allowances, and credits.

## Decisions made

1. Provider-centric. Accounts are grouped by provider and found automatically from the credentials the harnesses already store:
   - Single-tenant harnesses: Claude Code (`anthropic`), Codex (`openai`), Antigravity (`google`), Copilot CLI (`github-copilot`).
   - Multi-tenant harnesses: opencode and pi (any provider they hold a login for).
   - The same account seen from several harnesses is shown once, with the harnesses listed. For example, one ChatGPT account used by Codex, opencode and pi.
2. Credentials are read into memory only. They are never printed, logged, cached or written back.
3. The output uses the Catppuccin styles of `internal/term`, like the other `chezmoi-agents` subcommands.
4. Shared setup with `chezmoi agents models`: see [Shared setup](#shared-setup-with-chezmoi-agents-models). It was agreed with that work and is recorded in both plans.

## What each provider exposes

Every subscription usage endpoint below is undocumented: it is the one the vendor's own CLI calls, and it can change without notice. The parsers treat every field as optional and turn any failure into "unavailable" for that account. HTTP 401, 403 and 429 are reported as such, never retried.

### Anthropic (Claude Pro/Max login)

- Credentials:
  - macOS: Keychain item `Claude Code-credentials`. When `CLAUDE_CONFIG_DIR` is set, the name gets the suffix `-<first 8 hex of sha256(dir)>`.
  - Linux and Windows: `~/.claude/.credentials.json`, mode 0600.
  - Keys (checked on this machine): `claudeAiOauth.{accessToken, refreshToken, expiresAt (ms), scopes, subscriptionType, rateLimitTier}`.
  - The item also holds `mcpOAuth`; ignore it.
  - Docs: <https://code.claude.com/docs/en/authentication>.
- Usage: `GET https://api.anthropic.com/api/oauth/usage` with `Authorization: Bearer` and `anthropic-beta: oauth-2025-04-20`.
  - Response: `five_hour`, `seven_day`, `seven_day_opus`, `seven_day_sonnet` and others, each `{utilization 0-100, resets_at}`; `extra_usage {is_enabled, monthly_limit, used_credits, utilization}`.
  - Plan: `subscriptionType` from the credentials (`max` here).
- Official, but in-session only: the statusline JSON `rate_limits.{five_hour, seven_day}` (<https://code.claude.com/docs/en/statusline>). It reaches only a statusline process, so it is not a source for this command.
- opencode no longer supports Claude logins. pi does (`anthropic` oauth), with the same usage endpoint.

### OpenAI (ChatGPT login: Codex, opencode, pi)

- Credentials:
  - Codex: `$CODEX_HOME/auth.json` (default `~/.codex`), with `auth_mode`, `tokens.{access_token, account_id, id_token, refresh_token}`, `OPENAI_API_KEY`. Checked on this machine.
  - With `cli_auth_credentials_store = keyring`, Codex keeps them in the OS store under service `Codex Auth` instead.
  - Source: <https://raw.githubusercontent.com/openai/codex/main/codex-rs/login/src/auth/storage.rs>.
  - opencode: `openai` oauth `{access, refresh, expires, accountId}`.
  - pi: `openai` oauth. Its account id is the `chatgpt_account_id` claim of the access JWT.
  - The account id is the key used to merge the three into one account.
- Usage: `GET https://chatgpt.com/backend-api/wham/usage` with `Authorization: Bearer`, `ChatGPT-Account-Id` and `originator: codex_cli_rs`.
  - Response: `plan_type`; `rate_limit.{primary_window, secondary_window}` each `{used_percent, limit_window_seconds, reset_at}`; `credits.{has_credits, unlimited, balance}`.
  - Model: <https://raw.githubusercontent.com/openai/codex/main/codex-rs/codex-backend-openapi-models/src/models/rate_limit_status_payload.rs>.
- Offline source (official protocol, checked on this machine): Codex session logs `~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl`.
  - The lines with `type: event_msg` and `payload.type: token_count` carry `payload.rate_limits.{plan_type, primary, secondary}`, each `{used_percent, window_minutes, resets_at}`, plus `credits`.
  - The newest snapshot is stale by definition. A window whose `resets_at` has passed shows as reset.
  - Source: <https://raw.githubusercontent.com/openai/codex/main/codex-rs/protocol/src/protocol.rs>.

### GitHub Copilot (the special case)

- Why it differs from the other providers:
  - One GitHub account owns a monthly allowance shared by every model and every surface: the CLI, VS Code, opencode and pi.
  - There is no rolling 5-hour or weekly window.
  - The seat can belong to an organization (Business, Enterprise). Then the user has no billing view of their own.
  - GitHub moved Copilot to usage-based billing (AI Credits) on 2026-06-01, so the shape of the allowance must be checked live during implementation.
- Credentials:
  - Copilot CLI: the OS store (Keychain service `copilot-cli` on this machine). `~/.copilot/config.json` (JSONC) lists `loggedInUsers` and `lastLoggedInUser`, with no token.
  - Environment: `COPILOT_GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_TOKEN`.
  - opencode and pi: `github-copilot` oauth, whose `refresh` is the GitHub OAuth token.
  - The account key is the GitHub login.
  - Docs: <https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference>.
- Usage (undocumented, the one the Copilot clients use): `GET https://api.github.com/copilot_internal/user` with the GitHub token.
  - Response: `copilot_plan`, `quota_reset_date`, `organization_login_list`, and `quota_snapshots.{premium_interactions, chat, completions}`, each `{entitlement, remaining, percent_remaining, unlimited, overage_permitted, overage_count}`.
  - On paid plans only `premium_interactions` is a real limit.
  - It works for organization seats too.
- Official alternative: `GET /users/{username}/settings/billing/premium_request/usage` (<https://docs.github.com/en/rest/billing/usage>).
  - It returns spend, not the allowance or what remains.
  - It covers only a plan the user bought personally.
  - So it does not answer "how much is left".

### Google (Antigravity, Gemini CLI)

- Usage (undocumented Code Assist API): `POST https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist` gives `currentTier`, `paidTier.name` and `cloudaicompanionProject`.
- Then `POST .../v1internal:retrieveUserQuota {"project": ...}` gives `buckets[]`, each `{modelId, remainingFraction, resetTime}`.
- Source: <https://raw.githubusercontent.com/google-gemini/gemini-cli/main/packages/core/src/code_assist/types.ts>.
- Credentials:
  - Gemini CLI: `~/.gemini/oauth_creds.json` `{access_token, refresh_token, expiry_date}`. Since 2026-06-18 Gemini CLI OAuth no longer serves individual, AI Pro and Ultra accounts (Standard, Enterprise and Workspace still work).
  - Antigravity keeps its own login in the OS store under a name not yet identified. Community tools read it from Windows Credential Manager `gemini:antigravity`. On macOS the Keychain item's access list is limited to `agy`, so another process triggers a prompt.
- Logs: `~/.gemini/antigravity-cli/cli.log` shows only `Cache(loadCodeAssistResponse)` refresh lines and no quota values, so it is not a source. This machine is currently logged out of Antigravity.
- Conclusion: Google is the weakest provider (open question 6).

### API-key accounts (opencode, pi, environment)

- A plain API key has no subscription and no remaining-quota endpoint at Anthropic or OpenAI. Their organization Usage and Cost APIs need an admin key.
- Three providers have documented balance endpoints that accept a normal key:
  - OpenRouter: `GET /api/v1/key`, giving `limit`, `limit_remaining`, `usage`.
  - DeepSeek: `GET /user/balance`.
  - Moonshot: `GET /v1/users/me/balance`.
- Docs: <https://openrouter.ai/docs/api-reference/limits>, <https://api-docs.deepseek.com/api/get-user-balance>, <https://platform.kimi.ai/docs/api/balance>. Whether to support them is open question 5.

### Multi-tenant stores

- opencode:
  - V2 keeps credentials in SQLite: `~/.local/share/opencode/opencode.db` (`XDG_DATA_HOME`, `OPENCODE_DB`), table `credential` (`integration_id`, `label`, `active`, `value` JSON `{type: oauth|key|external, access, refresh, expires, metadata}`).
  - It can hold several accounts per provider, one of them active.
  - `auth.json` beside it is the legacy V1 store. It is imported once and not updated afterwards.
  - Rows with `mcp_` ids are MCP logins; ignore them.
  - Docs: <https://github.com/anomalyco/opencode/blob/dev/services/site/src/docs/content/cli/providers.mdx>.
- pi: `~/.pi/agent/auth.json` (`PI_CODING_AGENT_DIR`), one entry per provider id: `{type: oauth, access, refresh, expires}` or `{type: api_key, key}`, where `key` may be `!command`.
  - pi never runs a `!command` key here: such an account is listed as "key from command, skipped".
  - Docs: <https://github.com/badlogic/pi-mono/blob/main/packages/coding-agent/docs/configuration.md>.
- Logged in on this machine: opencode `openai` and `github-copilot` (oauth); pi `openai` and `github-copilot` (oauth).

## Command

```text
chezmoi agents usage [-provider <id>] [-offline]
```

- Discovers accounts, queries each provider in parallel (one goroutine per account, 10 s timeout each, overall 15 s), prints, and writes the cache.
- `-provider` limits the output to one provider id.
- `-offline` skips the network. It shows discovered accounts, the cached snapshot with its age, and Codex session snapshots.
- Exit code: 0 when it printed, even when some accounts are unavailable; 1 only on a usage error.
- Root `Makefile`: `usage: agents ## Show subscription usage per provider account (network)`.

### Output (Catppuccin via `internal/term`)

```text
chezmoi agents usage — subscription usage per provider account

anthropic        Claude Max                              claude
  5h             ██████░░░░░░░░░░░░░░   31%   resets in 2h 14m
  7d             ███░░░░░░░░░░░░░░░░░   14%   resets Mon 09:00
  7d opus        ░░░░░░░░░░░░░░░░░░░░    0%
openai           ChatGPT Plus · me@example.com           codex, opencode, pi
  5h             ███████░░░░░░░░░░░░░   33%   resets in 3h 02m
  7d             █░░░░░░░░░░░░░░░░░░░    5%   resets Thu 18:30
github-copilot   Pro · octocat                           copilot, opencode, pi
  premium        ██████████████░░░░░░   70%   90 of 300 left · resets Dec 1
google           Antigravity                             antigravity
  ! not logged in (run agy)
openrouter       API key                                 pi
  ! no usage API
```

- Provider in blue (`B`). Plan and account in mauve (`M`); the account is an email or login, never a token or account id.
- Harnesses and reset times in dim (`D`).
- Bars use `term.Bar` in green below 70%, yellow from 70%, red from 90%.
- Warnings use `term.Style.Warn` (owned by the models work).
- Columns are aligned with `term.Width`/`Pad` (owned by the models work).

## Shared setup with `chezmoi agents models`

Agreed with the models-sync work on 2026-11-06 and recorded in both WIP plans. There is one of each:

| Piece | Owner | Notes |
|---|---|---|
| Provider ids | shared | models.dev provider keys: `anthropic`, `openai`, `github-copilot`, `google`, plus models.dev ids for API-key providers. No second list. |
| Provider and account registry | usage (this plan) | `internal/provider`, Go code, no data file and no `chezmoi init` prompt. Plan tier and account come from the credentials at run time. Open question 8 offers a `harness.toml` table instead. |
| Usage queries and snapshot | usage | `internal/usage`; pure `usage.Advisories(snaps, now) []string` for the warning surface |
| Cache dir | usage | `internal/cache.Dir(getenv)`: `$XDG_CACHE_HOME`, otherwise `os.UserCacheDir()`, plus `/chezmoi-agents`. Usage writes `usage.json`. |
| Model catalog | models | `internal/catalog`; it does not import `provider` (the glue is in `cli`) |
| `term` helpers | models: `table.go` (`Width`, `Pad`, `Style.Table`, `Style.Warn`, pushed at 02ad341); usage: `bar.go` (`Bar`) | `term.go` untouched |
| Apply/update warning | models owns the one subcommand `chezmoi agents doctor` and the one script `run_after_agents-doctor.{sh,ps1}.tmpl`, in a later PR | Usage contributes capacity advisories from the cached snapshot. Open question 7. |
| `cli.go`, `docs/OS.md`, AGENTS files | both, minimal hunks | One `case` and its usage lines each. `docs/OS.md` gets a parent section "Models and subscriptions" (created by models) with a subsection per command. Whoever merges second rebases. |
| `edit.toml` | models only (`[edit.models]`) | Usage has no data file, so it has no entry. |

## Package layout

Following the `tools/agents` layering: pure packages know nothing of processes or the network, and `cli` wires them.

- `internal/provider` (pure). Readers take an injected `Store` interface (`ReadFile`, `Secret(service, account)`, `Getenv`, `Home`) and return accounts:
  - `provider.go`: `Provider` (id, name, kind: subscription or API key), the fixed harness-to-provider facts, and `Account`.
    - `Account` fields: provider, label, plan, harnesses, auth kind, account key for merging, `expires`, and an unexported credential.
    - `String`, `GoString`, `MarshalJSON` and `LogValue` redact the credential.
  - One reader per harness: `claude.go`, `codex.go`, `copilot.go`, `gemini.go`, `opencode.go`, `pi.go`.
  - `discover.go`: `Discover(ctx, Store, harnesses []string) ([]Account, []Problem)`. It merges by (provider, account key) and keeps the most recently refreshed credential.
- `internal/usage` (pure):
  - `Snapshot` (provider, account label, plan, windows, allowance, credits, fetched-at, status).
  - `Window` (name, used %, resets at).
  - One parser per response: `anthropic.go`, `openai.go` (HTTP and session JSONL), `copilot.go`, `google.go`, plus `balance.go` if open question 5 says so.
  - `Advisories`, and `Marshal`/`Parse` for the cache.
- `internal/cache` (pure apart from `os.UserCacheDir`): `Dir`.
- `internal/term/bar.go`: `Bar(st Style, percent float64, width int) string`.
- `internal/cli/usage.go`: flags; the real `Store` (files plus the OS credential store, open question 2); the HTTP client (timeouts, contexts, closed bodies, base URLs overridable for tests); parallel fetch with `sync.WaitGroup.Go`; rendering; the cache write (0600, atomic).
- Expired tokens: if `expires` has passed, the account shows "token expired, open <harness> to refresh" and is not queried (open question 3).

## Tests (offline)

- Every test uses `t.Parallel()`, `t.TempDir()`, `go-cmp`, and runs with `-race -shuffle=on`.
- Fixtures in `internal/provider/testdata/` hold every credential shape, with fake tokens:
  - Claude JSON, Codex `auth.json` (file and keyring), Copilot `config.json` (JSONC);
  - opencode `auth.json` and a small SQLite fixture created in the test, pi `auth.json`, Gemini `oauth_creds.json`.
- Fixtures in `internal/usage/testdata/` hold every response shape: trimmed real responses with identifiers replaced, and a Codex session JSONL.
- Table tests: merging, expiry, `!command` keys skipped, MCP rows ignored, redaction (`fmt.Sprint`, `%#v`, `json.Marshal` and `slog` of an `Account` never contain the fake token).
- CLI tests run against `httptest` servers; golden terminal output is checked with and without colour.

## Docs, runbook, AGENTS

- `docs/OS.md` gets a subsection "Subscription usage (`chezmoi-agents usage`)" under the shared parent section. It covers:
  - what is discovered, per harness and OS;
  - what each provider shows, and that the endpoints are undocumented;
  - that credentials stay in memory.
  - Runbook (When / Steps / Verify / Rollback): check capacity before fanning out agents, refresh an expired login, add a provider.
- Root `AGENTS.md`: one bullet linking to the subsection.
- `tools/agents/AGENTS.md`: add `provider`, `usage` and `cache` to the pure package list, and add a rule: "credentials are read into memory only; never printed, logged, cached or written back".
- The `host-hygiene` skill's quota line (`claude -p /usage`, `/status`, ...) could later point to `chezmoi agents usage`. That skill lives in the harness repo, so the change is out of scope here; noted in the report.
- Parity: macOS, Linux and Windows paths are covered in the readers. A source that cannot be built on one OS gets a GitHub issue with a priority label, filed before shipping.

## Release

- Independent of the models PR. The packages are disjoint, and the shared files take one hunk each.
- PR graph:
  1. PR 1 off `main`: `feat(agents): add usage subcommand for subscription usage per provider`. It contains the code, docs, Makefile target and AGENTS bullets, and deletes this WIP file. It does not depend on the models PR: if `term.Width`/`Pad`/`Warn` are not on `main` yet when this PR lands, it rebases onto the models PR and becomes a stack layer on it.
  2. The `doctor` hook comes later, in the models work's PR, after both land.
- Version: tag `agents-vX.Y.0` (the next minor after whatever the models PR takes), then `chore(packages): bump chezmoi-agents`. The hub releases.
- Implementation split, after the answers:
  1. On this branch, commit the types and signatures of `provider`, `usage` and `cache`.
  2. Then, in parallel:
     - worker A: `internal/provider` and its tests;
     - worker B: `internal/usage` and `internal/cache` and their tests;
     - worker C: `term/bar.go`, `cli/usage.go`, the dispatch and the CLI tests.
  3. Then docs and the Makefile.
- Done means `make check` passes at the root and in `tools/agents`.

## Open questions (for the user, one at a time)

1. Integration point: (a) a `chezmoi agents usage` subcommand in `tools/agents`; (b) a separate tool; (c) a statusline segment in the harness `hud` plugin. Recommended: (a). It reuses the term styles, release and packaging, and sits next to `models`. (c) sees only the running harness's own provider.
2. Reading the OS credential store, which macOS (Claude, Copilot, Codex keyring mode), Windows and Linux need:
   - (a) the dependency `github.com/zalando/go-keyring`: one API for all three OSes, but small dependencies, including D-Bus on Linux;
   - (b) executing the OS tools: `security` on macOS, `secret-tool` on Linux, no Windows equivalent without a module;
   - (c) files only.
   - Recommended: (a).
   - Related: opencode V2 credentials are in SQLite. Reading them needs either the dependency `modernc.org/sqlite` (pure Go, large) or the legacy `auth.json` only, which is stale after V2. Recommended: `modernc.org/sqlite`, read-only.
3. Expired access tokens:
   - (a) read-only: show "expired, open <harness>" and never refresh;
   - (b) refresh in memory without writing back.
   - Recommended: (a). Refresh tokens rotate (Claude, Codex), so a refresh outside the harness can log it out. Harnesses refresh often, so expiry is rare in practice.
4. Copilot:
   - (a) the allowance meter from `copilot_internal/user` (plan, premium allowance used and left, reset date), which works for organization seats;
   - (b) the official billing API: spend only, personal plans only;
   - (c) both.
   - Recommended: (a). The AI Credits shape is checked live first.
5. Providers with no usage API (plain Anthropic or OpenAI keys, any key-only provider):
   - (a) list the account with "no usage API";
   - (b) as (a), plus balances for OpenRouter, DeepSeek and Moonshot through their documented endpoints;
   - (c) hide them.
   - Recommended: (a) now. Add (b) only for a key the user actually holds; this machine has none.
6. Google:
   - (a) Gemini CLI credentials plus the Code Assist quota API. Antigravity's own login is shown as "usage not supported yet", with a GitHub issue to find its credential store per OS;
   - (b) also query agy's loopback quota listener while agy runs (undocumented, needs agy running);
   - (c) leave Google out.
   - Recommended: (a).
7. Apply-time capacity advisory, shared with the models `doctor` hook:
   - (a) cache only: `doctor` reads `usage.json` and warns when a window is at 90% or more and its reset is in the future, but never touches the network on apply;
   - (b) `doctor` queries live with a 2 s budget;
   - (c) no capacity advisory on apply.
   - Recommended: (a). Apply stays offline and fast, and the cache is as fresh as the last `chezmoi agents usage`.
   - The hook's name (`doctor` or `notice`) is the models plan's open question; do not answer it twice.
8. Harness-to-provider facts:
   - (a) Go code in `internal/provider`: nothing in the templates reads them, and opencode and pi are discovered at run time anyway;
   - (b) a `[harnessProviders]` table in `.chezmoidata/harness.toml`.
   - Recommended: (a).
