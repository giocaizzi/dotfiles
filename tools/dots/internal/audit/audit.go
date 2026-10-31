// Package audit finds the aliases live in an interactive shell that the alias
// registry does not account for.
package audit

import (
	"regexp"
	"slices"
	"strings"
)

var (
	renderedAlias = regexp.MustCompile(`(?m)^alias ([^=\s]+)=`)
	omzContext    = regexp.MustCompile(`(?m)^zstyle ':omz:plugins:([^']+)' aliases yes$`)
	omzAlias      = regexp.MustCompile(`(?m)^\s*alias (?:-g )?'?([^'=\s]*)'?=`)
)

// builtin names the aliases the shells and the rendered file define at runtime:
// ls and python are detected when the file is sourced, the others ship with zsh.
func builtin() []string { return []string{"ls", "python", "run-help", "which-command"} }

// Plugins lists the Oh My Zsh plugin contexts switched on in the rendered
// omz-aliases.zsh.
func Plugins(omzAliases string) []string {
	matches := omzContext.FindAllStringSubmatch(omzAliases, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m[1])
	}
	return out
}

// OMZAliases lists the alias names a plugin's .zsh files define.
func OMZAliases(text string) []string {
	var out []string
	for _, m := range omzAlias.FindAllStringSubmatch(text, -1) {
		if m[1] != "" {
			out = append(out, m[1])
		}
	}
	return out
}

// Known is every alias the registry accounts for: those rendered into aliases.sh,
// the runtime ones and the aliases of the Oh My Zsh plugins turned on.
func Known(aliasesSh string, omzPluginAliases []string) map[string]bool {
	known := map[string]bool{}
	for _, m := range renderedAlias.FindAllStringSubmatch(aliasesSh, -1) {
		known[m[1]] = true
	}
	for _, n := range slices.Concat(builtin(), omzPluginAliases) {
		known[n] = true
	}
	return known
}

// Unknown returns, sorted and deduplicated, the live names that are not known.
func Unknown(known map[string]bool, live string) []string {
	var out []string
	for name := range strings.FieldsSeq(live) {
		if !known[name] {
			out = append(out, name)
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}
