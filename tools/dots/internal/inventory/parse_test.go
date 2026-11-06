package inventory

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseVersion(t *testing.T) {
	t.Parallel()
	tests := []struct{ in, want string }{
		{"go version go1.27.3 darwin/arm64", "1.27.3"},
		{"jq-1.7.1", "1.7.1"},
		{"bws 2.1.0\n", "2.1.0"},
		{"no digits", ""},
		{"v7", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := ParseVersion(tt.in); got != tt.want {
			t.Errorf("ParseVersion(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestBrewList(t *testing.T) {
	t.Parallel()
	got := BrewList("jq 1.7.1\nopenssl@3 3.3.0 3.4.1\n\nlonely\n")
	want := map[string]string{"jq": "1.7.1", "openssl@3": "3.4.1"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BrewList() (-want +got):\n%s", diff)
	}
}

func TestBrewOutdated(t *testing.T) {
	t.Parallel()
	got, err := BrewOutdated([]byte(`{"formulae":[{"name":"jq","current_version":"1.8"}],"casks":[{"name":"ghostty","current_version":"1.3"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(map[string]string{"jq": "1.8", "ghostty": "1.3"}, got); diff != "" {
		t.Errorf("BrewOutdated() (-want +got):\n%s", diff)
	}
	if _, err := BrewOutdated([]byte("nope")); err == nil {
		t.Error("BrewOutdated(invalid): want an error")
	}
}

func TestAptPolicy(t *testing.T) {
	t.Parallel()
	out := `fd-find:
  Installed: 9.0.0-1
  Candidate: 9.0.0-2
  Version table:
 *** 9.0.0-1 500
bat:
  Installed: (none)
  Candidate: 0.24.0-1
`
	want := map[string]AptVersions{
		"fd-find": {Installed: "9.0.0-1", Candidate: "9.0.0-2"},
		"bat":     {Candidate: "0.24.0-1"},
	}
	if diff := cmp.Diff(want, AptPolicy(out)); diff != "" {
		t.Errorf("AptPolicy() (-want +got):\n%s", diff)
	}
}

func TestWingetExport(t *testing.T) {
	t.Parallel()
	got, err := WingetExport([]byte(`{"Sources":[{"Packages":[{"PackageIdentifier":"Git.Git","Version":"2.50.0"}]},{"Packages":[{"PackageIdentifier":"JQ.JQ","Version":"1.7"}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(map[string]string{"git.git": "2.50.0", "jq.jq": "1.7"}, got); diff != "" {
		t.Errorf("WingetExport() (-want +got):\n%s", diff)
	}
	if _, err := WingetExport([]byte("{")); err == nil {
		t.Error("WingetExport(invalid): want an error")
	}
}

func TestUVList(t *testing.T) {
	t.Parallel()
	got := UVList("pyright v1.1.414\n- pyright\n- pyright-langserver\n\nruff v0.9.1\n")
	if diff := cmp.Diff(map[string]string{"pyright": "1.1.414", "ruff": "0.9.1"}, got); diff != "" {
		t.Errorf("UVList() (-want +got):\n%s", diff)
	}
}

func TestUVOutdated(t *testing.T) {
	t.Parallel()
	got := UVOutdated("pyright v1.1.400 [latest: 1.1.405]\nruff v0.9.1\n")
	if diff := cmp.Diff(map[string]string{"pyright": "1.1.405"}, got); diff != "" {
		t.Errorf("UVOutdated() (-want +got):\n%s", diff)
	}
}

func TestNpm(t *testing.T) {
	t.Parallel()
	ls, err := NpmLs([]byte(`{"dependencies":{"typescript":{"version":"5.9.2"},"@a/b":{"version":"2.0.0"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(map[string]string{"typescript": "5.9.2", "@a/b": "2.0.0"}, ls); diff != "" {
		t.Errorf("NpmLs() (-want +got):\n%s", diff)
	}
	out, err := NpmOutdated([]byte(`{"typescript":{"current":"5.9.2","wanted":"5.9.3","latest":"6.0.1"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(map[string]string{"typescript": "6.0.1"}, out); diff != "" {
		t.Errorf("NpmOutdated() (-want +got):\n%s", diff)
	}
	empty, err := NpmOutdated(nil)
	if err != nil || len(empty) != 0 {
		t.Errorf("NpmOutdated(empty) = %v, %v, want an empty map", empty, err)
	}
	if _, err := NpmLs([]byte("x")); err == nil {
		t.Error("NpmLs(invalid): want an error")
	}
	if _, err := NpmOutdated([]byte("x")); err == nil {
		t.Error("NpmOutdated(invalid): want an error")
	}
}

func TestNpmSpec(t *testing.T) {
	t.Parallel()
	tests := []struct{ id, name, pin string }{
		{"typescript@5", "typescript", "5"},
		{"@scope/pkg@2", "@scope/pkg", "2"},
		{"@scope/pkg", "@scope/pkg", ""},
		{"plain", "plain", ""},
	}
	for _, tt := range tests {
		name, pin := NpmSpec(tt.id)
		if name != tt.name || pin != tt.pin {
			t.Errorf("NpmSpec(%q) = %q, %q, want %q, %q", tt.id, name, pin, tt.name, tt.pin)
		}
	}
}

func TestUVName(t *testing.T) {
	t.Parallel()
	tests := []struct{ id, want string }{
		{"pyright", "pyright"},
		{"pyright[extra]", "pyright"},
		{"ruff>=0.5", "ruff"},
		{"tool==1.0", "tool"},
		{"tool~=1.0", "tool"},
		{"tool@latest", "tool"},
		{"tool ; python_version>'3'", "tool"},
		{"tool<2", "tool"},
		{"tool!=1", "tool"},
	}
	for _, tt := range tests {
		if got := UVName(tt.id); got != tt.want {
			t.Errorf("UVName(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestGitCounts(t *testing.T) {
	t.Parallel()
	if a, b, ok := GitCounts("2\t5\n"); a != 2 || b != 5 || !ok {
		t.Errorf("GitCounts() = %d, %d, %v, want 2, 5, true", a, b, ok)
	}
	for _, bad := range []string{"", "1", "a b", "1 2 3"} {
		if _, _, ok := GitCounts(bad); ok {
			t.Errorf("GitCounts(%q) ok, want false", bad)
		}
	}
}
