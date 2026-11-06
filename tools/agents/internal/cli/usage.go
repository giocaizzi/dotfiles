package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
func (d Deps) runUsageWith(ctx context.Context, args []string, store provider.Store, f usageFetcher) int {
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
			meta = "(cached " + usg.HumanDuration(now.Sub(cached.WrittenAt)) + " ago; -live to refresh)"
		}
	}

	if *only != "" {
		snaps = filterProvider(snaps, *only)
	}
	renderUsage(d.Stdout, term.NewStyle(d.Stdout, d.Getenv), snaps, meta, now)
	return 0
}

// snapKey identifies the account of a snapshot: its stable id under the
// provider, never the harness list, which changes when a harness logs in or
// out. Without an id (a cache file from before ids) it falls back to the
// provider and label.
func snapKey(id, provider, label string) string {
	if id != "" {
		return provider + "\x00id\x00" + id
	}
	return provider + "\x00label\x00" + label
}

// acctRef is the identity of a discovered account or a live snapshot.
type acctRef struct{ id, provider, label string }

// cacheIndex finds the cached snapshot of an account. A cached snapshot with
// an id matches by id. One without (an old cache file) matches only by
// provider and label, and only when no other current account of that provider
// shares the label; otherwise it is dropped as unknown.
type cacheIndex struct {
	byKey  map[string]usg.Snapshot
	legacy map[string]usg.Snapshot
	shared map[string]int
}

func newCacheIndex(cached []usg.Snapshot, current []acctRef) cacheIndex {
	c := cacheIndex{
		byKey:  make(map[string]usg.Snapshot, len(cached)),
		legacy: make(map[string]usg.Snapshot),
		shared: make(map[string]int, len(current)),
	}
	for _, s := range cached {
		k := snapKey(s.ID, s.Provider, s.Account)
		dst := c.byKey
		if s.ID == "" {
			dst = c.legacy
		}
		if _, dup := dst[k]; !dup {
			dst[k] = s
		}
	}
	for _, r := range current {
		c.shared[snapKey("", r.provider, r.label)]++
	}
	return c
}

func (c cacheIndex) find(r acctRef) (usg.Snapshot, bool) {
	if r.id != "" {
		if s, ok := c.byKey[snapKey(r.id, r.provider, r.label)]; ok {
			return s, true
		}
	}
	k := snapKey("", r.provider, r.label)
	if c.shared[k] > 1 {
		return usg.Snapshot{}, false
	}
	s, ok := c.legacy[k]
	return s, ok
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

// cachedView is the default view: one row per discovered account, from its
// cached snapshot (with the account's current harnesses) when there is one.
// Cached snapshots of accounts no longer discovered are dropped. Without a
// cached snapshot, offline sources fill openai and google; an account that
// cannot be queried shows why; the rest show "no data yet".
func cachedView(accounts []provider.Account, cached []usg.Snapshot, f usageFetcher) []usg.Snapshot {
	refs := make([]acctRef, len(accounts))
	for i, a := range accounts {
		refs[i] = acctRef{a.ID, a.Provider, a.Label}
	}
	have := newCacheIndex(cached, refs)
	var out []usg.Snapshot
	seen := make(map[string]bool, len(accounts))
	for i, a := range accounts {
		key := snapKey(a.ID, a.Provider, a.Label)
		if seen[key] {
			continue
		}
		seen[key] = true
		if s, ok := have.find(refs[i]); ok {
			s.ID = a.ID
			s.Harnesses = a.Harnesses
			out = append(out, s)
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
// account, marked stale with the failure as its note. Only live accounts are
// kept, so an account no longer discovered leaves the rewritten cache.
func keepStale(live, cached []usg.Snapshot) []usg.Snapshot {
	refs := make([]acctRef, len(live))
	for i, s := range live {
		refs[i] = acctRef{s.ID, s.Provider, s.Account}
	}
	old := newCacheIndex(cached, refs)
	out := make([]usg.Snapshot, len(live))
	for i, s := range live {
		out[i] = s
		if s.Status != usg.StatusUnavailable {
			continue
		}
		prev, ok := old.find(refs[i])
		if !ok || (prev.Status != usg.StatusOK && prev.Status != usg.StatusStale) {
			continue
		}
		prev.ID = s.ID
		prev.Harnesses = s.Harnesses
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
