// Package inventory is the overview `chezmoi dots outdated` prints: every tool the
// package registry gives this machine, grouped by where it comes from, with its
// installed, pinned and latest version and the state they add up to. It is pure:
// the survey package fills it, the term package prints it.
package inventory

import (
	"strings"
	"time"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/pkgreg"
)

// Source is where a tool comes from; each is one table of the overview.
type Source string

// The sources, in the order the overview prints them.
const (
	Release Source = "release" // [release.*]: pinned binaries chezmoi installs as externals
	Brew    Source = "brew"    // Homebrew formulae and casks (macOS)
	Apt     Source = "apt"     // apt packages (Linux)
	Winget  Source = "winget"  // winget packages (Windows)
	UV      Source = "uv"      // uv tools
	Npm     Source = "npm"     // global npm packages
	Script  Source = "script"  // official installers run by the install script
)

// Sources lists every source in print order.
func Sources() []Source { return []Source{Release, Brew, Apt, Winget, UV, Npm, Script} }

// Tool is one row of the overview.
type Tool struct {
	Name      string
	ID        string // the package's name in its manager when it differs from Name ("fd-find"); "" otherwise
	Installed string // installed version; "" when not installed or not known (see Missing)
	Missing   bool   // known not to be installed
	Pinned    string // the registry's pin: a release version, or an npm version prefix ("5"); "" when unpinned
	Latest    string // newest version known; "" when unknown (offline, lookup failed, no lookup for this source)
}

// State is what a tool's versions add up to.
type State int

// The states, from fine to worst; Unknown is "cannot tell".
const (
	UpToDate State = iota // installed is the newest known (and matches the pin)
	Unknown               // installed or latest not known
	Outdated              // a newer version exists: bump the pin, or upgrade the package
	Drifted               // installed differs from the pin: chezmoi apply (or reinstall) brings it back
	Missing               // not installed
)

// Matches reports whether version satisfies pin: equal, or pin is a prefix of
// whole components ("5" matches "5.9.2", not "50.1").
func Matches(version, pin string) bool {
	return version == pin || strings.HasPrefix(version, pin+".")
}

// State classifies t. A release is outdated when upstream has a newer version than
// its pin (whatever is installed matches the pin); a package is outdated when its
// manager knows a newer version than the installed one that still matches its pin.
func (t Tool) State(src Source) State {
	switch {
	case t.Missing:
		return Missing
	case t.Installed != "" && t.Pinned != "" && !Matches(t.Installed, t.Pinned):
		return Drifted
	case src == Release && t.Latest != "" && pkgreg.Compare(t.Latest, t.Pinned) > 0:
		return Outdated
	case t.Installed == "" || t.Latest == "":
		return Unknown
	case src != Release && t.Latest != t.Installed && (t.Pinned == "" || Matches(t.Latest, t.Pinned)):
		return Outdated
	default:
		return UpToDate
	}
}

// Group is one source's table. Note says why it is empty or partial ("brew is not
// installed", "offline: latest versions not looked up"); it never holds tool output.
type Group struct {
	Source Source
	Tools  []Tool // sorted by name
	Note   string
}

// Tree is the dotfiles source checkout against its upstream, read from the last
// fetched refs.
type Tree struct {
	Known    bool      // false when git could not be read (no repository, no git)
	Branch   string    // "main"; "" when detached
	Upstream string    // "origin/main"; "" when the branch tracks nothing
	Ahead    int       // local commits not on the upstream
	Behind   int       // upstream commits not pulled
	Dirty    int       // changed or untracked files
	Fetched  time.Time // last fetch; zero when never fetched
}

// Report is the whole overview.
type Report struct {
	OS, Role string
	Offline  bool // latest versions of releases, uv and npm tools were not looked up
	Groups   []Group
	Tree     Tree
}

// SelectionTemplate renders, through chezmoi (execute-template), what the registry
// gives this machine: the same role, release and package selectors the externals and
// install scripts use, so the overview never re-implements their rules.
const SelectionTemplate = `{{- $sel := dict "os" .chezmoi.os "role" (includeTemplate "role" .) "release" (get . "release" | default dict) "releases" (includeTemplate "releases" . | fromJson) -}}
{{- range $m := list "brew" "cask" "apt" "winget" "uv" "npm" "script" -}}
{{-   $_ := set $sel $m (includeTemplate "packages" (dict "manager" $m "ctx" $) | fromJson) -}}
{{- end -}}
{{- $sel | toJson -}}
`

// Selection is SelectionTemplate's output, decoded by ParseSelection.
type Selection struct {
	OS, Role string
	Releases []ReleaseSel            // this machine's releases, sorted by name
	Packages map[Source][]PackageSel // Brew (formulae and casks), Apt, Winget, UV, Npm, Script; sorted by name
}

// ReleaseSel is one release this machine installs.
type ReleaseSel struct {
	Release    pkgreg.Release // name, pinned version, url (for the upstream lookup)
	Dest       string         // home-relative install path, with .exe on Windows where the registry says so
	VersionCmd []string       // argv printing the installed version; "{dest}" expands to the absolute Dest
}

// PackageSel is one package this machine installs.
type PackageSel struct {
	Name       string
	ID         string   // the manager's name for it (brew formula, cask, apt package, winget id, uv/npm spec)
	Cask       bool     // a Homebrew cask, not a formula
	Cmd        string   // the command it provides (registry "cmd", default Name)
	VersionCmd []string // argv printing the installed version (registry "versionCmd"), for Script
}
