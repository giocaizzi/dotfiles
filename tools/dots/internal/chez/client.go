package chez

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
)

// Overrides is data layered over a configuration (chezmoi --override-data).
type Overrides map[string]any

// Entry is one target of the target state (chezmoi dump).
type Entry struct {
	Type     string `json:"type"`
	Contents string `json:"contents"`
	Linkname string `json:"linkname"`
	Raw      string `json:"-"` // the entry's JSON, which decides whether two states differ
}

// State is the target state of a configuration: target path to entry.
type State map[string]Entry

// Client calls chezmoi. The zero value is not usable; build one with New.
type Client struct {
	run    Runner
	bin    string
	global []string
}

// New returns a client that runs bin through r with global flags (--config,
// --source, ...) before every subcommand.
func New(r Runner, bin string, global ...string) *Client {
	return &Client{run: r, bin: bin, global: global}
}

// With returns a client with more global flags; later flags win.
func (c *Client) With(global ...string) *Client {
	n := *c
	n.global = append(append([]string(nil), c.global...), global...)
	return &n
}

// render marks a read-only render of the source: .chezmoiexternal.toml.tmpl then
// declares no externals, so nothing is downloaded for a preview.
func render(o Overrides) Overrides { return o.Clone(Overrides{"dotsRender": true}) }

func (c *Client) args(o Overrides, sub ...string) ([]string, error) {
	a := append([]string{c.bin}, c.global...)
	if len(o) > 0 {
		j, err := json.Marshal(o)
		if err != nil {
			return nil, fmt.Errorf("encode overrides: %w", err)
		}
		a = append(a, "--override-data", string(j))
	}
	return append(a, sub...), nil
}

func (c *Client) output(ctx context.Context, o Overrides, stdin io.Reader, sub ...string) ([]byte, error) {
	a, err := c.args(o, sub...)
	if err != nil {
		return nil, err
	}
	out, err := Output(ctx, c.run, Cmd{Args: a, Stdin: stdin})
	if err != nil {
		return nil, fmt.Errorf("chezmoi %s: %w", sub[0], err)
	}
	return out, nil
}

func (c *Client) stream(ctx context.Context, stdout, stderr io.Writer, stdin io.Reader, sub ...string) error {
	a, err := c.args(nil, sub...)
	if err != nil {
		return err
	}
	return c.run.Run(ctx, Cmd{Args: a, Stdin: stdin, Stdout: stdout, Stderr: stderr})
}

// Data returns the template data (chezmoi data).
func (c *Client) Data(ctx context.Context, o Overrides) (map[string]any, error) {
	out, err := c.output(ctx, render(o), nil, "data", "--format", "json")
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		return nil, fmt.Errorf("decode chezmoi data: %w", err)
	}
	return m, nil
}

// Dump returns the target state without directories and externals.
func (c *Client) Dump(ctx context.Context, o Overrides) (State, error) {
	out, err := c.output(ctx, render(o), nil, "dump", "--exclude", "dirs,externals", "--format", "json")
	if err != nil {
		return nil, err
	}
	return DecodeState(out)
}

// DecodeState decodes chezmoi dump --format json.
func DecodeState(b []byte) (State, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("decode chezmoi dump: %w", err)
	}
	s := make(State, len(raw))
	for k, v := range raw {
		var e Entry
		if err := json.Unmarshal(v, &e); err != nil {
			return nil, fmt.Errorf("decode chezmoi dump entry %s: %w", k, err)
		}
		e.Raw = string(v)
		s[k] = e
	}
	return s, nil
}

// Changed lists, sorted, the targets whose entry differs between two states.
func Changed(before, after State) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range []State{before, after} {
		for k := range s {
			if seen[k] {
				continue
			}
			seen[k] = true
			if before[k].Raw != after[k].Raw {
				out = append(out, k)
			}
		}
	}
	slices.Sort(out)
	return out
}

// Managed maps each managed target to its source path relative to the source dir.
func (c *Client) Managed(ctx context.Context, o Overrides) (map[string]string, error) {
	out, err := c.output(ctx, render(o), nil, "managed", "-i", "files,symlinks,scripts", "-x", "externals", "-p", "all", "-f", "json")
	if err != nil {
		return nil, err
	}
	var raw map[string]struct {
		SourceRelative string `json:"sourceRelative"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("decode chezmoi managed: %w", err)
	}
	m := make(map[string]string, len(raw))
	for k, v := range raw {
		m[k] = v.SourceRelative
	}
	return m, nil
}

// Cat returns what chezmoi would write for target.
func (c *Client) Cat(ctx context.Context, target string, o Overrides) ([]byte, error) {
	return c.output(ctx, render(o), nil, "cat", target)
}

// SourcePath returns the source dir (target empty) or a target's source file.
func (c *Client) SourcePath(ctx context.Context, target string, o Overrides) (string, error) {
	sub := []string{"source-path"}
	if target != "" {
		sub = append(sub, target)
	}
	out, err := c.output(ctx, render(o), nil, sub...)
	return strings.TrimSpace(string(out)), err
}

// TargetPath returns the destination dir, normally the home directory.
func (c *Client) TargetPath(ctx context.Context) (string, error) {
	out, err := c.output(ctx, nil, nil, "target-path")
	return strings.TrimSpace(string(out)), err
}

// WorkingTree returns the root of the git repository that holds the source dir.
func (c *Client) WorkingTree(ctx context.Context) (string, error) {
	out, err := c.output(ctx, nil, nil, "dump-config", "--format", "json")
	if err != nil {
		return "", err
	}
	var cfg struct {
		WorkingTree string `json:"workingTree"`
	}
	if err := json.Unmarshal(out, &cfg); err != nil {
		return "", fmt.Errorf("decode chezmoi dump-config: %w", err)
	}
	return cfg.WorkingTree, nil
}

// Status returns chezmoi status for targets (empty when they are up to date).
func (c *Client) Status(ctx context.Context, targets ...string) (string, error) {
	out, err := c.output(ctx, nil, nil, append([]string{"status"}, targets...)...)
	return strings.TrimSpace(string(out)), err
}

// Diff writes chezmoi diff for targets to w.
func (c *Client) Diff(ctx context.Context, w, errw io.Writer, targets ...string) error {
	return c.stream(ctx, w, errw, nil, append([]string{"diff"}, targets...)...)
}

// Apply applies targets.
func (c *Client) Apply(ctx context.Context, stdout, stderr io.Writer, targets ...string) error {
	return c.stream(ctx, stdout, stderr, nil, append([]string{"apply"}, targets...)...)
}

// Encrypt encrypts plain with the configured encryption, armored as chezmoi writes it.
func (c *Client) Encrypt(ctx context.Context, plain []byte) ([]byte, error) {
	return c.output(ctx, nil, bytes.NewReader(plain), "encrypt")
}

// Decrypt decrypts the encrypted file at path.
func (c *Client) Decrypt(ctx context.Context, path string) ([]byte, error) {
	return c.output(ctx, nil, nil, "decrypt", path)
}

// Execute renders a template read from tmpl.
func (c *Client) Execute(ctx context.Context, o Overrides, tmpl []byte, init bool) ([]byte, error) {
	sub := []string{"execute-template"}
	if init {
		sub = append(sub, "--init", "--no-tty")
	}
	return c.output(ctx, render(o), bytes.NewReader(tmpl), sub...)
}

// Clone returns a copy of o with the entries of more added.
func (o Overrides) Clone(more Overrides) Overrides {
	n := Overrides{}
	maps.Copy(n, o)
	maps.Copy(n, more)
	return n
}
