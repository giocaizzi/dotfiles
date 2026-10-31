// Package aliasreg validates the alias registry (.chezmoidata/aliases.toml and the
// encrypted private registry, which share a schema).
package aliasreg

import (
	"regexp"
	"slices"
	"sort"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/finding"
)

var name = regexp.MustCompile(`^[A-Za-z0-9._][A-Za-z0-9._-]*$`)

func shells() []string { return []string{"posix", "powershell"} }
func oses() []string   { return []string{"darwin", "linux", "windows"} }
func roles() []string  { return []string{"personal", "work", "light"} }
func allowed() []string {
	return []string{"cmd", "posix", "powershell", "shells", "os", "machines", "why"}
}

// Validate checks the alias, aliasBlock and aliasExternal tables of data.
func Validate(data map[string]any) []finding.Finding {
	var out []finding.Finding
	aliases, _ := data["alias"].(map[string]any)
	for _, n := range sortedKeys(aliases) {
		a, _ := aliases[n].(map[string]any)
		out = append(out, validateAlias(n, a)...)
	}
	blocks, _ := data["aliasBlock"].(map[string]any)
	for _, n := range sortedKeys(blocks) {
		b, _ := blocks[n].(map[string]any)
		if _, isStr := b["message"].(string); !isStr || len(b) != 1 {
			out = append(out, finding.New("block "+n, "needs exactly a message string"))
		}
	}
	external, _ := data["aliasExternal"].(map[string]any)
	for _, sh := range sortedKeys(external) {
		set, _ := external[sh].(map[string]any)
		for _, k := range sortedKeys(set) {
			if _, isBool := set[k].(bool); !isBool {
				out = append(out, finding.New("external "+sh+"."+k, "must be true or false"))
			}
		}
	}
	return out
}

func validateAlias(n string, a map[string]any) []finding.Finding {
	var out []finding.Finding
	bad := func(format string, args ...any) { out = append(out, finding.New("alias "+n, format, args...)) }
	if !name.MatchString(n) {
		bad("invalid name")
	}
	for _, k := range sortedKeys(a) {
		if !slices.Contains(allowed(), k) {
			bad("unknown key %s", k)
		}
	}
	_, hasCmd := a["cmd"]
	_, hasPosix := a["posix"]
	_, hasPS := a["powershell"]
	if hasCmd && (hasPosix || hasPS) {
		bad("use cmd or posix/powershell, not both")
	}
	on := enabledShells(a, bad)
	for _, s := range on {
		if _, has := a[s]; !hasCmd && !has {
			bad("no %s command (add one, or limit shells with why)", s)
		}
	}
	for _, s := range shells() {
		if _, has := a[s]; has && !slices.Contains(on, s) {
			bad("%s command set but %s not in shells", s, s)
		}
	}
	checkSubset(a, "os", oses(), bad)
	checkSubset(a, "machines", roles(), bad)
	return out
}

// enabledShells returns the shells an alias targets, reporting a bad "shells" list.
func enabledShells(a map[string]any, bad func(string, ...any)) []string {
	listed, hasShells := a["shells"]
	if !hasShells {
		return shells()
	}
	got, ok := stringList(listed)
	if !ok || len(got) == 0 || !subset(got, shells()) {
		bad("shells must be a non-empty subset of %v", shells())
		return nil
	}
	if len(got) < 2 {
		if _, why := a["why"]; !why {
			bad("limited to %v without why", got)
		}
	}
	return got
}

// checkSubset reports a key whose value is not a list within the allowed set.
func checkSubset(a map[string]any, key string, allowedSet []string, bad func(string, ...any)) {
	v, ok := a[key]
	if !ok {
		return
	}
	if got, isList := stringList(v); !isList || !subset(got, allowedSet) {
		bad("%s must be a subset of %v", key, allowedSet)
	}
}

func stringList(v any) ([]string, bool) {
	list, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(list))
	for _, x := range list {
		s, isStr := x.(string)
		if !isStr {
			return nil, false
		}
		out = append(out, s)
	}
	return out, true
}

func subset(got, allowed []string) bool {
	for _, g := range got {
		if !slices.Contains(allowed, g) {
			return false
		}
	}
	return true
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
