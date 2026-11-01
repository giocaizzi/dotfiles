// Package cli implements the chezmoi-agents command-line interface.
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

	"github.com/giocaizzi/dotfiles/tools/agents/internal/finding"
	"github.com/giocaizzi/dotfiles/tools/agents/internal/policy"
)

// ErrUsage indicates incorrect command invocation.
var ErrUsage = errors.New("usage error")

// Deps encapsulates external dependencies for testability.
type Deps struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Getenv  func(string) string
	GOOS    string
	Version string
}

// Run dispatches subcommands.
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

// Run executes the command using the given dependencies.
func (d Deps) Run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		d.usage()
		return 0
	}

	switch args[0] {
	case "check":
		return d.runCheck(ctx, args[1:])
	case "render":
		return d.runRender(ctx, args[1:])
	case "prompt":
		return d.runPrompt(ctx, args[1:])
	case "version", "-v", "--version":
		fmt.Fprintf(d.Stdout, "chezmoi-agents %s\n", d.Version)
		return 0
	case "help", "-h", "--help":
		d.usage()
		return 0
	default:
		fmt.Fprintf(d.Stderr, "unknown command: %s\n", args[0])
		d.usage()
		return 2
	}
}

func (d Deps) usage() {
	fmt.Fprintln(d.Stdout, "Usage: chezmoi-agents <command> [flags]")
	fmt.Fprintln(d.Stdout, "\nCommands:")
	fmt.Fprintln(d.Stdout, "  check    Validate harness policy and configuration parity")
	fmt.Fprintln(d.Stdout, "  render   Render client-native configuration (JSON/TOML)")
	fmt.Fprintln(d.Stdout, "  prompt   Compose or validate layered system prompt")
	fmt.Fprintln(d.Stdout, "  version  Show version information")
	fmt.Fprintln(d.Stdout, "  help     Show this help message")
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

func newFlags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func formatFindings(fs []finding.Finding) string {
	return finding.Lines(fs)
}
