// Package lint runs the checks on the dotfiles source: the registries, the config
// template and the rendered templates.
package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/aliasreg"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/finding"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/harnessreg"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/matrix"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/pkgreg"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/registry"
)

// Checks lists the check names, in the order `lint` runs them.
func Checks() []string {
	return []string{"registry", "aliases", "packages", "plugins", "matrix", "templates"}
}

// Linter runs the checks against one source dir.
type Linter struct {
	Src        string       // the source dir
	Iso        *chez.Client // chezmoi with an empty config, so no machine answers leak in
	Real       *chez.Client // chezmoi with this machine's config
	Runner     chez.Runner
	LookPath   func(string) (string, error)
	HarnessDir string // a checkout of giocaizzi/harness; empty asks GitHub through gh
	Scratch    string
	Errw       io.Writer
}

// Run runs the named check.
func (l Linter) Run(ctx context.Context, name string) ([]finding.Finding, error) {
	switch name {
	case "registry":
		return l.registry(ctx)
	case "aliases":
		return l.aliases(ctx)
	case "packages":
		return l.packages(ctx)
	case "plugins":
		return l.plugins(ctx)
	case "matrix":
		return l.matrix(ctx)
	case "templates":
		return l.templates(ctx)
	}
	return nil, fmt.Errorf("unknown check %q (checks: %s)", name, strings.Join(Checks(), ", "))
}

func (l Linter) registry(ctx context.Context) ([]finding.Finding, error) {
	data, err := l.Iso.Data(ctx, nil)
	if err != nil {
		return nil, err
	}
	reg, fs := registry.Parse(data)
	return append(fs, reg.Validate(os.DirFS(l.Src))...), nil
}

func (l Linter) aliases(ctx context.Context) ([]finding.Finding, error) {
	data, err := l.Iso.Data(ctx, nil)
	if err != nil {
		return nil, err
	}
	out := aliasreg.Validate(data)
	// The private registry needs this machine's age key; where there is none it is skipped.
	realData, err := l.Real.Data(ctx, nil)
	if err != nil {
		return nil, err
	}
	if identity := ageIdentity(realData); identity != "" {
		if _, statErr := os.Stat(identity); statErr == nil {
			private, privErr := l.privateAliases(ctx)
			if privErr != nil {
				return nil, privErr
			}
			out = append(out, private...)
		}
	}
	return out, nil
}

// privateAliases validates the encrypted private registry, decrypted with this machine's key.
func (l Linter) privateAliases(ctx context.Context) ([]finding.Finding, error) {
	tmpl := `{{ include ".aliases.private.toml.age" | decrypt | fromToml | toJson }}`
	b, err := l.Real.Execute(ctx, nil, []byte(tmpl), false)
	if err != nil {
		return nil, err
	}
	var private map[string]any
	if err = json.Unmarshal(b, &private); err != nil {
		return nil, fmt.Errorf("decode private aliases: %w", err)
	}
	var out []finding.Finding
	for _, f := range aliasreg.Validate(private) {
		out = append(out, finding.New("private "+f.Where, "%s", f.Message))
	}
	return out, nil
}

func ageIdentity(data map[string]any) string {
	c, _ := data["chezmoi"].(map[string]any)
	cfg, _ := c["config"].(map[string]any)
	age, _ := cfg["age"].(map[string]any)
	id, _ := age["identity"].(string)
	return id
}

func (l Linter) packages(ctx context.Context) ([]finding.Finding, error) {
	data, err := l.Iso.Data(ctx, nil)
	if err != nil {
		return nil, err
	}
	return pkgreg.Validate(data), nil
}

func (l Linter) plugins(ctx context.Context) ([]finding.Finding, error) {
	data, err := l.Iso.Data(ctx, nil)
	if err != nil {
		return nil, err
	}
	plugins := map[string][]string{}
	raw, _ := data["harnessPlugins"].(map[string]any)
	for p, v := range raw {
		list, _ := v.([]any)
		for _, h := range list {
			s, _ := h.(string)
			plugins[p] = append(plugins[p], s)
		}
	}
	var personal []string
	list, _ := data["harnessPersonalPlugins"].([]any)
	for _, p := range list {
		s, _ := p.(string)
		personal = append(personal, s)
	}
	tmpl, err := os.ReadFile(filepath.Join(l.Src, ".chezmoi.toml.tmpl"))
	if err != nil {
		return nil, fmt.Errorf("read config template: %w", err)
	}
	choices, err := matrix.Harnesses(tmpl)
	if err != nil {
		return nil, err
	}
	upstream, err := l.upstreamPlugins(ctx)
	if err != nil {
		return nil, err
	}
	return harnessreg.Check(plugins, personal, choices, upstream), nil
}

// upstreamPlugins lists the plugin directories of giocaizzi/harness.
func (l Linter) upstreamPlugins(ctx context.Context) ([]string, error) {
	if l.HarnessDir != "" {
		ents, err := os.ReadDir(filepath.Join(l.HarnessDir, "plugins"))
		if err != nil {
			return nil, fmt.Errorf("read harness plugins: %w", err)
		}
		var out []string
		for _, e := range ents {
			if e.IsDir() {
				out = append(out, e.Name())
			}
		}
		return out, nil
	}
	out, err := chez.Output(ctx, l.Runner, chez.Cmd{Args: []string{"gh", "api", "repos/giocaizzi/harness/contents/plugins", "--jq", `.[] | select(.type == "dir") | .name`}})
	if err != nil {
		return nil, fmt.Errorf("list harness plugins (set --harness-dir to a checkout, or sign in with gh): %w", err)
	}
	return strings.Fields(string(out)), nil
}

// matrix checks the config template against what the matrix seeds and the harness
// targets: they are written apart and must agree.
func (l Linter) matrix(ctx context.Context) ([]finding.Finding, error) {
	tmpl, err := os.ReadFile(filepath.Join(l.Src, ".chezmoi.toml.tmpl"))
	if err != nil {
		return nil, fmt.Errorf("read config template: %w", err)
	}
	var out []finding.Finding
	var seeded []string
	for k := range matrix.Seed("personal", nil, nil) {
		seeded = append(seeded, k)
	}
	slices.Sort(seeded)
	if got := matrix.DataKeys(tmpl); !slices.Equal(got, seeded) {
		out = append(out, finding.New(".chezmoi.toml.tmpl", "[data] keys %v differ from the keys the configuration matrix seeds %v (internal/matrix Seed)", got, seeded))
	}
	data, err := l.Iso.Data(ctx, nil)
	if err != nil {
		return nil, err
	}
	choices, err := matrix.Harnesses(tmpl)
	if err != nil {
		return nil, err
	}
	targets, _ := data["harnessTargets"].(map[string]any)
	var owned []string
	for h := range targets {
		owned = append(owned, h)
	}
	owned = append(owned, "copilot") // its targets are shared, so .chezmoiignore spells them out
	slices.Sort(owned)
	want := slices.Clone(choices)
	slices.Sort(want)
	if !slices.Equal(want, owned) {
		out = append(out, finding.New(".chezmoi.toml.tmpl", "harness choices %v differ from [harnessTargets] plus copilot %v", want, owned))
	}
	return out, nil
}

// templates renders the script and alias templates for every configuration and runs
// the linter of their language over the result; a template that renders empty does
// not apply to that configuration and is skipped, like chezmoi skips it.
func (l Linter) templates(ctx context.Context) ([]finding.Finding, error) {
	entries, err := matrix.Build(ctx, l.Real, l.Src, l.Scratch)
	if err != nil {
		return nil, err
	}
	// Also a single harness, which gates the harness scripts differently.
	for _, e := range slices.Clone(entries) {
		if e.Role != "light" {
			e.Overrides = e.Overrides.Clone(chez.Overrides{"harnesses": []string{"opencode"}})
			entries = append(entries, e)
		}
	}
	files, err := l.templateFiles()
	if err != nil {
		return nil, err
	}
	var (
		out []finding.Finding
		mu  sync.Mutex
		g   errgroup.Group
	)
	g.SetLimit(8)
	for _, f := range files {
		tmpl, err := os.ReadFile(filepath.Join(l.Src, filepath.FromSlash(f)))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f, err)
		}
		for i, e := range entries {
			g.Go(func() error {
				fs := l.renderAndLint(ctx, f, tmpl, e, i)
				mu.Lock()
				out = append(out, fs...)
				mu.Unlock()
				return nil
			})
		}
	}
	_ = g.Wait()
	return out, nil
}

// templateFiles lists the templates worth rendering whole: every script, the alias
// files and the JSON settings. Partials need arguments and the secrets files read the
// Secrets Manager, so they are checked through the templates that use them.
func (l Linter) templateFiles() ([]string, error) {
	out := []string{
		"dot_config/shell/aliases.sh.tmpl", "dot_config/shell/omz-aliases.zsh.tmpl", "dot_config/powershell/aliases.ps1.tmpl",
		".chezmoitemplates/vscode-mcp.json", ".chezmoitemplates/claude-settings.json", ".chezmoitemplates/copilot-settings.json", ".chezmoitemplates/antigravity-settings.json",
		"dot_codex/modify_private_config.toml",
	}
	scripts, err := filepath.Glob(filepath.Join(l.Src, ".chezmoiscripts", "*.tmpl"))
	if err != nil {
		return nil, fmt.Errorf("glob scripts: %w", err)
	}
	for _, p := range scripts {
		rel, _ := filepath.Rel(l.Src, p)
		out = append(out, filepath.ToSlash(rel))
	}
	return out, nil
}

func (l Linter) renderAndLint(ctx context.Context, file string, tmpl []byte, e matrix.Entry, n int) []finding.Finding {
	where := fmt.Sprintf("%s (%s)", file, e.Label())
	rendered, err := l.Iso.Execute(ctx, e.Overrides, tmpl, false)
	if err != nil {
		return []finding.Finding{finding.New(where, "does not render: %v", err)}
	}
	if strings.TrimSpace(string(rendered)) == "" {
		return nil
	}
	ext := filepath.Ext(strings.TrimSuffix(file, ".tmpl"))
	path := filepath.Join(l.Scratch, fmt.Sprintf("render-%d-%s%s", n, strings.ReplaceAll(file, "/", "_"), ext))
	if err := os.WriteFile(path, rendered, 0o600); err != nil { //nolint:gosec // path is inside the linter's own scratch dir
		return []finding.Finding{finding.New(where, "write render: %v", err)}
	}
	var argv []string
	switch {
	case ext == ".json":
		if !json.Valid(rendered) {
			return []finding.Finding{finding.New(where, "renders invalid JSON")}
		}
		return nil
	case ext == ".sh" && strings.Contains(file, "aliases"):
		argv = []string{"shellcheck", "-s", "bash", path}
	case ext == ".sh":
		argv = []string{"shellcheck", "-e", "SC2034", path}
	case ext == ".zsh":
		argv = []string{"zsh", "-n", path}
	case ext == ".ps1" && e.OS == "windows":
		argv = []string{"pwsh", "-NoProfile", "-Command", fmt.Sprintf(
			"$e = $null; [void][System.Management.Automation.Language.Parser]::ParseFile('%s', [ref]$null, [ref]$e); if ($e) { $e; exit 1 }", path)}
	default:
		return nil
	}
	if _, err := l.LookPath(argv[0]); err != nil {
		return nil // the tool is not installed here; CI has it
	}
	if _, err := chez.Output(ctx, l.Runner, chez.Cmd{Args: argv}); err != nil {
		return []finding.Finding{finding.New(where, "%v", err)}
	}
	return nil
}
