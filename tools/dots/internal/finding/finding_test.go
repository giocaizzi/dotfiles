package finding

import "testing"

func TestLines(t *testing.T) {
	t.Parallel()
	got := Lines([]Finding{New("b", "x %d", 1), New("a", "y"), New("", "plain"), New("a", "y")})
	if want := "a: y\nb: x 1\nplain"; got != want {
		t.Errorf("Lines() = %q, want %q (sorted, deduplicated)", got, want)
	}
	if got := Lines(nil); got != "" {
		t.Errorf("Lines(nil) = %q, want empty", got)
	}
}
