package matrix

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      string
		want    Config
		wantErr bool
	}{
		{"darwin", Config{"darwin", "personal"}, false},
		{"linux,work", Config{"linux", "work"}, false},
		{"windows/light", Config{"windows", "light"}, false},
		{"plan9", Config{}, true},
		{"linux,boss", Config{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, err := Parse(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestAll(t *testing.T) {
	t.Parallel()
	all := All()
	if len(all) != 9 || all[0].Label() != "darwin/personal" || all[8].Label() != "windows/light" {
		t.Errorf("All() = %v, want 9 configurations from darwin/personal to windows/light", all)
	}
}

func TestHarnesses(t *testing.T) {
	t.Parallel()
	got, err := Harnesses([]byte(`{{- $all := list "claude" "copilot" "vscode" -}}`))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"claude", "copilot", "vscode"}, got); diff != "" {
		t.Errorf("Harnesses() (-want +got):\n%s", diff)
	}
	if _, err := Harnesses([]byte("nothing")); err == nil {
		t.Error("Harnesses() without a list: want an error")
	}
}

func TestDataKeys(t *testing.T) {
	t.Parallel()
	tmpl := "x = 1\n[data]\n    gitUser = {{ $u }}\n    light = {{ $l }}\n{{- if x }}\n    harnesses = []\n{{- end }}\n\n[git]\n    autoAdd = true\n"
	if diff := cmp.Diff([]string{"gitUser", "harnesses", "light"}, DataKeys([]byte(tmpl))); diff != "" {
		t.Errorf("DataKeys() (-want +got):\n%s", diff)
	}
}

func TestSeed(t *testing.T) {
	t.Parallel()
	local := map[string]any{"gitUser": "me", "gitEmail": "", "githubUser": "gh"}
	got := Seed("work", local, []string{"claude"})
	want := map[string]any{
		"light": false, "work": true, "harnesses": []string{"claude"}, "bwsProjectId": "",
		"gitUser": "me", "gitEmail": "<gitEmail>", "workEmail": "<workEmail>",
		"githubUser": "gh", "personalGitRemote": "", "otelEndpoint": "",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Seed() (-want +got):\n%s", diff)
	}
}

func TestTOML(t *testing.T) {
	t.Parallel()
	got := TOML(map[string]any{"b": true, "a": `x"y`, "c": []string{"p", "q"}})
	want := "[data]\na = \"x\\\"y\"\nb = true\nc = [\"p\", \"q\"]\n"
	if got != want {
		t.Errorf("TOML() = %q, want %q", got, want)
	}
}
