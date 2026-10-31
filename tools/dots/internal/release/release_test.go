package release

import "testing"

func TestNewestTag(t *testing.T) {
	t.Parallel()
	ls := "a\trefs/tags/v1.2.0\nb\trefs/tags/v1.10.0\nc\trefs/tags/v1.9.0\nd\trefs/tags/v1.10.0-rc1\ne\trefs/tags/other-v9.0.0\nf\trefs/heads/main\n"
	got, err := newestTag(ls, "v")
	if err != nil || got != "1.10.0" {
		t.Errorf("newestTag() = %q, %v, want 1.10.0", got, err)
	}
	if _, err := newestTag(ls, "nope-v"); err == nil {
		t.Error("newestTag() without a matching tag: want an error")
	}
}
