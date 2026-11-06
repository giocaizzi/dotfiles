package cli

import (
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	_ "modernc.org/sqlite"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/provider"
)

func testStore(t *testing.T) osStore {
	t.Helper()
	return osStore{
		getenv: func(k string) string { return map[string]string{"HOME_TEST": "yes"}[k] },
		goos:   "linux",
		home:   t.TempDir(),
		now:    func() time.Time { return time.Unix(1700000000, 0) },
	}
}

func TestOSStoreFiles(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	path := filepath.Join(s.Home(), "a.txt")
	if s.Exists(path) {
		t.Fatal("Exists true for a missing file")
	}
	if _, err := s.ReadFile(path); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("ReadFile missing: err = %v, want fs.ErrNotExist", err)
	}
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	b, err := s.ReadFile(path)
	if err != nil || string(b) != "hello" || !s.Exists(path) {
		t.Fatalf("ReadFile = %q, %v", b, err)
	}
	if s.GOOS() != "linux" || s.Getenv("HOME_TEST") != "yes" || s.Now().Unix() != 1700000000 {
		t.Error("Deps passthroughs are wrong")
	}
}

func TestOSStoreSecretEmptyAccountNeedsDarwin(t *testing.T) {
	t.Parallel()
	_, err := testStore(t).Secret("svc", "")
	if err == nil {
		t.Fatal("expected an error for an empty account off macOS")
	}
}

func TestDecodeKeyringValue(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"plain":                                 "plain",
		"go-keyring-base64:" + b64("{\"a\":1}"): "{\"a\":1}",
		"go-keyring-encoded:" + hex.EncodeToString([]byte("sek")): "sek",
		"go-keyring-base64:!!notbase64":                           "go-keyring-base64:!!notbase64",
	}
	for in, want := range tests {
		if got := decodeKeyringValue(in); got != want {
			t.Errorf("decodeKeyringValue(%q) = %q, want %q", in, got, want)
		}
	}
}

func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

func TestOpencodeCredentials(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	path := filepath.Join(s.Home(), "opencode.db")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	for _, q := range []string{
		`CREATE TABLE credential (id TEXT, integration_id TEXT, label TEXT, active INTEGER, value TEXT)`,
		`INSERT INTO credential VALUES ('c1', 'openai', 'me@example.com', 1, '{"type":"oauth","access":"fake"}')`,
		`INSERT INTO credential VALUES ('c2', 'mcp_foo', NULL, 0, '{"type":"key"}')`,
	} {
		if _, err = db.ExecContext(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := s.OpencodeCredentials(path)
	if err != nil {
		t.Fatalf("OpencodeCredentials: %v", err)
	}
	want := []provider.OpencodeRow{
		{ID: "c1", IntegrationID: "openai", Label: "me@example.com", Active: true, Value: []byte(`{"type":"oauth","access":"fake"}`)},
		{ID: "c2", IntegrationID: "mcp_foo", Active: false, Value: []byte(`{"type":"key"}`)},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
}

func TestOpencodeCredentialsMissingDB(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	_, err := s.OpencodeCredentials(filepath.Join(s.Home(), "nope.db"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v, want fs.ErrNotExist", err)
	}
}
