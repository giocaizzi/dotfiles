# WIP: `chezmoi agents usage` (subscription usage per provider account)

Status: approved, in implementation. Sources checked on 2026-11-06. Delete this file when the feature ships; its lasting content moves to `docs/OS.md`.

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
5. Integration point: a `chezmoi agents usage` subcommand in `tools/agents`.
6. New Go dependencies:
   - `github.com/zalando/go-keyring` reads the OS credential store (macOS Keychain, Windows Credential Manager, Linux Secret Service).
   - `modernc.org/sqlite` reads opencode's V2 credential store, opened read-only (`mode=ro`).
7. Expired access tokens are never refreshed: the account shows "token expired, open <harness>" and is not queried.
8. Copilot shows both sources:
   - the allowance meter from `copilot_internal/user`;
   - the spend this month from the official billing API.
   The billing API covers only a plan the user pays for personally; for an organization seat it shows "managed by org".
9. API-key accounts with no usage API are listed as "no usage API".
10. Google is supported through Antigravity's own `/usage`. The primary source is `agy -p /usage --output-format json`; the fallback is the quota-exhausted lines in agy's log. See [Google](#google-antigravity).
11. Cached by default, live on request:
    - `chezmoi agents usage` reads the cache `usage.json` by default; `-live` queries every provider and rewrites the cache.
    - The apply hook (`chezmoi agents doctor`, owned by the models work) works the same way: cache by default, `--live` queries with a 2 s cap per provider, and failures degrade quietly to the cache.
12. Harness-to-provider facts are Go code in `internal/provider`. There is no data table and no `chezmoi init` prompt.
13. Every plan and billing model is handled generically, including future ones; see [Plans and billing models](#plans-and-billing-models):
    - an unmapped plan string shows as a labelled `unknown (<raw>)` row and never crashes;
    - data a seat holder cannot see shows "managed by org" or "not visible", never a guess.
14. No employer or organization names appear in code, fixtures or docs. Examples use generic wording (`me@example.com`, `octocat`, "your organization").

Status: approved on 2026-11-06 with the answers above; implementation started.

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

### Google (Antigravity)

Evidence from this machine (2026-11-06; agy is logged in; earlier "logged out" lines were from older sessions):

- `agy -p /usage --output-format json` answers locally:
  - It takes about 4 s.
  - `num_turns: 0` and zero tokens, so no model turn runs.
  - `command.name: "usage"`.
  - `command.data.groups[]` is `{name, description, buckets[]}`. Each bucket is `{id, name, window: "5h"|"weekly", remaining_fraction 0-1, reset_time RFC 3339, disabled?, description}`.
  - Groups seen: "Gemini Models", "Claude and GPT models", "Gemini 4 Argon". Bucket ids seen: `gemini-weekly`, `gemini-5h`, `3p-weekly`, `3p-5h`, `gemini-argon-wtu-weekly`.
  - The plain `-p /usage` prints the same as TSV: group, bucket, remaining percent or `disabled`, reset time.
- agy stores its login with `zalando/go-keyring` (the same library this tool uses), through a `KeyringTokenStorage` with a file fallback.
  - The service name is not visible.
  - On macOS the item's access list is limited to agy.
  - Refreshing needs agy's own OAuth client.
- agy calls `v1internal:retrieveUserQuotaSummary` (`QuotaSummary` buckets) on `cloudcode-pa.googleapis.com`, plus `loadCodeAssist` and `fetchAvailableModels`.
- `quota_manager.go` logs `doRefreshQuota: starting reload | skipped (throttled) | skipped (not logged in)` with no values.
- Quota-exhausted errors are logged as:
  `E1101 13:07:11.896026 2480 errorreport.go:224] ... RESOURCE_EXHAUSTED (code 429): Individual quota reached. ... Resets in 4h58m35s.`
  The line has no year; the file name `log/cli-YYYYMMDD_HHMMSS.log` gives it.
- `agy -p /status` is not a local command: it starts a model turn. So the plan name is not exposed, and the plan is `unknown (antigravity)` unless `/usage` says more later.

Design (decided on this evidence):

1. Primary source: run `agy -p /usage --output-format json --print-timeout 10s` with a 10 s context (`usage -live`).
   - The answer is accepted only if `command.name == "usage"` and `num_turns == 0`. Otherwise it shows "unavailable: agy did not answer /usage" and is never retried. This guards against a future agy that would send `/usage` to the model.
   - agy keeps owning its credentials and their refresh: nothing is read from its keychain item.
2. Fallback: when agy is missing, fails or times out, the newest log file's last `RESOURCE_EXHAUSTED ... Resets in <duration>` line still in the future gives an exhausted window with that reset (source `agy-log`). Otherwise the cached snapshot is shown.
3. Rejected: replicating `retrieveUserQuotaSummary` with agy's token. The keychain item is limited to agy (a prompt every run), the service name is unknown, and refresh needs agy's private OAuth client.
4. With the doctor's 2 s `--live` cap, agy (about 4 s) degrades quietly to the cache. That is expected.
5. Mapping: each bucket becomes a window limit named `<group> · 5h|7d`.
   - `UsedPct = 100 * (1 - remaining_fraction)`.
   - `disabled: true` maps to `Disabled`.
   - Scope is account.

### API-key accounts (opencode, pi, environment)

- A plain API key has no subscription and no remaining-quota endpoint at Anthropic or OpenAI. Their organization Usage and Cost APIs need an admin key.
- Three providers have documented balance endpoints that accept a normal key:
  - OpenRouter: `GET /api/v1/key`, giving `limit`, `limit_remaining`, `usage`.
  - DeepSeek: `GET /user/balance`.
  - Moonshot: `GET /v1/users/me/balance`.
- Docs: <https://openrouter.ai/docs/api-reference/limits>, <https://api-docs.deepseek.com/api/get-user-balance>, <https://platform.kimi.ai/docs/api/balance>. They are not supported for now (decision 9): such accounts show "no usage API".

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

## Plans and billing models

Every account is described by one generic model in `internal/usage` (code: `tools/agents/internal/usage/usage.go`). Provider mappings only fill it in.

- `Plan{Kind, Billing, Payer, Name, Raw}`:
  - Kind: `free`, `individual`, `team`, `enterprise`, `api`, `cloud`, `gateway`, `unknown`.
  - Billing: `flat` (an allowance included in the plan or seat), `metered` (pay per use), `mixed` (allowance plus overage or credits), `unknown`.
  - Payer: `self`, `org`, `cloud`, `unknown`.
  - Raw: the provider's own string. An unmapped value gives Kind `unknown`, shown as `unknown (<raw>)`.
- `Limit{Name, Kind, Scope, UsedPct, Used, Cap, Remaining, ResetsAt, Overage, Unlimited, Disabled, Hidden, Note}`:
  - Kind: `window` (5 h, weekly), `allowance` (monthly), `spend` (against a cap or budget), `credits` (a balance).
  - Scope: `seat`, `pool` (organization), `account`.
  - Every amount is optional and carries its unit (`requests`, `credits`, `USD`).
  - `Hidden` means "managed by org, not visible": the row is shown with that label and no number.
- `Snapshot.Status`:
  - `ok`;
  - `stale` (cache or an offline source);
  - `expired`;
  - `unavailable` (the call failed);
  - `no-api`;
  - `not-visible` (usage lives with the organization, the cloud console or the gateway);
  - `not-logged-in`;
  - `skipped`.
- Unknown JSON fields are ignored.
  - An unknown window key in a known payload becomes a generic window row named after the key (for example a new `seven_day_*` from Anthropic).
  - Nothing panics on an unexpected shape: that account becomes `unavailable` with the parse error, and the other accounts still show.

Mappings (checked 2026-11-06; values not in the tables map to `unknown (<raw>)`):

| Provider | How it is detected | Plan mapping | What a seat holder sees |
|---|---|---|---|
| anthropic, subscription | `claudeAiOauth` login; `subscriptionType`, `rateLimitTier` | `pro`, `max` (name from `rateLimitTier`, e.g. `default_claude_max_5x` → "Max 5x"): individual / flat / self, extra usage → mixed. `team`: team / flat / org. `enterprise`: enterprise / unknown / org (seat- and usage-based cannot be told apart locally). | `oauth/usage` windows (5 h, 7 d, 7 d per model family, Scope `seat` on team or enterprise). `extra_usage` as a spend limit when present. A window that comes back null on a team or enterprise seat shows "no per-seat window (org-managed or usage-based)". Organization spend limits show as hidden rows. |
| anthropic, API key | `ANTHROPIC_API_KEY` or `ANTHROPIC_AUTH_TOKEN` in the environment or in the `env` block of `~/.claude/settings.json`, or `apiKeyHelper` | api / metered / unknown | `no-api`: "usage in the Console" |
| anthropic, cloud | `CLAUDE_CODE_USE_BEDROCK`, `CLAUDE_CODE_USE_VERTEX` or `CLAUDE_CODE_USE_FOUNDRY` (environment or settings `env`) | cloud / metered / cloud | `not-visible`: "usage in your cloud billing console" |
| anthropic, gateway | `ANTHROPIC_BASE_URL` with a token or `apiKeyHelper` | gateway / unknown / org | `not-visible`: "usage in your organization's gateway" |
| openai, subscription | ChatGPT login (Codex, opencode, pi); `plan_type` | `free`: free. `go`, `plus`, `pro`, `prolite`, `promax`: individual / flat / self. `team`, `business`, `self_serve_business_*`: team / org; `*_usage_based` → metered. `enterprise*`, `ent*`, `edu*`, `sci*`: enterprise / org; `*_usage_based` → metered. | `wham/usage` 5 h and weekly windows, `credits`, `spend_control` (shown as hidden when present but without numbers) |
| openai, API key | `auth_mode: apikey` or `OPENAI_API_KEY` | api / metered | `no-api` |
| github-copilot | GitHub login (Copilot CLI keychain, opencode, pi); `copilot_plan`, `organization_login_list` | `free`: free. `individual`, `pro`, `pro_plus`, `max`: individual / mixed / self. `business`: team / mixed / org. `enterprise`: enterprise / mixed / org. | Allowance (`premium_interactions` or AI credits, monthly, Scope `seat` or `pool` for organization seats) from `copilot_internal/user`. Spend this month from the billing API on personal plans; a 403 or 404 there shows "managed by org". |
| google | Antigravity (`agy /usage`) | unknown (antigravity), since agy does not expose the tier | Windows per model group from `/usage` |
| any API-key-only provider | opencode or pi `key` credentials | api / metered | `no-api` |

## Command

```text
chezmoi agents usage [-live] [-provider <id>]
```

- By default it shows the cached `usage.json`, with each snapshot's age.
  - Accounts discovered since the last live run but missing from the cache show "no data yet (run with -live)".
  - The exception is offline sources, which are read without network: the newest Codex session snapshot and the agy log.
- `-live` discovers accounts and queries each in parallel: one goroutine per account, 10 s each (agy included). It prints the results and rewrites the cache (0600, atomic). A failed account keeps its cached snapshot, marked stale.
- `-provider` limits the output to one provider id.
- Exit code: 0 whenever it printed, even when some accounts are unavailable; 2 on a flag error.
- Root `Makefile`: `usage: agents ## Show subscription usage per provider account (live)`, which runs `usage -live`.

### Output (Catppuccin via `internal/term`)

```text
chezmoi agents usage — subscription usage per provider account   (cached 12m ago; -live to refresh)

anthropic        Enterprise seat · me@example.com        claude
  5h             ██████░░░░░░░░░░░░░░   31%   resets in 2h 14m
  7d             ███░░░░░░░░░░░░░░░░░   14%   resets Mon 09:00
  spend limit    managed by org
openai           Plus · me@example.com                   codex, opencode, pi
  5h             ███████░░░░░░░░░░░░░   33%   resets in 3h 02m
  7d             █░░░░░░░░░░░░░░░░░░░    5%   resets Thu 18:30
github-copilot   Business seat · octocat                 copilot, opencode, pi
  premium        ██████████████░░░░░░   70%   pool · resets Dec 1
  spend          managed by org
google           unknown (antigravity)                   antigravity
  Gemini · 7d    ████████████████████  100%   resets Sat 22:48
  Claude/GPT·7d  ███████████████░░░░░   77%   resets Sun 13:05
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
| Provider and account registry | usage (this plan) | `internal/provider`, Go code (decision 12), no data file and no `chezmoi init` prompt. Plan tier and account come from the credentials at run time. |
| Usage queries and snapshot | usage | `internal/usage`: `Plan`, `Limit`, `Snapshot`, and the pure `usage.Summaries(snaps, now) []Summary{Provider, Plans, Headroom}` and `usage.Advisories(snaps, now) []Advisory{Provider, Level, Text}` for the doctor |
| Plan and billing judgement | models | The doctor's "better config available" rule reads `Summary`: a provider with no account means no recommendation; flat billing with healthy headroom allows a stronger in-plan model; metered billing, or an organization payer, means the cost-efficient model in the same tier (catalog $/Mtok); headroom under 10% means the same-tier model from a provider with headroom (info only); an unknown plan or a hidden limit means rank only. Sent to the models work on 2026-11-06. |
| Cache dir | usage | `internal/cache.Dir(getenv)`: `$XDG_CACHE_HOME`, otherwise `os.UserCacheDir()`, plus `/chezmoi-agents`. Usage writes `usage.json`. |
| Model catalog | models | `internal/catalog`; it does not import `provider` (the glue is in `cli`) |
| `term` helpers | models: `table.go` (`Width`, `Pad`, `Style.Table`, `Style.Warn`, pushed at 02ad341); usage: `bar.go` (`Bar`) | `term.go` untouched |
| Apply/update warning | models owns the one subcommand `chezmoi agents doctor` and the one script `run_after_agents-doctor.{sh,ps1}.tmpl`, in a later PR | Usage contributes `Advisories` from the cache, or live with `--live` (2 s cap per provider) (decision 11). |
| `cli.go`, `docs/OS.md`, AGENTS files | both, minimal hunks | One `case` and its usage lines each. `docs/OS.md` gets a parent section "Models and subscriptions" (created by models) with a subsection per command. Whoever merges second rebases. |
| `edit.toml` | models only (`[edit.models]`) | Usage has no data file, so it has no entry. |

## Package layout

Following the `tools/agents` layering: pure packages know nothing of processes or the network, and `cli` wires them. The groundwork commit holds the types and signatures.

- `internal/usage` (pure): `usage.go` holds the model (`Plan`, `Limit`, `Snapshot`, `Summary`, `Advisory`). `api.go` holds the signatures, which the units replace:
  - the cache: `MarshalCache`, `ParseCache`;
  - `Summaries`, `Advisories`;
  - the plan mappings: `AnthropicPlan`, `OpenAIPlan`, `CopilotPlan`;
  - one parser per payload: `ParseAnthropic`, `ParseOpenAI`, `ParseCodexSession`, `ParseCopilotUser`, `ParseCopilotBilling`, `ParseAgyUsage`, `ParseAgyLog`.
- `internal/provider` (pure): `Harnesses()` (the fixed facts), `Account`, `Credential`, the `Store` interface and `Discover(Store) []Account`.
  - `Credential` redacts itself in `String`, `GoString`, `MarshalJSON` and `LogValue`.
  - There is one reader per harness: `claude.go`, `codex.go`, `copilot.go`, `antigravity.go`, `opencode.go`, `pi.go`.
  - Accounts are merged by (provider, key), and the newest credential wins.
- `internal/cache`: `Dir(getenv)`.
- `internal/term/bar.go`: `Style.Bar(percent, width)`.
- `internal/term/table.go` comes from the models work (02ad341), carried unchanged so the branches merge cleanly.
- `internal/cli/usage.go`:
  - flags;
  - the real `Store`: files, `go-keyring` (with a macOS fallback for an unknown keychain account), and opencode SQLite opened read-only;
  - fetchers: an HTTP client with timeouts and contexts, base URLs overridable for tests; agy runs through `os/exec` with a context timeout;
  - parallel fetch with `sync.WaitGroup.Go`, rendering, and the cache read and write.

## Tests (offline)

- Every test uses `t.Parallel()`, `t.TempDir()` and `go-cmp`, and runs with `-race -shuffle=on`.
- `internal/provider/testdata/` holds every credential shape, with fake tokens.
- `internal/usage/testdata/` holds every payload, with identifiers replaced:
  - the trimmed responses, and the agy `/usage` JSON as captured above;
  - a Codex session JSONL;
  - an agy log excerpt;
  - team, enterprise and unknown plan variants.
- Table tests cover:
  - plan mapping, including unknown values;
  - merging, expiry, skipped `!command` keys and ignored MCP rows;
  - hidden limits;
  - no panic on malformed payloads;
  - redaction: `fmt.Sprint`, `%#v`, `json.Marshal` and `slog` never contain the fake token.
- CLI tests use `httptest` servers and a fake agy script (Unix only; Windows skips the exec test). Golden output is checked with and without colour.

## Docs, runbook, AGENTS

- `docs/OS.md` gets the subsection "Subscription usage (`chezmoi-agents usage`)" under "Models and subscriptions". If the models PR has not created that parent yet, this PR adds the parent heading, and the second PR to merge rebases. The subsection covers:
  - what is discovered, per harness and OS;
  - the plan and billing model;
  - cache vs `-live`;
  - that the endpoints are undocumented;
  - that credentials stay in memory.
  - Runbook (When / Steps / Verify / Rollback): check capacity before fanning out agents, an expired login, an account showing "managed by org".
- Root `AGENTS.md`: one bullet. `tools/agents/AGENTS.md`: `provider`, `usage` and `cache` in the pure package list, plus the in-memory credential rule.
- Out of scope: the `host-hygiene` skill's quota line (harness repo) could point to `chezmoi agents usage` later.

## Release

- PR 1 off `main`: `feat(agents): add usage subcommand for subscription usage per provider`.
  - It is independent of the models PR. `term/table.go` is carried identically, so the add/add merges cleanly.
  - It deletes this WIP file.
- After it merges, the hub tags `agents-vX.Y.0` (the next minor, coordinated with the models release), then `chore(packages): bump chezmoi-agents` (`chezmoi dots bump chezmoi-agents`).
- The `doctor` hook comes later, from the models work.
- Implementation split:
  1. Groundwork on this branch: the types, signatures, `cache.Dir`, `term.Bar`, and `term/table.go` carried over.
  2. Three parallel workers, each in its own worktree branched from the groundwork:
     - U1, `internal/usage`: parsers, mappings, cache, `Summaries`, `Advisories`, tests.
     - U2, `internal/provider`: readers, `Discover`, tests.
     - U3, `internal/cli/usage.go`: dispatch, the real `Store`, fetchers, rendering, tests, the Makefile target and the docs. It works against the signatures and fake data, and the integration replaces the fakes.
  3. Integrate and run `make check` at the root and in `tools/agents`.

## Open questions

None: all eight are answered (decisions 5-12).
