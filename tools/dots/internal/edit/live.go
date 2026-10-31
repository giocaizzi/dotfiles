package edit

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/term"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/udiff"
)

const doneFile = "done"

// LiveChezmoi is what the live view needs of chezmoi.
type LiveChezmoi interface {
	Dump(ctx context.Context, o chez.Overrides) (chez.State, error)
	Encrypt(ctx context.Context, plain []byte) ([]byte, error)
}

// Live re-renders this machine from a private copy of the source every time the
// edited file is saved, and shows the diff of each target the edit changes so far.
// The copy keeps the real source untouched; paths into it read as the real source dir.
type Live struct {
	CZ       LiveChezmoi // renders the shadow source
	File     string      // source-relative file being edited
	Edited   string      // the file the editor has open
	Shadow   string      // private copy of the source dir
	RealSrc  string
	Dir      string // scratch dir; a file called "done" in it ends the loop
	Out      io.Writer
	Style    term.Style
	Interval time.Duration
}

// Loop runs until ctx ends or the edit is over.
func (l Live) Loop(ctx context.Context) error {
	base, err := l.render(ctx)
	if err != nil {
		return err
	}
	var last [sha256.Size]byte
	first := true
	for {
		if _, err := os.Stat(filepath.Join(l.Dir, doneFile)); err == nil {
			return nil
		}
		if cur, err := hash(l.Edited); err == nil && (first || cur != last) {
			first, last = false, cur
			l.show(ctx, base)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(l.Interval):
		}
	}
}

func (l Live) show(ctx context.Context, base chez.State) {
	fmt.Fprint(l.Out, "\x1b[H\x1b[2J")
	fmt.Fprintf(l.Out, "%sdots live preview%s: %s on this machine (refreshes on save)\n", l.Style.B, l.Style.O, l.File)
	if err := l.sync(ctx); err != nil {
		fmt.Fprintln(l.Out, err)
		return
	}
	now, err := l.render(ctx)
	if err != nil {
		fmt.Fprintln(l.Out, err)
		return
	}
	for _, t := range chez.Changed(base, now) {
		fmt.Fprint(l.Out, l.Style.Diff(udiff.Diff(t+" (before)", t+" (now)", text(base[t]), text(now[t]))))
	}
}

// sync copies the edited file into the shadow source, encrypted again when it is.
func (l Live) sync(ctx context.Context) error {
	b, err := os.ReadFile(l.Edited)
	if err != nil {
		return fmt.Errorf("read %s: %w", l.Edited, err)
	}
	if strings.HasSuffix(l.File, ".age") {
		if b, err = l.CZ.Encrypt(ctx, b); err != nil {
			return fmt.Errorf("encrypt %s: %w", l.File, err)
		}
	}
	dst := filepath.Join(l.Shadow, filepath.FromSlash(l.File))
	if err := os.WriteFile(dst, b, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}
	return nil
}

func (l Live) render(ctx context.Context) (chez.State, error) {
	st, err := l.CZ.Dump(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}
	for k, e := range st {
		e.Contents = strings.ReplaceAll(e.Contents, l.Shadow, l.RealSrc)
		e.Linkname = strings.ReplaceAll(e.Linkname, l.Shadow, l.RealSrc)
		e.Raw = strings.ReplaceAll(e.Raw, jsonPath(l.Shadow), jsonPath(l.RealSrc))
		st[k] = e
	}
	return st, nil
}

// jsonPath is p as it appears inside a JSON string.
func jsonPath(p string) string { return strings.ReplaceAll(p, `\`, `\\`) }

func text(e chez.Entry) string {
	if e.Linkname != "" {
		return e.Linkname
	}
	return e.Contents
}

func hash(path string) ([sha256.Size]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("read %s: %w", path, err)
	}
	return sha256.Sum256(b), nil
}

// CopyTree copies the regular files of src into dst.
func CopyTree(src, dst string) error {
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		switch {
		case d.IsDir():
			return os.MkdirAll(target, 0o700)
		case d.Type().IsRegular():
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			return os.WriteFile(target, b, 0o600)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("copy %s: %w", src, err)
	}
	return nil
}

// Pane shows the live view in a side pane of the terminal multiplexer: tmux, or
// Windows Terminal on Windows. Without either, no preview is shown.
type Pane struct {
	Runner   chez.Runner
	Getenv   func(string) string
	LookPath func(string) (string, error)
	Exe      string   // this program, run again as `live`
	Global   []string // chezmoi flags the live process must use too
	Bin      string   // the chezmoi binary
	Src      string   // the real source dir
}

// Start implements Preview.
func (p Pane) Start(ctx context.Context, a PreviewArgs) func() {
	none := func() {}
	tmux := p.Getenv("TMUX") != "" && has(p.LookPath, "tmux")
	wt := runtime.GOOS == "windows" && p.Getenv("WT_SESSION") != "" && has(p.LookPath, "wt")
	if !tmux && !wt {
		return none
	}
	shadow := filepath.Join(a.Scratch, "shadow")
	if err := CopyTree(p.Src, shadow); err != nil {
		return none
	}
	args := []string{p.Exe, "live", "--bin", p.Bin, "--src", shadow, "--real-src", p.Src,
		"--file", a.File, "--edited", a.Edited, "--dir", a.Scratch}
	for _, g := range p.Global {
		args = append(args, "--global", g)
	}
	finish := func() { _ = os.WriteFile(filepath.Join(a.Scratch, doneFile), nil, 0o600) }
	if tmux {
		out, err := chez.Output(ctx, p.Runner, chez.Cmd{Args: []string{"tmux", "split-window", "-h", "-d", "-P", "-F", "#{pane_id}", shellJoin(args)}})
		if err != nil {
			return none
		}
		pane := strings.TrimSpace(string(out))
		return func() {
			finish()
			_ = p.Runner.Run(ctx, chez.Cmd{Args: []string{"tmux", "kill-pane", "-t", pane}, Stdout: io.Discard, Stderr: io.Discard})
		}
	}
	if err := p.Runner.Run(ctx, chez.Cmd{Args: append([]string{"wt", "-w", "0", "split-pane", "-V", "--"}, args...), Stdout: io.Discard, Stderr: io.Discard}); err != nil {
		return none
	}
	return finish
}

func has(look func(string) (string, error), name string) bool {
	_, err := look(name)
	return err == nil
}

// shellJoin quotes words for a POSIX shell, which is what tmux runs the command with.
func shellJoin(words []string) string {
	q := make([]string, len(words))
	for i, w := range words {
		q[i] = "'" + strings.ReplaceAll(w, "'", `'\''`) + "'"
	}
	return strings.Join(q, " ")
}
