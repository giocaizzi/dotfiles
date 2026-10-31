package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/edit"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/lint"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/pkgreg"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/release"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/term"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/upgrade"
)

func (d Deps) lint(ctx context.Context, args []string) error {
	fs := newFlags("lint")
	harnessDir := fs.String("harness-dir", d.Getenv("HARNESS_DIR"), "a checkout of giocaizzi/harness (default: ask GitHub through gh)")
	checks, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(checks) == 0 {
		checks = lint.Checks()
	}
	cz, _ := d.chezmoi()
	src, err := cz.SourcePath(ctx, "", nil)
	if err != nil {
		return fmt.Errorf("find the chezmoi source dir: %w", err)
	}
	scratch, err := os.MkdirTemp("", "dots-lint-")
	if err != nil {
		return fmt.Errorf("make scratch dir: %w", err)
	}
	defer os.RemoveAll(scratch)
	empty := filepath.Join(scratch, "empty.toml")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		return fmt.Errorf("write empty config: %w", err)
	}
	l := lint.Linter{
		Src: src, Iso: cz.With("--config", empty), Real: cz, Runner: d.Runner, LookPath: d.LookPath,
		HarnessDir: *harnessDir, Scratch: scratch, Errw: d.Stderr,
	}
	st := term.NewStyle(d.Stderr, d.Getenv)
	failed := false
	for _, c := range checks {
		found, err := l.Run(ctx, c)
		if err != nil {
			return fmt.Errorf("lint %s: %w", c, err)
		}
		if len(found) == 0 {
			fmt.Fprintf(d.Stderr, "%s✓%s %s\n", st.G, st.O, c)
			continue
		}
		failed = true
		fmt.Fprintf(d.Stderr, "%s✗%s %s\n%s\n", st.R, st.O, c, joinLines(found))
	}
	if failed {
		return ErrReported
	}
	return nil
}

func (d Deps) finder() release.Finder {
	return release.Finder{Runner: d.Runner, HTTP: d.HTTP}
}

func (d Deps) releases(ctx context.Context) (string, []pkgreg.Release, error) {
	cz, _ := d.chezmoi()
	src, err := cz.SourcePath(ctx, "", nil)
	if err != nil {
		return "", nil, fmt.Errorf("find the chezmoi source dir: %w", err)
	}
	data, err := cz.Data(ctx, nil)
	if err != nil {
		return "", nil, err
	}
	return filepath.Join(src, ".chezmoidata", "packages.toml"), pkgreg.Releases(data), nil
}

func (d Deps) outdated(ctx context.Context, args []string) error {
	if pos, err := parse(newFlags("outdated"), args); err != nil || len(pos) > 0 {
		return ErrUsage
	}
	_, rels, err := d.releases(ctx)
	if err != nil {
		return err
	}
	st := term.NewStyle(d.Stdout, d.Getenv)
	n := 0
	for _, s := range d.finder().Check(ctx, rels) {
		switch {
		case s.Err != nil:
			fmt.Fprintf(d.Stdout, "%s%-16s%s %s✗ %v%s\n", st.M, s.Release.Name, st.O, st.R, s.Err, st.O)
		case s.Outdated():
			n++
			fmt.Fprintf(d.Stdout, "%s%-16s%s %s → %s%s%s\n", st.M, s.Release.Name, st.O, s.Release.Version, st.G, s.Latest, st.O)
		default:
			fmt.Fprintf(d.Stdout, "%s%-16s %s up to date%s\n", st.D, s.Release.Name, s.Release.Version, st.O)
		}
	}
	if n > 0 {
		fmt.Fprintf(d.Stdout, "\n%d outdated: chezmoi dots bump pins the new versions and checksums\n", n)
	}
	return nil
}

func (d Deps) bump(ctx context.Context, args []string) error {
	fs := newFlags("bump")
	dry := fs.Bool("dry-run", false, "show what would change, write nothing")
	names, err := parse(fs, args)
	if err != nil {
		return err
	}
	file, rels, err := d.releases(ctx)
	if err != nil {
		return err
	}
	if len(names) > 0 {
		rels = slices.DeleteFunc(rels, func(r pkgreg.Release) bool { return !slices.Contains(names, r.Name) })
		if len(rels) != len(names) {
			return fmt.Errorf("%w: unknown release in %v", ErrUsage, names)
		}
	}
	text, err := os.ReadFile(file) //nolint:gosec // file is the package registry in the user's own dotfiles source
	if err != nil {
		return fmt.Errorf("read %s: %w", file, err)
	}
	updated, f := string(text), d.finder()
	bumped := 0
	for _, s := range f.Check(ctx, rels) {
		if s.Err != nil {
			return fmt.Errorf("release %s: %w", s.Release.Name, s.Err)
		}
		if !s.Outdated() {
			continue
		}
		sums, err := f.Sums(ctx, s.Release, s.Latest)
		if err != nil {
			return err
		}
		if updated, err = pkgreg.SetRelease(updated, s.Release.Name, s.Latest, sums); err != nil {
			return err
		}
		bumped++
		fmt.Fprintf(d.Stdout, "%s %s → %s\n", s.Release.Name, s.Release.Version, s.Latest)
	}
	switch {
	case bumped == 0:
		fmt.Fprintln(d.Stdout, "all releases are up to date")
	case *dry:
		fmt.Fprintln(d.Stdout, "dry run: nothing written")
	default:
		if err := os.WriteFile(file, []byte(updated), 0o644); err != nil { //nolint:gosec // a source file like the others
			return fmt.Errorf("write %s: %w", file, err)
		}
		fmt.Fprintf(d.Stdout, "pinned in %s; run chezmoi apply to install\n", file)
	}
	return nil
}

func (d Deps) upgrade(ctx context.Context, args []string) error {
	fs := newFlags("upgrade")
	dry := fs.Bool("dry-run", false, "show the commands, run nothing")
	only, err := parse(fs, args)
	if err != nil {
		return err
	}
	bin := d.Getenv("CHEZMOI_EXECUTABLE")
	if bin == "" {
		bin = "chezmoi"
	}
	steps, err := upgrade.Select(upgrade.Plan(upgrade.Env{GOOS: d.GOOS, Home: d.Home, Bin: bin, LookPath: d.LookPath}), only)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUsage, err)
	}
	if err := upgrade.Run(ctx, d.Runner, steps, *dry, d.Stdout, d.Stderr); err != nil {
		return err
	}
	if len(only) == 0 || slices.Contains(only, "releases") {
		fmt.Fprintln(d.Stdout, "==> releases: pinned binaries")
		return d.outdated(ctx, nil)
	}
	return nil
}

func (d Deps) live(ctx context.Context, args []string) error {
	fs := newFlags("live")
	bin := fs.String("bin", "chezmoi", "")
	src := fs.String("src", "", "")
	realSrc := fs.String("real-src", "", "")
	file := fs.String("file", "", "")
	edited := fs.String("edited", "", "")
	dir := fs.String("dir", "", "")
	var global stringList
	fs.Var(&global, "global", "")
	if _, err := parse(fs, args); err != nil {
		return err
	}
	cz := chez.New(d.Runner, *bin, append(slices.Clone(global), "--source", *src)...)
	l := edit.Live{
		CZ: cz, File: *file, Edited: *edited, Shadow: *src, RealSrc: *realSrc, Dir: *dir,
		Out: d.Stdout, Style: term.NewStyle(d.Stdout, d.Getenv), Interval: 500 * time.Millisecond,
	}
	return l.Loop(ctx)
}

type stringList []string

func (s *stringList) String() string     { return strings.Join(*s, " ") }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }
