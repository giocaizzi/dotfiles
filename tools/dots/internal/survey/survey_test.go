package survey

import (
	"context"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/inventory"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/pkgreg"
)

const secret = "SECRET-TOKEN-IN-STDERR"

type reply struct {
	stdout, stderr string
	err            error
	write          func(args []string) // a side effect, e.g. winget writing its export
}

// fakeRunner answers commands by their joined argv; an unknown command fails.
type fakeRunner struct {
	mu      sync.Mutex
	replies map[string]reply
	calls   []string
	envs    map[string][]string
	stdin   map[string]bool
}

func newRunner(r map[string]reply) *fakeRunner {
	return &fakeRunner{replies: r, envs: map[string][]string{}, stdin: map[string]bool{}}
}

func (f *fakeRunner) Run(ctx context.Context, c chez.Cmd) error {
	key := strings.Join(c.Args, " ")
	f.mu.Lock()
	f.calls = append(f.calls, key)
	f.envs[key] = c.Env
	f.stdin[key] = c.Stdin != nil
	r, ok := f.replies[key]
	f.mu.Unlock()
	if _, has := ctx.Deadline(); !has {
		return errors.New("no timeout on " + key)
	}
	if !ok {
		_, _ = c.Stderr.Write([]byte(secret))
		return errors.New("unexpected command " + key)
	}
	if r.write != nil {
		r.write(c.Args)
	}
	_, _ = c.Stdout.Write([]byte(r.stdout))
	_, _ = c.Stderr.Write([]byte(r.stderr))
	return r.err
}

func (f *fakeRunner) called(prefix string) bool {
	return slices.ContainsFunc(f.calls, func(c string) bool { return strings.HasPrefix(c, prefix) })
}

type fakeInfo struct {
	fs.FileInfo

	mod time.Time
}

func (i fakeInfo) ModTime() time.Time { return i.mod }

func statOf(files map[string]time.Time) func(string) (fs.FileInfo, error) {
	return func(p string) (fs.FileInfo, error) {
		if m, ok := files[p]; ok {
			return fakeInfo{mod: m}, nil
		}
		return nil, fs.ErrNotExist
	}
}

func lookPath(have ...string) func(string) (string, error) {
	return func(n string) (string, error) {
		if slices.Contains(have, n) {
			return "/bin/" + n, nil
		}
		return "", errors.New("not found")
	}
}

func base(r *fakeRunner) Survey {
	return Survey{
		Runner:   r,
		Stat:     statOf(nil),
		LookPath: lookPath(),
		Home:     filepath.Join(string(filepath.Separator), "home", "u"),
		Latest: func(context.Context, pkgreg.Upstream) (string, error) {
			return "", errors.New("latest not stubbed")
		},
	}
}

func group(t *testing.T, rep inventory.Report, src inventory.Source) inventory.Group {
	t.Helper()
	for _, g := range rep.Groups {
		if g.Source == src {
			return g
		}
	}
	t.Fatalf("no %s group in %v", src, rep.Groups)
	return inventory.Group{}
}

func checkGroup(t *testing.T, got inventory.Group, wantNote string, want []inventory.Tool) {
	t.Helper()
	if got.Note != wantNote {
		t.Errorf("Note = %q, want %q", got.Note, wantNote)
	}
	if strings.Contains(got.Note, secret) {
		t.Errorf("Note %q leaks command output", got.Note)
	}
	if diff := cmp.Diff(want, got.Tools); diff != "" {
		t.Errorf("Tools (-want +got):\n%s", diff)
	}
}

func pkg(name, id string) inventory.PackageSel {
	return inventory.PackageSel{Name: name, ID: id, Cmd: name, VersionCmd: []string{name, "--version"}}
}

func TestRunGroupsInOrder(t *testing.T) {
	t.Parallel()
	r := newRunner(map[string]reply{"uv tool list": {stdout: "b v1.0.0\n"}, "uv tool list --outdated": {}})
	s := base(r)
	s.LookPath = lookPath("npm")
	sel := inventory.Selection{
		OS: "plan9", Role: "work",
		Packages: map[inventory.Source][]inventory.PackageSel{
			inventory.Script: {pkg("z", "z")},
			inventory.UV:     {pkg("b", "b"), pkg("a", "a")},
		},
	}
	rep := s.Run(t.Context(), sel)
	got := make([]inventory.Source, 0, len(rep.Groups))
	for _, g := range rep.Groups {
		got = append(got, g.Source)
	}
	if diff := cmp.Diff([]inventory.Source{inventory.UV, inventory.Script}, got); diff != "" {
		t.Errorf("groups (-want +got):\n%s", diff)
	}
	if rep.OS != "plan9" || rep.Role != "work" || rep.Offline {
		t.Errorf("report header = %q %q %v", rep.OS, rep.Role, rep.Offline)
	}
	if names := []string{group(t, rep, inventory.UV).Tools[0].Name, group(t, rep, inventory.UV).Tools[1].Name}; names[0] != "a" || names[1] != "b" {
		t.Errorf("tools not sorted by name: %v", names)
	}
	for k, v := range r.stdin {
		if v {
			t.Errorf("%s ran with stdin", k)
		}
	}
}

func relSel(name, pinned string) inventory.ReleaseSel {
	return inventory.ReleaseSel{
		Release:    pkgreg.Release{Name: name, Version: pinned, URL: "https://github.com/o/" + name + "/releases/download/v{version}/x.tgz"},
		Dest:       ".local/bin/" + name,
		VersionCmd: []string{"{dest}", "--version"},
	}
}

func TestReleases(t *testing.T) {
	t.Parallel()
	s := base(nil)
	home := s.Home
	dest := func(n string) string { return filepath.Join(home, ".local", "bin", n) }
	r := newRunner(map[string]reply{
		dest("a") + " --version": {stdout: "a version 1.2.3\n"},
		dest("b") + " --version": {stderr: "b 2.0.0"},
		dest("c") + " --version": {err: errors.New("boom"), stderr: secret},
	})
	s.Runner = r
	s.Stat = statOf(map[string]time.Time{dest("a"): {}, dest("b"): {}, dest("c"): {}})
	s.Latest = func(_ context.Context, up pkgreg.Upstream) (string, error) {
		if strings.HasSuffix(up.Repo, "/b.git") {
			return "", errors.New(secret)
		}
		return "1.3.0", nil
	}
	sel := inventory.Selection{Releases: []inventory.ReleaseSel{relSel("c", "3.0.0"), relSel("a", "1.2.3"), relSel("b", "2.0.0"), relSel("gone", "1.0.0")}}
	g := s.releases(t.Context(), sel.Releases)
	checkGroup(t, g, "latest unknown for 1 (no network or lookup failed)", []inventory.Tool{
		{Name: "a", Installed: "1.2.3", Pinned: "1.2.3", Latest: "1.3.0"},
		{Name: "b", Installed: "2.0.0", Pinned: "2.0.0"},
		{Name: "c", Pinned: "3.0.0", Latest: "1.3.0"},
		{Name: "gone", Missing: true, Pinned: "1.0.0"},
	})
}

func TestOfflineRunsNoNetworkCommand(t *testing.T) {
	t.Parallel()
	home := filepath.Join(string(filepath.Separator), "home", "u")
	bin := filepath.Join(home, ".local", "bin", "a")
	r := newRunner(map[string]reply{
		bin + " --version": {stdout: "1.0.0"},
		"uv tool list":     {stdout: "pyright v1.1.1\n"},
		"npm ls -g --depth=0 --json --prefix " + filepath.Join(home, ".local"): {stdout: `{"dependencies":{"ts":{"version":"5.1.0"}}}`},
		"git -C /wt rev-parse --abbrev-ref HEAD":                               {stdout: "main\n"},
	})
	s := base(r)
	s.Offline, s.Fetch, s.GOOS = true, true, "linux"
	s.WorkingTree = "/wt"
	s.LookPath = lookPath("npm")
	s.Stat = statOf(map[string]time.Time{bin: {}})
	s.Latest = func(context.Context, pkgreg.Upstream) (string, error) {
		t.Error("Latest called offline")
		return "", nil
	}
	sel := inventory.Selection{
		Releases: []inventory.ReleaseSel{relSel("a", "1.0.0")},
		Packages: map[inventory.Source][]inventory.PackageSel{
			inventory.UV:  {pkg("pyright", "pyright")},
			inventory.Npm: {pkg("ts", "ts@5")},
		},
	}
	rep := s.Run(t.Context(), sel)
	for _, bad := range []string{"uv tool list --outdated", "npm outdated", "git -C /wt fetch"} {
		if r.called(bad) {
			t.Errorf("offline run called %q: %v", bad, r.calls)
		}
	}
	for _, src := range []inventory.Source{inventory.Release, inventory.UV, inventory.Npm} {
		if got := group(t, rep, src).Note; got != offlineNote {
			t.Errorf("%s note = %q", src, got)
		}
	}
	if !rep.Offline || !rep.Tree.Known {
		t.Errorf("report = %+v", rep)
	}
	checkGroup(t, group(t, rep, inventory.Npm), offlineNote, []inventory.Tool{{Name: "ts", ID: "ts@5", Installed: "5.1.0", Pinned: "5"}})
}

func TestBrew(t *testing.T) {
	t.Parallel()
	pkgs := []inventory.PackageSel{pkg("jq", "jq"), pkg("old", "old"), pkg("gone", "gone"), {Name: "app", ID: "app-cask", Cask: true}}
	ok := map[string]reply{
		"brew list --formula --versions": {stdout: "jq 1.7.1\nold 1.0 1.1\n"},
		"brew list --cask --versions":    {stdout: "app-cask 2.0\n"},
		"brew outdated --json=v2":        {stdout: `{"formulae":[{"name":"old","current_version":"1.2"}],"casks":[]}`},
	}
	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		r := newRunner(ok)
		s := base(r)
		s.LookPath = lookPath("brew")
		checkGroup(t, s.brew(t.Context(), pkgs), "", []inventory.Tool{
			{Name: "app", ID: "app-cask", Installed: "2.0", Latest: "2.0"},
			{Name: "gone", Missing: true},
			{Name: "jq", Installed: "1.7.1", Latest: "1.7.1"},
			{Name: "old", Installed: "1.1", Latest: "1.2"},
		})
		want := []string{"HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_ENV_HINTS=1"}
		if diff := cmp.Diff(want, r.envs["brew outdated --json=v2"]); diff != "" {
			t.Errorf("env (-want +got):\n%s", diff)
		}
	})
	t.Run("not installed", func(t *testing.T) {
		t.Parallel()
		s := base(newRunner(nil))
		checkGroup(t, s.brew(t.Context(), pkgs[:1]), "brew is not installed", []inventory.Tool{{Name: "jq"}})
	})
	t.Run("outdated fails", func(t *testing.T) {
		t.Parallel()
		m := map[string]reply{}
		maps.Copy(m, ok)
		m["brew outdated --json=v2"] = reply{err: errors.New("x"), stderr: secret}
		s := base(newRunner(m))
		s.LookPath = lookPath("brew")
		checkGroup(t, s.brew(t.Context(), pkgs[:1]), "latest unknown: brew outdated failed", []inventory.Tool{{Name: "jq", Installed: "1.7.1"}})
	})
}

func TestApt(t *testing.T) {
	t.Parallel()
	cmd := "apt-cache policy fd-find bat"
	r := newRunner(map[string]reply{cmd: {stdout: "fd-find:\n  Installed: 9.0-1\n  Candidate: 9.0-2\nbat:\n  Installed: (none)\n  Candidate: 0.24\n"}})
	s := base(r)
	s.LookPath = lookPath("apt-cache")
	pkgs := []inventory.PackageSel{pkg("fd", "fd-find"), pkg("bat", "bat")}
	checkGroup(t, s.apt(t.Context(), pkgs), "", []inventory.Tool{
		{Name: "bat", Missing: true},
		{Name: "fd", ID: "fd-find", Installed: "9.0-1", Latest: "9.0-2"},
	})
	if diff := cmp.Diff([]string{"LC_ALL=C"}, r.envs[cmd]); diff != "" {
		t.Errorf("env (-want +got):\n%s", diff)
	}
	checkGroup(t, base(newRunner(nil)).apt(t.Context(), pkgs[:1]), "apt is not installed", []inventory.Tool{{Name: "fd", ID: "fd-find"}})
	s.Runner = newRunner(nil)
	checkGroup(t, s.apt(t.Context(), pkgs[:1]), "installed versions unknown: apt-cache policy failed", []inventory.Tool{{Name: "fd", ID: "fd-find"}})
}

func TestWinget(t *testing.T) {
	t.Parallel()
	pkgs := []inventory.PackageSel{pkg("git", "Git.Git"), pkg("jq", "JQ.JQ")}
	run := func(content string, err error) Survey {
		r := &wingetRunner{content: content, err: err}
		s := base(nil)
		s.Runner = r
		return s
	}
	checkGroup(t, run(`{"Sources":[{"Packages":[{"PackageIdentifier":"git.git","Version":"2.50.0"}]}]}`, nil).winget(t.Context(), pkgs),
		"latest not looked up for winget (winget upgrade lists it)", []inventory.Tool{
			{Name: "git", ID: "Git.Git", Installed: "2.50.0"},
			{Name: "jq", ID: "JQ.JQ", Missing: true},
		})
	checkGroup(t, run("", errors.New("x")).winget(t.Context(), pkgs[:1]),
		"installed versions unknown: winget export failed", []inventory.Tool{{Name: "git", ID: "Git.Git"}})
}

// wingetRunner writes the --output file like winget does.
type wingetRunner struct {
	content string
	err     error
}

func (w *wingetRunner) Run(_ context.Context, c chez.Cmd) error {
	i := slices.Index(c.Args, "--output")
	if c.Args[0] != "winget" || i < 0 || !slices.Contains(c.Args, "--include-versions") || !slices.Contains(c.Args, "--disable-interactivity") {
		return errors.New("unexpected " + strings.Join(c.Args, " "))
	}
	if w.content != "" {
		if err := os.WriteFile(c.Args[i+1], []byte(w.content), 0o600); err != nil {
			return err
		}
	}
	_, _ = c.Stderr.Write([]byte(secret))
	return w.err
}

func TestUV(t *testing.T) {
	t.Parallel()
	pkgs := []inventory.PackageSel{pkg("pyright", "pyright[x]>=1"), pkg("ruff", "ruff"), pkg("gone", "gone")}
	list := reply{stdout: "pyright v1.1.400\n- pyright\nruff v0.9.1\n"}
	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		s := base(newRunner(map[string]reply{"uv tool list": list, "uv tool list --outdated": {stdout: "pyright v1.1.400 [latest: 1.1.405]\n"}}))
		checkGroup(t, s.uv(t.Context(), pkgs), "", []inventory.Tool{
			{Name: "gone", Missing: true},
			{Name: "pyright", ID: "pyright[x]>=1", Installed: "1.1.400", Latest: "1.1.405"},
			{Name: "ruff", Installed: "0.9.1", Latest: "0.9.1"},
		})
	})
	t.Run("outdated fails", func(t *testing.T) {
		t.Parallel()
		s := base(newRunner(map[string]reply{"uv tool list": list, "uv tool list --outdated": {err: errors.New("x"), stderr: secret}}))
		checkGroup(t, s.uv(t.Context(), pkgs[:2]), "latest unknown for 2 (no network or lookup failed)", []inventory.Tool{
			{Name: "pyright", ID: "pyright[x]>=1", Installed: "1.1.400"},
			{Name: "ruff", Installed: "0.9.1"},
		})
	})
	t.Run("list fails", func(t *testing.T) {
		t.Parallel()
		checkGroup(t, base(newRunner(nil)).uv(t.Context(), pkgs[1:2]), "installed versions unknown: uv tool list failed", []inventory.Tool{{Name: "ruff"}})
	})
}

func TestNpm(t *testing.T) {
	t.Parallel()
	pkgs := []inventory.PackageSel{pkg("ts", "typescript@5"), pkg("cli", "@s/cli"), pkg("gone", "gone")}
	ls := `{"dependencies":{"typescript":{"version":"5.9.2"},"@s/cli":{"version":"1.0.0"}}}`
	t.Run("exit 1 with JSON, linux prefix", func(t *testing.T) {
		t.Parallel()
		s := base(nil)
		s.GOOS = "linux"
		pre := " --prefix " + filepath.Join(s.Home, ".local")
		s.Runner = newRunner(map[string]reply{
			"npm ls -g --depth=0 --json" + pre: {stdout: ls, err: errors.New("exit 1")},
			"npm outdated -g --json" + pre:     {stdout: `{"typescript":{"latest":"6.0.0"}}`, err: errors.New("exit 1")},
		})
		s.LookPath = lookPath("npm")
		checkGroup(t, s.npm(t.Context(), pkgs), "", []inventory.Tool{
			{Name: "cli", ID: "@s/cli", Installed: "1.0.0", Latest: "1.0.0"},
			{Name: "gone", Missing: true},
			{Name: "ts", ID: "typescript@5", Installed: "5.9.2", Pinned: "5", Latest: "6.0.0"},
		})
	})
	t.Run("darwin has no prefix, outdated fails", func(t *testing.T) {
		t.Parallel()
		s := base(newRunner(map[string]reply{
			"npm ls -g --depth=0 --json": {stdout: ls},
			"npm outdated -g --json":     {err: errors.New("offline"), stderr: secret},
		}))
		s.GOOS = "darwin"
		s.LookPath = lookPath("npm")
		checkGroup(t, s.npm(t.Context(), pkgs[:1]), "latest unknown for 1 (no network or lookup failed)", []inventory.Tool{
			{Name: "ts", ID: "typescript@5", Installed: "5.9.2", Pinned: "5"},
		})
	})
	t.Run("nothing outdated", func(t *testing.T) {
		t.Parallel()
		s := base(newRunner(map[string]reply{"npm ls -g --depth=0 --json": {stdout: ls}, "npm outdated -g --json": {}}))
		s.LookPath = lookPath("npm")
		checkGroup(t, s.npm(t.Context(), pkgs[1:2]), "", []inventory.Tool{{Name: "cli", ID: "@s/cli", Installed: "1.0.0", Latest: "1.0.0"}})
	})
	t.Run("not installed", func(t *testing.T) {
		t.Parallel()
		checkGroup(t, base(newRunner(nil)).npm(t.Context(), pkgs[:1]), "npm is not installed", []inventory.Tool{{Name: "ts", ID: "typescript@5"}})
	})
	t.Run("ls fails", func(t *testing.T) {
		t.Parallel()
		s := base(newRunner(nil))
		s.LookPath = lookPath("npm")
		checkGroup(t, s.npm(t.Context(), pkgs[:1]), "installed versions unknown: npm ls failed", []inventory.Tool{{Name: "ts", ID: "typescript@5"}})
	})
}

func TestScripts(t *testing.T) {
	t.Parallel()
	s := base(newRunner(map[string]reply{
		"bws --version":   {stdout: "bws 2.1.0\n"},
		"flaky --version": {err: errors.New("x"), stderr: secret},
	}))
	s.LookPath = lookPath("bws", "flaky")
	pkgs := []inventory.PackageSel{pkg("bws", "bws"), pkg("absent", "absent"), pkg("flaky", "flaky")}
	checkGroup(t, s.scripts(t.Context(), pkgs), "installed by official installers; latest not looked up", []inventory.Tool{
		{Name: "absent", Missing: true},
		{Name: "bws", Installed: "2.1.0"},
		{Name: "flaky"},
	})
}

func git(args string) string { return "git -C /wt " + args }

func fullTree() map[string]reply {
	return map[string]reply{
		git("rev-parse --abbrev-ref HEAD"):                             {stdout: "main\n"},
		git("rev-parse --abbrev-ref --symbolic-full-name @{upstream}"): {stdout: "origin/main\n"},
		git("rev-list --left-right --count HEAD...@{upstream}"):        {stdout: "2\t5\n"},
		git("status --porcelain"):                                      {stdout: " M a\n?? b\n"},
		git("rev-parse --git-path FETCH_HEAD"):                         {stdout: ".git/FETCH_HEAD\n"},
	}
}

func TestTree(t *testing.T) {
	t.Parallel()
	fetchedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	fetchHead := filepath.Join("/wt", ".git", "FETCH_HEAD")
	t.Run("behind and ahead", func(t *testing.T) {
		t.Parallel()
		r := newRunner(fullTree())
		s := base(r)
		s.WorkingTree = "/wt"
		s.Stat = statOf(map[string]time.Time{fetchHead: fetchedAt})
		want := inventory.Tree{Known: true, Branch: "main", Upstream: "origin/main", Ahead: 2, Behind: 5, Dirty: 2, Fetched: fetchedAt}
		if diff := cmp.Diff(want, s.tree(t.Context())); diff != "" {
			t.Errorf("tree (-want +got):\n%s", diff)
		}
		if r.called("git -C /wt fetch") {
			t.Error("fetched without Fetch")
		}
		if diff := cmp.Diff([]string{"GIT_OPTIONAL_LOCKS=0"}, r.envs[git("status --porcelain")]); diff != "" {
			t.Errorf("env (-want +got):\n%s", diff)
		}
	})
	t.Run("no upstream, detached, never fetched", func(t *testing.T) {
		t.Parallel()
		m := fullTree()
		m[git("rev-parse --abbrev-ref HEAD")] = reply{stdout: "HEAD\n"}
		delete(m, git("rev-parse --abbrev-ref --symbolic-full-name @{upstream}"))
		s := base(newRunner(m))
		s.WorkingTree = "/wt"
		want := inventory.Tree{Known: true, Dirty: 2}
		if diff := cmp.Diff(want, s.tree(t.Context())); diff != "" {
			t.Errorf("tree (-want +got):\n%s", diff)
		}
	})
	t.Run("unknown", func(t *testing.T) {
		t.Parallel()
		if got := base(newRunner(nil)).tree(t.Context()); got.Known {
			t.Errorf("no working tree: Known")
		}
		s := base(newRunner(nil))
		s.WorkingTree = "/wt"
		if got := s.tree(t.Context()); got.Known {
			t.Errorf("git failing: Known")
		}
	})
}

func TestTreeFetch(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		fetch, offline, want bool
	}{{true, false, true}, {true, true, false}, {false, false, false}} {
		m := fullTree()
		m[git("fetch --quiet")] = reply{err: errors.New("no network"), stderr: secret}
		r := newRunner(m)
		s := base(r)
		s.WorkingTree, s.Fetch, s.Offline = "/wt", tt.fetch, tt.offline
		tree := s.tree(t.Context())
		if got := r.called("git -C /wt fetch"); got != tt.want {
			t.Errorf("fetch=%v offline=%v: fetched = %v, want %v", tt.fetch, tt.offline, got, tt.want)
		}
		if !tree.Known {
			t.Errorf("a failed fetch must not make the tree unknown")
		}
		if tt.want {
			if diff := cmp.Diff([]string{"GIT_TERMINAL_PROMPT=0"}, r.envs[git("fetch --quiet")]); diff != "" {
				t.Errorf("env (-want +got):\n%s", diff)
			}
		}
	}
}

func TestCapBuf(t *testing.T) {
	t.Parallel()
	var c capBuf
	chunk := make([]byte, 700<<10)
	for range 3 {
		if n, err := c.Write(chunk); n != len(chunk) || err != nil {
			t.Fatalf("Write = %d, %v", n, err)
		}
	}
	if c.b.Len() != outputCap {
		t.Errorf("kept %d bytes, want %d", c.b.Len(), outputCap)
	}
}
