package term

import (
	"fmt"
	"io"
	"strings"
)

// Symbols for matrix preview grids and overview tables.
const (
	SymSquareFilled = "■"
	SymSquareEmpty  = "□"
	SymDot          = "·"
	SymCross        = "✗"
)

// Palette returns the distinct highlight colours for matrix preview diffs:
// Green, Mauve, Cyan, Yellow, Blue.
func (s Style) Palette() []string {
	return []string{s.G, s.M, s.C, s.Y, s.B}
}

// MatrixGrid prints a 3×3 OS × Role block.
func (s Style) MatrixGrid(w io.Writer, cell func(os, role string) (glyph, color string)) {
	oses := []string{"darwin", "linux", "windows"}
	roles := []string{"personal", "work", "light"}

	fmt.Fprintln(w, "            personal    work     light")
	for _, osName := range oses {
		var b strings.Builder
		fmt.Fprintf(&b, "%-12s", osName)
		for i, role := range roles {
			glyph, color := cell(osName, role)
			if i == 0 {
				b.WriteString("   ")
			} else {
				b.WriteString("         ")
			}
			if color != "" && s.O != "" {
				b.WriteString(color)
				b.WriteString(glyph)
				b.WriteString(s.O)
			} else {
				b.WriteString(glyph)
			}
		}
		fmt.Fprintln(w, b.String())
	}
}

// OverviewRow represents one row in the matrix preview overview table.
type OverviewRow struct {
	Entry   string // entry name (e.g. "ghostty")
	Darwin  string // 3-symbol block e.g. "■ ■ □" or "· · ·"
	Linux   string // 3-symbol block e.g. "■ ■ □"
	Windows string // 3-symbol block e.g. "□ □ □"
	Targets string // description e.g. "2 targets (macOS, Linux)" or "1 target (all)"
}

// OverviewTable prints a formatted table of all entries across the matrix.
func (s Style) OverviewTable(w io.Writer, rows []OverviewRow) {
	header := "  ENTRY                 DARWIN    LINUX     WINDOWS   TARGETS"
	if s.D != "" {
		fmt.Fprintf(w, "%s%s%s\n", s.D, header, s.O)
	} else {
		fmt.Fprintln(w, header)
	}
	for _, r := range rows {
		entry := r.Entry
		if s.M != "" && !strings.Contains(entry, "\x1b[") {
			entry = s.M + entry + s.O
		}
		targets := r.Targets
		if s.D != "" && !strings.Contains(targets, "\x1b[") {
			targets = s.D + targets + s.O
		}
		fmt.Fprintf(w, "  %s%s%s%s%s\n",
			padRight(entry, 22),
			padRight(r.Darwin, 10),
			padRight(r.Linux, 10),
			padRight(r.Windows, 10),
			targets,
		)
	}
}

func visibleWidth(s string) int {
	inEscape := false
	w := 0
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		w++
	}
	return w
}

func padRight(s string, width int) string {
	vw := visibleWidth(s)
	if vw >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vw)
}
