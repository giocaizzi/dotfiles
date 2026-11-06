package usage

import (
	"errors"
	"io"
	"time"
)

// ErrNotImplemented marks the groundwork stubs; the unit that owns a function
// replaces it.
var ErrNotImplemented = errors.New("not implemented")

// Cache is the content of usage.json in the chezmoi-agents cache dir. It never
// holds a credential.
type Cache struct {
	Schema    int        `json:"schema"`
	WrittenAt time.Time  `json:"writtenAt"`
	Snapshots []Snapshot `json:"snapshots"`
}

// MarshalCache encodes snapshots deterministically (2-space indent).
func MarshalCache(snaps []Snapshot, now time.Time) ([]byte, error) {
	return nil, ErrNotImplemented
}

// ParseCache decodes usage.json; unknown fields are ignored.
func ParseCache(b []byte) (Cache, error) { return Cache{}, ErrNotImplemented }

// Summaries reduces snapshots to one Summary per provider, sorted by provider.
func Summaries(snaps []Snapshot, now time.Time) []Summary { return nil }

// Advisories lists capacity facts: a visible, enabled limit at 90% or more whose
// reset is in the future (warn), an exhausted allowance with overage on (info).
// Unknown plans and hidden limits give none.
func Advisories(snaps []Snapshot, now time.Time) []Advisory { return nil }

// AnthropicPlan maps Claude Code's claudeAiOauth.subscriptionType and
// rateLimitTier to a Plan.
func AnthropicPlan(subscriptionType, rateLimitTier string) Plan { return Plan{} }

// ParseAnthropic parses GET api.anthropic.com/api/oauth/usage.
func ParseAnthropic(body []byte, plan Plan) ([]Limit, error) { return nil, ErrNotImplemented }

// OpenAIPlan maps a ChatGPT plan_type to a Plan.
func OpenAIPlan(planType string) Plan { return Plan{} }

// ParseOpenAI parses GET chatgpt.com/backend-api/wham/usage.
func ParseOpenAI(body []byte) (Plan, []Limit, error) { return Plan{}, nil, ErrNotImplemented }

// ParseCodexSession reads a Codex rollout JSONL and returns the plan and limits
// of its last token_count rate_limits snapshot, and the time of that line.
func ParseCodexSession(r io.Reader) (Plan, []Limit, time.Time, error) {
	return Plan{}, nil, time.Time{}, ErrNotImplemented
}

// CopilotPlan maps copilot_plan (and whether the seat comes from an
// organization) to a Plan.
func CopilotPlan(copilotPlan string, orgSeat bool) Plan { return Plan{} }

// ParseCopilotUser parses GET api.github.com/copilot_internal/user.
func ParseCopilotUser(body []byte) (Plan, []Limit, error) { return Plan{}, nil, ErrNotImplemented }

// ParseCopilotBilling parses the user premium request / AI credit usage report
// of the GitHub billing REST API into one spend limit for the current month.
func ParseCopilotBilling(body []byte) (Limit, error) { return Limit{}, ErrNotImplemented }

// ParseAgyUsage parses `agy -p /usage --output-format json`. It fails unless
// command.name is "usage" and num_turns is 0.
func ParseAgyUsage(body []byte) ([]Limit, error) { return nil, ErrNotImplemented }

// ParseAgyLog scans an Antigravity CLI log for the last
// "RESOURCE_EXHAUSTED (code 429): ... Resets in 33h5m2s." line and returns an
// exhausted window resetting then. logStart is the time in the log file name
// (cli-YYYYMMDD_HHMMSS.log); log lines carry month, day and time but no year.
func ParseAgyLog(r io.Reader, logStart time.Time) ([]Limit, time.Time, error) {
	return nil, time.Time{}, ErrNotImplemented
}
