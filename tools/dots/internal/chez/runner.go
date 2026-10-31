// Package chez is a typed client of the chezmoi command line. chezmoi's own Go
// packages are internal, so its commands (data, dump, managed, cat, ...) are the
// stable interface; every call goes through a Runner, which tests replace.
package chez

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Cmd is one process to run.
type Cmd struct {
	Args   []string
	Dir    string
	Env    []string // extra KEY=VALUE entries, on top of the environment
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Runner runs a Cmd to completion.
type Runner interface {
	Run(ctx context.Context, c Cmd) error
}

// ExecRunner runs commands as operating system processes.
type ExecRunner struct{}

// Run implements Runner.
func (ExecRunner) Run(ctx context.Context, c Cmd) error {
	if len(c.Args) == 0 {
		return errors.New("run: empty command")
	}
	cmd := exec.CommandContext(ctx, c.Args[0], c.Args[1:]...)
	cmd.Dir = c.Dir
	cmd.Stdin = c.Stdin
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
	if len(c.Env) > 0 {
		cmd.Env = append(cmd.Environ(), c.Env...)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", c.Args[0], err)
	}
	return nil
}

// Output runs c and returns its standard output. On failure the error names the
// command and carries what it wrote to standard error, or to standard output when
// that was empty (linters report there).
func Output(ctx context.Context, r Runner, c Cmd) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	c.Stdout, c.Stderr = &stdout, &stderr
	if err := r.Run(ctx, c); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %s", err, msg)
	}
	return stdout.Bytes(), nil
}
