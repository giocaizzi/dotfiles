package inventory

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/pkgreg"
)

const selJSON = `{
 "os": %q, "role": "personal",
 "release": {
  "zed": {"version": "1.0.0", "url": "https://github.com/o/zed/releases/download/v{version}/z.tgz", "target": {"linux/amd64": "L"}},
  "alpha": {"version": "2.0.0", "url": "u", "versionCmd": ["{dest}", "version"]},
  "skipped": {"version": "3.0.0", "url": "u"}
 },
 "releases": {"zed": {"dest": ".local/bin/zed"}, "alpha": {"dest": ".local/bin/alpha.exe"}},
 "brew": [{"name": "jq", "id": "jq", "entry": {"brew": "jq"}}, {"name": "fd", "id": "fd", "entry": {"brew": "fd", "cmd": "fd"}}],
 "cask": [{"name": "ghostty", "id": "ghostty", "entry": {"cask": "ghostty"}}],
 "apt": [{"name": "fd", "id": "fd-find", "entry": {"apt": "fd-find"}}],
 "winget": [{"name": "git", "id": "Git.Git", "entry": {"winget": "Git.Git"}}],
 "uv": [{"name": "pyright", "id": "pyright", "entry": {"uv": "pyright"}}],
 "npm": [{"name": "ts", "id": "typescript@5", "entry": {"npm": "typescript@5"}}],
 "script": [
  {"name": "bws", "id": "bws", "entry": {"script": ["linux"], "versionCmd": ["bws", "-V"]}},
  {"name": "brewed", "id": "brewed", "entry": {"script": ["darwin"], "brew": "brewed"}},
  {"name": "aptd", "id": "aptd", "entry": {"script": ["linux"], "apt": "aptd", "cmd": "ap"}}
 ]
}`

func TestParseSelection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		os   string
		want map[Source][]PackageSel
	}{
		{"darwin", map[Source][]PackageSel{
			Brew: {
				{Name: "fd", ID: "fd", Cmd: "fd", VersionCmd: []string{"fd", "--version"}},
				{Name: "ghostty", ID: "ghostty", Cask: true, Cmd: "ghostty", VersionCmd: []string{"ghostty", "--version"}},
				{Name: "jq", ID: "jq", Cmd: "jq", VersionCmd: []string{"jq", "--version"}},
			},
			UV:  {{Name: "pyright", ID: "pyright", Cmd: "pyright", VersionCmd: []string{"pyright", "--version"}}},
			Npm: {{Name: "ts", ID: "typescript@5", Cmd: "ts", VersionCmd: []string{"ts", "--version"}}},
			Script: {
				{Name: "aptd", ID: "aptd", Cmd: "ap", VersionCmd: []string{"ap", "--version"}},
				{Name: "bws", ID: "bws", Cmd: "bws", VersionCmd: []string{"bws", "-V"}},
			},
		}},
		{"linux", map[Source][]PackageSel{
			Apt: {{Name: "fd", ID: "fd-find", Cmd: "fd", VersionCmd: []string{"fd", "--version"}}},
			UV:  {{Name: "pyright", ID: "pyright", Cmd: "pyright", VersionCmd: []string{"pyright", "--version"}}},
			Npm: {{Name: "ts", ID: "typescript@5", Cmd: "ts", VersionCmd: []string{"ts", "--version"}}},
			Script: {
				{Name: "brewed", ID: "brewed", Cmd: "brewed", VersionCmd: []string{"brewed", "--version"}},
				{Name: "bws", ID: "bws", Cmd: "bws", VersionCmd: []string{"bws", "-V"}},
			},
		}},
		{"windows", map[Source][]PackageSel{
			Winget: {{Name: "git", ID: "Git.Git", Cmd: "git", VersionCmd: []string{"git", "--version"}}},
			UV:     {{Name: "pyright", ID: "pyright", Cmd: "pyright", VersionCmd: []string{"pyright", "--version"}}},
			Npm:    {{Name: "ts", ID: "typescript@5", Cmd: "ts", VersionCmd: []string{"ts", "--version"}}},
			Script: {
				{Name: "aptd", ID: "aptd", Cmd: "ap", VersionCmd: []string{"ap", "--version"}},
				{Name: "brewed", ID: "brewed", Cmd: "brewed", VersionCmd: []string{"brewed", "--version"}},
				{Name: "bws", ID: "bws", Cmd: "bws", VersionCmd: []string{"bws", "-V"}},
			},
		}},
		{"plan9", map[Source][]PackageSel{
			UV:  {{Name: "pyright", ID: "pyright", Cmd: "pyright", VersionCmd: []string{"pyright", "--version"}}},
			Npm: {{Name: "ts", ID: "typescript@5", Cmd: "ts", VersionCmd: []string{"ts", "--version"}}},
			Script: {
				{Name: "aptd", ID: "aptd", Cmd: "ap", VersionCmd: []string{"ap", "--version"}},
				{Name: "brewed", ID: "brewed", Cmd: "brewed", VersionCmd: []string{"brewed", "--version"}},
				{Name: "bws", ID: "bws", Cmd: "bws", VersionCmd: []string{"bws", "-V"}},
			},
		}},
	}
	wantReleases := []ReleaseSel{
		{
			Release:    pkgreg.Release{Name: "alpha", Version: "2.0.0", URL: "u", Targets: map[string]string{}},
			Dest:       ".local/bin/alpha.exe",
			VersionCmd: []string{"{dest}", "version"},
		},
		{
			Release:    pkgreg.Release{Name: "zed", Version: "1.0.0", URL: "https://github.com/o/zed/releases/download/v{version}/z.tgz", Targets: map[string]string{"linux/amd64": "L"}},
			Dest:       ".local/bin/zed",
			VersionCmd: []string{"{dest}", "--version"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.os, func(t *testing.T) {
			t.Parallel()
			got, err := ParseSelection([]byte(fmt.Sprintf(selJSON, tt.os)))
			if err != nil {
				t.Fatal(err)
			}
			if got.OS != tt.os || got.Role != "personal" {
				t.Errorf("OS, Role = %q, %q", got.OS, got.Role)
			}
			if diff := cmp.Diff(wantReleases, got.Releases); diff != "" {
				t.Errorf("Releases (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got.Packages); diff != "" {
				t.Errorf("Packages (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseSelectionNilSafe(t *testing.T) {
	t.Parallel()
	got, err := ParseSelection([]byte(`{"os":"linux"}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Releases) != 0 || len(got.Packages[Apt]) != 0 || len(got.Packages[Script]) != 0 {
		t.Errorf("ParseSelection(minimal) = %+v, want nothing selected", got)
	}
}

func TestParseSelectionErrors(t *testing.T) {
	t.Parallel()
	for _, in := range []string{"", "{", `{"role":"work"}`} {
		if _, err := ParseSelection([]byte(in)); err == nil {
			t.Errorf("ParseSelection(%q): want an error", in)
		}
	}
}
