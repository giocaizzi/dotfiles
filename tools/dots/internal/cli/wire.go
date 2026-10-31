package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/edit"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/finding"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/matrix"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/registry"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/term"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/tree"
)

// chezmoi is the chezmoi this run talks to: the one that started the plugin, with
// the config and source it was started with.
func (d Deps) chezmoi() (*chez.Client, []string) {
	bin := d.Getenv("CHEZMOI_EXECUTABLE")
	if bin == "" {
		bin = "chezmoi"
	}
	var global []string
	if v := d.Getenv("CHEZMOI_CONFIG_FILE"); v != "" {
		global = append(global, "--config", v)
	}
	if v := d.Getenv("CHEZMOI_SOURCE_DIR"); v != "" {
		global = append(global, "--source", v)
	}
	return chez.New(d.Runner, bin, global...), append([]string{bin}, global...)
}

// service builds the edit/view use case. style picks the colours of its output.
func (d Deps) service(ctx context.Context, style term.Style) (*edit.Service, error) {
	cz, _ := d.chezmoi()
	src, err := cz.SourcePath(ctx, "", nil)
	if err != nil {
		return nil, fmt.Errorf("find the chezmoi source dir: %w", err)
	}
	data, err := cz.Data(ctx, nil)
	if err != nil {
		return nil, err
	}
	reg, problems := registry.Parse(data)
	if len(problems) > 0 {
		return nil, fmt.Errorf("the edit registry is invalid (chezmoi dots lint registry):\n%s", finding.Lines(problems))
	}
	prompt := &term.Prompter{In: d.Stdin, Out: d.Stderr, Stderr: d.Stderr, Runner: d.Runner, LookPath: d.LookPath}
	_, global := d.chezmoi()
	return &edit.Service{
		CZ:       cz,
		Registry: reg,
		Matrix: func(ctx context.Context, scratch string) ([]matrix.Entry, error) {
			return matrix.Build(ctx, cz, src, scratch)
		},
		Tree:   tree.New(os.DirFS(src)),
		Editor: edit.ExecEditor{Runner: d.Runner, Getenv: d.Getenv, Stdin: d.Stdin, Stdout: d.Stderr, Stderr: d.Stderr},
		Check:  edit.ExecChecker{Runner: d.Runner, Stderr: d.Stderr},
		Prompt: prompt,
		Git:    edit.ExecGit{Runner: d.Runner},
		Preview: edit.Pane{
			Runner: d.Runner, Getenv: d.Getenv, LookPath: d.LookPath, Exe: d.Exe,
			Bin: global[0], Global: global[1:], Src: src,
		},
		Src:   src,
		Out:   d.Stdout,
		Err:   d.Stderr,
		Style: style,
	}, nil
}

func (d Deps) edit(ctx context.Context, args []string) error {
	pos, err := parse(newFlags(cmdEdit), args)
	if err != nil {
		return err
	}
	if len(pos) > 1 {
		return ErrUsage
	}
	s, err := d.service(ctx, term.NewStyle(d.Stderr, d.Getenv))
	if err != nil {
		return err
	}
	what := ""
	if len(pos) == 1 {
		what = pos[0]
	}
	return s.Run(ctx, what)
}

func (d Deps) list(ctx context.Context, args []string) error {
	if pos, err := parse(newFlags("list"), args); err != nil || len(pos) > 0 {
		return ErrUsage
	}
	s, err := d.service(ctx, term.NewStyle(d.Stdout, d.Getenv))
	if err != nil {
		return err
	}
	s.List()
	return nil
}

func (d Deps) view(ctx context.Context, args []string) error {
	fs := newFlags("view")
	as := fs.String("as", "", "render one configuration")
	all := fs.Bool("matrix", false, "render every configuration")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 || (*as != "" && *all) {
		return ErrUsage
	}
	mode, cfg := edit.ViewHere, matrix.Config{}
	switch {
	case *all:
		mode = edit.ViewMatrix
	case *as != "":
		if cfg, err = matrix.Parse(*as); err != nil {
			return fmt.Errorf("%w: %w", ErrUsage, err)
		}
		mode = edit.ViewAs
	}
	s, err := d.service(ctx, term.NewStyle(d.Stdout, d.Getenv))
	if err != nil {
		return err
	}
	s.Err = d.Stderr
	return s.View(ctx, pos[0], mode, cfg)
}

func joinLines(fs []finding.Finding) string {
	return "  " + strings.ReplaceAll(finding.Lines(fs), "\n", "\n  ")
}
