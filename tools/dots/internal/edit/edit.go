// Package edit is the edit flow of `chezmoi dots`: change a dotfiles source file,
// see what the change does on every machine type, then apply it here.
package edit

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/matrix"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/registry"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/term"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/tree"
)

// ErrReported marks a failure the flow already explained to the user.
var ErrReported = errors.New("failed")

// ErrUsage marks a request the flow cannot make sense of.
var ErrUsage = errors.New("usage")

// Chezmoi is the part of chezmoi the flow uses; *chez.Client implements it.
type Chezmoi interface {
	SourcePath(ctx context.Context, target string, o chez.Overrides) (string, error)
	TargetPath(ctx context.Context) (string, error)
	WorkingTree(ctx context.Context) (string, error)
	Dump(ctx context.Context, o chez.Overrides) (chez.State, error)
	Managed(ctx context.Context, o chez.Overrides) (map[string]string, error)
	Cat(ctx context.Context, target string, o chez.Overrides) ([]byte, error)
	Status(ctx context.Context, targets ...string) (string, error)
	Diff(ctx context.Context, w, errw io.Writer, targets ...string) error
	Apply(ctx context.Context, stdout, stderr io.Writer, targets ...string) error
	Encrypt(ctx context.Context, plain []byte) ([]byte, error)
	Decrypt(ctx context.Context, path string) ([]byte, error)
}

// Editor opens a file for the user and returns when they are done.
type Editor interface {
	Edit(ctx context.Context, path string) error
}

// Checker runs an entry's check command; a non-nil error fails the check.
type Checker interface {
	Check(ctx context.Context, dir string, argv []string) error
}

// Prompter asks the user.
type Prompter interface {
	Confirm(question string, def bool) (bool, error)
	Pick(ctx context.Context, label string, items []string) (string, error)
}

// Git reports the uncommitted changes of a working tree.
type Git interface {
	Status(ctx context.Context, dir string) (string, error)
}

// Preview is a live view of what the edit changes on this machine, shown beside the
// editor while it is open. Start returns the function that closes it.
type Preview interface {
	Start(ctx context.Context, p PreviewArgs) (stop func())
}

// PreviewArgs says what a Preview shows.
type PreviewArgs struct {
	File     string // source-relative file being edited
	Edited   string // the file the editor has open
	Scratch  string // private dir the preview may use
	Baseline chez.State
}

// Service runs the flow. Build one with every field set (Preview may be nil).
type Service struct {
	CZ       Chezmoi
	Registry registry.Registry
	Matrix   func(ctx context.Context, scratch string) ([]matrix.Entry, error)
	Tree     *tree.Tree
	Editor   Editor
	Check    Checker
	Prompt   Prompter
	Git      Git
	Preview  Preview
	Src      string // the source dir
	Out      io.Writer
	Err      io.Writer
	Style    term.Style
}

// Run edits what names: a registry entry, a file of a local entry's folder, a source
// file (source-relative) or a managed target. Empty picks a registry entry.
func (s *Service) Run(ctx context.Context, what string) error {
	if what == "" {
		names := s.Registry.Names()
		items := make([]string, 0, len(names))
		for _, n := range names {
			e, _ := s.Registry.Get(n)
			items = append(items, n+"\t"+e.Desc)
		}
		pick, err := s.Prompt.Pick(ctx, "dots", items)
		if err != nil {
			return fmt.Errorf("pick: %w", err)
		}
		if pick == "" {
			return ErrReported
		}
		what, _, _ = strings.Cut(pick, "\t")
	}
	home, err := s.CZ.TargetPath(ctx)
	if err != nil {
		return fmt.Errorf("target path: %w", err)
	}
	if e, ok := s.Registry.Get(what); ok {
		if e.Local != "" {
			return s.editLocal(ctx, e, filepath.Join(home, filepath.FromSlash(e.Local)), "")
		}
		return s.editSource(ctx, &e, e.Source)
	}
	if rel, ok := strings.CutPrefix(filepath.ToSlash(what), filepath.ToSlash(home)+"/"); ok {
		if e, found := s.Registry.ByLocal(rel); found {
			dir := filepath.Join(home, filepath.FromSlash(e.Local))
			return s.editLocal(ctx, e, dir, strings.TrimPrefix(rel, e.Local+"/"))
		}
	}
	if fi, statErr := os.Stat(filepath.Join(s.Src, filepath.FromSlash(what))); statErr == nil && fi.Mode().IsRegular() {
		return s.editFile(ctx, filepath.ToSlash(what))
	}
	file, err := s.pickNode(ctx, what)
	if err != nil {
		return err
	}
	return s.editFile(ctx, file)
}

func (s *Service) editFile(ctx context.Context, file string) error {
	if e, ok := s.Registry.BySource(file); ok {
		return s.editSource(ctx, &e, file)
	}
	return s.editSource(ctx, nil, file)
}

// pickNode lets the user pick which source file a target is made of.
func (s *Service) pickNode(ctx context.Context, target string) (string, error) {
	source, err := s.CZ.SourcePath(ctx, target, nil)
	if err != nil {
		fmt.Fprintf(s.Err, "%s✗ dots: %s is no registry entry, source file or managed target%s (chezmoi dots help)\n", s.Style.R, target, s.Style.O)
		return "", ErrUsage
	}
	rel, err := filepath.Rel(s.Src, source)
	if err != nil {
		return "", fmt.Errorf("source path of %s: %w", target, err)
	}
	var files, items []string
	for _, n := range s.Tree.Of(filepath.ToSlash(rel)) {
		if fi, statErr := os.Stat(filepath.Join(s.Src, filepath.FromSlash(n.Path))); statErr != nil || !fi.Mode().IsRegular() {
			continue
		}
		files = append(files, n.Path)
		item := n.Path
		if e, ok := s.Registry.BySource(n.Path); ok {
			item += "\t[" + e.Name + "]"
		}
		items = append(items, item)
	}
	switch len(files) {
	case 0:
		return "", fmt.Errorf("%s: no source file to edit: %w", target, ErrUsage)
	case 1:
		return files[0], nil
	}
	pick, err := s.Prompt.Pick(ctx, target, items)
	if err != nil {
		return "", fmt.Errorf("pick: %w", err)
	}
	if pick == "" {
		return "", ErrReported
	}
	pick, _, _ = strings.Cut(pick, "\t")
	return pick, nil
}

type result struct {
	state chez.State
	err   error
}

type snapshot struct {
	here result
	conf []result
}

// snapshot renders the target state of this machine and of every configuration,
// in parallel; a render that fails is recorded, not returned.
func (s *Service) snapshot(ctx context.Context, entries []matrix.Entry) snapshot {
	snap := snapshot{conf: make([]result, len(entries))}
	var g errgroup.Group
	g.SetLimit(4)
	g.Go(func() error {
		snap.here.state, snap.here.err = s.CZ.Dump(ctx, nil)
		return nil
	})
	for i, e := range entries {
		g.Go(func() error {
			snap.conf[i].state, snap.conf[i].err = s.CZ.Dump(ctx, e.Overrides)
			return nil
		})
	}
	_ = g.Wait()
	return snap
}

func (s *Service) editSource(ctx context.Context, e *registry.Entry, file string) error {
	scratch, err := os.MkdirTemp("", "dots-")
	if err != nil {
		return fmt.Errorf("make scratch dir: %w", err)
	}
	defer os.RemoveAll(scratch)
	path := filepath.Join(s.Src, filepath.FromSlash(file))

	entries, err := s.Matrix(ctx, scratch)
	if err != nil {
		return fmt.Errorf("build the configuration matrix: %w", err)
	}
	before := s.snapshot(ctx, entries)

	edited := path
	encrypted := strings.HasSuffix(file, ".age")
	if encrypted {
		if edited, err = s.privateCopy(ctx, path, file, scratch); err != nil {
			return err
		}
	}
	changed, err := s.openEditor(ctx, file, path, edited, scratch, before.here)
	if err != nil {
		return err
	}
	if encrypted && changed {
		if err = s.reencrypt(ctx, edited, path); err != nil {
			return err
		}
	}
	if err = s.runCheck(ctx, e, edited, path); err != nil {
		return err
	}
	after := s.snapshot(ctx, entries)
	if !s.report(entries, before, after) || after.here.err != nil {
		if after.here.err != nil {
			fmt.Fprintln(s.Err, after.here.err)
		}
		fmt.Fprintf(s.Err, "%s✗ A configuration no longer renders, nothing applied.%s The source keeps the edit: run the command again to fix it (%s).\n", s.Style.R, s.Style.O, path)
		return ErrReported
	}
	return s.apply(ctx, e, before, after)
}

// privateCopy decrypts the source file into the scratch dir and returns the copy's path.
func (s *Service) privateCopy(ctx context.Context, path, file, scratch string) (string, error) {
	plain, err := s.CZ.Decrypt(ctx, path)
	if err != nil {
		return "", fmt.Errorf("decrypt %s: %w", file, err)
	}
	copyPath := filepath.Join(scratch, "edit")
	if err = os.WriteFile(copyPath, plain, 0o600); err != nil {
		return "", fmt.Errorf("write private copy: %w", err)
	}
	return copyPath, nil
}

// openEditor runs the editor on edited, with the preview beside it, and reports
// whether the file changed.
func (s *Service) openEditor(ctx context.Context, file, path, edited, scratch string, here result) (bool, error) {
	sumBefore, err := sum(edited)
	if err != nil {
		return false, err
	}
	stop := func() {}
	if s.Preview != nil && here.err == nil {
		stop = s.Preview.Start(ctx, PreviewArgs{File: file, Edited: edited, Scratch: scratch, Baseline: here.state})
	}
	editErr := s.Editor.Edit(ctx, edited)
	stop()
	if editErr != nil {
		fmt.Fprintf(s.Err, "%s✗ The editor failed: %s unchanged.%s\n", s.Style.R, path, s.Style.O)
		return false, ErrReported
	}
	sumAfter, err := sum(edited)
	if err != nil {
		return false, err
	}
	return sumAfter != sumBefore, nil
}

// runCheck runs the entry's check command, if it has one, on the edited file.
func (s *Service) runCheck(ctx context.Context, e *registry.Entry, edited, path string) error {
	if e == nil || len(e.Check) == 0 {
		return nil
	}
	wt, err := s.CZ.WorkingTree(ctx)
	if err != nil {
		return fmt.Errorf("working tree: %w", err)
	}
	if err = s.Check.Check(ctx, wt, withFile(e.Check, edited)); err != nil {
		fmt.Fprintf(s.Err, "%s✗ Check failed, nothing applied.%s The source keeps the edit: run the command again to fix it (%s).\n", s.Style.R, s.Style.O, path)
		return ErrReported
	}
	return nil
}

func (s *Service) reencrypt(ctx context.Context, edited, path string) error {
	plain, err := os.ReadFile(edited) //nolint:gosec // edited is the source file being edited or its private copy in the scratch dir
	if err != nil {
		return fmt.Errorf("read %s: %w", edited, err)
	}
	enc, err := s.CZ.Encrypt(ctx, plain)
	if err != nil {
		return fmt.Errorf("encrypt %s: %w", path, err)
	}
	tmp := path + ".new"
	defer os.Remove(tmp)
	if err := os.WriteFile(tmp, enc, 0o644); err != nil { //nolint:gosec // an encrypted source file is as readable as the others
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}

// apply shows this machine's changed targets and applies them on confirmation.
func (s *Service) apply(ctx context.Context, e *registry.Entry, before, after snapshot) error {
	home, err := s.CZ.TargetPath(ctx)
	if err != nil {
		return fmt.Errorf("target path: %w", err)
	}
	var targets, scripts []string
	if before.here.err == nil {
		targets, scripts = changedTargets(home, before.here.state, after.here.state)
	}
	st := s.Style
	if len(scripts) > 0 {
		fmt.Fprintf(s.Err, "%s▸ scripts changed:%s %s %s(they run on the next full chezmoi apply)%s\n", st.B, st.O, strings.Join(scripts, " "), st.D, st.O)
	}
	pending := ""
	if len(targets) > 0 {
		if pending, err = s.CZ.Status(ctx, targets...); err != nil {
			return fmt.Errorf("status: %w", err)
		}
	}
	if pending == "" {
		fmt.Fprintf(s.Err, "%s✓ Targets up to date: nothing to apply.%s\n", st.G, st.O)
	} else if err = s.offerApply(ctx, e, targets); err != nil {
		return err
	}
	return s.reportSourceChanges(ctx)
}

// changedTargets splits what the edit changed into target paths under home and
// the names of the scripts chezmoi runs on its next full apply.
func changedTargets(home string, before, after chez.State) (targets, scripts []string) {
	for _, t := range chez.Changed(before, after) {
		if name, ok := strings.CutPrefix(t, ".chezmoiscripts/"); ok {
			scripts = append(scripts, name)
		} else {
			targets = append(targets, filepath.Join(home, filepath.FromSlash(t)))
		}
	}
	return targets, scripts
}

// offerApply shows the diff of the targets and applies them when the user agrees.
func (s *Service) offerApply(ctx context.Context, e *registry.Entry, targets []string) error {
	st := s.Style
	show := targets
	if e != nil && e.First {
		show = targets[:1]
	}
	if err := s.CZ.Diff(ctx, s.Out, s.Err, show...); err != nil {
		return fmt.Errorf("diff: %w", err)
	}
	yes, err := s.Prompt.Confirm(st.Y+"apply now?"+st.O, false)
	if err != nil {
		return fmt.Errorf("confirm: %w", err)
	}
	if !yes {
		return nil
	}
	if err = s.CZ.Apply(ctx, s.Out, s.Err, targets...); err != nil {
		return fmt.Errorf("apply: %w", err)
	}
	fmt.Fprintf(s.Err, "%s✓ applied%s\n", st.G, st.O)
	if e != nil && e.Note != "" {
		fmt.Fprintf(s.Err, "%s%s%s\n", st.D, e.Note, st.O)
	}
	return nil
}

// reportSourceChanges tells the user the dotfiles source has uncommitted changes.
func (s *Service) reportSourceChanges(ctx context.Context) error {
	st := s.Style
	wt, err := s.CZ.WorkingTree(ctx)
	if err != nil {
		return fmt.Errorf("working tree: %w", err)
	}
	if changes, gitErr := s.Git.Status(ctx, wt); gitErr == nil && changes != "" {
		fmt.Fprintf(s.Err, "%s▸ the dotfiles source changed:%s ship a PR (e.g. /ship) from %s\n", st.B, st.O, wt)
		for line := range strings.SplitSeq(changes, "\n") {
			fmt.Fprintf(s.Err, "  %s%s%s\n", st.D, line, st.O)
		}
	}
	return nil
}

// report prints, per configuration, the targets the edit changed ("same as" an
// earlier configuration with the same list) and returns false when a configuration
// that rendered before no longer renders. One that already failed is only reported.
func (s *Service) report(entries []matrix.Entry, before, after snapshot) bool {
	st := s.Style
	fmt.Fprintf(s.Err, "%s▸ changed by this edit%s %s(per configuration, every harness selected)%s\n", st.B, st.O, st.D, st.O)
	ok := true
	changed := make([]string, len(entries))
	row := func(label, text string) { fmt.Fprintf(s.Err, "  %s%-17s%s %s\n", st.M, label, st.O, text) }
	for i, e := range entries {
		label := e.Label()
		b, a := before.conf[i], after.conf[i]
		switch {
		case a.err != nil && b.err == nil:
			row(label, st.R+"✗ render FAILED"+st.O)
			fmt.Fprintf(s.Err, "    %s\n", strings.ReplaceAll(a.err.Error(), "\n", "\n    "))
			ok = false
			continue
		case a.err != nil:
			row(label, st.D+"✗ render failed, already before this edit"+st.O)
			continue
		case b.err != nil:
			row(label, st.G+"✓ renders again"+st.O)
			continue
		}
		changed[i] = strings.Join(chez.Changed(b.state, a.state), ", ")
		same := ""
		for m := range i {
			if changed[i] != "" && changed[m] == changed[i] {
				same = entries[m].Label()
				break
			}
		}
		switch {
		case same != "":
			row(label, st.D+"same as "+same+st.O)
		case changed[i] != "":
			row(label, changed[i])
		default:
			row(label, st.D+"nothing"+st.O)
		}
	}
	return ok
}

func withFile(argv []string, file string) []string {
	out := slices.Clone(argv)
	for i, a := range out {
		out[i] = strings.ReplaceAll(a, "{file}", file)
	}
	return out
}

func sum(path string) ([32]byte, error) {
	b, err := os.ReadFile(path) //nolint:gosec // path is the source file being edited or its private copy
	if err != nil {
		return [32]byte{}, fmt.Errorf("read %s: %w", path, err)
	}
	return sha256.Sum256(b), nil
}
