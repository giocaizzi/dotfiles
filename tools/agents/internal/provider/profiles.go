package provider

import (
	"path/filepath"
	"slices"
	"strings"
)

// maxGlobProfiles caps the extra profile directories found by glob, per harness.
const maxGlobProfiles = 8

// profileSpec describes where one harness keeps its login profiles: a config
// directory selected by an environment variable (one login per directory), the
// default directory, and well-known sibling directories found by glob.
type profileSpec struct {
	harness string
	envVar  string
	def     []string // default directory, relative to home
	glob    []string // glob pattern, relative to home
	prefix  string   // stripped from a directory name to get the profile name
	markers []string // a globbed directory is a profile only if it holds one of these
}

func claudeSpec() profileSpec {
	return profileSpec{harness: harnessClaude, envVar: "CLAUDE_CONFIG_DIR", def: []string{".claude"}, glob: []string{".claude-*"}, prefix: "claude-", markers: []string{".credentials.json", ".claude.json"}}
}

func codexSpec() profileSpec {
	return profileSpec{harness: "codex", envVar: "CODEX_HOME", def: []string{".codex"}, glob: []string{".codex-*"}, prefix: "codex-", markers: []string{"auth.json", "config.toml"}}
}

func piSpec() profileSpec {
	return profileSpec{harness: "pi", envVar: "PI_CODING_AGENT_DIR", def: []string{".pi", "agent"}, glob: []string{".pi", "agent-*"}, prefix: "agent-", markers: []string{"auth.json"}}
}

// profile is one login directory of a harness.
type profile struct {
	harness   string
	dir       string // as found: the env value, the default, or the glob match
	name      string // empty for the default directory
	active    bool   // the profile the process environment belongs to
	isDefault bool
	envSet    bool // dir came from the environment variable
}

// label is the harness name qualified by the profile name: "claude:work".
func (p profile) label() string {
	if p.name == "" {
		return p.harness
	}
	return p.harness + ":" + p.name
}

// keySuffix qualifies a synthetic merge key so two profiles never merge on it.
func (p profile) keySuffix() string {
	if p.name == "" {
		return ""
	}
	return ":" + p.name
}

// explicit reports whether the harness was pointed at dir (environment
// variable) or dir is not the default: the cases where Claude Code hashes the
// directory into its keychain service.
func (p profile) explicit() bool { return p.envSet || !p.isDefault }

// silent reports whether an unreadable credential is skipped without a row: a
// globbed profile was never asked for.
func (p profile) silent() bool { return !p.envSet && !p.isDefault }

func (s profileSpec) profileName(dir string) string {
	base := strings.TrimPrefix(filepath.Base(dir), ".")
	if n := strings.TrimPrefix(base, s.prefix); n != "" {
		return n
	}
	return base
}

func (s profileSpec) homePath(st Store, rel []string) string {
	return filepath.Join(append([]string{st.Home()}, rel...)...)
}

// profiles lists the profiles of a harness: the environment directory, the
// default directory, then up to maxGlobProfiles globbed directories that hold
// one of the harness's files. Directories are deduplicated by canonical path;
// the environment directory wins.
func (s profileSpec) profiles(st Store) []profile {
	def := s.homePath(st, s.def)
	seen := map[string]bool{}
	var out []profile

	env := st.Getenv(s.envVar)
	if env != "" {
		p := profile{harness: s.harness, dir: env, active: true, envSet: true}
		if st.Canonical(env) == st.Canonical(def) {
			p.isDefault = true
		} else {
			p.name = s.profileName(env)
		}
		seen[st.Canonical(env)] = true
		out = append(out, p)
	}
	if !seen[st.Canonical(def)] {
		seen[st.Canonical(def)] = true
		out = append(out, profile{harness: s.harness, dir: def, isDefault: true, active: env == ""})
	}

	matches := st.Glob(s.homePath(st, s.glob))
	slices.Sort(matches)
	globbed := 0
	for _, dir := range matches {
		if globbed == maxGlobProfiles {
			break
		}
		if seen[st.Canonical(dir)] || !s.hasMarker(st, dir) {
			continue
		}
		seen[st.Canonical(dir)] = true
		globbed++
		out = append(out, profile{harness: s.harness, dir: dir, name: s.profileName(dir)})
	}
	return out
}

func (s profileSpec) hasMarker(st Store, dir string) bool {
	for _, m := range s.markers {
		if st.Exists(filepath.Join(dir, m)) {
			return true
		}
	}
	return false
}

// joinProfiles concatenates the accounts of each profile. When exactly one
// profile of the harness produced accounts, names are not shown: its
// harness label becomes the plain harness name. Account ids stay qualified.
func joinProfiles(harness string, per [][]Account) []Account {
	var out []Account
	found := 0
	for _, accs := range per {
		if len(accs) > 0 {
			found++
		}
	}
	for _, accs := range per {
		for _, a := range accs {
			if found == 1 {
				a.Harnesses = slices.Clone(a.Harnesses)
				for i, h := range a.Harnesses {
					if strings.HasPrefix(h, harness+":") {
						a.Harnesses[i] = harness
					}
				}
			}
			out = append(out, a)
		}
	}
	return out
}
