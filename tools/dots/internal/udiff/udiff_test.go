package udiff

import (
	"strings"
	"testing"
)

func TestDiff(t *testing.T) {
	t.Parallel()
	if got := Diff("a", "b", "x\ny\n", "x\ny\n"); got != "" {
		t.Errorf("Diff(equal) = %q, want empty", got)
	}
	got := Diff("old", "new", "1\n2\n3\n4\n5\n6\n7\n8\n9\n", "1\n2\n3\n4\nFIVE\n6\n7\n8\n9\n")
	for _, want := range []string{"--- old\n+++ new\n", "-5\n", "+FIVE\n", " 4\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("Diff() = %q, want it to contain %q", got, want)
		}
	}
	if strings.Contains(got, " 1\n") {
		t.Errorf("Diff() = %q, want context limited to 3 lines", got)
	}
}
