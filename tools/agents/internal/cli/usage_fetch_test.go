package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/provider"
	usg "github.com/giocaizzi/dotfiles/tools/agents/internal/usage"
)

// secret is the fake token every test uses; it must never reach a snapshot.
const secret = "tok-SECRET-0123456789"

var fixedNow = time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

// Parser-dependent assertions are loose while internal/usage is a stub: after
// merging the parser units, re-check the tests marked RECHECK and assert the
// parsed limits, plans and the stale/ok statuses exactly.

// seen is what the fake servers keep of a request.
type seen struct {
	Method string
	URL    *url.URL
	Header http.Header
}

type recorder struct {
	mu   sync.Mutex
	reqs []seen
}

func (r *recorder) add(req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reqs = append(r.reqs, seen{req.Method, req.URL, req.Header.Clone()})
}

func (r *recorder) all() []seen {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]seen(nil), r.reqs...)
}

func testFetcher(t *testing.T, srv *httptest.Server) fetcher {
	t.Helper()
	f := newFetcher(t.TempDir(), func(string) string { return "" })
	f.now = func() time.Time { return fixedNow }
	f.agyPath = func() (string, error) { return "", os.ErrNotExist }
	if srv != nil {
		f.anthropicURL, f.chatgptURL, f.githubURL = srv.URL, srv.URL, srv.URL
		f.client = srv.Client()
	}
	return f
}

func oauthAccount(prov string) provider.Account {
	return provider.Account{
		Provider: prov, Label: "me@example.com", Harnesses: []string{"h1", "h2"},
		Auth: provider.AuthOAuth, Plan: usg.Plan{Kind: usg.KindIndividual, Name: "Plus"},
		Cred: provider.NewCredential(secret, "acct-42"),
	}
}

// assertNoSecret fails when the snapshot, as JSON or text, holds the token.
func assertNoSecret(t *testing.T, s usg.Snapshot) {
	t.Helper()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), secret) || strings.Contains(s.Note, secret) {
		t.Errorf("snapshot leaks the token: %s", b)
	}
}

func stubbed(s usg.Snapshot) bool { return strings.Contains(s.Note, usg.ErrNotImplemented.Error()) }

func TestFetchAnthropicHeaders(t *testing.T) {
	t.Parallel()
	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.add(r)
		_, _ = w.Write([]byte(`{"five_hour":{"utilization":31,"resets_at":"2026-11-06T14:00:00Z"}}`))
	}))
	defer srv.Close()

	snap := testFetcher(t, srv).fetch(t.Context(), oauthAccount(provider.Anthropic), time.Second)

	reqs := rec.all()
	if len(reqs) != 1 {
		t.Fatalf("requests = %d, want 1", len(reqs))
	}
	r := reqs[0]
	if r.URL.Path != "/api/oauth/usage" || r.Method != http.MethodGet {
		t.Errorf("request = %s %s", r.Method, r.URL.Path)
	}
	for k, want := range map[string]string{
		"Authorization":  "Bearer " + secret,
		"Anthropic-Beta": "oauth-2025-04-20",
		"User-Agent":     "chezmoi-agents",
	} {
		if got := r.Header.Get(k); got != want {
			t.Errorf("header %s = %q, want %q", k, got, want)
		}
	}
	if snap.Provider != provider.Anthropic || snap.Account != "me@example.com" ||
		!cmp.Equal(snap.Harnesses, []string{"h1", "h2"}) || snap.Plan.Name != "Plus" || !snap.FetchedAt.Equal(fixedNow) {
		t.Errorf("snapshot identity wrong: %+v", snap)
	}
	if !stubbed(snap) { // RECHECK: assert StatusOK, Source "api" and the parsed 5h limit.
		if snap.Status != usg.StatusOK || snap.Source != "api" {
			t.Errorf("status/source = %s/%s", snap.Status, snap.Source)
		}
	}
	assertNoSecret(t, snap)
}

func TestFetchOpenAIHeaders(t *testing.T) {
	t.Parallel()
	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.add(r)
		_, _ = w.Write([]byte(`{"plan_type":"plus"}`))
	}))
	defer srv.Close()

	snap := testFetcher(t, srv).fetch(t.Context(), oauthAccount(provider.OpenAI), time.Second)

	r := rec.all()[0]
	if r.URL.Path != "/backend-api/wham/usage" {
		t.Errorf("path = %s", r.URL.Path)
	}
	for k, want := range map[string]string{
		"Authorization":      "Bearer " + secret,
		"Chatgpt-Account-Id": "acct-42",
		"Originator":         "codex_cli_rs",
	} {
		if got := r.Header.Get(k); got != want {
			t.Errorf("header %s = %q, want %q", k, got, want)
		}
	}
	assertNoSecret(t, snap)
}

func TestFetchOpenAINoAccountIDHeader(t *testing.T) {
	t.Parallel()
	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.add(r)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	acct := oauthAccount(provider.OpenAI)
	acct.Cred = provider.NewCredential(secret, "")

	testFetcher(t, srv).fetch(t.Context(), acct, time.Second)

	if _, ok := rec.all()[0].Header["Chatgpt-Account-Id"]; ok {
		t.Error("ChatGPT-Account-Id sent without an account id")
	}
}

func TestFetchCopilotAuthRetry(t *testing.T) {
	t.Parallel()
	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.add(r)
		if strings.HasPrefix(r.Header.Get("Authorization"), "token ") {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("bad credentials " + secret))
			return
		}
		_, _ = w.Write([]byte(`{"copilot_plan":"business"}`))
	}))
	defer srv.Close()
	acct := oauthAccount(provider.GitHubCopilot)
	acct.Login = "octocat"

	snap := testFetcher(t, srv).fetch(t.Context(), acct, time.Second)

	reqs := rec.all()
	if len(reqs) < 2 {
		t.Fatalf("requests = %d, want the token request and the Bearer retry", len(reqs))
	}
	if got := reqs[0].Header.Get("Authorization"); got != "token "+secret {
		t.Errorf("first Authorization = %q", got)
	}
	if got := reqs[1].Header.Get("Authorization"); got != "Bearer "+secret {
		t.Errorf("retry Authorization = %q", got)
	}
	if got := reqs[0].Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q", got)
	}
	// RECHECK: once ParseCopilotUser exists, assert the billing call
	// (/users/octocat/settings/billing/premium_request/usage, X-GitHub-Api-Version
	// 2026-03-10) and the "managed by org" hidden spend limit on 403/404.
	assertNoSecret(t, snap)
}

func TestFetchHTTPErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code int
		note string
	}{
		{http.StatusUnauthorized, "unauthorized (401)"},
		{http.StatusForbidden, "unauthorized (403)"},
		{http.StatusTooManyRequests, "rate limited (429)"},
		{http.StatusInternalServerError, "HTTP 500"},
	}
	for _, tt := range tests {
		t.Run(tt.note, func(t *testing.T) {
			t.Parallel()
			var calls int
			var mu sync.Mutex
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				mu.Lock()
				calls++
				mu.Unlock()
				w.Header().Set("X-Echo", secret)
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(`{"error":"` + secret + `"}`))
			}))
			defer srv.Close()

			for _, prov := range []string{provider.Anthropic, provider.OpenAI} {
				snap := testFetcher(t, srv).fetch(t.Context(), oauthAccount(prov), time.Second)
				if snap.Status != usg.StatusUnavailable || snap.Note != tt.note {
					t.Errorf("%s: status/note = %s/%q, want unavailable/%q", prov, snap.Status, snap.Note, tt.note)
				}
				assertNoSecret(t, snap)
			}
			mu.Lock()
			defer mu.Unlock()
			if calls != 2 {
				t.Errorf("calls = %d, want 2 (never retried)", calls)
			}
		})
	}
}

func TestFetchNetworkError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.NotFoundHandler())
	f := testFetcher(t, srv)
	srv.Close() // connection refused

	snap := f.fetch(t.Context(), oauthAccount(provider.Anthropic), time.Second)

	if snap.Status != usg.StatusUnavailable || !strings.HasPrefix(snap.Note, "unavailable: ") {
		t.Errorf("status/note = %s/%q", snap.Status, snap.Note)
	}
	if strings.Contains(snap.Note, "http://") || strings.Contains(snap.Note, srv.URL) {
		t.Errorf("note carries the URL: %q", snap.Note)
	}
	assertNoSecret(t, snap)
}

func TestFetchTimeout(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)

	snap := testFetcher(t, srv).fetch(t.Context(), oauthAccount(provider.Anthropic), 50*time.Millisecond)

	if snap.Status != usg.StatusUnavailable || !strings.HasPrefix(snap.Note, "unavailable: ") {
		t.Errorf("status/note = %s/%q", snap.Status, snap.Note)
	}
}

func TestFetchPresetStatusMakesNoRequest(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request %s", r.URL.Path)
	}))
	defer srv.Close()

	for _, st := range []usg.Status{usg.StatusExpired, usg.StatusNoAPI, usg.StatusNotVisible, usg.StatusSkipped, usg.StatusNotLoggedIn} {
		acct := oauthAccount(provider.Anthropic)
		acct.Status, acct.Note = st, "note for "+string(st)
		snap := testFetcher(t, srv).fetch(t.Context(), acct, time.Second)
		if snap.Status != st || snap.Note != "note for "+string(st) || len(snap.Limits) != 0 {
			t.Errorf("%s: snapshot = %+v", st, snap)
		}
	}
}

func TestFetchUnsupportedAccountIsNoAPI(t *testing.T) {
	t.Parallel()
	acct := provider.Account{Provider: "openrouter", Label: "k", Auth: provider.AuthAPIKey, Cred: provider.NewCredential(secret, "")}
	snap := testFetcher(t, nil).fetch(t.Context(), acct, time.Second)
	if snap.Status != usg.StatusNoAPI {
		t.Errorf("status = %s, want no-api", snap.Status)
	}
}

func TestFetchBodyLimit(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 3<<20)))
	}))
	defer srv.Close()
	body, err := testFetcher(t, srv).get(t.Context(), srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) != maxBody {
		t.Errorf("body = %d bytes, want %d", len(body), maxBody)
	}
}

func writeScript(t *testing.T, dir, body string) string {
	t.Helper()
	path := filepath.Join(dir, "agy")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func googleAccount() provider.Account {
	return provider.Account{
		Provider: provider.Google, Harnesses: []string{"antigravity"}, Auth: provider.AuthCLI,
		Plan: usg.Plan{Kind: usg.KindUnknown, Raw: "antigravity"},
	}
}

func TestFetchAgyRunsUsageCommand(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("shell script fake")
	}
	dir := t.TempDir()
	script := writeScript(t, dir, `echo "$@" > "`+dir+`/args"
pwd > "`+dir+`/cwd"
printf '{"num_turns":0,"command":{"name":"usage","data":{"groups":[]}}}'
`)
	f := testFetcher(t, nil)
	f.agyPath = func() (string, error) { return script, nil }

	snap := f.fetch(t.Context(), googleAccount(), 10*time.Second)

	args, err := os.ReadFile(filepath.Join(dir, "args"))
	if err != nil {
		t.Fatalf("fake agy was not run: %v", err)
	}
	if got, want := strings.TrimSpace(string(args)), "-p /usage --output-format json --print-timeout 10s"; got != want {
		t.Errorf("agy args = %q, want %q", got, want)
	}
	cwd, _ := os.ReadFile(filepath.Join(dir, "cwd"))
	wantCwd, _ := filepath.EvalSymlinks(os.TempDir()) //nolint:usetesting // asserts the production cwd
	gotCwd, _ := filepath.EvalSymlinks(strings.TrimSpace(string(cwd)))
	if gotCwd != wantCwd {
		t.Errorf("agy cwd = %q, want %q", gotCwd, wantCwd)
	}
	// RECHECK: with ParseAgyUsage implemented assert StatusOK, Source "agy" and the buckets.
	if snap.Status == usg.StatusOK && snap.Source != "agy" {
		t.Errorf("source = %q, want agy", snap.Source)
	}
	if snap.Status != usg.StatusOK && snap.Note != agyUnavailable {
		t.Errorf("status/note = %s/%q", snap.Status, snap.Note)
	}
}

func TestFetchAgyFailureIsUnavailable(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("shell script fake")
	}
	for name, body := range map[string]string{
		"exit code": "exit 3\n",
		"garbage":   "echo not json\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			script := writeScript(t, t.TempDir(), body)
			f := testFetcher(t, nil)
			f.agyPath = func() (string, error) { return script, nil }
			snap := f.fetch(t.Context(), googleAccount(), 5*time.Second)
			if snap.Status != usg.StatusUnavailable || snap.Note != agyUnavailable {
				t.Errorf("status/note = %s/%q", snap.Status, snap.Note)
			}
		})
	}
}

func TestFetchAgyMissingFallsBackToLog(t *testing.T) {
	t.Parallel()
	f := testFetcher(t, nil) // agyPath: not found
	logDir := filepath.Join(f.home, ".gemini", "antigravity-cli", "log")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	line := "E1106 11:07:11.896026 2480 errorreport.go:224] RESOURCE_EXHAUSTED (code 429): Individual quota reached. Resets in 4h58m35s.\n"
	for _, name := range []string{"cli-20261105_090000.log", "cli-20261106_100000.log"} {
		if err := os.WriteFile(filepath.Join(logDir, name), []byte(line), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	snap := f.fetch(t.Context(), googleAccount(), time.Second)

	if snap.Status == usg.StatusStale { // RECHECK: with the real ParseAgyLog this is the expected outcome.
		if snap.Source != "agy-log" || len(snap.Limits) == 0 {
			t.Errorf("fallback snapshot = %+v", snap)
		}
		for _, l := range snap.Limits {
			if l.ResetsAt == nil || !l.ResetsAt.After(fixedNow) {
				t.Errorf("kept a limit that already reset: %+v", l)
			}
		}
		return
	}
	if snap.Status != usg.StatusUnavailable || snap.Note != agyUnavailable {
		t.Errorf("status/note = %s/%q", snap.Status, snap.Note)
	}
}

func TestAgyLogFallbackWithoutLogs(t *testing.T) {
	t.Parallel()
	if _, ok := testFetcher(t, nil).agyLogFallback(googleAccount()); ok {
		t.Error("fallback found a log in an empty home")
	}
}

func TestCodexSessionFallback(t *testing.T) {
	t.Parallel()
	f := testFetcher(t, nil)
	if _, ok := f.codexSessionFallback(oauthAccount(provider.OpenAI)); ok {
		t.Fatal("fallback found a session in an empty home")
	}

	day := filepath.Join(f.home, ".codex", "sessions", "2026", "11", "05")
	if err := os.MkdirAll(day, 0o755); err != nil {
		t.Fatal(err)
	}
	line := `{"type":"event_msg","payload":{"type":"token_count","rate_limits":{"plan_type":"plus","primary":{"used_percent":12,"window_minutes":300,"resets_at":1793000000}}}}` + "\n"
	if err := os.WriteFile(filepath.Join(day, "rollout-2026-11-05T10-00-00-abc.jsonl"), []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}

	snap, ok := f.codexSessionFallback(oauthAccount(provider.OpenAI))
	if ok { // RECHECK: with ParseCodexSession implemented, ok must be true.
		if snap.Status != usg.StatusStale || snap.Source != "codex-session" {
			t.Errorf("snapshot = %+v", snap)
		}
	}
}

func TestCodexSessionFallbackRespectsCodexHome(t *testing.T) {
	t.Parallel()
	f := testFetcher(t, nil)
	other := t.TempDir()
	f.getenv = func(k string) string {
		if k == "CODEX_HOME" {
			return other
		}
		return ""
	}
	// An old session far outside the 7-day window must be ignored either way.
	old := filepath.Join(other, "sessions", "2026", "01", "01")
	if err := os.MkdirAll(old, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, ok := f.codexSessionFallback(oauthAccount(provider.OpenAI)); ok {
		t.Error("fallback used a session with no files")
	}
}

func TestOpenAIFailureUsesSessionFallback(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	snap := testFetcher(t, srv).fetch(t.Context(), oauthAccount(provider.OpenAI), time.Second)

	// No session on disk: the failure stays unavailable.
	if snap.Status != usg.StatusUnavailable || snap.Note != "rate limited (429)" {
		t.Errorf("status/note = %s/%q", snap.Status, snap.Note)
	}
}

func TestLiveSnapshotsKeepOrderAndRunInParallel(t *testing.T) {
	t.Parallel()
	var started sync.WaitGroup
	started.Add(3)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		started.Done()
		started.Wait() // all three requests must be in flight together
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	accts := []provider.Account{oauthAccount(provider.Anthropic), oauthAccount(provider.OpenAI), oauthAccount(provider.GitHubCopilot)}
	snaps := liveSnapshots(t.Context(), accts, testFetcher(t, srv), 5*time.Second)

	got := make([]string, 0, len(snaps))
	for _, s := range snaps {
		got = append(got, s.Provider)
	}
	want := []string{provider.Anthropic, provider.OpenAI, provider.GitHubCopilot}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("order (-want +got):\n%s", diff)
	}
}
