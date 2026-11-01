// Package tree finds what a source file is made of: the templates it includes, the
// files it includes, the symlink target it points at and the data files it reads.
// The edges are read from the template text with patterns, which is exact for how
// this repository writes templates; a branch is listed whichever role or OS takes it.
package tree

import (
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"
)

// Node is one file of a composition tree.
type Node struct {
	Depth int
	How   string // source, template, include, symlink or data
	Path  string // source-relative
}

const maxDepth = 8

var (
	symlinkRe  = regexp.MustCompile(`\.chezmoi\.sourceDir\s*\}\}/[^\s{}]+`)
	templateRe = regexp.MustCompile(`(?:template|includeTemplate)\s+"([^"]+)"`)
	includeRe  = regexp.MustCompile(`include\s+"([^"]+)"`)
	printfRe   = regexp.MustCompile(`include\s+\(printf\s+"([^"]+)"`)
	tableRe    = regexp.MustCompile(`^\[\[?([A-Za-z0-9_]+)`)
	keyRe      = regexp.MustCompile(`^([A-Za-z0-9_]+)\s*=`)
)

// Tree reads composition trees from a source dir.
type Tree struct {
	src  fs.FS
	data map[string][]string // .chezmoidata file to its top-level keys
}

// New indexes the data files of src.
func New(src fs.FS) *Tree {
	t := &Tree{src: src, data: map[string][]string{}}
	files, _ := fs.Glob(src, ".chezmoidata/*.toml")
	for _, f := range files {
		if strings.HasPrefix(path.Base(f), ".") {
			continue
		}
		if b, err := fs.ReadFile(src, f); err == nil {
			t.data[f] = topKeys(string(b))
		}
	}
	return t
}

func topKeys(toml string) []string {
	var keys []string
	inTable := false
	for line := range strings.SplitSeq(toml, "\n") {
		if m := tableRe.FindStringSubmatch(line); m != nil {
			inTable = true
			keys = append(keys, m[1])
		} else if m := keyRe.FindStringSubmatch(line); m != nil && !inTable {
			keys = append(keys, m[1])
		}
	}
	slices.Sort(keys)
	return slices.Compact(keys)
}

// Of lists the nodes of file's tree, depth first, file itself first.
func (t *Tree) Of(file string) []Node {
	out := []Node{{Depth: 0, How: "source", Path: file}}
	return t.walk(out, file, 0, nil)
}

func (t *Tree) walk(out []Node, file string, depth int, ancestors []string) []Node {
	if strings.HasSuffix(file, ".age") || depth >= maxDepth {
		return out
	}
	b, err := fs.ReadFile(t.src, file)
	if err != nil {
		return out
	}
	text := string(b)
	seen := append(slices.Clone(ancestors), file)
	for _, n := range t.edges(file, text) {
		if slices.Contains(seen, n.Path) {
			continue
		}
		n.Depth = depth + 1
		out = append(out, n)
		out = t.walk(out, n.Path, depth+1, seen)
	}
	return out
}

func (t *Tree) edges(file, text string) []Node {
	var nodes []Node
	add := func(how, p string) {
		if !slices.Contains(nodes, Node{How: how, Path: p}) {
			nodes = append(nodes, Node{How: how, Path: p})
		}
	}
	if strings.HasPrefix(path.Base(file), "symlink_") {
		for _, m := range symlinkRe.FindAllString(text, -1) {
			add("symlink", m[strings.Index(m, "}}/")+3:])
		}
	}
	for _, m := range templateRe.FindAllStringSubmatch(text, -1) {
		add("template", ".chezmoitemplates/"+m[1])
	}
	for _, m := range includeRe.FindAllStringSubmatch(text, -1) {
		add("include", m[1])
	}
	for _, m := range printfRe.FindAllStringSubmatch(text, -1) {
		matches, _ := fs.Glob(t.src, strings.ReplaceAll(m[1], "%s", "*"))
		for _, p := range matches {
			add("include", p)
		}
	}
	if !strings.HasPrefix(file, ".chezmoidata/") {
		for _, f := range slices.Sorted(mapsKeys(t.data)) {
			for _, k := range t.data[f] {
				if regexp.MustCompile(`([\s($]\.` + k + `\b|get\s+\S+\s+"` + k + `")`).MatchString(text) {
					add("data", f)
				}
			}
		}
	}
	return nodes
}

func mapsKeys(m map[string][]string) func(func(string) bool) {
	return func(yield func(string) bool) {
		for k := range m {
			if !yield(k) {
				return
			}
		}
	}
}

// Invert builds a reverse index from each source-tree dependency to the sorted,
// unique targets that depend on it. managed maps each target to its source path.
func (t *Tree) Invert(managed map[string]string) map[string][]string {
	inv := make(map[string][]string)
	for target, source := range managed {
		for _, n := range t.Of(source) {
			inv[n.Path] = append(inv[n.Path], target)
		}
	}
	for p, targets := range inv {
		slices.Sort(targets)
		inv[p] = slices.Compact(targets)
	}
	return inv
}

// Feeds lists, sorted, the targets whose tree holds file; managed maps each target
// to its source path.
func (t *Tree) Feeds(managed map[string]string, file string) []string {
	return t.Invert(managed)[file]
}
