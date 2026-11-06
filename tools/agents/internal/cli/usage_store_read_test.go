package cli

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestOSStoreReadFileLimits(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	write := func(name string, n int) string {
		p := filepath.Join(s.Home(), name)
		if err := os.WriteFile(p, bytes.Repeat([]byte("x"), n), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}

	if b, err := s.ReadFile(write("ok.json", 100)); err != nil || len(b) != 100 {
		t.Errorf("normal file: len %d, err %v", len(b), err)
	}
	if b, err := s.ReadFile(write("max.json", maxCredentialFile)); err != nil || len(b) != maxCredentialFile {
		t.Errorf("file of exactly the cap: len %d, err %v", len(b), err)
	}
	if _, err := s.ReadFile(write("big.json", maxCredentialFile+1)); !errors.Is(err, errTooLarge) {
		t.Errorf("oversized file: err = %v, want errTooLarge", err)
	}
	if _, err := s.ReadFile(s.Home()); !errors.Is(err, errNotRegular) {
		t.Errorf("directory: err = %v, want errNotRegular", err)
	}
	if _, err := s.ReadFile(filepath.Join(s.Home(), "missing")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("missing file: err = %v, want fs.ErrNotExist", err)
	}
}
