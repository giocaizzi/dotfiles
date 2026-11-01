package term

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPalette(t *testing.T) {
	t.Parallel()
	s := Style{G: "green", M: "mauve", C: "cyan", Y: "yellow", B: "blue"}
	want := []string{"green", "mauve", "cyan", "yellow", "blue"}
	if diff := cmp.Diff(want, s.Palette()); diff != "" {
		t.Errorf("Palette() mismatch (-want +got):\n%s", diff)
	}
}

func TestMatrixGrid_NoColor(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	s := Style{}

	cell := func(osName, _ string) (string, string) {
		switch osName {
		case "darwin":
			return SymSquareFilled, ""
		case "linux":
			return SymSquareEmpty, ""
		case "windows":
			return SymDot, ""
		default:
			return SymCross, ""
		}
	}

	s.MatrixGrid(&buf, cell)

	want := "            personal    work     light\n" +
		"darwin         ■         ■         ■\n" +
		"linux          □         □         □\n" +
		"windows        ·         ·         ·\n"

	if diff := cmp.Diff(want, buf.String()); diff != "" {
		t.Errorf("MatrixGrid() mismatch (-want +got):\n%s", diff)
	}
}

func TestMatrixGrid_Colored(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	s := Style{G: "\x1b[32m", O: "\x1b[0m"}

	cell := func(_, _ string) (string, string) {
		return SymSquareFilled, s.G
	}

	s.MatrixGrid(&buf, cell)

	g := "\x1b[32m" + SymSquareFilled + "\x1b[0m"
	tail := g + "         " + g + "         " + g + "\n"
	want := "            personal    work     light\n" +
		"darwin         " + tail +
		"linux          " + tail +
		"windows        " + tail

	if diff := cmp.Diff(want, buf.String()); diff != "" {
		t.Errorf("MatrixGrid() colored mismatch (-want +got):\n%s", diff)
	}
}

func TestOverviewTable_NoColor(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	s := Style{}

	rows := []OverviewRow{
		{
			Entry:   "ghostty",
			Darwin:  "■ ■ □",
			Linux:   "■ ■ □",
			Windows: "· · ·",
			Targets: "2 targets (macOS, Linux)",
		},
		{
			Entry:   "git",
			Darwin:  "■ ■ ■",
			Linux:   "■ ■ ■",
			Windows: "■ ■ ■",
			Targets: "1 target (all)",
		},
	}

	s.OverviewTable(&buf, rows)

	want := "  ENTRY                 DARWIN    LINUX     WINDOWS   TARGETS\n" +
		"  ghostty               ■ ■ □     ■ ■ □     · · ·     2 targets (macOS, Linux)\n" +
		"  git                   ■ ■ ■     ■ ■ ■     ■ ■ ■     1 target (all)\n"

	if diff := cmp.Diff(want, buf.String()); diff != "" {
		t.Errorf("OverviewTable() mismatch (-want +got):\n%s", diff)
	}
}

func TestOverviewTable_Colored(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	s := Style{M: "\x1b[35m", D: "\x1b[2m", O: "\x1b[0m"}

	twoFilledOneEmpty := "\x1b[32m" + SymSquareFilled + "\x1b[0m \x1b[32m" + SymSquareFilled + "\x1b[0m \x1b[35m" + SymSquareEmpty + "\x1b[0m"
	rows := []OverviewRow{
		{
			Entry:   "ghostty",
			Darwin:  twoFilledOneEmpty,
			Linux:   twoFilledOneEmpty,
			Windows: "· · ·",
			Targets: "2 targets (macOS, Linux)",
		},
	}

	s.OverviewTable(&buf, rows)

	want := "\x1b[2m  ENTRY                 DARWIN    LINUX     WINDOWS   TARGETS\x1b[0m\n" +
		"  \x1b[35mghostty\x1b[0m               " + twoFilledOneEmpty + "     " + twoFilledOneEmpty + "     · · ·     \x1b[2m2 targets (macOS, Linux)\x1b[0m\n"

	if diff := cmp.Diff(want, buf.String()); diff != "" {
		t.Errorf("OverviewTable() colored mismatch (-want +got):\n%s", diff)
	}
}
