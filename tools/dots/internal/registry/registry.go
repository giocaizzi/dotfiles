// Package registry is the edit registry (.chezmoidata/edit.toml): the named
// components `chezmoi dots` edits.
package registry

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/finding"
)

// Entry is one editable component.
type Entry struct {
	Name   string
	Desc   string
	Source string   // source-relative file to edit; empty for a local entry
	Local  string   // home-relative folder of unmanaged files; empty for a source entry
	Check  []string // command run with {file} replaced by the edited file
	Note   string   // printed after an apply
	First  bool     // diff: show the first changed target only
}

// Registry is the parsed registry.
type Registry struct {
	entries map[string]Entry
}

// Names returns the entry names, sorted.
func (r Registry) Names() []string {
	names := make([]string, 0, len(r.entries))
	for n := range r.entries {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Get returns the entry called name.
func (r Registry) Get(name string) (Entry, bool) {
	e, ok := r.entries[name]
	return e, ok
}

// BySource returns the entry whose source is the given source-relative path.
func (r Registry) BySource(source string) (Entry, bool) {
	for _, n := range r.Names() {
		if e := r.entries[n]; e.Source != "" && e.Source == source {
			return e, true
		}
	}
	return Entry{}, false
}

// ByLocal returns the local entry whose folder holds the home-relative path.
func (r Registry) ByLocal(rel string) (Entry, bool) {
	for _, n := range r.Names() {
		if e := r.entries[n]; e.Local != "" && strings.HasPrefix(rel, e.Local+"/") {
			return e, true
		}
	}
	return Entry{}, false
}

var localFolder = regexp.MustCompile(`^(/|~|\.\.?/)|/\.\./`)

// Parse reads the registry from chezmoi data (the "edit" table) and validates its
// shape. A registry with findings must not be used.
func Parse(data map[string]any) (Registry, []finding.Finding) {
	var fs []finding.Finding
	reg := Registry{entries: map[string]Entry{}}
	table, _ := data["edit"].(map[string]any)
	for name, raw := range table {
		m, ok := raw.(map[string]any)
		if !ok {
			fs = append(fs, finding.New(name, "not a table"))
			continue
		}
		e, ef := parseEntry(name, m)
		fs = append(fs, ef...)
		reg.entries[name] = e
	}
	return reg, fs
}

func parseEntry(name string, m map[string]any) (Entry, []finding.Finding) {
	var out []finding.Finding
	bad := func(format string, args ...any) { out = append(out, finding.New(name, format, args...)) }
	str := func(k string) string { v, _ := m[k].(string); return v }
	e := Entry{Name: name, Desc: str("desc"), Source: str("source"), Local: str("local"), Note: str("note")}
	for k := range m {
		if !slices.Contains([]string{"desc", "source", "local", "check", "note", "diff"}, k) {
			bad("unknown key %s", k)
		}
	}
	if e.Desc == "" {
		bad("desc is required")
	}
	_, hasLocal := m["local"]
	switch {
	case hasLocal:
		if e.Local == "" || localFolder.MatchString(e.Local) {
			bad("local must be a folder relative to the home dir")
		}
		for _, k := range []string{"source", "diff", "note"} {
			if _, ok := m[k]; ok {
				bad("local excludes source, diff and note")
				break
			}
		}
	case e.Source == "":
		bad("source is required")
	}
	if d, ok := m["diff"]; ok {
		if d != "first" {
			bad(`diff must be "first"`)
		}
		e.First = d == "first"
	}
	if c, ok := m["check"]; ok {
		list, isList := c.([]any)
		if !isList || len(list) == 0 {
			bad("check must be a command as a list: [\"make\", \"lint\"], with {file} for the edited file")
		}
		for _, a := range list {
			s, isStr := a.(string)
			if !isStr {
				bad("check must be a list of strings")
				break
			}
			e.Check = append(e.Check, s)
		}
	}
	return e, out
}

// Validate checks the registry against the source tree: every source exists, and
// every .chezmoidata file and every encrypted file has an entry.
func (r Registry) Validate(src fs.FS) []finding.Finding {
	var out []finding.Finding
	covered := map[string]bool{}
	for _, n := range r.Names() {
		e := r.entries[n]
		if e.Source == "" {
			continue
		}
		covered[e.Source] = true
		if _, err := fs.Stat(src, e.Source); err != nil {
			out = append(out, finding.New(n, "source %s does not exist", e.Source))
		}
	}
	_ = fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil //nolint:nilerr // an unreadable entry is not a registry problem
		}
		if (strings.HasPrefix(p, ".chezmoidata/") || path.Ext(p) == ".age") && !covered[p] {
			out = append(out, finding.New(p, "no edit registry entry"))
		}
		return nil
	})
	return out
}

// Describe formats the entry's one-line description.
func (e Entry) Describe() string { return fmt.Sprintf("%s  %s", e.Name, e.Desc) }
