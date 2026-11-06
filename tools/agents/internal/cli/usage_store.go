package cli

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/zalando/go-keyring"
	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver

	"github.com/giocaizzi/dotfiles/tools/agents/internal/provider"
)

const secretTimeout = 5 * time.Second

// osStore is the real provider.Store: files, the OS credential store and
// opencode's SQLite database, all read-only and into memory.
type osStore struct {
	getenv func(string) string
	goos   string
	home   string
	now    func() time.Time
}

// newOSStore builds the store from the CLI dependencies.
func newOSStore(d Deps) osStore {
	home, _ := os.UserHomeDir()
	return osStore{getenv: d.Getenv, goos: d.GOOS, home: home, now: time.Now}
}

var _ provider.Store = osStore{}

func (s osStore) ReadFile(path string) ([]byte, error) {
	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return b, nil
}

func (s osStore) Exists(path string) bool {
	_, err := os.Stat(filepath.Clean(path))
	return err == nil
}

func (s osStore) Getenv(key string) string { return s.getenv(key) }
func (s osStore) Home() string             { return s.home }
func (s osStore) GOOS() string             { return s.goos }
func (s osStore) Now() time.Time           { return s.now() }

// Secret reads one credential-store item. An empty account means "any account":
// only macOS supports it, through the security tool. Errors never carry the
// secret.
func (s osStore) Secret(service, account string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), secretTimeout)
	defer cancel()

	if account == "" {
		if s.goos != "darwin" {
			return "", fmt.Errorf("credential %q: an account is required on %s", service, s.goos)
		}
		return securityFindPassword(ctx, service)
	}

	type result struct {
		val string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		v, err := keyring.Get(service, account)
		ch <- result{v, err}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			return "", fmt.Errorf("credential %q: %w", service, r.err)
		}
		return r.val, nil
	case <-ctx.Done():
		return "", fmt.Errorf("credential %q: timed out", service)
	}
}

func securityFindPassword(ctx context.Context, service string) (string, error) {
	cmd := exec.CommandContext(ctx, "/usr/bin/security", "find-generic-password", "-s", service, "-w")
	out, err := cmd.Output()
	if err != nil {
		// Only the exit status is known here; stderr is never echoed.
		return "", fmt.Errorf("credential %q: not found", service)
	}
	return decodeKeyringValue(strings.TrimSpace(string(out))), nil
}

// decodeKeyringValue undoes go-keyring's storage encodings on macOS.
func decodeKeyringValue(v string) string {
	const (
		b64Prefix = "go-keyring-base64:"
		encPrefix = "go-keyring-encoded:"
	)
	switch {
	case strings.HasPrefix(v, b64Prefix):
		if b, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(v, b64Prefix)); err == nil {
			return string(b)
		}
	case strings.HasPrefix(v, encPrefix):
		if b, err := hex.DecodeString(strings.TrimPrefix(v, encPrefix)); err == nil {
			return string(b)
		}
	}
	return v
}

// OpencodeCredentials reads opencode's V2 credential table, read-only.
func (s osStore) OpencodeCredentials(dbPath string) ([]provider.OpencodeRow, error) {
	path := filepath.Clean(dbPath)
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("opencode database: %w", err)
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro&_pragma=busy_timeout(2000)")
	if err != nil {
		return nil, fmt.Errorf("open opencode database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), secretTimeout)
	defer cancel()
	rows, err := db.QueryContext(ctx, `SELECT id, integration_id, label, active, value FROM credential`)
	if err != nil {
		return nil, fmt.Errorf("query opencode credentials: %w", err)
	}
	defer rows.Close()

	var out []provider.OpencodeRow
	for rows.Next() {
		var (
			id, integration, label sql.NullString
			active                 sql.NullBool
			value                  []byte
		)
		if err := rows.Scan(&id, &integration, &label, &active, &value); err != nil {
			return nil, fmt.Errorf("read opencode credential: %w", err)
		}
		out = append(out, provider.OpencodeRow{
			ID:            id.String,
			IntegrationID: integration.String,
			Label:         label.String,
			Active:        active.Bool,
			Value:         value,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read opencode credentials: %w", err)
	}
	return out, nil
}
