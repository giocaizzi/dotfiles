// Package upgrade plans and runs the update of everything a machine installs:
// the dotfiles, the system packages, the language tools.
package upgrade

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
)

// Step is one update: a name, why it exists and the commands that do it.
type Step struct {
	Name string
	Desc string
	Cmds [][]string
	Skip string // non-empty: why the step does not apply here
}

// Env is what the plan depends on.
type Env struct {
	GOOS     string
	Home     string
	Bin      string // the chezmoi binary
	LookPath func(string) (string, error)
}

// Plan lists the steps for this machine, in order.
func Plan(e Env) []Step {
	steps := []Step{{
		Name: "dotfiles", Desc: "pull the dotfiles and apply them (refreshes due externals)",
		Cmds: [][]string{{e.Bin, "update", "--refresh-externals"}},
	}}
	switch e.GOOS {
	case "darwin":
		steps = append(steps, Step{
			Name: "brew", Desc: "Homebrew formulae and casks",
			Cmds: [][]string{{"brew", "update"}, {"brew", "upgrade"}, {"brew", "upgrade", "--cask"}},
		})
	case "linux":
		steps = append(steps, Step{
			Name: "apt", Desc: "apt packages",
			Cmds: [][]string{{"sudo", "apt-get", "update"}, {"sudo", "apt-get", "upgrade", "-y"}},
		})
	case "windows":
		steps = append(steps, Step{
			Name: "winget", Desc: "winget packages",
			Cmds: [][]string{{"winget", "upgrade", "--all", "--silent", "--accept-package-agreements", "--accept-source-agreements"}},
		})
	}
	uv := Step{Name: "uv", Desc: "uv tools (and uv itself where its own installer put it)", Cmds: [][]string{{"uv", "tool", "upgrade", "--all"}}}
	if e.GOOS != "darwin" {
		uv.Cmds = append([][]string{{"uv", "self", "update"}}, uv.Cmds...)
	}
	npm := []string{"npm", "update", "-g"}
	if e.GOOS == "linux" {
		npm = append(npm, "--prefix", filepath.Join(e.Home, ".local"))
	}
	steps = append(steps, uv, Step{Name: "npm", Desc: "global npm tools", Cmds: [][]string{npm}})
	if e.GOOS == "windows" {
		steps = append(steps, Step{
			Name: "psmodules", Desc: "PowerShell modules",
			Cmds: [][]string{{"pwsh", "-NoProfile", "-Command", "Update-Module"}},
		})
	}
	for i := range steps {
		if _, err := e.LookPath(steps[i].Cmds[0][0]); err != nil {
			steps[i].Skip = steps[i].Cmds[0][0] + " is not installed"
		}
	}
	return steps
}

// Names lists the step names.
func Names(steps []Step) []string {
	out := make([]string, len(steps))
	for i, s := range steps {
		out[i] = s.Name
	}
	return out
}

// Select keeps the steps named in only (all when empty); an unknown name is an error.
func Select(steps []Step, only []string) ([]Step, error) {
	if len(only) == 0 {
		return steps, nil
	}
	for _, o := range only {
		if !slices.Contains(Names(steps), o) {
			return nil, fmt.Errorf("unknown step %q (steps: %v)", o, Names(steps))
		}
	}
	var out []Step
	for _, s := range steps {
		if slices.Contains(only, s.Name) {
			out = append(out, s)
		}
	}
	return out, nil
}

// Run executes the steps, printing each. A failing step does not stop the others;
// the returned error joins every failure. With dryRun nothing runs.
func Run(ctx context.Context, r chez.Runner, steps []Step, dryRun bool, out, errw io.Writer) error {
	var errs []error
	for _, s := range steps {
		if s.Skip != "" {
			fmt.Fprintf(out, "==> %s: skipped, %s\n", s.Name, s.Skip)
			continue
		}
		fmt.Fprintf(out, "==> %s: %s\n", s.Name, s.Desc)
		for _, c := range s.Cmds {
			fmt.Fprintf(out, "    $ %v\n", c)
			if dryRun {
				continue
			}
			if err := r.Run(ctx, chez.Cmd{Args: c, Stdin: nil, Stdout: out, Stderr: errw}); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", s.Name, err))
				break
			}
		}
	}
	return errors.Join(errs...)
}
