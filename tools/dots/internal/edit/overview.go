package edit

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/term"
)

type configManaged struct {
	inverted map[string][]string
}

// ViewOverview displays the full 9-configuration matrix overview for all registry entries.
func (s *Service) ViewOverview(ctx context.Context) error {
	scratch, err := mkScratch()
	if err != nil {
		return err
	}
	defer scratch.remove()

	entries, err := s.Matrix(ctx, scratch.dir)
	if err != nil {
		return fmt.Errorf("build configuration matrix: %w", err)
	}

	managedList := make([]configManaged, len(entries))
	var g errgroup.Group
	g.SetLimit(4)
	for i, e := range entries {
		g.Go(func() error {
			m, err := s.CZ.Managed(ctx, e.Overrides)
			if err != nil {
				return err
			}
			managedList[i] = configManaged{inverted: s.Tree.Invert(m)}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("fetch managed configurations: %w", err)
	}

	entryIndex := map[string]int{}
	for i, e := range entries {
		entryIndex[e.Label()] = i
	}

	rows := s.buildOverviewRows(managedList, entryIndex)

	fmt.Fprintf(s.Out, "%s▸ matrix overview (9 configurations: personal · work · light)%s\n", s.Style.B, s.Style.O)
	s.Style.OverviewTable(s.Out, rows)
	fmt.Fprintf(s.Out, "\n  %sLegend:%s\n", s.Style.D, s.Style.O)
	fmt.Fprintf(s.Out, "    %s%s%s reached   %s%s%s unmanaged   %s%s%s local folder\n",
		s.Style.G, term.SymSquareFilled, s.Style.O,
		s.Style.D, term.SymSquareEmpty, s.Style.O,
		s.Style.D, term.SymDot, s.Style.O)
	return nil
}

func (s *Service) buildOverviewRows(managedList []configManaged, entryIndex map[string]int) []term.OverviewRow {
	names := s.Registry.Names()
	rows := make([]term.OverviewRow, 0, len(names))
	for _, n := range names {
		e, _ := s.Registry.Get(n)
		if e.Source == "" {
			dotBlock := fmt.Sprintf("%s%s %s %s%s", s.Style.D, term.SymDot, term.SymDot, term.SymDot, s.Style.O)
			rows = append(rows, term.OverviewRow{
				Entry:   n,
				Darwin:  dotBlock,
				Linux:   dotBlock,
				Windows: dotBlock,
				Targets: "local unmanaged folder",
			})
			continue
		}
		rows = append(rows, s.overviewRowForEntry(n, e.Source, managedList, entryIndex))
	}
	return rows
}

func (s *Service) overviewRowForEntry(name, source string, managedList []configManaged, entryIndex map[string]int) term.OverviewRow {
	allTargets := map[string]bool{}
	formatOS := func(osName string) string {
		var syms []string
		for _, r := range []string{"personal", "work", "light"} {
			idx, ok := entryIndex[osName+"/"+r]
			if !ok {
				syms = append(syms, s.Style.D+term.SymSquareEmpty+s.Style.O)
				continue
			}
			reached := managedList[idx].inverted[source]
			if len(reached) > 0 {
				syms = append(syms, s.Style.G+term.SymSquareFilled+s.Style.O)
				for _, t := range reached {
					allTargets[t] = true
				}
			} else {
				syms = append(syms, s.Style.D+term.SymSquareEmpty+s.Style.O)
			}
		}
		return strings.Join(syms, " ")
	}

	darwinCol := formatOS("darwin")
	linuxCol := formatOS("linux")
	windowsCol := formatOS("windows")

	var desc string
	switch len(allTargets) {
	case 0:
		desc = "nothing reached"
	case 1:
		desc = "1 target"
	default:
		desc = fmt.Sprintf("%d targets", len(allTargets))
	}

	return term.OverviewRow{
		Entry:   name,
		Darwin:  darwinCol,
		Linux:   linuxCol,
		Windows: windowsCol,
		Targets: desc,
	}
}
