package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/usage"
)

const claudeKeychainService = "Claude Code-credentials"

type claudeSettings struct {
	Env          map[string]any `json:"env"`
	APIKeyHelper string         `json:"apiKeyHelper"`
}

type claudeCreds struct {
	ClaudeAiOauth struct {
		AccessToken      string  `json:"accessToken"`
		ExpiresAt        float64 `json:"expiresAt"`
		SubscriptionType string  `json:"subscriptionType"`
		RateLimitTier    string  `json:"rateLimitTier"`
	} `json:"claudeAiOauth"`
}

func truthy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// claudeOverride returns the account of a cloud, gateway or API setup, if the
// environment or settings.json holds one. It overrides the OAuth login, as in
// Claude Code itself.
func claudeOverride(st Store, settings claudeSettings) (Account, bool) {
	env := func(k string) string {
		if v := st.Getenv(k); v != "" {
			return v
		}
		s, _ := settings.Env[k].(string)
		return s
	}
	a := Account{Provider: Anthropic, Harnesses: []string{harnessClaude}, Source: srcEnv, Status: usage.StatusNotVisible}
	for _, c := range []struct{ env, name string }{
		{"CLAUDE_CODE_USE_BEDROCK", "Bedrock"},
		{"CLAUDE_CODE_USE_VERTEX", "Vertex AI"},
		{"CLAUDE_CODE_USE_FOUNDRY", "Foundry"},
	} {
		if truthy(env(c.env)) {
			a.Key = "claude-cloud"
			a.Auth = AuthCloud
			a.Plan = usage.Plan{Kind: usage.KindCloud, Billing: usage.BillingMetered, Payer: usage.PayerCloud, Name: c.name}
			a.Note = "usage in your cloud billing console"
			return a, true
		}
	}

	hasToken := env("ANTHROPIC_AUTH_TOKEN") != ""
	if env("ANTHROPIC_BASE_URL") != "" && (hasToken || settings.APIKeyHelper != "") {
		a.Key = "claude-gateway"
		a.Auth = AuthGateway
		a.Plan = usage.Plan{Kind: usage.KindGateway, Billing: usage.BillingUnknown, Payer: usage.PayerOrg}
		a.Note = "usage in your organization's gateway"
		return a, true
	}
	if env("ANTHROPIC_API_KEY") != "" || hasToken || settings.APIKeyHelper != "" {
		a.Key = "claude-api-key"
		a.Auth = AuthAPIKey
		a.Plan = usage.Plan{Kind: usage.KindAPI, Billing: usage.BillingMetered, Payer: usage.PayerUnknown}
		a.Status = usage.StatusNoAPI
		a.Note = "usage in the Console"
		return a, true
	}
	return Account{}, false
}

// claudeSecret reads the OAuth credentials JSON: the keychain on darwin, else
// <dir>/.credentials.json. A missing source returns an error matching fs.ErrNotExist.
func claudeSecret(st Store, dir, cfgDir string) (raw []byte, source string, err error) {
	if st.GOOS() != "darwin" {
		b, rerr := st.ReadFile(filepath.Join(dir, ".credentials.json"))
		if rerr != nil {
			rerr = fmt.Errorf("read credentials file: %w", rerr)
		}
		return b, srcFile, rerr
	}
	service := claudeKeychainService
	if cfgDir != "" {
		sum := sha256.Sum256([]byte(cfgDir))
		service += "-" + hex.EncodeToString(sum[:])[:8]
	}
	s, err := st.Secret(service, "")
	if err == nil && s == "" {
		err = fs.ErrNotExist
	}
	return []byte(s), srcKeychain, err
}

// readClaude finds the Claude Code account.
func readClaude(st Store) []Account {
	cfgDir := st.Getenv("CLAUDE_CONFIG_DIR")
	dir := cfgDir
	if dir == "" {
		dir = filepath.Join(st.Home(), ".claude")
	}

	var settings claudeSettings
	_, _ = readJSON(st, filepath.Join(dir, "settings.json"), &settings) // malformed settings count as absent
	if a, ok := claudeOverride(st, settings); ok {
		return []Account{a}
	}

	oauth := Account{Provider: Anthropic, Key: "claude-oauth", Harnesses: []string{harnessClaude}, Auth: AuthOAuth}
	raw, source, err := claudeSecret(st, dir, cfgDir)
	oauth.Source = source
	if err != nil && (isNotExist(err) || source == srcKeychain) {
		return nil // no credentials; a keychain error (locked, denied) is the same to us
	}
	var c claudeCreds
	if err != nil || unmarshalQuiet(raw, &c) != nil {
		oauth.Plan = unknownPlan()
		oauth.Status = usage.StatusUnavailable
		oauth.Note = noteUnreadable
		return []Account{oauth}
	}
	o := c.ClaudeAiOauth
	if o.AccessToken == "" {
		return nil
	}
	oauth.Cred = NewCredential(o.AccessToken, "")
	oauth.Plan = usage.AnthropicPlan(o.SubscriptionType, o.RateLimitTier)
	oauth.Expires = msTime(o.ExpiresAt)
	expiredStatus(&oauth, st.Now(), harnessClaude)
	return []Account{oauth}
}
