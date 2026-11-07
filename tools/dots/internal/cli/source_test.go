package cli_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/cli"
)

// fakeChezmoi answers `chezmoi data` with no releases and `source-path` with path
// (an error when empty), and records the subcommands it was asked for.
type fakeChezmoi struct {
	path string
	mu   sync.Mutex
	subs []string
}

func (f *fakeChezmoi) Run(_ context.Context, c chez.Cmd) error {
	sub := ""
	for _, a := range c.Args[1:] {
		if a == "data" || a == "source-path" {
			sub = a
		}
	}
	f.mu.Lock()
	f.subs = append(f.subs, sub)
	f.mu.Unlock()
	switch sub {
	case "data":
		_, err := fmt.Fprint(c.Stdout, `{}`)
		return err
	case "source-path":
		if f.path == "" {
			return errors.New(`exec: "chezmoi": executable file not found in $PATH`)
		}
		_, err := fmt.Fprintln(c.Stdout, f.path)
		return err
	}
	return fmt.Errorf("unexpected chezmoi call %v", c.Args)
}

// repo makes a repo whose .chezmoiroot points at home/, which holds the registry.
func repo(t *testing.T) (root, home string) {
	t.Helper()
	root = t.TempDir()
	home = filepath.Join(root, "home")
	if err := os.MkdirAll(filepath.Join(home, ".chezmoidata"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".chezmoiroot"), []byte("home\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".chezmoidata", "packages.toml"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	return root, home
}

func TestBumpSourceDir(t *testing.T) {
	t.Parallel()
	root, home := repo(t)
	tests := []struct {
		name     string
		env      string
		path     string
		wantSubs []string
		wantErr  string
	}{
		{name: "env names the source state", env: home, wantSubs: []string{"data"}},
		{name: "env names the repo root", env: root, wantSubs: []string{"data"}},
		{name: "env unset asks chezmoi", path: home, wantSubs: []string{"source-path", "data"}},
		{name: "env unset without chezmoi", wantSubs: []string{"source-path"}, wantErr: "find the chezmoi source dir"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fake := &fakeChezmoi{path: tt.path}
			var out bytes.Buffer
			d := cli.Deps{
				Getenv: func(k string) string {
					if k == "CHEZMOI_SOURCE_DIR" {
						return tt.env
					}
					return ""
				},
				Stdout: &out, Stderr: &bytes.Buffer{}, Runner: fake,
			}
			err := cli.Run(t.Context(), d, []string{"bump"})
			if (err != nil) != (tt.wantErr != "") || (err != nil && !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("bump error = %v, want %q", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.wantSubs, fake.subs); diff != "" {
				t.Errorf("chezmoi calls (-want +got):\n%s", diff)
			}
			if err == nil && !strings.Contains(out.String(), "all releases are up to date\n") {
				t.Errorf("bump output = %q, want it to read the registry", out.String())
			}
		})
	}
}
