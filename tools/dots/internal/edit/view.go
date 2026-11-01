package edit

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/matrix"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/term"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/udiff"
)

// ViewMode says which configurations View renders.
type ViewMode int

// The view modes.
const (
	ViewHere   ViewMode = iota // this machine
	ViewAs                     // one named configuration
	ViewMatrix                 // every configuration
)

// List prints the registry, one entry per line.
func (s *Service) List() {
	for _, n := range s.Registry.Names() {
		e, _ := s.Registry.Get(n)
		fmt.Fprintf(s.Out, "%s%-20s%s %s\n", s.Style.M, n, s.Style.O, e.Desc)
	}
}

type run struct {
	label string
	o     chez.Overrides
}

func (s *Service) runs(ctx context.Context, mode ViewMode, as matrix.Config, scratch string) ([]run, error) {
	if mode == ViewHere {
		return []run{{label: "this machine"}}, nil
	}
	entries, err := s.Matrix(ctx, scratch)
	if err != nil {
		return nil, fmt.Errorf("build the configuration matrix: %w", err)
	}
	var out []run
	for _, e := range entries {
		if mode == ViewMatrix || e.Config == as {
			out = append(out, run{label: e.Label(), o: e.Overrides})
		}
	}
	return out, nil
}

func (s *Service) entrySources() map[string]string {
	m := map[string]string{}
	for _, n := range s.Registry.Names() {
		if e, _ := s.Registry.Get(n); e.Source != "" {
			m[e.Source] = n
		}
	}
	return m
}

// View shows what a registry entry reaches, or what a target is made of and renders to.
func (s *Service) View(ctx context.Context, what string, mode ViewMode, as matrix.Config) error {
	scratch, err := mkScratch()
	if err != nil {
		return err
	}
	defer scratch.remove()
	runs, err := s.runs(ctx, mode, as, scratch.dir)
	if err != nil {
		return err
	}
	if e, ok := s.Registry.Get(what); ok && e.Source != "" {
		return s.viewEntry(ctx, e.Name, e.Source, e.Desc, runs)
	}
	return s.viewTarget(ctx, what, mode, runs)
}

func (s *Service) viewEntry(ctx context.Context, name, source, desc string, runs []run) error {
	st := s.Style
	fmt.Fprintf(s.Out, "%s▸ %s%s %s\n", st.B, name, st.O, desc)
	st.Tree(s.Out, s.Tree.Of(source), s.entrySources())
	reach := make([]string, len(runs))
	fail := make([]error, len(runs))
	var g errgroup.Group
	g.SetLimit(4)
	for i, r := range runs {
		g.Go(func() error {
			reach[i], fail[i] = s.reaches(ctx, r.o, source)
			return nil
		})
	}
	_ = g.Wait()
	if len(runs) == 1 {
		fmt.Fprintf(s.Out, "%s▸ reaches%s\n", st.B, st.O)
		switch {
		case fail[0] != nil:
			fmt.Fprintf(s.Out, "  %s%-17s%s %s✗ render failed%s\n", st.M, runs[0].label, st.O, st.R, st.O)
		case reach[0] == "":
			fmt.Fprintf(s.Out, "  %s%-17s%s %snothing%s\n", st.M, runs[0].label, st.O, st.D, st.O)
		default:
			fmt.Fprintf(s.Out, "  %s%-17s%s %s\n", st.M, runs[0].label, st.O, reach[0])
		}
		return nil
	}

	s.renderMatrixReaches(runs, reach, fail)
	return nil
}

func (s *Service) renderMatrixReaches(runs []run, reach []string, fail []error) {
	st := s.Style
	palette := st.Palette()
	uniquePaths := []string{}
	pathColors := map[string]string{}
	runIndex := map[string]int{}
	for i, r := range runs {
		runIndex[r.label] = i
		if fail[i] == nil && reach[i] != "" {
			if _, exists := pathColors[reach[i]]; !exists {
				color := palette[len(uniquePaths)%len(palette)]
				pathColors[reach[i]] = color
				uniquePaths = append(uniquePaths, reach[i])
			}
		}
	}

	fmt.Fprintf(s.Out, "%s▸ reaches%s\n", st.B, st.O)
	st.MatrixGrid(s.Out, func(osName, role string) (glyph, color string) {
		idx, ok := runIndex[osName+"/"+role]
		if !ok {
			return term.SymSquareEmpty, st.D
		}
		switch {
		case fail[idx] != nil:
			return term.SymCross, st.R
		case reach[idx] == "":
			return term.SymSquareEmpty, st.D
		default:
			return term.SymSquareFilled, pathColors[reach[idx]]
		}
	})

	s.renderMatrixLegend(uniquePaths, pathColors, runs, reach, fail)
}

func (s *Service) renderMatrixLegend(uniquePaths []string, pathColors map[string]string, runs []run, reach []string, fail []error) {
	st := s.Style
	fmt.Fprintf(s.Out, "\n")
	for _, p := range uniquePaths {
		var matched []string
		for i, r := range runs {
			if reach[i] == p && fail[i] == nil {
				matched = append(matched, r.label)
			}
		}
		fmt.Fprintf(s.Out, "  %s%s%s %s (%s)\n", pathColors[p], term.SymSquareFilled, st.O, p, strings.Join(matched, ", "))
	}
	var unmanaged []string
	var failed []string
	for i, r := range runs {
		if fail[i] != nil {
			failed = append(failed, r.label)
		} else if reach[i] == "" {
			unmanaged = append(unmanaged, r.label)
		}
	}
	if len(unmanaged) > 0 {
		fmt.Fprintf(s.Out, "  %s%s%s unmanaged (%s)\n", st.D, term.SymSquareEmpty, st.O, strings.Join(unmanaged, ", "))
	}
	if len(failed) > 0 {
		fmt.Fprintf(s.Out, "  %s%s%s failed (%s)\n", st.R, term.SymCross, st.O, strings.Join(failed, ", "))
	}
}

// reaches lists, comma-separated, the targets one configuration builds from source.
func (s *Service) reaches(ctx context.Context, o chez.Overrides, source string) (string, error) {
	managed, err := s.CZ.Managed(ctx, o)
	if err != nil {
		return "", fmt.Errorf("managed targets: %w", err)
	}
	return strings.Join(s.Tree.Feeds(managed, source), ", "), nil
}

// renderGroup is a set of configurations a target renders identically on.
type renderGroup struct {
	labels []string
	text   string
}

// groupRenders groups the runs by rendered text; rendered[i] is nil when run i does
// not manage the target, and its label goes to missing.
func groupRenders(runs []run, rendered []*string) (groups []*renderGroup, missing []string) {
	for i, r := range runs {
		if rendered[i] == nil {
			missing = append(missing, r.label)
			continue
		}
		found := slices.IndexFunc(groups, func(g *renderGroup) bool { return g.text == *rendered[i] })
		if found < 0 {
			groups = append(groups, &renderGroup{labels: []string{r.label}, text: *rendered[i]})
		} else {
			groups[found].labels = append(groups[found].labels, r.label)
		}
	}
	return groups, missing
}

// firstSource finds the source file of target, on this machine or else on any run.
func (s *Service) firstSource(ctx context.Context, target string, runs []run) string {
	for _, r := range append([]run{{}}, runs...) {
		if p, err := s.CZ.SourcePath(ctx, target, r.o); err == nil {
			return p
		}
	}
	return ""
}

func (s *Service) viewTarget(ctx context.Context, target string, mode ViewMode, runs []run) error {
	st := s.Style
	source := s.firstSource(ctx, target, runs)
	if source == "" {
		fmt.Fprintf(s.Err, "%s✗ dots: %s is no registry entry or managed target%s\n", st.R, target, st.O)
		return ErrReported
	}
	rel, err := filepath.Rel(s.Src, source)
	if err != nil {
		return fmt.Errorf("source path of %s: %w", target, err)
	}
	fmt.Fprintf(s.Out, "%s▸ tree%s\n", st.B, st.O)
	st.Tree(s.Out, s.Tree.Of(filepath.ToSlash(rel)), s.entrySources())
	if mode == ViewHere {
		out, err := s.CZ.Cat(ctx, target, nil)
		if err != nil {
			return fmt.Errorf("render %s: %w", target, err)
		}
		fmt.Fprintf(s.Out, "%s▸ rendered on this machine%s\n", st.B, st.O)
		_, err = s.Out.Write(out)
		return err //nolint:wrapcheck // a write to the terminal
	}
	rendered := make([]*string, len(runs))
	var g errgroup.Group
	g.SetLimit(4)
	for i, r := range runs {
		g.Go(func() error {
			if out, err := s.CZ.Cat(ctx, target, r.o); err == nil {
				text := string(out)
				rendered[i] = &text
			}
			return nil
		})
	}
	_ = g.Wait()
	groups, missing := groupRenders(runs, rendered)
	if mode == ViewMatrix {
		s.renderTargetMatrix(groups, missing)
	}
	for i, g := range groups {
		if i == 0 {
			fmt.Fprintf(s.Out, "%s▸ rendered on %s%s\n%s", st.B, strings.Join(g.labels, ", "), st.O, g.text)
			continue
		}
		fmt.Fprintf(s.Out, "%s▸ rendered on %s%s %s(against the first)%s\n", st.B, strings.Join(g.labels, ", "), st.O, st.D, st.O)
		fmt.Fprint(s.Out, st.Diff(udiff.Diff(groups[0].labels[0], g.labels[0], groups[0].text, g.text)))
	}
	if len(missing) > 0 {
		fmt.Fprintf(s.Out, "%s▸ not managed on: %s%s\n", st.D, strings.Join(missing, ", "), st.O)
	}
	return nil
}

func (s *Service) renderTargetMatrix(groups []*renderGroup, missing []string) {
	st := s.Style
	palette := st.Palette()
	fmt.Fprintf(s.Out, "%s▸ matrix%s\n", st.B, st.O)
	labelGroup := map[string]int{}
	for gIdx, g := range groups {
		for _, lbl := range g.labels {
			labelGroup[lbl] = gIdx
		}
	}
	st.MatrixGrid(s.Out, func(osName, role string) (glyph, color string) {
		lbl := osName + "/" + role
		if gIdx, ok := labelGroup[lbl]; ok {
			return term.SymSquareFilled, palette[gIdx%len(palette)]
		}
		return term.SymSquareEmpty, st.D
	})
	fmt.Fprintf(s.Out, "\n")
	for gIdx, g := range groups {
		fmt.Fprintf(s.Out, "  %s%s%s variant %d (%s)\n", palette[gIdx%len(palette)], term.SymSquareFilled, st.O, gIdx+1, strings.Join(g.labels, ", "))
	}
	if len(missing) > 0 {
		fmt.Fprintf(s.Out, "  %s%s%s unmanaged (%s)\n", st.D, term.SymSquareEmpty, st.O, strings.Join(missing, ", "))
	}
	fmt.Fprintf(s.Out, "\n")
}
