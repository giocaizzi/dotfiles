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
	"github.com/giocaizzi/dotfiles/tools/dots/internal/term"
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

const cmdEdit = "edit"

// Run runs one command line (the arguments after `chezmoi dots`).
func Run(ctx context.Context, d Deps, args []string) error {
	cmd, rest := cmdEdit, args
	if len(args) > 0 {
		switch args[0] {
		case cmdEdit, "view", "list", "policy", "lint", "audit", "outdated", "bump", "upgrade", "live", "version", "help":
			cmd, rest = args[0], args[1:]
		case "-h", "--help":
			cmd = "help"
		}
	}
	switch cmd {
	case cmdEdit:
		return d.edit(ctx, rest)
	case "view":
		return d.view(ctx, rest)
	case "list":
		return d.list(ctx, rest)
	case "policy":
		return d.policy(ctx, rest)
	case "lint":
		return d.lint(ctx, rest)
	case "audit":
		return d.audit(ctx, rest)
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
		s := term.NewStyle(d.Stdout, d.Getenv)
		fmt.Fprint(d.Stdout, usage(s))
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

func usage(s term.Style) string {
	b := s.B // bold blue
	m := s.M // mauve
	y := s.Y // yellow
	d := s.D // dim overlay0
	o := s.O // reset

	var sb strings.Builder
	p := func(format string, a ...any) {
		fmt.Fprintf(&sb, format, a...)
	}

	p("%schezmoi dots%s %s— Manage, verify and render dotfiles across 9 OS/role matrix configurations%s\n\n", b, o, d, o)

	p("%sUSAGE%s\n", b, o)
	p("  %schezmoi dots%s %s[edit]%s %s[<what>]%s\n", m, o, y, o, d, o)
	p("  %schezmoi dots%s %sview%s %s<what>%s %s[--as <os>[,<role>] | --matrix]%s\n", m, o, m, o, y, o, d, o)
	p("  %schezmoi dots%s %spolicy%s %srender|check%s %s[--harness <name>] [--part <part>] [--os <os>]%s\n", m, o, m, o, y, o, d, o)
	p("  %schezmoi dots%s %slist%s\n", m, o, m, o)
	p("  %schezmoi dots%s %slint%s %s[registry|aliases|packages|plugins|matrix|templates|policy]...%s\n", m, o, m, o, d, o)
	p("  %schezmoi dots%s %saudit%s %saliases%s\n", m, o, m, o, y, o)
	p("  %schezmoi dots%s %soutdated%s\n", m, o, m, o)
	p("  %schezmoi dots%s %sbump%s %s[--dry-run]%s %s[<release>]...%s\n", m, o, m, o, d, o, d, o)
	p("  %schezmoi dots%s %supgrade%s %s[--dry-run]%s %s[<step>]...%s\n", m, o, m, o, d, o, d, o)
	p("  %schezmoi dots%s %sversion%s %s│%s %shelp%s\n\n", m, o, m, o, d, o, m, o)

	p("%sCOMMANDS%s\n", b, o)
	p("  %s%-9s%s %s%-27s%s %s\n", m, "edit", o, d, "[<what>]", o, "Edit dotfiles source and verify across 9 matrix configs")
	p("  %s%-9s%s %s%-27s%s %s\n", m, "view", o, y, "<what>", o, "Preview rendered target or composition tree")
	p("  %s%-9s%s %s%-27s%s %s\n", m, "policy", o, y, "render|check", o, "Translate or validate cross-harness security policies")
	p("  %s%-9s%s %-27s %s\n", m, "list", o, "", "List configured editable registry components")
	p("  %s%-9s%s %s%-27s%s %s\n", m, "lint", o, d, "[<check>...]", o, "Check registries, config templates, and rendered scripts")
	p("  %s%-9s%s %s%-27s%s %s\n", m, "audit", o, y, "aliases", o, "Detect unmanaged live shell aliases")
	p("  %s%-9s%s %-27s %s\n", m, "outdated", o, "", "Check for newer releases of pinned tools")
	p("  %s%-9s%s %s%-27s%s %s\n", m, "bump", o, d, "[--dry-run] [<release>...]", o, "Update pinned checksums and tool versions")
	p("  %s%-9s%s %s%-27s%s %s\n", m, "upgrade", o, d, "[--dry-run] [<step>...]", o, "Execute sequential system/dotfile upgrade steps")
	p("  %s%-9s%s %-27s %s\n", m, "version", o, "", "Print chezmoi-dots version information")
	p("  %s%-9s%s %-27s %s\n\n", m, "help", o, "", "Display this help message")

	p("%sFLAGS%s\n", b, o)
	p("  %s--as%s %s<os>[,<role>]%s    Render single matrix target (os: darwin|linux|windows; role: personal|work|light)\n", y, o, d, o)
	p("  %s--matrix%s              Render all matrix configurations as comparative unified diffs\n", y, o)
	p("  %s--harness%s %s<name>%s      Target AI harness (claude, antigravity, copilot, codex)\n", y, o, d, o)
	p("  %s--part%s %s<part>%s         Output granularity (permissions, settings, all)\n", y, o, d, o)
	p("  %s--os%s %s<os>%s             Target operating system (darwin, linux, windows)\n", y, o, d, o)
	p("  %s--file%s %s<path>%s         Explicit path to harness-policy.toml\n", y, o, d, o)
	p("  %s--dry-run%s             Simulate modifications without writing to disk\n\n", y, o)

	p("%sEXAMPLES%s\n", b, o)
	p("  %sdots%s                             Interactive picker for editable components\n", m, o)
	p("  %sdots agents/work%s                 Edit an entry in private temp copy with live pane diff\n", m, o)
	p("  %sdots view ~/.claude/CLAUDE.md%s    Display composition tree and target content\n", m, o)
	p("  %sdots view dot_config/git/config.tmpl --matrix%s\n                                Render and diff all 9 matrix configurations\n", m, o)
	p("  %sdots policy render --harness claude --os windows%s\n                                Render native permissions for Claude Code on Windows\n", m, o)
	p("  %sdots policy check%s                Validate cross-harness policy invariants\n", m, o)
	p("  %sdots lint%s                        Validate all dotfiles registries and rendered templates\n\n", m, o)

	p("%sDOCUMENTATION%s\n", b, o)
	p("  %sdocs/OS.md#editing-your-dotfiles%s %s·%s %sdocs/OS.md#packages%s %s·%s %sdocs/OS.md#shared-harness-policy%s\n", d, o, d, o, d, o, d, o, d, o)

	return sb.String()
}
