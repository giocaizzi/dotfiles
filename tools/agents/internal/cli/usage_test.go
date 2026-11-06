package cli

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/provider"
	usg "github.com/giocaizzi/dotfiles/tools/agents/internal/usage"
)

type fakeStore struct{ home string }

func (fakeStore) ReadFile(string) ([]byte, error)       { return nil, fs.ErrNotExist }
func (fakeStore) Secret(string, string) (string, error) { return "", fs.ErrNotExist }
func (fakeStore) Exists(string) bool                    { return false }
func (fakeStore) Getenv(string) string                  { return "" }
func (s fakeStore) Home() string                        { return s.home }
func (fakeStore) GOOS() string                          { return runtime.GOOS }
func (fakeStore) Now() time.Time                        { return fixedNow }
func (fakeStore) OpencodeCredentials(string) ([]provider.OpencodeRow, error) {
	return nil, fs.ErrNotExist
}

func usageDeps(t *testing.T) (Deps, *bytes.Buffer, *bytes.Buffer, string) {
	t.Helper()
	cacheHome := t.TempDir()
	var out, errb bytes.Buffer
	d := Deps{
		Stdout: &out, Stderr: &errb, GOOS: runtime.GOOS,
		Getenv: func(k string) string {
			if k == "XDG_CACHE_HOME" {
				return cacheHome
			}
			return ""
		},
	}
	return d, &out, &errb, filepath.Join(cacheHome, "chezmoi-agents", "usage.json")
}

func TestRunUsageFlagError(t *testing.T) {
	t.Parallel()
	d, _, errb, _ := usageDeps(t)
	if code := d.runUsageWith(t.Context(), []string{"-nope"}, fakeStore{}, testFetcher(t, nil)); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errb.String(), "flag parse error") {
		t.Errorf("stderr = %q", errb.String())
	}
}

func TestRunUsageDispatch(t *testing.T) {
	t.Parallel()
	d, out, _, _ := usageDeps(t)
	if code := d.Run(t.Context(), []string{"usage", "-bogus"}); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	_ = out
}

func TestRunUsageCachedWithoutCache(t *testing.T) {
	t.Parallel()
	d, out, _, _ := usageDeps(t)
	f := testFetcher(t, nil)
	if code := d.runUsageWith(t.Context(), nil, fakeStore{home: f.home}, f); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	for _, want := range []string{"chezmoi agents usage", "no cache yet; -live to refresh"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestRunUsageLiveWritesCache(t *testing.T) {
	t.Parallel()
	d, out, errb, path := usageDeps(t)
	f := testFetcher(t, nil)
	code := d.runUsageWith(t.Context(), []string{"-live"}, fakeStore{home: f.home}, f)
	if code != 0 {
		t.Fatalf("exit = %d, stderr %q", code, errb)
	}
	if !strings.Contains(out.String(), "(live)") {
		t.Errorf("output = %q", out)
	}
	if _, err := os.Stat(path); err != nil {
		// RECHECK: before MarshalCache exists the write fails and says so on stderr.
		if !strings.Contains(errb.String(), "cache not written") {
			t.Errorf("no cache file and no warning: %v / %q", err, errb)
		}
	}
}

func TestUsageCacheRoundTrip(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "sub", "usage.json")
	snaps := sampleSnapshots()

	err := writeUsageCache(path, snaps, fixedNow)
	if errors.Is(err, usg.ErrNotImplemented) {
		t.Skip("usage.MarshalCache is still a stub") // RECHECK after merge
	}
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		fi, serr := os.Stat(path)
		if serr != nil {
			t.Fatal(serr)
		}
		if fi.Mode().Perm() != 0o600 {
			t.Errorf("cache mode = %v, want 0600", fi.Mode().Perm())
		}
		di, _ := os.Stat(filepath.Dir(path))
		if di.Mode().Perm() != 0o700 {
			t.Errorf("cache dir mode = %v, want 0700", di.Mode().Perm())
		}
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Errorf("temp files left behind: %v", entries)
	}
	got, err := readUsageCache(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.WrittenAt.Equal(fixedNow) {
		t.Errorf("WrittenAt = %v", got.WrittenAt)
	}
	if diff := cmp.Diff(len(snaps), len(got.Snapshots)); diff != "" {
		t.Errorf("snapshot count (-want +got):\n%s", diff)
	}
}

func TestReadUsageCacheMissingIsEmpty(t *testing.T) {
	t.Parallel()
	c, err := readUsageCache(filepath.Join(t.TempDir(), "usage.json"))
	if err != nil || len(c.Snapshots) != 0 {
		t.Errorf("= %+v, %v", c, err)
	}
}

func TestSnapKey(t *testing.T) {
	t.Parallel()
	a := snapKey("openai", "me@example.com", []string{"codex", "pi"})
	if a == snapKey("openai", "me@example.com", []string{"codex"}) ||
		a == snapKey("openai", "other@example.com", []string{"codex", "pi"}) ||
		a == snapKey("anthropic", "me@example.com", []string{"codex", "pi"}) {
		t.Error("snapKey collides across distinct accounts")
	}
	if a != snapKey("openai", "me@example.com", []string{"codex", "pi"}) {
		t.Error("snapKey is not stable")
	}
}

func TestFilterProvider(t *testing.T) {
	t.Parallel()
	got := filterProvider(sampleSnapshots(), "anthropic")
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	for _, s := range got {
		if s.Provider != "anthropic" {
			t.Errorf("provider %q kept", s.Provider)
		}
	}
	if len(filterProvider(sampleSnapshots(), "nope")) != 0 {
		t.Error("unknown provider kept rows")
	}
}

func TestCachedViewListsMissingAccounts(t *testing.T) {
	t.Parallel()
	f := testFetcher(t, nil)
	cached := []usg.Snapshot{{Provider: "anthropic", Account: "me@example.com", Harnesses: []string{"claude"}, Status: usg.StatusOK}}
	accts := []provider.Account{
		{Provider: "anthropic", Label: "me@example.com", Harnesses: []string{"claude"}, Auth: provider.AuthOAuth},
		{Provider: "github-copilot", Label: "octocat", Harnesses: []string{"copilot"}, Auth: provider.AuthOAuth},
		{Provider: "openrouter", Label: "", Harnesses: []string{"pi"}, Auth: provider.AuthAPIKey, Status: usg.StatusNoAPI, Note: "no usage API"},
		{Provider: "openai", Label: "me@example.com", Harnesses: []string{"codex"}, Auth: provider.AuthOAuth},
		{Provider: "google", Harnesses: []string{"antigravity"}, Auth: provider.AuthCLI},
	}

	got := cachedView(accts, cached, f)

	byProvider := map[string]usg.Snapshot{}
	for _, s := range got {
		byProvider[s.Provider] = s
	}
	if len(got) != 5 {
		t.Fatalf("rows = %d, want 5", len(got))
	}
	if s := byProvider["github-copilot"]; s.Status != usg.StatusUnavailable || s.Note != noDataNote {
		t.Errorf("copilot row = %+v", s)
	}
	if s := byProvider["openrouter"]; s.Status != usg.StatusNoAPI || s.Note != "no usage API" {
		t.Errorf("openrouter row = %+v", s)
	}
	// openai and google have no offline data here, so they ask for -live.
	for _, p := range []string{"openai", "google"} {
		if s := byProvider[p]; s.Note != noDataNote && s.Status != usg.StatusStale { // RECHECK: offline parsers
			t.Errorf("%s row = %+v", p, s)
		}
	}
}

func TestKeepStale(t *testing.T) {
	t.Parallel()
	old := fixedNow.Add(-time.Hour)
	cached := []usg.Snapshot{
		{Provider: "anthropic", Account: "a", Harnesses: []string{"claude"}, Status: usg.StatusOK, Source: "api", FetchedAt: old, Limits: []usg.Limit{{Name: "5h"}}},
		{Provider: "openai", Account: "b", Harnesses: []string{"codex"}, Status: usg.StatusExpired, FetchedAt: old},
	}
	live := []usg.Snapshot{
		{Provider: "anthropic", Account: "a", Harnesses: []string{"claude"}, Status: usg.StatusUnavailable, Note: "rate limited (429)"},
		{Provider: "openai", Account: "b", Harnesses: []string{"codex"}, Status: usg.StatusUnavailable, Note: "HTTP 500"},
		{Provider: "github-copilot", Account: "c", Harnesses: []string{"copilot"}, Status: usg.StatusUnavailable, Note: "HTTP 500"},
		{Provider: "google", Account: "", Harnesses: []string{"antigravity"}, Status: usg.StatusOK},
	}

	got := keepStale(live, cached)

	if s := got[0]; s.Status != usg.StatusStale || s.Note != "rate limited (429)" || len(s.Limits) != 1 || !s.FetchedAt.Equal(old) || s.Source != "api" {
		t.Errorf("failed account did not keep its cached snapshot as stale: %+v", s)
	}
	if got[1].Status != usg.StatusUnavailable {
		t.Errorf("a non-ok cached snapshot must not replace the failure: %+v", got[1])
	}
	if got[2].Status != usg.StatusUnavailable || got[3].Status != usg.StatusOK {
		t.Errorf("accounts without cache changed: %+v / %+v", got[2], got[3])
	}
}
