package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
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
	if *live {
		accounts = provider.Merge(f.resolveIdentities(ctx, accounts))
	} else {
		accounts = provider.Merge(bindCached(accounts, cached.Snapshots))
	}

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
// out.
func snapKey(id, provider string) string {
	return provider + "\x00" + id
}

// acctRef is the identity of a discovered account or a live snapshot.
type acctRef struct {
	id, provider, label string
	kind                usg.Kind
	fingerprints        []string
	unknown             bool // identity not proven: an id match also needs a shared fingerprint
}

// isMetered reports plans billed per use (API key, cloud, gateway) rather than
// by a subscription.
func isMetered(k usg.Kind) bool {
	return k == usg.KindAPI || k == usg.KindCloud || k == usg.KindGateway
}

// cacheIndex finds the cached snapshot of an account by id. A snapshot without
// an id never matches. An id match is a miss when the account changed under the
// same id: both labels set and different (switched accounts), the auth mode
// moved between subscription and metered, or an unproven identity shares no
// credential fingerprint with the snapshot.
type cacheIndex map[string]usg.Snapshot

func newCacheIndex(cached []usg.Snapshot) cacheIndex {
	c := make(cacheIndex, len(cached))
	for _, s := range cached {
		if s.ID == "" {
			continue
		}
		k := snapKey(s.ID, s.Provider)
		if _, dup := c[k]; !dup {
			c[k] = s
		}
	}
	return c
}

func (c cacheIndex) find(r acctRef) (usg.Snapshot, bool) {
	if r.id == "" {
		return usg.Snapshot{}, false
	}
	s, ok := c[snapKey(r.id, r.provider)]
	if !ok {
		return usg.Snapshot{}, false
	}
	if r.label != "" && s.Account != "" && r.label != s.Account {
		return usg.Snapshot{}, false
	}
	if isMetered(r.kind) != isMetered(s.Plan.Kind) {
		return usg.Snapshot{}, false
	}
	if r.unknown && !sharesFingerprint(r.fingerprints, s.Fingerprints) {
		return usg.Snapshot{}, false
	}
	return s, true
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
		refs[i] = acctRef{id: a.ID, provider: a.Provider, label: a.Label, kind: a.Plan.Kind, fingerprints: a.Fingerprints, unknown: identityUnknown(a)}
	}
	have := newCacheIndex(cached)
	var out []usg.Snapshot
	seen := make(map[string]bool, len(accounts))
	for i, a := range accounts {
		key := snapKey(a.ID, a.Provider)
		if seen[key] {
			continue
		}
		seen[key] = true
		if s, ok := have.find(refs[i]); ok {
			s.ID = a.ID
			s.Harnesses = a.Harnesses
			s.Fingerprints = a.Fingerprints
			s.IdentityUnknown = identityUnknown(a)
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
		case a.Provider == provider.OpenAI && slices.Contains(a.Harnesses, "codex"):
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
		refs[i] = acctRef{id: s.ID, provider: s.Provider, label: s.Account, kind: s.Plan.Kind, fingerprints: s.Fingerprints, unknown: s.IdentityUnknown}
	}
	old := newCacheIndex(cached)
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
		prev.Fingerprints = s.Fingerprints
		prev.IdentityUnknown = s.IdentityUnknown
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
