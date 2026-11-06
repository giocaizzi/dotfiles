package provider

import (
	"cmp"
	"maps"
	"path/filepath"
	"slices"
)

type piEntry struct {
	Type    string  `json:"type"` // oauth | api_key
	Access  string  `json:"access"`
	Refresh string  `json:"refresh"`
	Expires float64 `json:"expires"`
	Key     string  `json:"key"`
	// ChatGPT subscription logins ("openai-codex") store the account id.
	AccountID string `json:"accountId"`
	// GitHub Copilot enterprise logins store the domain as enterpriseUrl.
	// Source: https://github.com/earendil-works/pi/blob/HEAD/packages/ai/src/auth/oauth/github-copilot.ts
	// (formerly badlogic/pi-mono). enterpriseDomain is read too, leniently.
	EnterpriseURL    string `json:"enterpriseUrl"`
	EnterpriseDomain string `json:"enterpriseDomain"`
}

// piCodexKey is the store key of pi's ChatGPT subscription login.
// Source: https://github.com/earendil-works/pi/blob/HEAD/packages/ai/src/providers/openai-codex.ts
const piCodexKey = "openai-codex"

func sortedKeys[V any](m map[string]V) []string {
	return slices.SortedFunc(maps.Keys(m), cmp.Compare[string])
}

// readPi reads the auth.json of every pi profile, one entry per provider. A
// key starting with "!" is a command that is never run.
func readPi(st Store) []Account {
	profiles := piSpec().profiles(st)
	per := make([][]Account, 0, len(profiles))
	for _, p := range profiles {
		per = append(per, readPiProfile(st, p))
	}
	return joinProfiles("pi", per)
}

func readPiProfile(st Store, p profile) []Account {
	var entries map[string]piEntry
	found, err := readJSON(st, filepath.Join(p.dir, "auth.json"), &entries)
	if !found {
		return nil
	}
	if err != nil {
		return unreadableProfile(p, "pi", "")
	}
	var out []Account
	for _, id := range sortedKeys(entries) {
		e := entries[id]
		prov := id
		if id == piCodexKey {
			prov = OpenAI
		}
		out = append(out, tenantEntry{
			harness:   p.label(),
			provider:  prov,
			storeKey:  id,
			accountID: e.AccountID,
			typ:       normalizeTenantType(e.Type),
			access:    e.Access,
			refresh:   e.Refresh,
			key:       e.Key,
			expires:   msTime(e.Expires),
			host:      cmp.Or(e.EnterpriseURL, e.EnterpriseDomain),
			source:    srcFile,
		}.account(st.Now()))
	}
	return out
}
