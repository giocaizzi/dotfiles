// Package term is the terminal side of the commands: colours, the composition tree
// and diff printers, a picker and a yes/no prompt.
package term

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/tree"
)

// Style holds the colour escapes for one output stream: blue headings (B), mauve
// labels (M), red (R), green (G), yellow (Y), cyan (C), dim (D) and reset (O). All are empty
// when the stream is no terminal, NO_COLOR is set or TERM is dumb. They are the
// basic ANSI colours, which the terminal's Catppuccin theme maps.
type Style struct{ B, M, R, G, Y, C, D, O string }

// NewStyle decides the colours for w.
func NewStyle(w io.Writer, getenv func(string) string) Style {
	f, ok := w.(*os.File)
	if !ok || getenv("NO_COLOR") != "" || getenv("TERM") == "dumb" {
		return Style{}
	}
	if fi, err := f.Stat(); err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		return Style{}
	}
	const esc = "\x1b["
	return Style{B: esc + "1;34m", M: esc + "35m", R: esc + "31m", G: esc + "32m", Y: esc + "33m", C: esc + "36m", D: esc + "2m", O: esc + "0m"}
}

// Tree prints nodes with ├─/└─ connectors, naming each node's registry entry
// (entryOf maps a source path to it) and marking encrypted nodes.
func (s Style) Tree(w io.Writer, nodes []tree.Node, entryOf map[string]string) {
	last := make([]bool, len(nodes))
	open := map[int]bool{}
	maxDepth, maxW := 0, 0
	for _, n := range nodes {
		maxDepth = max(maxDepth, n.Depth)
	}
	for i := len(nodes) - 1; i >= 0; i-- {
		d := nodes[i].Depth
		last[i] = !open[d]
		open[d] = true
		for j := d + 1; j <= maxDepth; j++ {
			open[j] = false
		}
		maxW = max(maxW, 3*d+len([]rune(nodes[i].Path)))
	}
	lastAt := map[int]bool{}
	for i, n := range nodes {
		var pre strings.Builder
		for j := 1; j < n.Depth; j++ {
			pre.WriteString(s.D + map[bool]string{true: "   ", false: "│  "}[lastAt[j]] + s.O)
		}
		if n.Depth > 0 {
			pre.WriteString(s.D + map[bool]string{true: "└─ ", false: "├─ "}[last[i]] + s.O)
		}
		lastAt[n.Depth] = last[i]
		tag := ""
		if e, ok := entryOf[n.Path]; ok {
			tag = s.M + e + s.O
		}
		if strings.HasSuffix(n.Path, ".age") {
			if tag != "" {
				tag += s.D + " · " + s.O
			}
			tag += s.Y + "encrypted" + s.O
		}
		pad := ""
		if tag != "" {
			pad = strings.Repeat(" ", maxW-3*n.Depth-len([]rune(n.Path))+2) + tag
		}
		fmt.Fprintf(w, "  %s%s%s\n", pre.String(), n.Path, pad)
	}
}

// Diff colours a unified diff: added green, removed red, headers and hunks dim.
func (s Style) Diff(diff string) string {
	if s.O == "" {
		return diff
	}
	var b strings.Builder
	for line := range strings.SplitSeq(strings.TrimSuffix(diff, "\n"), "\n") {
		switch {
		case strings.HasPrefix(line, "+++ "), strings.HasPrefix(line, "--- "), strings.HasPrefix(line, "@@"):
			b.WriteString(s.D + line + s.O)
		case strings.HasPrefix(line, "+"):
			b.WriteString(s.G + line + s.O)
		case strings.HasPrefix(line, "-"):
			b.WriteString(s.R + line + s.O)
		default:
			b.WriteString(line)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// Prompter asks the user: fzf for picks when it is installed, numbered prompts
// otherwise.
type Prompter struct {
	In       io.Reader
	Out      io.Writer // where prompts go, normally standard error
	Stderr   io.Writer
	Runner   chez.Runner
	LookPath func(string) (string, error)
	reader   *bufio.Reader
}

func (p *Prompter) line() (string, error) {
	if p.reader == nil {
		p.reader = bufio.NewReader(p.In)
	}
	s, err := p.reader.ReadString('\n')
	if err != nil && s == "" {
		return "", fmt.Errorf("read answer: %w", err)
	}
	return strings.TrimSpace(s), nil
}

// Confirm asks a yes/no question; an empty answer is def.
func (p *Prompter) Confirm(question string, def bool) (bool, error) {
	hint := "[y/N]"
	if def {
		hint = "[Y/n]"
	}
	fmt.Fprintf(p.Out, "%s %s ", question, hint)
	a, err := p.line()
	if err != nil {
		return false, err
	}
	if a == "" {
		return def, nil
	}
	return strings.EqualFold(a[:1], "y"), nil
}

// Pick returns the chosen item, or "" when the user picked nothing valid.
func (p *Prompter) Pick(ctx context.Context, label string, items []string) (string, error) {
	if _, err := p.LookPath("fzf"); err == nil {
		var out bytes.Buffer
		err := p.Runner.Run(ctx, chez.Cmd{
			Args:  []string{"fzf", "--prompt=" + label + "> ", "--height=~100%"},
			Stdin: strings.NewReader(strings.Join(items, "\n") + "\n"), Stdout: &out, Stderr: p.Stderr,
		})
		if err != nil {
			return "", nil //nolint:nilerr // fzf exits non-zero when the user cancels
		}
		return strings.TrimRight(out.String(), "\n"), nil
	}
	for i, it := range items {
		fmt.Fprintf(p.Out, "%d) %s\n", i+1, it)
	}
	fmt.Fprintf(p.Out, "%s [1-%d]: ", label, len(items))
	a, err := p.line()
	if err != nil {
		return "", err
	}
	if n, convErr := strconv.Atoi(a); convErr == nil && n >= 1 && n <= len(items) {
		return items[n-1], nil
	}
	return "", nil
}
