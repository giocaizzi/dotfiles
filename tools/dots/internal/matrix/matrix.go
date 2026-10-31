// Package matrix is the set of configurations a dotfiles edit is rendered for: every
// OS and machine role. Each configuration's data comes from rendering the real
// .chezmoi.toml.tmpl with that role's answers, so it cannot drift from what
// `chezmoi init` stores.
package matrix

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
)

// Config is one OS and machine role.
type Config struct {
	OS   string
	Role string
}

// OSes are the supported operating systems.
func OSes() []string { return []string{"darwin", "linux", "windows"} }

// Roles are the machine roles.
func Roles() []string { return []string{"personal", roleWork, roleLight} }

const (
	roleWork  = "work"
	roleLight = "light"
)

// All returns every configuration, OS-major.
func All() []Config {
	all := make([]Config, 0, len(OSes())*len(Roles()))
	for _, o := range OSes() {
		for _, r := range Roles() {
			all = append(all, Config{OS: o, Role: r})
		}
	}
	return all
}

// Label is "<os>/<role>".
func (c Config) Label() string { return c.OS + "/" + c.Role }

// Parse reads "<os>", "<os>,<role>" or "<os>/<role>"; the role defaults to personal.
func Parse(s string) (Config, error) {
	os, role, _ := strings.Cut(strings.NewReplacer(",", "/").Replace(s), "/")
	if role == "" {
		role = "personal"
	}
	if !slices.Contains(OSes(), os) || !slices.Contains(Roles(), role) {
		return Config{}, fmt.Errorf("unknown configuration %q (os %s; role %s)", s, strings.Join(OSes(), "|"), strings.Join(Roles(), "|"))
	}
	return Config{OS: os, Role: role}, nil
}

// Entry is a configuration with the data it renders with.
type Entry struct {
	Config

	Overrides chez.Overrides
}

var harnessList = regexp.MustCompile(`\$all\s*:=\s*list((?:\s+"[a-z]+")+)`)

// Harnesses reads the harness choices offered by the config template.
func Harnesses(tmpl []byte) ([]string, error) {
	m := harnessList.FindSubmatch(tmpl)
	if m == nil {
		return nil, errors.New(`config template: no ` + "`$all := list \"…\"`" + ` harness list`)
	}
	return strings.Fields(strings.ReplaceAll(string(m[1]), `"`, "")), nil
}

// DataKeys lists the keys of the template's [data] table.
func DataKeys(tmpl []byte) []string {
	var keys []string
	in := false
	key := regexp.MustCompile(`^\s+([A-Za-z]+)\s*=`)
	for line := range strings.SplitSeq(string(tmpl), "\n") {
		switch {
		case strings.HasPrefix(line, "["):
			in = strings.HasPrefix(line, "[data]")
		case in:
			if m := key.FindStringSubmatch(line); m != nil {
				keys = append(keys, m[1])
			}
		}
	}
	sort.Strings(keys)
	return keys
}

// Seed is the answers `chezmoi init` is given for a role. Identity answers this
// machine lacks become <name> placeholders; the other answers keep this machine's
// value or stay blank (feature off), and no secrets project is set, so the
// parallel renders never reach the Secrets Manager.
func Seed(role string, local map[string]any, harnesses []string) map[string]any {
	str := func(k string) string { v, _ := local[k].(string); return v }
	seed := map[string]any{
		"light":        role == roleLight,
		"work":         role == roleWork,
		"harnesses":    harnesses,
		"bwsProjectId": "",
	}
	for _, k := range []string{"gitUser", "gitEmail", "workEmail"} {
		seed[k] = cmpOr(str(k), "<"+k+">")
	}
	for _, k := range []string{"githubUser", "personalGitRemote", "otelEndpoint"} {
		seed[k] = str(k)
	}
	return seed
}

func cmpOr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// TOML renders a seed as the [data] table of a chezmoi config file.
func TOML(seed map[string]any) string {
	keys := make([]string, 0, len(seed))
	for k := range seed {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("[data]\n")
	for _, k := range keys {
		switch v := seed[k].(type) {
		case bool:
			fmt.Fprintf(&b, "%s = %t\n", k, v)
		case string:
			fmt.Fprintf(&b, "%s = %s\n", k, strconv.Quote(v))
		case []string:
			q := make([]string, len(v))
			for i, s := range v {
				q[i] = strconv.Quote(s)
			}
			fmt.Fprintf(&b, "%s = [%s]\n", k, strings.Join(q, ", "))
		}
	}
	return b.String()
}

// Build renders the config template once per role and returns every configuration
// with its overrides. cz must target the source dir src; scratch is a temp dir.
func Build(ctx context.Context, cz *chez.Client, src, scratch string) ([]Entry, error) {
	tmpl, err := os.ReadFile(filepath.Join(src, ".chezmoi.toml.tmpl")) //nolint:gosec // src is the user's own dotfiles source dir
	if err != nil {
		return nil, fmt.Errorf("read config template: %w", err)
	}
	harnesses, err := Harnesses(tmpl)
	if err != nil {
		return nil, err
	}
	local, err := cz.Data(ctx, nil)
	if err != nil {
		return nil, err
	}
	data := map[string]chez.Overrides{}
	for _, role := range Roles() {
		seed := Seed(role, local, harnesses)
		seedPath := filepath.Join(scratch, "seed-"+role+".toml")
		if err := os.WriteFile(seedPath, []byte(TOML(seed)), 0o600); err != nil {
			return nil, fmt.Errorf("write seed config: %w", err)
		}
		rendered, err := cz.With("--config", seedPath).Execute(ctx, nil, tmpl, true)
		if err != nil {
			return nil, fmt.Errorf("render config template for %s: %w", role, err)
		}
		renderedPath := filepath.Join(scratch, "config-"+role+".toml")
		if err = os.WriteFile(renderedPath, rendered, 0o600); err != nil { //nolint:gosec // renderedPath is inside the scratch dir this process created
			return nil, fmt.Errorf("write rendered config: %w", err)
		}
		got, err := cz.With("--config", renderedPath).Data(ctx, nil)
		if err != nil {
			return nil, err
		}
		o := chez.Overrides{}
		for k, zero := range seed {
			if v, ok := got[k]; ok {
				o[k] = v
			} else {
				o[k] = zeroOf(zero) // a key the role's render leaves out is off, not inherited
			}
		}
		data[role] = o
	}
	var entries []Entry
	for _, c := range All() {
		o := data[c.Role].Clone(chez.Overrides{"chezmoi": map[string]any{"os": c.OS}})
		entries = append(entries, Entry{Config: c, Overrides: o})
	}
	return entries, nil
}

func zeroOf(v any) any {
	switch v.(type) {
	case bool:
		return false
	case []string:
		return []string{}
	default:
		return ""
	}
}
