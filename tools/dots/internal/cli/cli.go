// Package cli is the command line of `chezmoi dots`: it parses arguments, wires the
// adapters into the use cases and prints their results. The logic lives elsewhere.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/edit"
)

// Deps is everything the commands touch outside the process: tests replace it.
type Deps struct {
	Getenv   func(string) string
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
	Runner   chez.Runner
	LookPath func(string) (string, error)
	HTTP     *http.Client
	Exe      string // this program's path, re-run for the live pane
	Version  string
	GOOS     string
	Home     string
}

// ErrReported is returned for a failure that was already printed.
var ErrReported = edit.ErrReported

// ErrUsage is returned for a command line that makes no sense.
var ErrUsage = edit.ErrUsage

// Run runs one command line (the arguments after `chezmoi dots`).
func Run(ctx context.Context, d Deps, args []string) error {
	cmd, rest := "edit", args
	if len(args) > 0 {
		switch args[0] {
		case "edit", "view", "list", "lint", "outdated", "bump", "upgrade", "live", "version", "help":
			cmd, rest = args[0], args[1:]
		case "-h", "--help":
			cmd = "help"
		}
	}
	switch cmd {
	case "edit":
		return d.edit(ctx, rest)
	case "view":
		return d.view(ctx, rest)
	case "list":
		return d.list(ctx, rest)
	case "lint":
		return d.lint(ctx, rest)
	case "outdated":
		return d.outdated(ctx, rest)
	case "bump":
		return d.bump(ctx, rest)
	case "upgrade":
		return d.upgrade(ctx, rest)
	case "live":
		return d.live(ctx, rest)
	case "version":
		fmt.Fprintln(d.Stdout, "chezmoi-dots", d.Version)
		return nil
	default:
		fmt.Fprint(d.Stdout, usage())
		return nil
	}
}

// parse parses flags that may come before or after the positional arguments.
func parse(fs *flag.FlagSet, args []string) ([]string, error) {
	fs.SetOutput(io.Discard)
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil, ErrUsage
			}
			return nil, fmt.Errorf("%w: %w", ErrUsage, err)
		}
		if fs.NArg() == 0 {
			return pos, nil
		}
		pos = append(pos, fs.Arg(0))
		args = fs.Args()[1:]
	}
}

func usage() string {
	return strings.TrimLeft(`
usage: chezmoi dots [edit] [<what>]
       chezmoi dots view <what> [--as <os>[,<role>] | --matrix]
       chezmoi dots list
       chezmoi dots lint [registry|aliases|packages|plugins|matrix|templates]... [--harness-dir <dir>]
       chezmoi dots outdated
       chezmoi dots bump [--dry-run] [<release>]...
       chezmoi dots upgrade [--dry-run] [<step>]...
       chezmoi dots version | help

Edit any part of the dotfiles in the chezmoi source, then see what it changes.

<what> is, in this order: a registry entry (.chezmoidata/edit.toml, dots list), a file
of a local entry's folder, a source file (source-relative) or a managed target.

  dots                           pick a registry entry
  dots agents/work               edit an entry
  dots ~/.claude/CLAUDE.md       edit what a target is made of: pick a node of its
                                 composition tree (layers, partials, data files)
  dots dot_config/git/config.tmpl
                                 edit a source file by its source-relative path
  dots ~/.ssh/config.d/<name>    edit a local file in place (a new name creates it)

Every source edit renders the target state of this machine and of 9 configurations
(darwin, linux, windows x personal, work, light; every harness selected), opens
$EDITOR (inside tmux, or Windows Terminal, a side pane shows this machine's changes
on every save), runs the entry's check, renders again and lists the changed targets
per configuration. A configuration that stops rendering fails the edit: nothing is
applied. Then, for this machine only: chezmoi diff of the changed targets, "apply
now? [y/N]", and a reminder to ship a PR. Encrypted files are edited in a private
temp copy and re-encrypted only when changed. A local file is edited in a private temp
copy that replaces it once the entry's check passes.

  view <what>                    its composition tree and its rendered content
                                 (an entry: the targets it reaches)
  --as <os>[,<role>]             render one configuration instead of this machine:
                                 os darwin|linux|windows, role personal (default),
                                 work or light; a preview only, never applied
  --matrix                       every configuration: identical renders grouped,
                                 the others as diffs against the first
  list                           the registry entries

  lint                           check the registries, the config template and every
                                 rendered script and alias template (all checks, or
                                 the ones named)
  outdated                       the pinned release binaries with a newer version
  bump                           pin the newer versions and their checksums in
                                 .chezmoidata/packages.toml (all, or the ones named)
  upgrade                        update the dotfiles, system packages and language
                                 tools (steps: dotfiles, brew|apt|winget, uv, npm,
                                 psmodules)

Docs: docs/OS.md#editing-your-dotfiles, docs/OS.md#packages
`, "\n")
}
