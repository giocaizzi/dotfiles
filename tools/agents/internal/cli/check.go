package cli

import (
	"context"
	"fmt"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/parity"
)

func (d Deps) runCheck(_ context.Context, args []string) int {
	fs := newFlags("check")
	fileFlag := fs.String("file", "", "path to harness-policy.toml")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(d.Stderr, "flag parse error: %v\n", err)
		return 2
	}

	p, path, err := d.loadPolicy(*fileFlag)
	if err != nil {
		fmt.Fprintf(d.Stderr, "error loading policy: %v\n", err)
		return 1
	}

	findings := p.Validate(path)
	findings = append(findings, parity.CheckParity(p)...)
	if len(findings) > 0 {
		fmt.Fprintln(d.Stderr, formatFindings(findings))
		return 1
	}

	fmt.Fprintf(d.Stdout, "✓ policy & harness parity (%s)\n", path)
	return 0
}
