package inventory

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var dotted = regexp.MustCompile(`\d+(?:\.\d+)+`)

// ParseVersion returns the first dotted version in a command's output ("go version
// go1.27.3 darwin/arm64" gives "1.27.3"); "" when there is none.
func ParseVersion(out string) string { return dotted.FindString(out) }

// BrewList reads `brew list --formula|--cask --versions`: "name v1 [v2...]" lines,
// keeping the last version of each.
func BrewList(out string) map[string]string {
	m := map[string]string{}
	for line := range strings.SplitSeq(out, "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 {
			m[f[0]] = f[len(f)-1]
		}
	}
	return m
}

// BrewOutdated reads `brew outdated --json=v2` into name to current version.
func BrewOutdated(b []byte) (map[string]string, error) {
	type entry struct {
		Name           string `json:"name"`
		CurrentVersion string `json:"current_version"`
	}
	var doc struct {
		Formulae []entry `json:"formulae"`
		Casks    []entry `json:"casks"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("brew outdated: %w", err)
	}
	m := map[string]string{}
	for _, e := range append(doc.Formulae, doc.Casks...) {
		m[e.Name] = e.CurrentVersion
	}
	return m, nil
}

// AptVersions is what `apt-cache policy` says of one package; "" for "(none)".
type AptVersions struct{ Installed, Candidate string }

// AptPolicy reads `apt-cache policy p1 p2`.
func AptPolicy(out string) map[string]AptVersions {
	m := map[string]AptVersions{}
	cur := ""
	for line := range strings.SplitSeq(out, "\n") {
		trim := strings.TrimSpace(line)
		switch {
		case line != "" && !strings.HasPrefix(line, " ") && strings.HasSuffix(trim, ":"):
			cur = strings.TrimSuffix(trim, ":")
			m[cur] = AptVersions{}
		case cur != "" && strings.HasPrefix(trim, "Installed:"):
			v := m[cur]
			v.Installed = aptValue(trim, "Installed:")
			m[cur] = v
		case cur != "" && strings.HasPrefix(trim, "Candidate:"):
			v := m[cur]
			v.Candidate = aptValue(trim, "Candidate:")
			m[cur] = v
		}
	}
	return m
}

func aptValue(line, key string) string {
	v := strings.TrimSpace(strings.TrimPrefix(line, key))
	if v == "(none)" {
		return ""
	}
	return v
}

// WingetExport reads `winget export --include-versions` into lower-cased package
// id to version.
func WingetExport(b []byte) (map[string]string, error) {
	var doc struct {
		Sources []struct {
			Packages []struct {
				PackageIdentifier string `json:"PackageIdentifier"`
				Version           string `json:"Version"`
			} `json:"Packages"`
		} `json:"Sources"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("winget export: %w", err)
	}
	m := map[string]string{}
	for _, s := range doc.Sources {
		for _, p := range s.Packages {
			m[strings.ToLower(p.PackageIdentifier)] = p.Version
		}
	}
	return m, nil
}

// UVList reads `uv tool list`: "name vX.Y.Z" lines; "- exe" lines are skipped.
func UVList(out string) map[string]string {
	m := map[string]string{}
	for line := range strings.SplitSeq(out, "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 && f[0] != "-" && strings.HasPrefix(f[1], "v") {
			m[f[0]] = strings.TrimPrefix(f[1], "v")
		}
	}
	return m
}

var uvLatest = regexp.MustCompile(`^(\S+)\s+v\S+\s+\[latest:\s*([^\]\s]+)\s*\]`)

// UVOutdated reads `uv tool list --outdated`: "name vX [latest: Y]" into name to Y.
func UVOutdated(out string) map[string]string {
	m := map[string]string{}
	for line := range strings.SplitSeq(out, "\n") {
		if g := uvLatest.FindStringSubmatch(strings.TrimSpace(line)); g != nil {
			m[g[1]] = g[2]
		}
	}
	return m
}

// NpmLs reads `npm ls -g --depth=0 --json` into name to installed version.
func NpmLs(b []byte) (map[string]string, error) {
	var doc struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("npm ls: %w", err)
	}
	m := make(map[string]string, len(doc.Dependencies))
	for n, d := range doc.Dependencies {
		m[n] = d.Version
	}
	return m, nil
}

// NpmOutdated reads `npm outdated -g --json` into name to latest version; empty
// input means nothing is outdated.
func NpmOutdated(b []byte) (map[string]string, error) {
	m := map[string]string{}
	if strings.TrimSpace(string(b)) == "" {
		return m, nil
	}
	var doc map[string]struct {
		Latest string `json:"latest"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("npm outdated: %w", err)
	}
	for n, d := range doc {
		m[n] = d.Latest
	}
	return m, nil
}

// NpmSpec splits an npm spec: "typescript@5" is typescript and 5, "@scope/pkg" is
// itself and no pin.
func NpmSpec(id string) (name, pin string) {
	if i := strings.LastIndex(id, "@"); i > 0 {
		return id[:i], id[i+1:]
	}
	return id, ""
}

// UVName is the tool name of a uv spec: up to the first of `[<>=~!@ ;`.
func UVName(id string) string {
	if i := strings.IndexAny(id, "[<>=~!@ ;"); i >= 0 {
		return id[:i]
	}
	return id
}

// GitCounts reads `git rev-list --left-right --count HEAD...@{upstream}`: "2\t5".
func GitCounts(out string) (ahead, behind int, ok bool) {
	f := strings.Fields(out)
	if len(f) != 2 {
		return 0, 0, false
	}
	a, errA := strconv.Atoi(f[0])
	b, errB := strconv.Atoi(f[1])
	if errA != nil || errB != nil {
		return 0, 0, false
	}
	return a, b, true
}
