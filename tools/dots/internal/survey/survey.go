// Package survey fills the inventory: it asks each package manager, the installed
// release binaries and git what this machine has, with timeouts and no shell.
package survey

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/inventory"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/pkgreg"
)

const (
	outputCap       = 1 << 20 // bytes kept of a command's output
	releaseLookups  = 6       // concurrent upstream lookups
	defaultTimeout  = 10 * time.Second
	defaultNetwork  = 15 * time.Second
	offlineNote     = "offline: latest versions not looked up"
	unknownLatestFn = "latest unknown for %d (no network or lookup failed)"
)

// Survey gathers the overview of this machine.
type Survey struct {
	Runner      chez.Runner
	Latest      func(context.Context, pkgreg.Upstream) (string, error) // release.Finder.Latest
	Stat        func(string) (fs.FileInfo, error)                      // os.Stat
	LookPath    func(string) (string, error)
	Home        string
	GOOS        string
	WorkingTree string        // the dotfiles git checkout; "" when unknown
	Offline     bool          // no network: no upstream or registry lookups, no fetch
	Fetch       bool          // git fetch the working tree first
	Timeout     time.Duration // per local command; 0 means 10s
	NetTimeout  time.Duration // per network lookup; 0 means 15s
}

// Run surveys sel. It never fails: what cannot be read is reported as unknown.
func (s Survey) Run(ctx context.Context, sel inventory.Selection) inventory.Report {
	rep := inventory.Report{OS: sel.OS, Role: sel.Role, Offline: s.Offline}
	srcs := inventory.Sources()
	groups := make([]*inventory.Group, len(srcs))
	var wg sync.WaitGroup
	for i, src := range srcs {
		if src == inventory.Release && len(sel.Releases) == 0 || src != inventory.Release && len(sel.Packages[src]) == 0 {
			continue
		}
		wg.Go(func() {
			g := s.group(ctx, src, sel)
			g.Source = src
			groups[i] = &g
		})
	}
	wg.Go(func() { rep.Tree = s.tree(ctx) })
	wg.Wait()
	for _, g := range groups {
		if g != nil {
			rep.Groups = append(rep.Groups, *g)
		}
	}
	return rep
}

func (s Survey) group(ctx context.Context, src inventory.Source, sel inventory.Selection) inventory.Group {
	pkgs := sel.Packages[src]
	switch src {
	case inventory.Release:
		return s.releases(ctx, sel.Releases)
	case inventory.Brew:
		return s.brew(ctx, pkgs)
	case inventory.Apt:
		return s.apt(ctx, pkgs)
	case inventory.Winget:
		return s.winget(ctx, pkgs)
	case inventory.UV:
		return s.uv(ctx, pkgs)
	case inventory.Npm:
		return s.npm(ctx, pkgs)
	case inventory.Script:
		return s.scripts(ctx, pkgs)
	}
	return inventory.Group{}
}

// capBuf keeps the first outputCap bytes written to it and drops the rest.
type capBuf struct{ b bytes.Buffer }

func (c *capBuf) Write(p []byte) (int, error) {
	if room := outputCap - c.b.Len(); room > 0 {
		c.b.Write(p[:min(room, len(p))])
	}
	return len(p), nil
}

func (s Survey) local() time.Duration { return orDefault(s.Timeout, defaultTimeout) }
func (s Survey) net() time.Duration   { return orDefault(s.NetTimeout, defaultNetwork) }

func orDefault(d, def time.Duration) time.Duration {
	if d == 0 {
		return def
	}
	return d
}

// exec runs argv (never through a shell) with its output captured and capped.
func (s Survey) exec(ctx context.Context, timeout time.Duration, env []string, args ...string) (stdout, stderr string, err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var out, errb capBuf
	err = s.Runner.Run(ctx, chez.Cmd{Args: args, Env: env, Stdout: &out, Stderr: &errb})
	return out.b.String(), errb.b.String(), err
}

func sorted(tools []inventory.Tool) []inventory.Tool {
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
	return tools
}

func pkgTool(p inventory.PackageSel) inventory.Tool {
	t := inventory.Tool{Name: p.Name}
	if p.ID != p.Name {
		t.ID = p.ID
	}
	return t
}

// unknown is a group whose rows are all "not known", with the reason.
func unknown(pkgs []inventory.PackageSel, note string) inventory.Group {
	g := inventory.Group{Note: note}
	for _, p := range pkgs {
		g.Tools = append(g.Tools, pkgTool(p))
	}
	g.Tools = sorted(g.Tools)
	return g
}

func (s Survey) releases(ctx context.Context, rels []inventory.ReleaseSel) inventory.Group {
	tools := make([]inventory.Tool, len(rels))
	failed := make([]bool, len(rels))
	sem := make(chan struct{}, releaseLookups)
	var wg sync.WaitGroup
	for i, r := range rels {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			tools[i], failed[i] = s.release(ctx, r)
		})
	}
	wg.Wait()
	g := inventory.Group{Tools: sorted(tools)}
	switch n := countTrue(failed); {
	case s.Offline:
		g.Note = offlineNote
	case n > 0:
		g.Note = fmt.Sprintf(unknownLatestFn, n)
	}
	return g
}

func countTrue(b []bool) int {
	n := 0
	for _, x := range b {
		if x {
			n++
		}
	}
	return n
}

// release reads one release; failed says its latest version lookup failed.
func (s Survey) release(ctx context.Context, r inventory.ReleaseSel) (t inventory.Tool, failed bool) {
	t = inventory.Tool{Name: r.Release.Name, Pinned: r.Release.Version}
	dest := filepath.Join(s.Home, filepath.FromSlash(r.Dest))
	if _, err := s.Stat(dest); errors.Is(err, fs.ErrNotExist) {
		t.Missing = true
		return t, false
	} else if err == nil {
		argv := make([]string, len(r.VersionCmd))
		for i, a := range r.VersionCmd {
			argv[i] = strings.ReplaceAll(a, "{dest}", dest)
		}
		if len(argv) > 0 {
			argv[0] = filepath.FromSlash(argv[0])
			stdout, stderr, _ := s.exec(ctx, s.local(), nil, argv...)
			t.Installed = inventory.ParseVersion(stdout + "\n" + stderr)
		}
	}
	if s.Offline {
		return t, false
	}
	up, err := r.Release.Upstream()
	if err == nil && s.Latest != nil {
		lctx, cancel := context.WithTimeout(ctx, s.net())
		defer cancel()
		t.Latest, err = s.Latest(lctx, up)
	} else {
		err = errors.New("no lookup")
	}
	if err != nil {
		t.Latest = ""
		return t, true
	}
	return t, false
}

func (s Survey) brew(ctx context.Context, pkgs []inventory.PackageSel) inventory.Group {
	if _, err := s.LookPath("brew"); err != nil {
		return unknown(pkgs, "brew is not installed")
	}
	env := []string{"HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_ENV_HINTS=1"}
	var formulae, casks, outdated map[string]string
	var errF, errC, errO error
	var wg sync.WaitGroup
	list := func(kind string, dst *map[string]string, dstErr *error) {
		wg.Go(func() {
			out, _, err := s.exec(ctx, s.local(), env, "brew", "list", kind, "--versions")
			*dst, *dstErr = inventory.BrewList(out), err
		})
	}
	list("--formula", &formulae, &errF)
	list("--cask", &casks, &errC)
	wg.Go(func() {
		out, _, err := s.exec(ctx, s.local(), env, "brew", "outdated", "--json=v2")
		if err == nil {
			outdated, err = inventory.BrewOutdated([]byte(out))
		}
		errO = err
	})
	wg.Wait()

	g := inventory.Group{}
	for _, p := range pkgs {
		t := pkgTool(p)
		installed, listErr := formulae, errF
		if p.Cask {
			installed, listErr = casks, errC
		}
		if listErr == nil {
			v, ok := installed[p.ID]
			t.Installed, t.Missing = v, !ok
		}
		if !t.Missing && errO == nil && t.Installed != "" {
			t.Latest = t.Installed
			if v, ok := outdated[p.ID]; ok {
				t.Latest = v
			}
		}
		g.Tools = append(g.Tools, t)
	}
	g.Tools = sorted(g.Tools)
	switch {
	case errF != nil || errC != nil:
		g.Note = "installed versions unknown: brew list failed"
	case errO != nil:
		g.Note = "latest unknown: brew outdated failed"
	}
	return g
}

func (s Survey) apt(ctx context.Context, pkgs []inventory.PackageSel) inventory.Group {
	if _, err := s.LookPath("apt-cache"); err != nil {
		return unknown(pkgs, "apt is not installed")
	}
	args := []string{"apt-cache", "policy"}
	for _, p := range pkgs {
		args = append(args, p.ID)
	}
	out, _, err := s.exec(ctx, s.local(), []string{"LC_ALL=C"}, args...)
	if err != nil {
		return unknown(pkgs, "installed versions unknown: apt-cache policy failed")
	}
	policy := inventory.AptPolicy(out)
	g := inventory.Group{}
	for _, p := range pkgs {
		t := pkgTool(p)
		v := policy[p.ID]
		if v.Installed == "" {
			t.Missing = true
		} else {
			t.Installed, t.Latest = v.Installed, v.Candidate
		}
		g.Tools = append(g.Tools, t)
	}
	g.Tools = sorted(g.Tools)
	return g
}

func (s Survey) winget(ctx context.Context, pkgs []inventory.PackageSel) inventory.Group {
	dir, err := os.MkdirTemp("", "dots-winget-")
	if err != nil {
		return unknown(pkgs, "installed versions unknown: winget export failed")
	}
	defer func() { _ = os.RemoveAll(dir) }()
	file := filepath.Join(dir, "winget.json")
	// winget exits non-zero when some sources are unreachable yet still writes the
	// export, so a readable export wins over the exit status.
	_, _, err = s.exec(ctx, s.local(), nil, "winget", "export", "--output", file,
		"--include-versions", "--accept-source-agreements", "--disable-interactivity")
	_ = err
	var installed map[string]string
	if b, err := os.ReadFile(file); err == nil { //nolint:gosec // a file in our own temp dir
		installed, err = inventory.WingetExport(b)
		if err != nil {
			installed = nil
		}
	}
	if installed == nil {
		return unknown(pkgs, "installed versions unknown: winget export failed")
	}
	g := inventory.Group{Note: "latest not looked up for winget (winget upgrade lists it)"}
	for _, p := range pkgs {
		t := pkgTool(p)
		v, ok := installed[strings.ToLower(p.ID)]
		t.Installed, t.Missing = v, !ok
		g.Tools = append(g.Tools, t)
	}
	g.Tools = sorted(g.Tools)
	return g
}

// latestNote is the note of a group whose latest versions come from one network call.
func (s Survey) latestNote(g inventory.Group, outdatedFailed bool) string {
	switch {
	case s.Offline:
		return offlineNote
	case outdatedFailed:
		n := 0
		for _, t := range g.Tools {
			if !t.Missing && t.Installed != "" {
				n++
			}
		}
		return fmt.Sprintf(unknownLatestFn, n)
	}
	return ""
}

func (s Survey) uv(ctx context.Context, pkgs []inventory.PackageSel) inventory.Group {
	out, _, err := s.exec(ctx, s.local(), nil, "uv", "tool", "list")
	if err != nil {
		return unknown(pkgs, "installed versions unknown: uv tool list failed")
	}
	installed := inventory.UVList(out)
	var latest map[string]string
	outdatedFailed := false
	if !s.Offline {
		out, _, err := s.exec(ctx, s.net(), nil, "uv", "tool", "list", "--outdated")
		latest, outdatedFailed = inventory.UVOutdated(out), err != nil
	}
	g := inventory.Group{}
	for _, p := range pkgs {
		t := pkgTool(p)
		name := inventory.UVName(p.ID)
		v, ok := installed[name]
		t.Installed, t.Missing = v, !ok
		if ok && !s.Offline && !outdatedFailed {
			t.Latest = v
			if l, ok := latest[name]; ok {
				t.Latest = l
			}
		}
		g.Tools = append(g.Tools, t)
	}
	g.Tools = sorted(g.Tools)
	g.Note = s.latestNote(g, outdatedFailed)
	return g
}

// npmJSON runs an npm command whose JSON stdout counts even when it exits 1
// (outdated packages, extraneous ones). Empty output after a failure is a failure.
func (s Survey) npmJSON(ctx context.Context, timeout time.Duration, args []string) ([]byte, bool) {
	out, _, err := s.exec(ctx, timeout, nil, args...)
	if err != nil && strings.TrimSpace(out) == "" {
		return nil, false
	}
	return []byte(out), true
}

func (s Survey) npm(ctx context.Context, pkgs []inventory.PackageSel) inventory.Group {
	if _, err := s.LookPath("npm"); err != nil {
		return unknown(pkgs, "npm is not installed")
	}
	var prefix []string
	if s.GOOS == "linux" {
		prefix = []string{"--prefix", filepath.Join(s.Home, ".local")}
	}
	var installed, latest map[string]string
	b, ok := s.npmJSON(ctx, s.local(), slices.Concat([]string{"npm", "ls", "-g", "--depth=0", "--json"}, prefix))
	if ok {
		var err error
		if installed, err = inventory.NpmLs(b); err != nil {
			ok = false
		}
	}
	if !ok {
		return unknown(pkgs, "installed versions unknown: npm ls failed")
	}
	outdatedFailed := false
	if !s.Offline {
		b, ok := s.npmJSON(ctx, s.net(), slices.Concat([]string{"npm", "outdated", "-g", "--json"}, prefix))
		var err error
		if ok {
			latest, err = inventory.NpmOutdated(b)
		}
		outdatedFailed = !ok || err != nil
	}
	g := inventory.Group{}
	for _, p := range pkgs {
		t := pkgTool(p)
		name, pin := inventory.NpmSpec(p.ID)
		t.Pinned = pin
		v, ok := installed[name]
		t.Installed, t.Missing = v, !ok
		if ok && !s.Offline && !outdatedFailed {
			t.Latest = v
			if l, ok := latest[name]; ok {
				t.Latest = l
			}
		}
		g.Tools = append(g.Tools, t)
	}
	g.Tools = sorted(g.Tools)
	g.Note = s.latestNote(g, outdatedFailed)
	return g
}

func (s Survey) scripts(ctx context.Context, pkgs []inventory.PackageSel) inventory.Group {
	g := inventory.Group{Note: "installed by official installers; latest not looked up"}
	for _, p := range pkgs {
		t := pkgTool(p)
		if len(p.VersionCmd) == 0 {
			g.Tools = append(g.Tools, t)
			continue
		}
		if _, err := s.LookPath(p.VersionCmd[0]); err != nil {
			t.Missing = true
		} else {
			stdout, stderr, _ := s.exec(ctx, s.local(), nil, p.VersionCmd...)
			t.Installed = inventory.ParseVersion(stdout + "\n" + stderr)
		}
		g.Tools = append(g.Tools, t)
	}
	g.Tools = sorted(g.Tools)
	return g
}

func (s Survey) tree(ctx context.Context) inventory.Tree {
	wt := s.WorkingTree
	if wt == "" {
		return inventory.Tree{}
	}
	if s.Fetch && !s.Offline {
		_, _, err := s.exec(ctx, s.net(), []string{"GIT_TERMINAL_PROMPT=0"}, "git", "-C", wt, "fetch", "--quiet")
		_ = err // a failed fetch only leaves the last fetched refs in use
	}
	env := []string{"GIT_OPTIONAL_LOCKS=0"}
	git := func(args ...string) (string, error) {
		out, _, err := s.exec(ctx, s.local(), env, slices.Concat([]string{"git", "-C", wt}, args)...)
		return strings.TrimSpace(out), err
	}
	branch, err := git("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return inventory.Tree{}
	}
	t := inventory.Tree{Known: true, Branch: branch}
	if branch == "HEAD" {
		t.Branch = ""
	}
	if up, err := git("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}"); err == nil {
		t.Upstream = up
		if out, err := git("rev-list", "--left-right", "--count", "HEAD...@{upstream}"); err == nil {
			t.Ahead, t.Behind, _ = inventory.GitCounts(out)
		}
	}
	if out, err := git("status", "--porcelain"); err == nil {
		t.Dirty = countLines(out)
	}
	if p, err := git("rev-parse", "--git-path", "FETCH_HEAD"); err == nil && p != "" {
		if !filepath.IsAbs(p) {
			p = filepath.Join(wt, p)
		}
		if fi, err := s.Stat(p); err == nil {
			t.Fetched = fi.ModTime()
		}
	}
	return t
}

func countLines(out string) int {
	n := 0
	for line := range strings.SplitSeq(out, "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}
