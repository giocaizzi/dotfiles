package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/audit"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
)

// audit lists the aliases live in this machine's interactive shells that the registry
// does not account for.
func (d Deps) audit(ctx context.Context, args []string) error {
	pos, err := parse(newFlags("audit"), args)
	if err != nil {
		return err
	}
	if len(pos) != 1 || pos[0] != "aliases" {
		return fmt.Errorf("%w: audit takes aliases", ErrUsage)
	}
	known, err := d.knownAliases()
	if err != nil {
		return err
	}
	shells := []struct{ name, script string }{
		{"zsh", "print -l ${(k)aliases}"},
		{"bash", "compgen -a"},
	}
	failed := false
	for _, sh := range shells {
		if _, err := d.LookPath(sh.name); err != nil {
			continue
		}
		out, err := chez.Output(ctx, d.Runner, chez.Cmd{Args: []string{sh.name, "-ic", sh.script}})
		if err != nil {
			return fmt.Errorf("list %s aliases: %w", sh.name, err)
		}
		unknown := audit.Unknown(known, string(bytes.TrimSpace(out)))
		if len(unknown) == 0 {
			fmt.Fprintf(d.Stdout, "%s: all aliases accounted for\n", sh.name)
			continue
		}
		failed = true
		fmt.Fprintf(d.Stdout, "%s: aliases not in the registry:\n", sh.name)
		for _, n := range unknown {
			fmt.Fprintf(d.Stdout, "  %s\n", n)
		}
	}
	if failed {
		return ErrReported
	}
	return nil
}

// knownAliases reads the rendered alias files and the Oh My Zsh plugins they turn on.
func (d Deps) knownAliases() (map[string]bool, error) {
	cfg := d.Getenv("XDG_CONFIG_HOME")
	if cfg == "" {
		cfg = filepath.Join(d.Home, ".config")
	}
	shellDir := filepath.Join(cfg, "shell")
	rendered, err := os.ReadFile(filepath.Join(shellDir, "aliases.sh")) //nolint:gosec // the rendered alias file in the user's config dir
	if err != nil {
		return nil, fmt.Errorf("read the rendered aliases (chezmoi apply?): %w", err)
	}
	var plugin []string
	if omz, err := os.ReadFile(filepath.Join(shellDir, "omz-aliases.zsh")); err == nil { //nolint:gosec // as above
		for _, p := range audit.Plugins(string(omz)) {
			files, _ := filepath.Glob(filepath.Join(d.Home, ".oh-my-zsh", "plugins", p, "*.zsh"))
			for _, f := range files {
				if b, err := os.ReadFile(f); err == nil { //nolint:gosec // an Oh My Zsh plugin file
					plugin = append(plugin, audit.OMZAliases(string(b))...)
				}
			}
		}
	}
	return audit.Known(string(rendered), plugin), nil
}
