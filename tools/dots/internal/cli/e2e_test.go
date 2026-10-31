package cli_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/cli"
)

// These tests drive the real command line against a real chezmoi and a small fixture
// source; they skip when chezmoi is not installed.

const configTemplate = `{{- $roles := dict -}}
{{- range list "light" "work" }}
{{- if kindIs "bool" (get $ .) }}{{ $_ := set $roles . (ternary "yes" "no" (get $ .)) }}{{ end }}
{{- end }}
{{- $light := eq (promptChoiceOnce $roles "light" "Light?" (list "yes" "no") (get $roles "light" | default "no")) "yes" -}}
{{- $work := eq (promptChoiceOnce $roles "work" "Work?" (list "yes" "no") (get $roles "work" | default "no")) "yes" -}}
{{- $user := promptStringOnce . "gitUser" "Name" "" -}}
{{- $harnesses := list -}}
{{- if not $light }}
{{- $all := list "claude" "vscode" -}}
{{- $harnesses = promptMultichoiceOnce . "harnesses" "Harnesses" $all (get . "harnesses" | default list) -}}
{{- end }}
[data]
    gitUser = {{ $user | quote }}
    light = {{ $light }}
    work = {{ $work }}
{{- if not $light }}
    harnesses = {{ $harnesses | toJson }}
{{- end }}
`

const registry = `[edit.note]
desc = "A note"
source = "dot_note.tmpl"
check = ["chezmoi", "execute-template", "--file", "{file}"]
note = "reload the note"

[edit.registry]
desc = "The registry"
source = ".chezmoidata/edit.toml"
`

type env struct {
	src, dest string
	vars      map[string]string
}

func newEnv(t *testing.T) *env {
	t.Helper()
	if _, err := exec.LookPath("chezmoi"); err != nil {
		t.Skip("chezmoi is not installed")
	}
	root := t.TempDir()
	e := &env{src: filepath.Join(root, "src"), dest: filepath.Join(root, "dest")}
	write := func(path, text string) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(e.src, ".chezmoi.toml.tmpl"), configTemplate)
	write(filepath.Join(e.src, ".chezmoidata", "edit.toml"), registry)
	write(filepath.Join(e.src, "dot_note.tmpl"), "hello {{ if .work }}work{{ else }}home{{ end }}\n")
	cfg := filepath.Join(root, "chezmoi.toml")
	write(cfg, "destDir = "+quote(e.dest)+"\npersistentState = "+quote(filepath.Join(root, "state.boltdb"))+"\n[data]\n    work = false\n    light = false\n    harnesses = []\n")
	if err := os.MkdirAll(e.dest, 0o755); err != nil {
		t.Fatal(err)
	}
	e.vars = map[string]string{
		"CHEZMOI_EXECUTABLE":  "chezmoi",
		"CHEZMOI_SOURCE_DIR":  e.src,
		"CHEZMOI_CONFIG_FILE": cfg,
		"HOME":                root,
	}
	return e
}

func quote(p string) string { return `"` + filepath.ToSlash(p) + `"` }

func (e *env) run(t *testing.T, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errw bytes.Buffer
	// A file, as on a terminal: the editor subprocess then shares it instead of
	// draining a reader that the confirmation prompt needs afterwards.
	pr, pw, perr := os.Pipe()
	if perr != nil {
		t.Fatal(perr)
	}
	t.Cleanup(func() { _ = pr.Close() })
	if _, perr = pw.WriteString(stdin); perr != nil {
		t.Fatal(perr)
	}
	_ = pw.Close()
	d := cli.Deps{
		Getenv: func(k string) string { return e.vars[k] },
		Stdin:  pr, Stdout: &out, Stderr: &errw,
		Runner: chez.ExecRunner{}, LookPath: exec.LookPath,
		GOOS: runtime.GOOS, Home: e.vars["HOME"], Version: "test",
	}
	err = cli.Run(t.Context(), d, args)
	return out.String(), errw.String(), err
}

// editor makes the test binary the editor: it appends text to the file it is given.
func (e *env) editor(t *testing.T, text string) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	e.vars["EDITOR"] = "'" + exe + "' -test.run=TestHelperEditor -- e2e-helper '" + text + "'"
}

// TestHelperEditor is the editor of the edit tests, run as a subprocess.
func TestHelperEditor(t *testing.T) {
	t.Parallel()
	n := len(os.Args)
	if n < 3 || os.Args[n-3] != "e2e-helper" {
		t.Skip("helper process")
	}
	path := os.Args[n-1]
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(os.Args[n-2]); err != nil {
		t.Fatal(err)
	}
}

func TestListShowsTheRegistry(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	out, _, err := e.run(t, "", "list")
	if err != nil || !strings.Contains(out, "note") || !strings.Contains(out, "A note") {
		t.Errorf("list = %q, %v, want the note entry", out, err)
	}
}

func TestViewRendersAConfiguration(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	out, _, err := e.run(t, "", "view", filepath.Join(e.dest, ".note"), "--as", "linux,work")
	if err != nil || !strings.Contains(out, "hello work") {
		t.Errorf("view --as linux,work = %q, %v, want the work render", out, err)
	}
}

func TestEditAppliesWhenConfirmed(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.editor(t, "more")
	_, stderr, err := e.run(t, "y\n", "edit", "note")
	if err != nil {
		t.Fatalf("edit = %v\n%s", err, stderr)
	}
	got, _ := os.ReadFile(filepath.Join(e.dest, ".note"))
	if string(got) != "hello home\nmore" {
		t.Errorf("applied note = %q, want the edited render", got)
	}
	src, _ := os.ReadFile(filepath.Join(e.src, "dot_note.tmpl"))
	if !strings.HasSuffix(string(src), "more") {
		t.Errorf("source = %q, want the edit", src)
	}
	for _, want := range []string{"linux/work", "reload the note"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr = %q, want it to mention %q", stderr, want)
		}
	}
}

func TestEditDeclinedLeavesTheTarget(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.editor(t, "more")
	if _, stderr, err := e.run(t, "n\n", "edit", "note"); err != nil {
		t.Fatalf("edit = %v\n%s", err, stderr)
	}
	if _, err := os.Stat(filepath.Join(e.dest, ".note")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("target exists after declining (stat error %v)", err)
	}
}

func TestEditWithABrokenTemplateAppliesNothing(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.editor(t, "{{ if }}")
	_, stderr, err := e.run(t, "y\n", "edit", "note")
	if !errors.Is(err, cli.ErrReported) {
		t.Fatalf("edit = %v, want ErrReported\n%s", err, stderr)
	}
	if _, err := os.Stat(filepath.Join(e.dest, ".note")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("target exists after a failed check (stat error %v)", err)
	}
	if !strings.Contains(stderr, "nothing applied") {
		t.Errorf("stderr = %q, want the nothing-applied notice", stderr)
	}
}

func TestLintRegistryOnTheFixture(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	if _, stderr, err := e.run(t, "", "lint", "registry"); err != nil {
		t.Errorf("lint registry = %v\n%s", err, stderr)
	}
}
