package cli

import (
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
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

// writeCredentialDB creates opencode's credential table at path with two rows.
func writeCredentialDB(t *testing.T, path string) {
	t.Helper()
	// A plain file path is not safe in a DSN when it holds ?, # or %.
	dsn := (&url.URL{Scheme: "file", Path: sqlitePath(path)}).String()
	db, err := sql.Open("sqlite", dsn)
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
}

var wantCredentialRows = []provider.OpencodeRow{
	{ID: "c1", IntegrationID: "openai", Label: "me@example.com", Active: true, Value: []byte(`{"type":"oauth","access":"fake"}`)},
	{ID: "c2", IntegrationID: "mcp_foo", Active: false, Value: []byte(`{"type":"key"}`)},
}

func TestOpencodeCredentials(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	path := filepath.Join(s.Home(), "opencode.db")
	writeCredentialDB(t, path)

	got, err := s.OpencodeCredentials(path)
	if err != nil {
		t.Fatalf("OpencodeCredentials: %v", err)
	}
	if diff := cmp.Diff(wantCredentialRows, got); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
}

func TestOpencodeCredentialsAwkwardPath(t *testing.T) {
	t.Parallel()
	name := "odd#dir%41"
	if runtime.GOOS != "windows" {
		name += "?x"
	}
	dir := filepath.Join(t.TempDir(), name)
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "opencode.db")
	writeCredentialDB(t, path)

	got, err := testStore(t).OpencodeCredentials(path)
	if err != nil {
		t.Fatalf("OpencodeCredentials: %v", err)
	}
	if diff := cmp.Diff(wantCredentialRows, got); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
}

func TestSqliteDSN(t *testing.T) {
	t.Parallel()
	const q = "?_pragma=busy_timeout%282000%29&mode=ro"
	tests := []struct{ path, want string }{
		{"/home/me/opencode.db", "file:///home/me/opencode.db" + q},
		{"/a?b#c%d/x.db", "file:///a%3Fb%23c%25d/x.db" + q},
		{"C:/Users/me/x.db", "file:///C:/Users/me/x.db" + q},
	}
	for _, tt := range tests {
		if got := sqliteDSN(tt.path); got != tt.want {
			t.Errorf("sqliteDSN(%q) = %q, want %q", tt.path, got, tt.want)
		}
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

// stubSecurity points securityBin at a script that mimics the security tool:
// with -a it returns the item of that account only (the fresh one for wantAcct,
// exit 44 otherwise); without -a it returns the stale item of another account.
// It returns the script path.
func stubSecurity(t *testing.T, wantAcct string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell stub")
	}
	script := `#!/bin/sh
acct=""
while [ $# -gt 0 ]; do
  if [ "$1" = "-a" ]; then acct="$2"; fi
  shift
done
if [ -z "$acct" ]; then echo stale-token; exit 0; fi
if [ "$acct" = "` + wantAcct + `" ]; then echo fresh-token; exit 0; fi
exit 44
`
	path := filepath.Join(t.TempDir(), "security")
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSecretPrefersCurrentUserItem(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	s.goos = "darwin"
	s.securityBin = stubSecurity(t, currentUser(s.getenv))
	got, err := s.Secret("svc", "")
	if err != nil || got != "fresh-token" {
		t.Fatalf("Secret = %q, %v; want the current user's item", got, err)
	}
}

func TestSecretFallsBackToAnyAccount(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	s.goos = "darwin"
	s.securityBin = stubSecurity(t, "someone-else")
	got, err := s.Secret("svc", "")
	if err != nil || got != "stale-token" {
		t.Fatalf("Secret = %q, %v; want the service-only fallback", got, err)
	}
}
