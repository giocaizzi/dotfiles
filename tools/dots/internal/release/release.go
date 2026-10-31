// Package release finds newer versions of pinned release binaries and the checksums
// of their downloads.
package release

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/pkgreg"
)

// Finder looks up the newest published version and checksums downloads.
type Finder struct {
	Runner chez.Runner  // runs git ls-remote
	HTTP   *http.Client // go.dev and the downloads; needs a timeout
}

var numeric = regexp.MustCompile(`^[0-9]+(\.[0-9]+){1,2}$`)

// Latest returns the newest stable version published upstream.
func (f Finder) Latest(ctx context.Context, up pkgreg.Upstream) (string, error) {
	switch up.Kind {
	case "git":
		out, err := chez.Output(ctx, f.Runner, chez.Cmd{Args: []string{"git", "ls-remote", "--tags", "--refs", up.Repo}})
		if err != nil {
			return "", fmt.Errorf("list tags of %s: %w", up.Repo, err)
		}
		return newestTag(string(out), up.TagPrefix)
	case "godev":
		return f.latestGo(ctx)
	default:
		return "", fmt.Errorf("unknown upstream kind %q", up.Kind)
	}
}

// newestTag picks the highest numeric version among tags named <prefix><version>.
func newestTag(lsRemote, prefix string) (string, error) {
	best := ""
	for line := range strings.SplitSeq(lsRemote, "\n") {
		_, ref, ok := strings.Cut(strings.TrimSpace(line), "\trefs/tags/")
		if !ok || !strings.HasPrefix(ref, prefix) {
			continue
		}
		v := strings.TrimPrefix(ref, prefix)
		if numeric.MatchString(v) && (best == "" || pkgreg.Compare(v, best) > 0) {
			best = v
		}
	}
	if best == "" {
		return "", fmt.Errorf("no %s<version> tag", prefix)
	}
	return best, nil
}

func (f Finder) latestGo(ctx context.Context) (string, error) {
	body, err := f.get(ctx, "https://go.dev/dl/?mode=json")
	if err != nil {
		return "", err
	}
	defer body.Close()
	var rels []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}
	if err := json.NewDecoder(body).Decode(&rels); err != nil {
		return "", fmt.Errorf("decode go.dev releases: %w", err)
	}
	for _, r := range rels {
		if v := strings.TrimPrefix(r.Version, "go"); r.Stable && numeric.MatchString(v) {
			return v, nil
		}
	}
	return "", errors.New("go.dev lists no stable release")
}

// SHA256 downloads url and returns the SHA-256 of its body.
func (f Finder) SHA256(ctx context.Context, url string) (string, error) {
	body, err := f.get(ctx, url)
	if err != nil {
		return "", err
	}
	defer body.Close()
	h := sha256.New()
	if _, err := io.Copy(h, body); err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (f Finder) get(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", url, err)
	}
	resp, err := f.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("get %s: %s", url, resp.Status)
	}
	return resp.Body, nil
}

// Status is the outcome of looking up one release.
type Status struct {
	Release pkgreg.Release
	Latest  string
	Err     error
}

// Outdated reports whether a newer version exists.
func (s Status) Outdated() bool {
	return s.Err == nil && pkgreg.Compare(s.Latest, s.Release.Version) > 0
}

// Check looks up the newest version of every release, a few at a time.
func (f Finder) Check(ctx context.Context, rels []pkgreg.Release) []Status {
	out := make([]Status, len(rels))
	var g errgroup.Group
	g.SetLimit(6)
	for i, r := range rels {
		g.Go(func() error {
			out[i] = f.status(ctx, r)
			return nil
		})
	}
	_ = g.Wait()
	return out
}

// status looks up one release; a failure is recorded in the Status, not returned.
func (f Finder) status(ctx context.Context, r pkgreg.Release) Status {
	st := Status{Release: r}
	up, err := r.Upstream()
	if err != nil {
		st.Err = err
		return st
	}
	st.Latest, st.Err = f.Latest(ctx, up)
	return st
}

// Sums downloads version of r for every platform and returns the checksums.
func (f Finder) Sums(ctx context.Context, r pkgreg.Release, version string) (map[string]string, error) {
	platforms := r.Platforms()
	sums := make([]string, len(platforms))
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(3)
	for i, p := range platforms {
		g.Go(func() error {
			s, err := f.SHA256(ctx, r.AssetURL(version, p))
			sums[i] = s
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("checksum %s %s: %w", r.Name, version, err)
	}
	out := make(map[string]string, len(platforms))
	for i, p := range platforms {
		out[p] = sums[i]
	}
	return out, nil
}
