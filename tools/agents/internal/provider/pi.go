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
	// GitHub Copilot enterprise logins store the domain as enterpriseUrl.
	// Source: https://github.com/earendil-works/pi/blob/HEAD/packages/ai/src/auth/oauth/github-copilot.ts
	// (formerly badlogic/pi-mono). enterpriseDomain is read too, leniently.
	EnterpriseURL    string `json:"enterpriseUrl"`
	EnterpriseDomain string `json:"enterpriseDomain"`
}

func sortedKeys[V any](m map[string]V) []string {
	return slices.SortedFunc(maps.Keys(m), cmp.Compare[string])
}

// readPi reads pi's auth.json, one entry per provider. A key starting with "!"
// is a command that is never run.
func readPi(st Store) []Account {
	dir := st.Getenv("PI_CODING_AGENT_DIR")
	if dir == "" {
		dir = filepath.Join(st.Home(), ".pi", "agent")
	}
	var entries map[string]piEntry
	found, err := readJSON(st, filepath.Join(dir, "auth.json"), &entries)
	if !found {
		return nil
	}
	if err != nil {
		return []Account{unreadable("pi", "pi")}
	}
	var out []Account
	for _, id := range sortedKeys(entries) {
		e := entries[id]
		out = append(out, tenantEntry{
			harness:  "pi",
			provider: id,
			typ:      normalizeTenantType(e.Type),
			access:   e.Access,
			refresh:  e.Refresh,
			key:      e.Key,
			expires:  msTime(e.Expires),
			host:     cmp.Or(e.EnterpriseURL, e.EnterpriseDomain),
			source:   srcFile,
			hashKey:  true,
		}.account(st.Now()))
	}
	return out
}
