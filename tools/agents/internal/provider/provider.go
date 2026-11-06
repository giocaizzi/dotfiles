// Package provider knows which provider accounts the AI harnesses are logged
// into. It reads the credentials each harness already stores, through an
// injected Store, into memory only: a Credential never prints, marshals or logs
// its secret. It is pure: no network, no processes.
package provider

import (
	"log/slog"
	"time"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/usage"
)

// Provider ids are models.dev provider keys, shared with the models catalog.
const (
	Anthropic     = "anthropic"
	OpenAI        = "openai"
	GitHubCopilot = "github-copilot"
	Google        = "google"
)

// Harness is a fixed fact about one harness: the providers it can log into,
// and whether it holds several provider accounts at once.
type Harness struct {
	Name        string
	Providers   []string // fixed providers; empty for multi-tenant harnesses
	MultiTenant bool
}

// Harnesses lists the harnesses whose credentials are discovered.
func Harnesses() []Harness {
	return []Harness{
		{Name: "claude", Providers: []string{Anthropic}},
		{Name: "codex", Providers: []string{OpenAI}},
		{Name: "copilot", Providers: []string{GitHubCopilot}},
		{Name: "antigravity", Providers: []string{Google}},
		{Name: "opencode", MultiTenant: true},
		{Name: "pi", MultiTenant: true},
	}
}

// Auth is how an account authenticates.
type Auth string

// Auth kinds.
const (
	AuthOAuth   Auth = "oauth"   // subscription login; Credential holds the access token
	AuthAPIKey  Auth = "api-key" // provider API key
	AuthCloud   Auth = "cloud"   // a cloud platform account (Bedrock, Vertex, Foundry)
	AuthGateway Auth = "gateway" // an LLM gateway (custom base URL, key helper)
	AuthCLI     Auth = "cli"     // the harness CLI itself answers (agy /usage); no credential read
	AuthCommand Auth = "command" // a key produced by a command; never run, skipped
	AuthNone    Auth = "none"    // harness present, not logged in
)

// Credential is a secret held in memory. Its String, GoString, MarshalJSON and
// LogValue never reveal it.
type Credential struct {
	token     string
	accountID string // e.g. ChatGPT-Account-Id; not secret, but never displayed
}

// NewCredential wraps a token and an optional provider account id.
func NewCredential(token, accountID string) Credential {
	return Credential{token: token, accountID: accountID}
}

// Token returns the secret for an Authorization header. Callers never print it.
func (c Credential) Token() string { return c.token }

// AccountID returns the provider account id sent as a request header.
func (c Credential) AccountID() string { return c.accountID }

// Empty reports whether no token is held.
func (c Credential) Empty() bool { return c.token == "" }

const redacted = "[redacted]"

// String redacts the credential.
func (c Credential) String() string { return redacted }

// GoString redacts the credential for %#v.
func (c Credential) GoString() string { return redacted }

// MarshalJSON redacts the credential.
func (c Credential) MarshalJSON() ([]byte, error) { return []byte(`"` + redacted + `"`), nil }

// LogValue redacts the credential in slog.
func (c Credential) LogValue() slog.Value { return slog.StringValue(redacted) }

// Account is one provider account found on this machine.
type Account struct {
	Provider  string       // models.dev provider id
	Label     string       // email or login for display; never a token or an id
	Key       string       // merge key: provider account id, login or email; never displayed
	Harnesses []string     // harnesses logged into this account, sorted
	Auth      Auth         // how it authenticates
	Plan      usage.Plan   // from the credential when it says (subscriptionType, plan claims), else Kind unknown
	Expires   time.Time    // access token expiry; zero when unknown
	Status    usage.Status // set when the account cannot be queried: expired, not-logged-in, skipped, no-api, not-visible
	Note      string       // e.g. "usage lives in your cloud console"
	Cred      Credential   // the secret, in memory only
	Login     string       // GitHub login for Copilot billing; empty otherwise
	Source    string       // where it was read: "keychain", "file", "sqlite", "env"
	Refreshed time.Time    // when the harness last wrote it; the newest wins on merge
}

// Store is everything Discover reads from the machine. The real one lives in
// cli; tests use a fake.
type Store interface {
	// ReadFile reads a file; a missing file returns an error matching fs.ErrNotExist.
	ReadFile(path string) ([]byte, error)
	// Secret reads one item from the OS credential store (go-keyring). An empty
	// account means "any account" for that service.
	Secret(service, account string) (string, error)
	// OpencodeCredentials reads the rows of opencode's V2 credential table,
	// read-only. A missing database returns an error matching fs.ErrNotExist.
	OpencodeCredentials(dbPath string) ([]OpencodeRow, error)
	// Exists reports whether a path exists.
	Exists(path string) bool
	Getenv(key string) string
	Home() string
	GOOS() string
	Now() time.Time
}

// OpencodeRow is one row of opencode's `credential` table.
type OpencodeRow struct {
	ID            string
	IntegrationID string
	Label         string
	Active        bool
	Value         []byte // JSON {type, access, refresh, expires, key, metadata}
}

// Discover returns every account found, merged by (provider, key) with the
// harnesses that use it, sorted by provider then label.
func Discover(st Store) []Account { return nil }
