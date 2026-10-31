package tree

import (
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
)

func fixture() fstest.MapFS {
	return fstest.MapFS{
		"dot_zshrc.tmpl":                    {Data: []byte(`{{ template "role" . }}{{ includeTemplate "agents/core" . }}{{ if .light }}x{{ end }}`)},
		".chezmoitemplates/role":            {Data: []byte(`{{ template "role" . }}`)}, // a cycle
		".chezmoitemplates/agents/core":     {Data: []byte(`{{ include "docs/x.md" }}`)},
		"docs/x.md":                         {Data: []byte("x")},
		".chezmoidata/roles.toml":           {Data: []byte("light = false\n[machine]\nwork = true\n")},
		".chezmoidata/.hidden.toml":         {Data: []byte("light = 1\n")},
		"dot_config/app/symlink_conf.tmpl":  {Data: []byte(`{{ .chezmoi.sourceDir }}/dot_config/app/conf`)},
		"dot_config/app/conf":               {Data: []byte("c")},
		".chezmoitemplates/agents/layer.md": {Data: []byte("l")},
	}
}

func TestOf(t *testing.T) {
	t.Parallel()
	got := New(fixture()).Of("dot_zshrc.tmpl")
	want := []Node{
		{0, "source", "dot_zshrc.tmpl"},
		{1, "template", ".chezmoitemplates/role"},
		{1, "template", ".chezmoitemplates/agents/core"},
		{2, "include", "docs/x.md"},
		{1, "data", ".chezmoidata/roles.toml"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Of() (-want +got):\n%s", diff)
	}
}

func TestOfSymlink(t *testing.T) {
	t.Parallel()
	got := New(fixture()).Of("dot_config/app/symlink_conf.tmpl")
	want := []Node{{0, "source", "dot_config/app/symlink_conf.tmpl"}, {1, "symlink", "dot_config/app/conf"}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Of() (-want +got):\n%s", diff)
	}
}

func TestOfMissingFile(t *testing.T) {
	t.Parallel()
	got := New(fixture()).Of("nope")
	if diff := cmp.Diff([]Node{{0, "source", "nope"}}, got); diff != "" {
		t.Errorf("Of(missing) (-want +got):\n%s", diff)
	}
}

func TestFeeds(t *testing.T) {
	t.Parallel()
	managed := map[string]string{"/h/.zshrc": "dot_zshrc.tmpl", "/h/.other": "docs/x.md", "/h/.conf": "dot_config/app/symlink_conf.tmpl"}
	got := New(fixture()).Feeds(managed, "docs/x.md")
	if diff := cmp.Diff([]string{"/h/.other", "/h/.zshrc"}, got); diff != "" {
		t.Errorf("Feeds() (-want +got):\n%s", diff)
	}
}

func TestTopKeys(t *testing.T) {
	t.Parallel()
	got := topKeys("a = 1\n[b.c]\nd = 2\n[[e]]\nf = 3\n")
	if diff := cmp.Diff([]string{"a", "b", "e"}, got); diff != "" {
		t.Errorf("topKeys() (-want +got):\n%s", diff)
	}
}
