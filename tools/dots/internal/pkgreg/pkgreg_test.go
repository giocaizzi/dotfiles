package pkgreg

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/finding"
)

func TestCompare(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b string
		sign int
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.2.3", "1.2.10", -1},
		{"2.0", "1.99.9", 1},
		{"1.2", "1.2.0", 0},
		{"1.27.3", "1.27", 1},
	}
	for _, tt := range tests {
		t.Run(tt.a+"/"+tt.b, func(t *testing.T) {
			t.Parallel()
			got := Compare(tt.a, tt.b)
			if (got > 0) != (tt.sign > 0) || (got < 0) != (tt.sign < 0) {
				t.Errorf("Compare(%q, %q) = %d, want sign %d", tt.a, tt.b, got, tt.sign)
			}
		})
	}
}

func TestUpstream(t *testing.T) {
	t.Parallel()
	tests := []struct {
		url     string
		want    Upstream
		wantErr bool
	}{
		{"https://github.com/o/r/releases/download/v{version}/x_{target}.tgz", Upstream{"git", "https://github.com/o/r.git", "v"}, false},
		{"https://github.com/o/r/releases/download/bws-v{version}/x.zip", Upstream{"git", "https://github.com/o/r.git", "bws-v"}, false},
		{"https://gitlab.com/g/sub/cli/-/releases/v{version}/downloads/x.tgz", Upstream{"git", "https://gitlab.com/g/sub/cli.git", "v"}, false},
		{"https://go.dev/dl/go{version}.{target}.tar.gz", Upstream{Kind: "godev"}, false},
		{"https://example.com/{version}", Upstream{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel()
			got, err := Release{Name: "x", URL: tt.url}.Upstream()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Upstream() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Upstream() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReleases(t *testing.T) {
	t.Parallel()
	data := map[string]any{"release": map[string]any{
		"b": map[string]any{"version": "1", "url": "u", "target": map[string]any{"linux/amd64": "L", "darwin/arm64": "D"}},
		"a": map[string]any{"version": "2", "url": "v"},
	}}
	got := Releases(data)
	if len(got) != 2 || got[0].Name != "a" || got[1].Name != "b" {
		t.Fatalf("Releases() = %v, want a then b", got)
	}
	if diff := cmp.Diff([]string{"darwin/arm64", "linux/amd64"}, got[1].Platforms()); diff != "" {
		t.Errorf("Platforms() (-want +got):\n%s", diff)
	}
	r := Release{URL: "https://h/{version}/x-{target}.tgz", Targets: map[string]string{"linux/amd64": "L"}}
	if g, w := r.AssetURL("3", "linux/amd64"), "https://h/3/x-L.tgz"; g != w {
		t.Errorf("AssetURL() = %q, want %q", g, w)
	}
}

const sum = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func validData() map[string]any {
	return map[string]any{
		"release": map[string]any{"x": map[string]any{
			"version": "1.0.0",
			"url":     "https://github.com/o/r/releases/download/v{version}/x-{target}.tgz",
			"target":  map[string]any{"linux/amd64": "L"},
			"sha256":  map[string]any{"linux/amd64": sum},
		}},
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()
	if got := finding.Lines(Validate(validData())); got != "" {
		t.Fatalf("Validate(valid) = %q, want none", got)
	}
	tests := []struct {
		name   string
		mutate func(r map[string]any)
		want   string
	}{
		{"bad sum", func(r map[string]any) { r["sha256"] = map[string]any{"linux/amd64": "zz"} }, "sha256"},
		{"orphan sum", func(r map[string]any) {
			r["sha256"] = map[string]any{"linux/amd64": sum, "darwin/arm64": sum}
		}, "sha256 darwin/arm64 has no target"},
		{"missing sum", func(r map[string]any) { r["sha256"] = map[string]any{} }, "linux/amd64"},
		{"bad platform", func(r map[string]any) { r["target"] = map[string]any{"plan9/386": "P"} }, "must be <os>/<arch>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := validData()
			tt.mutate(d["release"].(map[string]any)["x"].(map[string]any))
			if got := finding.Lines(Validate(d)); !strings.Contains(got, tt.want) {
				t.Errorf("Validate() = %q, want it to contain %q", got, tt.want)
			}
		})
	}
}

const registry = `# releases
[release.x]
version = "1.0.0"  # pinned
url = "u"

[release.x.sha256]
"linux/amd64" = "aaa"  # keep
"darwin/arm64" = "bbb"

[release.y]
version = "9.9.9"
`

func TestSetRelease(t *testing.T) {
	t.Parallel()
	got, err := SetRelease(registry, "x", "2.0.0", map[string]string{"linux/amd64": "n1", "darwin/arm64": "n2"})
	if err != nil {
		t.Fatal(err)
	}
	want := strings.NewReplacer(`"1.0.0"`, `"2.0.0"`, `"aaa"`, `"n1"`, `"bbb"`, `"n2"`).Replace(registry)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("SetRelease() (-want +got):\n%s", diff)
	}
	if !strings.Contains(got, `version = "9.9.9"`) {
		t.Error("SetRelease() touched another release")
	}
}

func TestSetReleaseErrors(t *testing.T) {
	t.Parallel()
	if _, err := SetRelease(registry, "nope", "1", nil); err == nil {
		t.Error("unknown release: want an error")
	}
	if _, err := SetRelease(registry, "x", "2", map[string]string{"windows/amd64": "s"}); err == nil {
		t.Error("checksum without a line: want an error")
	}
}
