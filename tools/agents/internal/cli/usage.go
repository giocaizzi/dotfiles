package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/cache"
	"github.com/giocaizzi/dotfiles/tools/agents/internal/provider"
	"github.com/giocaizzi/dotfiles/tools/agents/internal/term"
	usg "github.com/giocaizzi/dotfiles/tools/agents/internal/usage"
)

const (
	usageCacheFile = "usage.json"
	noDataNote     = "no data yet (run with -live)"
)

func (d Deps) runUsage(ctx context.Context, args []string) int {
	store := newOSStore(d)
	return d.runUsageWith(ctx, args, store, newFetcher(store.home, d.Getenv))
}

// runUsageWith is runUsage with the credential store and the fetcher injected.
func (d Deps) runUsageWith(ctx context.Context, args []string, store provider.Store, f fetcher) int {
	flags := newFlags("usage")
	live := flags.Bool("live", false, "query every provider and rewrite the cache")
	only := flags.String("provider", "", "limit the output to one provider id")
	if err := flags.Parse(args); err != nil {
		fmt.Fprintf(d.Stderr, "flag parse error: %v\n", err)
		return 2
	}

	dir, err := cache.Dir(d.Getenv)
	if err != nil {
		fmt.Fprintf(d.Stderr, "usage: %v\n", err)
		return 1
	}
	path := filepath.Join(dir, usageCacheFile)
	now := f.now()

	cached, cacheErr := readUsageCache(path)
	if cacheErr != nil {
		fmt.Fprintf(d.Stderr, "usage: ignoring cache: %v\n", cacheErr)
	}

	accounts := provider.Discover(store)

	var snaps []usg.Snapshot
	var meta string
	if *live {
		snaps = liveSnapshots(ctx, accounts, f, liveTimeout)
		snaps = keepStale(snaps, cached.Snapshots)
		if err := writeUsageCache(path, snaps, now); err != nil {
			fmt.Fprintf(d.Stderr, "usage: cache not written: %v\n", err)
		}
		meta = "(live)"
	} else {
		snaps = cachedView(accounts, cached.Snapshots, f)
		switch {
		case cached.WrittenAt.IsZero():
			meta = "(no cache yet; -live to refresh)"
		default:
			meta = "(cached " + humanDuration(now.Sub(cached.WrittenAt)) + " ago; -live to refresh)"
		}
	}

	if *only != "" {
		snaps = filterProvider(snaps, *only)
	}
	renderUsage(d.Stdout, term.NewStyle(d.Stdout, d.Getenv), snaps, meta, now)
	return 0
}

// snapKey identifies the account of a snapshot.
func snapKey(provider, account string, harnesses []string) string {
	return provider + "\x00" + account + "\x00" + strings.Join(harnesses, ",")
}

func filterProvider(snaps []usg.Snapshot, id string) []usg.Snapshot {
	var out []usg.Snapshot
	for _, s := range snaps {
		if s.Provider == id {
			out = append(out, s)
		}
	}
	return out
}

// cachedView is the default view: the cached snapshots, plus one row per
// discovered account the cache lacks. Offline sources fill openai and google;
// an account that cannot be queried shows why; the rest show "no data yet".
func cachedView(accounts []provider.Account, cached []usg.Snapshot, f fetcher) []usg.Snapshot {
	have := make(map[string]bool, len(cached))
	for _, s := range cached {
		have[snapKey(s.Provider, s.Account, s.Harnesses)] = true
	}
	out := append([]usg.Snapshot(nil), cached...)
	for _, a := range accounts {
		if have[snapKey(a.Provider, a.Label, a.Harnesses)] {
			continue
		}
		var (
			snap usg.Snapshot
			ok   bool
		)
		switch {
		case a.Status != "" && a.Status != usg.StatusOK:
			snap, ok = f.base(a), true
			snap.Status, snap.Note = a.Status, a.Note
		case a.Provider == provider.OpenAI:
			snap, ok = f.codexSessionFallback(a)
		case a.Provider == provider.Google:
			snap, ok = f.agyLogFallback(a)
		}
		if !ok {
			snap = unavailable(f.base(a), noDataNote)
		}
		out = append(out, snap)
	}
	return out
}

// keepStale replaces a failed live snapshot by the cached one for the same
// account, marked stale with the failure as its note.
func keepStale(live, cached []usg.Snapshot) []usg.Snapshot {
	old := make(map[string]usg.Snapshot, len(cached))
	for _, s := range cached {
		old[snapKey(s.Provider, s.Account, s.Harnesses)] = s
	}
	out := make([]usg.Snapshot, len(live))
	for i, s := range live {
		out[i] = s
		if s.Status != usg.StatusUnavailable {
			continue
		}
		prev, ok := old[snapKey(s.Provider, s.Account, s.Harnesses)]
		if !ok || (prev.Status != usg.StatusOK && prev.Status != usg.StatusStale) {
			continue
		}
		prev.Status = usg.StatusStale
		prev.Note = s.Note
		if prev.Source == "" {
			prev.Source = "cache"
		}
		out[i] = prev
	}
	return out
}

func readUsageCache(path string) (usg.Cache, error) {
	b, err := os.ReadFile(filepath.Clean(path))
	if errors.Is(err, fs.ErrNotExist) {
		return usg.Cache{}, nil
	}
	if err != nil {
		return usg.Cache{}, fmt.Errorf("read cache: %w", err)
	}
	c, err := usg.ParseCache(b)
	if err != nil {
		return usg.Cache{}, fmt.Errorf("parse cache: %w", err)
	}
	return c, nil
}

// writeUsageCache replaces the cache atomically; it holds no credential.
func writeUsageCache(path string, snaps []usg.Snapshot, now time.Time) error {
	b, err := usg.MarshalCache(snaps, now)
	if err != nil {
		return fmt.Errorf("encode cache: %w", err)
	}
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, "usage-*.tmp")
	if err != nil {
		return fmt.Errorf("create cache file: %w", err)
	}
	defer os.Remove(tmp.Name())
	if err = fillTemp(tmp, b); err != nil {
		return err
	}
	if err = os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("replace cache: %w", err)
	}
	return nil
}

// fillTemp writes b to f with mode 0600 and closes it.
func fillTemp(f *os.File, b []byte) error {
	var errs []error
	if err := f.Chmod(0o600); err != nil && !errors.Is(err, errors.ErrUnsupported) {
		errs = append(errs, fmt.Errorf("chmod cache file: %w", err))
	}
	if _, err := f.Write(b); err != nil {
		errs = append(errs, fmt.Errorf("write cache file: %w", err))
	}
	if err := f.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close cache file: %w", err))
	}
	return errors.Join(errs...)
}
