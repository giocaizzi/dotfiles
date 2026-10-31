package edit

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/matrix"
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
			managed, err := s.CZ.Managed(ctx, r.o)
			if err != nil {
				fail[i] = err
				return nil
			}
			reach[i] = strings.Join(s.Tree.Feeds(managed, source), ", ")
			return nil
		})
	}
	_ = g.Wait()
	fmt.Fprintf(s.Out, "%s▸ reaches%s\n", st.B, st.O)
	for i, r := range runs {
		switch {
		case fail[i] != nil:
			fmt.Fprintf(s.Out, "  %s%-17s%s %s✗ render failed%s\n", st.M, r.label, st.O, st.R, st.O)
		case reach[i] == "":
			fmt.Fprintf(s.Out, "  %s%-17s%s %snothing%s\n", st.M, r.label, st.O, st.D, st.O)
		default:
			fmt.Fprintf(s.Out, "  %s%-17s%s %s\n", st.M, r.label, st.O, reach[i])
		}
	}
	return nil
}

func (s *Service) viewTarget(ctx context.Context, target string, mode ViewMode, runs []run) error {
	st := s.Style
	source := ""
	for _, r := range append([]run{{}}, runs...) {
		if p, err := s.CZ.SourcePath(ctx, target, r.o); err == nil {
			source = p
			break
		}
	}
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
	type group struct {
		labels []string
		text   string
	}
	var groups []*group
	var missing []string
	for i, r := range runs {
		if rendered[i] == nil {
			missing = append(missing, r.label)
			continue
		}
		var found *group
		for _, g := range groups {
			if g.text == *rendered[i] {
				found = g
				break
			}
		}
		if found == nil {
			groups = append(groups, &group{labels: []string{r.label}, text: *rendered[i]})
		} else {
			found.labels = append(found.labels, r.label)
		}
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
