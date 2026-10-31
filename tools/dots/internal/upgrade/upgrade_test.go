package upgrade

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
)

func env(goos string, missing ...string) Env {
	return Env{GOOS: goos, Home: "/h", Bin: "chezmoi", LookPath: func(n string) (string, error) {
		if slices.Contains(missing, n) {
			return "", errors.New("not found")
		}
		return "/bin/" + n, nil
	}}
}

func TestPlanPerOS(t *testing.T) {
	t.Parallel()
	tests := []struct {
		goos string
		want []string
	}{
		{"darwin", []string{"dotfiles", "brew", "uv", "npm"}},
		{"linux", []string{"dotfiles", "apt", "uv", "npm"}},
		{"windows", []string{"dotfiles", "winget", "uv", "npm", "psmodules"}},
	}
	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, Names(Plan(env(tt.goos)))); diff != "" {
				t.Errorf("Plan() names (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlanSkipsMissingTools(t *testing.T) {
	t.Parallel()
	for _, s := range Plan(env("darwin", "brew")) {
		if (s.Name == "brew") != (s.Skip != "") {
			t.Errorf("step %s skip = %q", s.Name, s.Skip)
		}
	}
}

func TestPlanLinuxNpmPrefix(t *testing.T) {
	t.Parallel()
	for _, s := range Plan(env("linux")) {
		if s.Name == "npm" && !strings.Contains(strings.Join(s.Cmds[0], " "), "--prefix /h/.local") {
			t.Errorf("npm cmd = %v, want the user prefix", s.Cmds[0])
		}
	}
}

func TestSelect(t *testing.T) {
	t.Parallel()
	steps := Plan(env("darwin"))
	got, err := Select(steps, []string{"npm", "brew"})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"brew", "npm"}, Names(got)); diff != "" {
		t.Errorf("Select() keeps plan order (-want +got):\n%s", diff)
	}
	if _, err := Select(steps, []string{"nope"}); err == nil {
		t.Error("Select(nope): want an error")
	}
}

type fakeRun struct {
	failOn string
	ran    []string
}

func (f *fakeRun) Run(_ context.Context, c chez.Cmd) error {
	f.ran = append(f.ran, c.Args[0])
	if c.Args[0] == f.failOn {
		return errors.New("exit 1")
	}
	return nil
}

func TestRun(t *testing.T) {
	t.Parallel()
	steps := Plan(env("darwin", "uv"))
	r := &fakeRun{failOn: "brew"}
	var out, errw bytes.Buffer
	err := Run(t.Context(), r, steps, false, &out, &errw)
	if err == nil || !strings.Contains(err.Error(), "brew") {
		t.Errorf("Run() = %v, want the brew failure", err)
	}
	// brew stops at its first failing command; dotfiles and npm still run; uv is skipped.
	if diff := cmp.Diff([]string{"chezmoi", "brew", "npm"}, r.ran); diff != "" {
		t.Errorf("commands run (-want +got):\n%s", diff)
	}
	if !strings.Contains(out.String(), "uv: skipped") {
		t.Errorf("output = %q, want the skipped step named", out.String())
	}
}

func TestRunDryRunRunsNothing(t *testing.T) {
	t.Parallel()
	r := &fakeRun{}
	var out bytes.Buffer
	if err := Run(t.Context(), r, Plan(env("linux")), true, &out, &out); err != nil {
		t.Fatal(err)
	}
	if len(r.ran) != 0 {
		t.Errorf("ran %v in a dry run", r.ran)
	}
}
