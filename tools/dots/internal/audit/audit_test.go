package audit

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestKnownAndUnknown(t *testing.T) {
	t.Parallel()
	sh := "alias gs='git status'\nalias ..='cd ..'\n# alias nope='x'\n"
	omz := "zstyle ':omz:plugins:git' aliases yes\nzstyle ':omz:plugins:docker' aliases no\n"
	if diff := cmp.Diff([]string{"git"}, Plugins(omz)); diff != "" {
		t.Errorf("Plugins() (-want +got):\n%s", diff)
	}
	plugin := "alias -g G='| grep'\nalias gco='git checkout'\n  alias 'gp'='git push'\n"
	if diff := cmp.Diff([]string{"G", "gco", "gp"}, OMZAliases(plugin)); diff != "" {
		t.Errorf("OMZAliases() (-want +got):\n%s", diff)
	}
	known := Known(sh, OMZAliases(plugin))
	got := Unknown(known, "gs\n..\nls\ngco\nrogue\nzzz\nrogue\n")
	if diff := cmp.Diff([]string{"rogue", "zzz"}, got); diff != "" {
		t.Errorf("Unknown() (-want +got):\n%s", diff)
	}
}
