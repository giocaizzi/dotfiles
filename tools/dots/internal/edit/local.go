package edit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/registry"
)

// editLocal edits a file of a local entry's folder: unmanaged, this machine only. The
// edit goes to a private copy that replaces the live file once the entry's check
// passes; no render, diff, apply or PR. name empty picks an existing file.
func (s *Service) editLocal(ctx context.Context, e registry.Entry, dir, name string) error {
	if name == "" {
		picked, err := s.pickLocal(ctx, e, dir)
		if err != nil {
			return err
		}
		name = picked
	}
	if strings.HasPrefix(name, ".") || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("%q is not a file name: %w", name, ErrUsage)
	}
	live := filepath.Join(dir, name)
	tmpDir, err := os.MkdirTemp("", "dots-")
	if err != nil {
		return fmt.Errorf("make scratch dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	tmp := filepath.Join(tmpDir, name)
	orig, err := os.ReadFile(live) //nolint:gosec // live is a file of the registry entry's own folder under the home dir
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", live, err)
	}
	if err = os.WriteFile(tmp, orig, 0o600); err != nil { //nolint:gosec // tmp is inside the scratch dir this process created
		return fmt.Errorf("write private copy: %w", err)
	}
	if err = s.editUntilValid(ctx, e, tmp, live); err != nil {
		return err
	}
	edited, err := os.ReadFile(tmp) //nolint:gosec // tmp is inside the scratch dir this process created
	if err != nil {
		return fmt.Errorf("read %s: %w", tmp, err)
	}
	if string(edited) == string(orig) {
		return nil
	}
	return replaceLive(dir, filepath.Base(live), edited)
}

// pickLocal lets the user pick an existing file of the entry's folder.
func (s *Service) pickLocal(ctx context.Context, e registry.Entry, dir string) (string, error) {
	st := s.Style
	files := regularFiles(dir)
	if len(files) == 0 {
		fmt.Fprintf(s.Err, "%s✗ dots: %s is empty;%s create a file with chezmoi dots %s/<name>\n", st.R, dir, st.O, dir)
		return "", ErrReported
	}
	pick, err := s.Prompt.Pick(ctx, e.Name, files)
	if err != nil {
		return "", fmt.Errorf("pick: %w", err)
	}
	if pick == "" {
		return "", ErrReported
	}
	return pick, nil
}

// editUntilValid opens tmp in the editor until the entry's check passes or the user gives up.
func (s *Service) editUntilValid(ctx context.Context, e registry.Entry, tmp, live string) error {
	st := s.Style
	for {
		if err := s.Editor.Edit(ctx, tmp); err != nil {
			fmt.Fprintf(s.Err, "%s✗ The editor failed: %s unchanged.%s\n", st.R, live, st.O)
			return ErrReported
		}
		if len(e.Check) == 0 || s.Check.Check(ctx, "", withFile(e.Check, tmp)) == nil {
			return nil
		}
		again, err := s.Prompt.Confirm(fmt.Sprintf("%s✗ Check failed, %s unchanged.%s %sEdit again?%s", st.R, live, st.O, st.Y, st.O), true)
		if err != nil || !again {
			return ErrReported
		}
	}
}

// replaceLive writes content over the live file, through a rename within its folder,
// so the live file is never half written.
func replaceLive(dir, name string, content []byte) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("make %s: %w", dir, err)
	}
	live := filepath.Join(dir, name)
	next := filepath.Join(dir, "."+name+".new")
	defer os.Remove(next)
	if err := os.WriteFile(next, content, 0o600); err != nil { //nolint:gosec // next is a sibling of the live file in the entry's own folder
		return fmt.Errorf("write %s: %w", next, err)
	}
	if err := os.Rename(next, live); err != nil {
		return fmt.Errorf("replace %s: %w", live, err)
	}
	return nil
}

func regularFiles(dir string) []string {
	ents, _ := os.ReadDir(dir)
	var out []string
	for _, e := range ents {
		if e.Type().IsRegular() {
			out = append(out, e.Name())
		}
	}
	return out
}
