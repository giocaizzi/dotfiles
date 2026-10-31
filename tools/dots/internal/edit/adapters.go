package edit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
)

// ExecEditor opens files in $EDITOR, then $VISUAL, then vi (notepad on Windows).
type ExecEditor struct {
	Runner chez.Runner
	Getenv func(string) string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Edit implements Editor. The editor command may carry arguments ("code --wait").
func (e ExecEditor) Edit(ctx context.Context, path string) error {
	cmd := e.Getenv("EDITOR")
	if cmd == "" {
		cmd = e.Getenv("VISUAL")
	}
	if cmd == "" {
		cmd = "vi"
		if runtime.GOOS == "windows" {
			cmd = "notepad"
		}
	}
	argv, err := SplitCommand(cmd)
	if err != nil {
		return err
	}
	if err := e.Runner.Run(ctx, chez.Cmd{Args: append(argv, path), Stdin: e.Stdin, Stdout: e.Stdout, Stderr: e.Stderr}); err != nil {
		return fmt.Errorf("editor: %w", err)
	}
	return nil
}

// SplitCommand splits a command line into words, honouring single and double quotes.
func SplitCommand(s string) ([]string, error) {
	var words []string
	var cur strings.Builder
	var quote rune
	started := false
	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '"' || r == '\'':
			quote, started = r, true
		case r == ' ' || r == '\t':
			if started || cur.Len() > 0 {
				words = append(words, cur.String())
				cur.Reset()
				started = false
			}
		default:
			cur.WriteRune(r)
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote in %q", s)
	}
	if started || cur.Len() > 0 {
		words = append(words, cur.String())
	}
	if len(words) == 0 {
		return nil, errors.New("empty editor command")
	}
	return words, nil
}

// ExecChecker runs check commands as processes; their output is hidden, their errors shown.
type ExecChecker struct {
	Runner chez.Runner
	Stderr io.Writer
}

// Check implements Checker.
func (c ExecChecker) Check(ctx context.Context, dir string, argv []string) error {
	return c.Runner.Run(ctx, chez.Cmd{Args: argv, Dir: dir, Stdout: io.Discard, Stderr: c.Stderr})
}

// ExecGit asks git for the status of a working tree.
type ExecGit struct{ Runner chez.Runner }

// Status implements Git.
func (g ExecGit) Status(ctx context.Context, dir string) (string, error) {
	out, err := chez.Output(ctx, g.Runner, chez.Cmd{Args: []string{"git", "-C", dir, "status", "--short"}})
	return strings.TrimRight(string(out), "\n"), err
}
