package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/cli"
)

func TestOutdatedUsage(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{
		{"outdated", "brew"},
		{"outdated", "--offline", "extra"},
		{"outdated", "--nope"},
	} {
		d := cli.Deps{Getenv: func(string) string { return "" }, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}
		if err := cli.Run(t.Context(), d, args); !errors.Is(err, cli.ErrUsage) {
			t.Errorf("Run(%v) error = %v, want ErrUsage", args, err)
		}
	}
}

func TestHelpMentionsOutdatedFlags(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	d := cli.Deps{Getenv: func(string) string { return "" }, Stdout: &out, Stderr: &bytes.Buffer{}}
	if err := cli.Run(t.Context(), d, []string{"help"}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"outdated [--offline] [--fetch] [--check]", "--check", "Every tool: installed, pinned and latest versions"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("help lacks %q", want)
		}
	}
}
