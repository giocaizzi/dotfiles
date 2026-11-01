package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/cli"
)

func TestCLIRun_Help(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	d := cli.Deps{
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	}

	code := d.Run(context.Background(), []string{"help"})
	if code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage: chezmoi-agents") {
		t.Errorf("expected usage output, got %q", stdout.String())
	}
}

func TestCLIRun_CheckAndRender(t *testing.T) {
	t.Parallel()
	policyPath := filepath.Join("..", "..", "..", "..", "home", ".chezmoidata", "harness-policy.toml")
	if _, err := os.Stat(policyPath); err != nil {
		t.Skip("skipping test; harness-policy.toml not accessible at relative path")
	}

	var stdout, stderr bytes.Buffer
	d := cli.Deps{
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
		Getenv: func(_ string) string { return "" },
	}

	code := d.Run(context.Background(), []string{"check", "-file", policyPath})
	if code != 0 {
		t.Fatalf("expected check exit 0, got %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "✓ policy & harness parity") {
		t.Errorf("expected check success output, got %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = d.Run(context.Background(), []string{"render", "-file", policyPath, "-harness", "pi"})
	if code != 0 {
		t.Fatalf("expected render exit 0, got %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "defaultProvider") {
		t.Errorf("expected pi output, got %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = d.Run(context.Background(), []string{"prompt", "-profile", "personal", "-check"})
	if code != 0 {
		t.Fatalf("expected prompt check exit 0, got %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "✓ prompt") {
		t.Errorf("expected prompt check success output, got %s", stdout.String())
	}
}
