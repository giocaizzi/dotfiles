package inventory

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/pkgreg"
)

type rawSelection struct {
	OS       string                    `json:"os"`
	Role     string                    `json:"role"`
	Release  map[string]map[string]any `json:"release"`
	Releases map[string]struct {
		Dest string `json:"dest"`
	} `json:"releases"`
	Brew   []rawPackage `json:"brew"`
	Cask   []rawPackage `json:"cask"`
	Apt    []rawPackage `json:"apt"`
	Winget []rawPackage `json:"winget"`
	UV     []rawPackage `json:"uv"`
	Npm    []rawPackage `json:"npm"`
	Script []rawPackage `json:"script"`
}

type rawPackage struct {
	Name  string         `json:"name"`
	ID    string         `json:"id"`
	Entry map[string]any `json:"entry"`
}

// ParseSelection decodes SelectionTemplate's output, keeping only the package
// managers that apply on its OS.
func ParseSelection(b []byte) (Selection, error) {
	var raw rawSelection
	if err := json.Unmarshal(b, &raw); err != nil {
		return Selection{}, fmt.Errorf("selection: %w", err)
	}
	if raw.OS == "" {
		return Selection{}, errors.New("selection: no os")
	}
	sel := Selection{OS: raw.OS, Role: raw.Role, Packages: map[Source][]PackageSel{}}

	table := make(map[string]any, len(raw.Release))
	for k, v := range raw.Release {
		table[k] = map[string]any(v)
	}
	byName := map[string]pkgreg.Release{}
	for _, r := range pkgreg.Releases(map[string]any{"release": table}) {
		byName[r.Name] = r
	}
	names := make([]string, 0, len(raw.Releases))
	for n := range raw.Releases {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		rel, ok := byName[n]
		if !ok {
			rel = pkgreg.Release{Name: n}
		}
		cmd := stringList(raw.Release[n]["versionCmd"])
		if len(cmd) == 0 {
			cmd = []string{"{dest}", "--version"}
		}
		sel.Releases = append(sel.Releases, ReleaseSel{Release: rel, Dest: raw.Releases[n].Dest, VersionCmd: cmd})
	}

	var managerKeys []string
	switch raw.OS {
	case "darwin":
		managerKeys = []string{"brew", "cask"}
		sel.Packages[Brew] = sortedPackages(slices.Concat(packages(raw.Brew, false), packages(raw.Cask, true)))
	case "linux":
		managerKeys = []string{"apt"}
		sel.Packages[Apt] = sortedPackages(packages(raw.Apt, false))
	case "windows":
		managerKeys = []string{"winget"}
		sel.Packages[Winget] = sortedPackages(packages(raw.Winget, false))
	}
	sel.Packages[UV] = sortedPackages(packages(raw.UV, false))
	sel.Packages[Npm] = sortedPackages(packages(raw.Npm, false))
	scripts := make([]rawPackage, 0, len(raw.Script))
	for _, p := range raw.Script {
		if !slices.ContainsFunc(managerKeys, func(k string) bool { _, ok := p.Entry[k]; return ok }) {
			scripts = append(scripts, p)
		}
	}
	sel.Packages[Script] = sortedPackages(packages(scripts, false))
	return sel, nil
}

func packages(in []rawPackage, cask bool) []PackageSel {
	out := make([]PackageSel, 0, len(in))
	for _, p := range in {
		ps := PackageSel{Name: p.Name, ID: p.ID, Cask: cask, Cmd: p.Name}
		if c, _ := p.Entry["cmd"].(string); c != "" {
			ps.Cmd = c
		}
		ps.VersionCmd = stringList(p.Entry["versionCmd"])
		if len(ps.VersionCmd) == 0 {
			ps.VersionCmd = []string{ps.Cmd, "--version"}
		}
		out = append(out, ps)
	}
	return out
}

func sortedPackages(p []PackageSel) []PackageSel {
	sort.SliceStable(p, func(i, j int) bool { return p[i].Name < p[j].Name })
	return p
}

func stringList(v any) []string {
	list, _ := v.([]any)
	out := make([]string, 0, len(list))
	for _, x := range list {
		s, ok := x.(string)
		if !ok {
			return nil
		}
		out = append(out, s)
	}
	return out
}
