package provider

import (
	"cmp"
	"path/filepath"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/usage"
)

const codexKeyringService = "Codex Auth"

type codexAuth struct {
	AuthMode string `json:"auth_mode"`
	APIKey   string `json:"OPENAI_API_KEY"`
	Tokens   struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
		AccountID   string `json:"account_id"`
	} `json:"tokens"`
}

func openaiAPIKeyAccount(harness string) Account {
	return Account{
		Provider:  OpenAI,
		ID:        harness,
		Harnesses: []string{harness},
		Auth:      AuthAPIKey,
		Plan:      usage.Plan{Kind: usage.KindAPI, Billing: usage.BillingMetered, Payer: usage.PayerUnknown},
		Status:    usage.StatusNoAPI,
		Note:      "no usage API for API keys",
	}
}

// openaiOAuth builds a ChatGPT-login account from an access token, an optional
// id_token and a known account id. The id_token is decoded and dropped.
func openaiOAuth(harness, access, idToken, accountID string) Account {
	a := Account{
		Provider:  OpenAI,
		ID:        harness,
		Harnesses: []string{harness},
		Auth:      AuthOAuth,
		Plan:      unknownPlan(),
	}
	accessClaims := jwtClaims(access)
	idClaims := jwtClaims(idToken)
	accessAuth, idAuth := openaiAuth(accessClaims), openaiAuth(idClaims)
	accountID = cmp.Or(accountID, claimString(accessAuth, "chatgpt_account_id"), claimString(idAuth, "chatgpt_account_id"))
	userID := cmp.Or(
		claimString(accessAuth, "chatgpt_user_id"), claimString(accessAuth, "user_id"),
		claimString(idAuth, "chatgpt_user_id"), claimString(idAuth, "user_id"),
	)
	a.Label = cmp.Or(
		claimString(idClaims, "email"), claimString(openaiProfile(idClaims), "email"),
		claimString(accessClaims, "email"), claimString(openaiProfile(accessClaims), "email"),
	)
	if plan := cmp.Or(claimString(idAuth, "chatgpt_plan_type"), claimString(accessAuth, "chatgpt_plan_type")); plan != "" {
		a.Plan = usage.OpenAIPlan(plan)
	}
	if accountID != "" {
		proven(&a, OpenAIKey(accountID, userID))
	} else {
		unproven(&a, access)
	}
	a.Expires = claimExp(accessClaims)
	a.Cred = NewCredential(access, accountID)
	return a
}

// readCodex finds the Codex login: auth.json, else the OS credential store
// (cli_auth_credentials_store = keyring), else OPENAI_API_KEY.
func readCodex(st Store) []Account {
	dir := st.Getenv("CODEX_HOME")
	if dir == "" {
		dir = filepath.Join(st.Home(), ".codex")
	}

	var raw []byte
	source := srcFile
	b, err := st.ReadFile(filepath.Join(dir, "auth.json"))
	switch {
	case err == nil:
		raw = b
	case isNotExist(err):
		if s, serr := st.Secret(codexKeyringService, ""); serr == nil && s != "" {
			raw, source = []byte(s), srcKeychain
		}
	default:
		a := unreadable(OpenAI, "codex")
		a.Source = source
		return []Account{a}
	}

	if raw == nil {
		if st.Getenv("OPENAI_API_KEY") != "" {
			a := openaiAPIKeyAccount("codex")
			a.Source = srcEnv
			return []Account{a}
		}
		return nil
	}

	var c codexAuth
	if err := unmarshalQuiet(raw, &c); err != nil {
		a := unreadable(OpenAI, "codex")
		a.Source = source
		return []Account{a}
	}

	if c.AuthMode == "chatgpt" || (c.AuthMode == "" && c.Tokens.AccessToken != "") {
		if c.Tokens.AccessToken == "" {
			return nil
		}
		a := openaiOAuth("codex", c.Tokens.AccessToken, c.Tokens.IDToken, c.Tokens.AccountID)
		a.Source = source
		expiredStatus(&a, st.Now(), "codex")
		return []Account{a}
	}
	if c.AuthMode == "apikey" || c.APIKey != "" || st.Getenv("OPENAI_API_KEY") != "" {
		a := openaiAPIKeyAccount("codex")
		a.Source = source
		return []Account{a}
	}
	return nil
}
