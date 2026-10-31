// Package pkgreg is the package registry (.chezmoidata/packages.toml): validation,
// the release table, and the text edit `chezmoi dots bump` makes to it.
package pkgreg

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/finding"
)

func roles() []string { return []string{"personal", "work", "light"} }

func installKeys() []string {
	return []string{"brew", "cask", "apt", "winget", "uv", "npm", "az", "psmodule", "python", "ompFont", "script"}
}

func packageKeys() []string {
	return append([]string{"roles", "scope", "cmd", "agents", "fontFile"}, installKeys()...)
}

func releaseKeys() []string {
	return []string{"roles", "version", "url", "type", "path", "dest", "stripComponents", "exe", "target", "sha256"}
}

var (
	platform = regexp.MustCompile(`^(darwin|linux|windows)/(amd64|arm64)$`)
	sha256   = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Validate checks the package and release tables of data.
func Validate(data map[string]any) []finding.Finding {
	var out []finding.Finding
	pkgs, _ := data["package"].(map[string]any)
	for _, n := range sortedKeys(pkgs) {
		p, _ := pkgs[n].(map[string]any)
		out = append(out, validatePackage(n, p)...)
	}
	rels, _ := data["release"].(map[string]any)
	for _, n := range sortedKeys(rels) {
		r, _ := rels[n].(map[string]any)
		out = append(out, validateRelease(n, r)...)
	}
	return out
}

func validatePackage(n string, p map[string]any) []finding.Finding {
	var out []finding.Finding
	bad := func(format string, args ...any) { out = append(out, finding.New("package "+n, format, args...)) }
	for _, k := range sortedKeys(p) {
		if !slices.Contains(packageKeys(), k) {
			bad("unknown key %s", k)
		}
	}
	if !slices.ContainsFunc(installKeys(), func(k string) bool { _, ok := p[k]; return ok }) && p["agents"] == nil {
		bad("no install method (one of %s)", strings.Join(installKeys(), ", "))
	}
	if r, ok := p["roles"]; ok && !subset(r, roles()) {
		bad("roles must be a subset of %v", roles())
	}
	if s, ok := p["script"]; ok && !subset(s, []string{"darwin", "linux", "windows"}) {
		bad("script must list operating systems")
	}
	if s, ok := p["scope"]; ok && s != "machine" {
		bad(`scope must be "machine"`)
	}
	if _, ok := p["ompFont"]; ok && p["fontFile"] == nil {
		bad("ompFont needs fontFile")
	}
	return out
}

func validateRelease(n string, r map[string]any) []finding.Finding {
	var out []finding.Finding
	bad := func(format string, args ...any) { out = append(out, finding.New("release "+n, format, args...)) }
	for _, k := range sortedKeys(r) {
		if !slices.Contains(releaseKeys(), k) {
			bad("unknown key %s", k)
		}
	}
	if v, _ := r["version"].(string); v == "" {
		bad("version is required")
	}
	if u, _ := r["url"].(string); !strings.Contains(u, "{version}") {
		bad("url must contain {version}")
	}
	if t, ok := r["type"]; ok && !slices.Contains([]string{"file", "archive", "archive-file"}, fmt.Sprint(t)) {
		bad("type must be file, archive or archive-file")
	}
	if ro, ok := r["roles"]; ok && !subset(ro, roles()) {
		bad("roles must be a subset of %v", roles())
	}
	targets, _ := r["target"].(map[string]any)
	if len(targets) == 0 {
		bad("target is required")
	}
	sums, _ := r["sha256"].(map[string]any)
	for _, t := range sortedKeys(targets) {
		if !platform.MatchString(t) {
			bad("target %s must be <os>/<arch>", t)
		}
		if s, _ := sums[t].(string); !sha256.MatchString(s) {
			bad("target %s has no valid sha256 (chezmoi dots bump)", t)
		}
	}
	for _, t := range sortedKeys(sums) {
		if _, ok := targets[t]; !ok {
			bad("sha256 %s has no target", t)
		}
	}
	return out
}

// Release is one pinned release binary.
type Release struct {
	Name    string
	Version string
	URL     string            // with {version} and {target}
	Targets map[string]string // "<os>/<arch>" to the release's name for it
}

// Releases reads the release table.
func Releases(data map[string]any) []Release {
	var out []Release
	rels, _ := data["release"].(map[string]any)
	for _, n := range sortedKeys(rels) {
		r, _ := rels[n].(map[string]any)
		rel := Release{Name: n, Targets: map[string]string{}}
		rel.Version, _ = r["version"].(string)
		rel.URL, _ = r["url"].(string)
		targets, _ := r["target"].(map[string]any)
		for k, v := range targets {
			rel.Targets[k], _ = v.(string)
		}
		out = append(out, rel)
	}
	return out
}

// Platforms lists the release's platforms, sorted.
func (r Release) Platforms() []string {
	out := make([]string, 0, len(r.Targets))
	for p := range r.Targets {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// AssetURL is the download URL of version for platform.
func (r Release) AssetURL(version, platform string) string {
	return strings.NewReplacer("{version}", version, "{target}", r.Targets[platform]).Replace(r.URL)
}

// Upstream says where a release's versions are published.
type Upstream struct {
	Kind      string // "git" (tags of a repository) or "godev"
	Repo      string // clone URL, for Kind "git"
	TagPrefix string // what precedes the version in a tag: "v", "bws-v", ...
}

var (
	githubURL = regexp.MustCompile(`^https://github\.com/([^/]+/[^/]+)/releases/download/([^/{]*)\{version\}/`)
	gitlabURL = regexp.MustCompile(`^https://gitlab\.com/(.+)/-/releases/([^/{]*)\{version\}/`)
)

// Upstream derives where to look for newer versions from the download URL.
func (r Release) Upstream() (Upstream, error) {
	if m := githubURL.FindStringSubmatch(r.URL); m != nil {
		return Upstream{Kind: "git", Repo: "https://github.com/" + m[1] + ".git", TagPrefix: m[2]}, nil
	}
	if m := gitlabURL.FindStringSubmatch(r.URL); m != nil {
		return Upstream{Kind: "git", Repo: "https://gitlab.com/" + m[1] + ".git", TagPrefix: m[2]}, nil
	}
	if strings.HasPrefix(r.URL, "https://go.dev/dl/go{version}") {
		return Upstream{Kind: "godev"}, nil
	}
	return Upstream{}, fmt.Errorf("release %s: cannot tell where %s is published", r.Name, r.URL)
}

// Compare orders dotted numeric versions: negative when a is older than b.
func Compare(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := range max(len(as), len(bs)) {
		x, y := segment(as, i), segment(bs, i)
		if x != y {
			return x - y
		}
	}
	return 0
}

func segment(s []string, i int) int {
	if i >= len(s) {
		return 0
	}
	n, _ := strconv.Atoi(s[i])
	return n
}

var (
	header  = regexp.MustCompile(`^\[([^\]]+)\]\s*$`)
	version = regexp.MustCompile(`^(\s*version\s*=\s*)"[^"]*"(.*)$`)
	sumLine = regexp.MustCompile(`^("([^"]+)"\s*=\s*)"[^"]*"(.*)$`)
)

// SetRelease rewrites the version and checksums of release name in the registry
// text, leaving everything else (comments, order) alone.
func SetRelease(text, name, ver string, sums map[string]string) (string, error) {
	lines := strings.Split(text, "\n")
	table, done := "", map[string]bool{}
	versionSet := false
	for i, line := range lines {
		if m := header.FindStringSubmatch(line); m != nil {
			table = m[1]
			continue
		}
		switch table {
		case "release." + name:
			if m := version.FindStringSubmatch(line); m != nil && !versionSet {
				lines[i] = m[1] + strconv.Quote(ver) + m[2]
				versionSet = true
			}
		case "release." + name + ".sha256":
			if m := sumLine.FindStringSubmatch(line); m != nil {
				if s, ok := sums[m[2]]; ok {
					lines[i] = m[1] + strconv.Quote(s) + m[3]
					done[m[2]] = true
				}
			}
		}
	}
	if !versionSet {
		return "", fmt.Errorf("release %s: no version line to rewrite", name)
	}
	for _, p := range sortedKeys(toAny(sums)) {
		if !done[p] {
			return "", errors.New("release " + name + ": no sha256 line for " + p)
		}
	}
	return strings.Join(lines, "\n"), nil
}

func toAny(m map[string]string) map[string]any {
	out := make(map[string]any, len(m))
	for k := range m {
		out[k] = true
	}
	return out
}

func subset(v any, allowed []string) bool {
	list, ok := v.([]any)
	if !ok {
		return false
	}
	for _, x := range list {
		s, isStr := x.(string)
		if !isStr || !slices.Contains(allowed, s) {
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
