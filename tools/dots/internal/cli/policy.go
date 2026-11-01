package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/finding"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/policy"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/term"
)

func (d Deps) policy(_ context.Context, args []string) error {
	action := "render"
	rest := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		action, rest = args[0], args[1:]
	}

	fs := newFlags("policy")
	harnessFlag := fs.String("harness", "", "target client harness (claude, antigravity, copilot, codex)")
	partFlag := fs.String("part", "", "output granularity (permissions, settings, all)")
	osFlag := fs.String("os", "", "target operating system (darwin, linux, windows)")
	fileFlag := fs.String("file", "", "path to harness-policy.toml")

	if _, err := parse(fs, rest); err != nil {
		return err
	}

	targetOS := *osFlag
	if targetOS == "" {
		if d.GOOS != "" {
			targetOS = d.GOOS
		} else {
			targetOS = runtime.GOOS
		}
	}

	p, sourcePath, err := d.loadPolicy(*fileFlag)
	if err != nil {
		return err
	}

	switch action {
	case "render":
		return d.renderPolicy(p, *harnessFlag, *partFlag, targetOS)
	case "check":
		return d.checkPolicy(p, sourcePath)
	default:
		return fmt.Errorf("%w: unknown policy action %q (expected render or check)", ErrUsage, action)
	}
}

func (d Deps) loadPolicy(explicitPath string) (*policy.Policy, string, error) {
	if explicitPath != "" {
		cleaned := filepath.Clean(explicitPath)
		f, err := os.Open(cleaned)
		if err != nil {
			return nil, "", fmt.Errorf("open policy file %s: %w", explicitPath, err)
		}
		defer f.Close()
		p, err := policy.ParseTOML(f)
		return p, explicitPath, err
	}

	// Try reading from stdin if redirected
	if f, ok := d.Stdin.(*os.File); ok {
		if fi, err := f.Stat(); err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
			p, err := policy.ParseTOML(d.Stdin)
			return p, "<stdin>", err
		}
	}

	// Look in candidate locations
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
			return p, cand, err
		}
	}

	return nil, "", errors.New("cannot locate harness-policy.toml; specify with --file")
}

func (d Deps) renderPolicy(p *policy.Policy, harness, part, targetOS string) error {
	var out any

	switch harness {
	case "claude":
		if part == "permissions" || part == "" {
			out = p.ClaudePermissions(targetOS)
		} else {
			out = p.ClaudeSettings(targetOS)
		}
	case "antigravity":
		if part == "permissions" || part == "" {
			out = p.AntigravityPermissions()
		} else {
			out = p.AntigravitySettings()
		}
	case "copilot":
		out = p.CopilotSettings()
	case "codex":
		out = p.CodexConfig()
	case "":
		out = map[string]any{
			"claude":      p.ClaudeSettings(targetOS),
			"antigravity": p.AntigravitySettings(),
			"copilot":     p.CopilotSettings(),
			"codex":       p.CodexConfig(),
		}
	default:
		return fmt.Errorf("%w: unknown harness %q (supported: claude, antigravity, copilot, codex)", ErrUsage, harness)
	}

	enc := json.NewEncoder(d.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encode policy json: %w", err)
	}
	return nil
}

func (d Deps) checkPolicy(p *policy.Policy, sourcePath string) error {
	findings := p.Validate(sourcePath)
	style := term.NewStyle(d.Stdout, d.Getenv)
	if len(findings) > 0 {
		fmt.Fprintf(d.Stderr, "%s\n", finding.Lines(findings))
		return ErrReported
	}
	fmt.Fprintf(d.Stdout, "%s✓%s policy (%s)\n", style.G, style.O, sourcePath)
	return nil
}
