// Package cli implements the chezmoi-agents command line interface.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/finding"
	"github.com/giocaizzi/dotfiles/tools/agents/internal/policy"
	"github.com/giocaizzi/dotfiles/tools/agents/internal/term"
)

// Deps holds the injected dependencies for the CLI.
type Deps struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Getenv  func(string) string
	GOOS    string
	Version string
}

// Run dispatches subcommands with standard OS dependencies.
func Run(ctx context.Context, args []string, version string) int {
	d := Deps{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Getenv:  os.Getenv,
		GOOS:    runtime.GOOS,
		Version: version,
	}
	return d.Run(ctx, args)
}

// Run executes the command line application and returns an exit code.
func (d Deps) Run(ctx context.Context, args []string) int {
	if d.GOOS == "" {
		d.GOOS = runtime.GOOS
	}
	if d.Getenv == nil {
		d.Getenv = os.Getenv
	}
	if d.Stdin == nil {
		d.Stdin = os.Stdin
	}
	if d.Stdout == nil {
		d.Stdout = os.Stdout
	}
	if d.Stderr == nil {
		d.Stderr = os.Stderr
	}

	if len(args) == 0 {
		st := term.NewStyle(d.Stdout, d.Getenv)
		fmt.Fprint(d.Stdout, usage(st))
		return 0
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "check":
		return d.runCheck(ctx, rest)
	case "render":
		return d.runRender(ctx, rest)
	case "prompt":
		return d.runPrompt(ctx, rest)
	case "version", "--version", "-v":
		v := d.Version
		if v == "" {
			v = "dev"
		}
		fmt.Fprintf(d.Stdout, "chezmoi-agents %s\n", v)
		return 0
	case "help", "--help", "-h":
		st := term.NewStyle(d.Stdout, d.Getenv)
		fmt.Fprint(d.Stdout, usage(st))
		return 0
	default:
		st := term.NewStyle(d.Stderr, d.Getenv)
		fmt.Fprintf(d.Stderr, "%sunknown command: %s%s\n\n", st.R, cmd, st.O)
		fmt.Fprint(d.Stderr, usage(st))
		return 2
	}
}

func usage(s term.Style) string {
	b := s.B // blue
	m := s.M // mauve
	y := s.Y // yellow
	d := s.D // dim overlay0
	o := s.O // reset

	var sb strings.Builder
	p := func(format string, a ...any) {
		fmt.Fprintf(&sb, format, a...)
	}

	p("%schezmoi agents%s %s— AI harness configuration, policy validation, parity and prompt engine%s\n\n", b, o, d, o)

	p("%sUSAGE%s\n", b, o)
	p("  %schezmoi agents%s %scheck%s %s[-file <path>] [-harness-dir <path>]%s\n", m, o, m, o, d, o)
	p("  %schezmoi agents%s %srender%s %s[-file <path>]%s %s-harness <name>%s %s[-target <file>]%s\n", m, o, m, o, d, o, y, o, d, o)
	p("  %schezmoi agents%s %sprompt%s %s[-profile <personal|work|light>]%s %s[-check]%s %s[-file <path>]%s\n", m, o, m, o, d, o, d, o, d, o)
	p("  %schezmoi agents%s %sversion%s %s│%s %shelp%s\n\n", m, o, m, o, d, o, m, o)

	p("%sCOMMANDS%s\n", b, o)
	p("  %s%-9s%s %s%-27s%s %s\n", m, "check", o, d, "[-file <path>]", o, "Validate harness policy, 6-way client parity, and plugins")
	p("  %s%-9s%s %s%-27s%s %s\n", m, "render", o, y, "-harness <name>", o, "Render client-native configuration (JSON/TOML)")
	p("  %s%-9s%s %s%-27s%s %s\n", m, "prompt", o, d, "[-profile <profile>]", o, "Compose or validate layered system prompt")
	p("  %s%-9s%s %-27s %s\n", m, "version", o, "", "Print chezmoi-agents version information")
	p("  %s%-9s%s %-27s %s\n\n", m, "help", o, "", "Display this help message")

	p("%sFLAGS%s\n", b, o)
	p("  %s-file%s %s<path>%s         Path to harness-policy.toml, harness.toml, or prompt layer\n", y, o, d, o)
	p("  %s-harness%s %s<name>%s      Target harness (claude|antigravity|copilot|codex|pi|opencode)\n", y, o, d, o)
	p("  %s-harness-dir%s %s<path>%s  Checkout of giocaizzi/harness (defaults to HARNESS_DIR)\n", y, o, d, o)
	p("  %s-profile%s %s<profile>%s   System prompt profile (personal|work|light)\n", y, o, d, o)
	p("  %s-check%s               Validate layer presence without printing output\n", y, o)
	p("  %s-target%s %s<file>%s       Target file to write rendered configuration\n\n", y, o, d, o)

	p("%sEXAMPLES%s\n", b, o)
	p("  %schezmoi-agents check%s                           Validate policy, client parity, and harness plugins\n", m, o)
	p("  %schezmoi-agents check -file .chezmoidata/harness.toml%s Validate plugin registry against upstream harness\n", m, o)
	p("  %schezmoi-agents render -harness pi%s              Render Pi native agent configuration\n", m, o)
	p("  %schezmoi-agents render -harness antigravity%s     Render Antigravity CLI native permissions\n", m, o)
	p("  %schezmoi-agents prompt -profile personal -check%s Verify personal prompt layer assembly\n", m, o)
	p("  %schezmoi-agents prompt -profile work%s            Display composed instructions for work profile\n\n", m, o)

	p("%sDOCUMENTATION%s\n", b, o)
	p("  %sdocs/OS.md#shared-harness-policy%s %s·%s %sdocs/OS.md#agent-instructions%s %s·%s %sdocs/OS.md#agent-plugins-giocaizzi-harness%s\n", d, o, d, o, d, o, d, o, d, o)

	return sb.String()
}

func (d Deps) loadPolicy(explicitPath string) (*policy.Policy, string, error) {
	if explicitPath != "" {
		cleaned := filepath.Clean(explicitPath)
		f, err := os.Open(cleaned)
		if err != nil {
			return nil, "", fmt.Errorf("open policy %s: %w", explicitPath, err)
		}
		defer f.Close()
		p, err := policy.ParseTOML(f)
		if err != nil {
			return nil, "", fmt.Errorf("parse %s: %w", explicitPath, err)
		}
		return p, explicitPath, nil
	}

	candidates := []string{}
	if src := d.Getenv("CHEZMOI_SOURCE_DIR"); src != "" {
		candidates = append(candidates, filepath.Join(src, ".chezmoidata", "harness-policy.toml"))
	}
	candidates = append(candidates,
		filepath.Join("home", ".chezmoidata", "harness-policy.toml"),
		filepath.Join(".chezmoidata", "harness-policy.toml"),
		filepath.Join("..", "home", ".chezmoidata", "harness-policy.toml"),
		filepath.Join("..", "..", "home", ".chezmoidata", "harness-policy.toml"),
	)

	for _, cand := range candidates {
		cleaned := filepath.Clean(cand)
		if _, err := os.Stat(cleaned); err == nil {
			f, err := os.Open(cleaned)
			if err != nil {
				return nil, "", fmt.Errorf("open %s: %w", cand, err)
			}
			defer f.Close()
			p, err := policy.ParseTOML(f)
			if err != nil {
				return nil, "", fmt.Errorf("parse %s: %w", cand, err)
			}
			return p, cand, nil
		}
	}

	return nil, "", errors.New("harness-policy.toml not found")
}

func (d Deps) findHarnessToml(explicitPath string) (string, error) {
	if explicitPath != "" {
		cleaned := filepath.Clean(explicitPath)
		if _, err := os.Stat(cleaned); err != nil {
			return "", fmt.Errorf("stat %s: %w", explicitPath, err)
		}
		return cleaned, nil
	}

	candidates := []string{}
	if src := d.Getenv("CHEZMOI_SOURCE_DIR"); src != "" {
		candidates = append(candidates, filepath.Join(src, ".chezmoidata", "harness.toml"))
	}
	candidates = append(candidates,
		filepath.Join("home", ".chezmoidata", "harness.toml"),
		filepath.Join(".chezmoidata", "harness.toml"),
		filepath.Join("..", "home", ".chezmoidata", "harness.toml"),
		filepath.Join("..", "..", "home", ".chezmoidata", "harness.toml"),
	)

	for _, cand := range candidates {
		cleaned := filepath.Clean(cand)
		if _, err := os.Stat(cleaned); err == nil {
			return cleaned, nil
		}
	}

	return "", errors.New("harness.toml not found")
}

func newFlags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func formatFindings(fs []finding.Finding) string {
	return finding.Lines(fs)
}
