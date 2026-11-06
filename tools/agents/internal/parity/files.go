package parity

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/finding"
	"github.com/giocaizzi/dotfiles/tools/agents/internal/policy"
)

const (
	vscodeSettingsRel = "dot_config/Code/User/settings.json"
	claudeTemplateRel = ".chezmoitemplates/claude-settings.json"
)

// CheckSourceFiles compares the native voice settings in the chezmoi source
// directory with the policy: the static VS Code settings file must hold the
// policy values, and the Claude template must read them from the policy.
func CheckSourceFiles(p *policy.Policy, sourceDir string) []finding.Finding {
	fs := checkVSCodeFile(p, filepath.Join(sourceDir, filepath.FromSlash(vscodeSettingsRel)))
	return append(fs, checkClaudeTemplate(filepath.Join(sourceDir, filepath.FromSlash(claudeTemplateRel)))...)
}

func checkVSCodeFile(p *policy.Policy, path string) []finding.Finding {
	const where = "vscode"
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return []finding.Finding{finding.New(where, "cannot read %s: %v", path, err)}
	}
	var settings map[string]any
	if err := json.Unmarshal(stripJSONC(data), &settings); err != nil {
		return []finding.Finding{finding.New(where, "cannot parse %s: %v", path, err)}
	}

	var fs []finding.Finding
	dictation := true // VS Code default when the key is absent
	if v, ok := settings["dictation.enabled"]; ok {
		b, isBool := v.(bool)
		if !isBool {
			fs = append(fs, finding.New(where, "%s: dictation.enabled must be a boolean; edit settings.json to match the policy", path))
		}
		dictation = b
	}
	if dictation != p.Voice.Enabled {
		fs = append(fs, finding.New(where,
			"%s: dictation.enabled=%v but policy voice.enabled=%v; edit settings.json to match the policy",
			path, dictation, p.Voice.Enabled))
	}

	lang, hasLang := settings["agents.voice.language"]
	if p.Voice.Language == "" {
		if hasLang && lang != "auto" {
			fs = append(fs, finding.New(where,
				"%s: agents.voice.language=%v but policy voice.language is empty; edit settings.json to match the policy (remove the key)",
				path, lang))
		}
	} else if lang != p.Voice.Language {
		fs = append(fs, finding.New(where,
			"%s: agents.voice.language=%v but policy voice.language=%q; edit settings.json to match the policy",
			path, lang, p.Voice.Language))
	}
	return fs
}

func checkClaudeTemplate(path string) []finding.Finding {
	const where = "claude"
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return []finding.Finding{finding.New(where, "cannot read %s: %v", path, err)}
	}
	text := string(data)

	var fs []finding.Finding
	for _, ref := range []string{
		".harnessPolicy.voice.enabled",
		".harnessPolicy.voice.mode",
		".harnessPolicy.voice.autoSubmit",
	} {
		if !strings.Contains(text, ref) {
			fs = append(fs, finding.New(where, "%s must read %s from the policy", path, ref))
		}
	}
	if strings.Contains(text, `"language":`) {
		fs = append(fs, finding.New(where,
			"%s sets a \"language\" key; it also sets Claude's reply language, so it is deliberately unmanaged", path))
	}
	return fs
}

// stripJSONC turns JSONC into JSON: it drops // and /* */ comments outside
// strings and trailing commas before } or ].
func stripJSONC(src []byte) []byte {
	out := make([]byte, 0, len(src))
	for i := 0; i < len(src); {
		switch {
		case src[i] == '"':
			end := stringEnd(src, i)
			out = append(out, src[i:end]...)
			i = end
		case bytes.HasPrefix(src[i:], []byte("//")):
			for i < len(src) && src[i] != '\n' {
				i++
			}
		case bytes.HasPrefix(src[i:], []byte("/*")):
			end := bytes.Index(src[i+2:], []byte("*/"))
			if end < 0 {
				return out
			}
			i += end + 4
		default:
			out = append(out, src[i])
			i++
		}
	}
	return dropTrailingCommas(out)
}

// stringEnd returns the index just past the JSON string that starts at src[start].
func stringEnd(src []byte, start int) int {
	for i := start + 1; i < len(src); i++ {
		switch src[i] {
		case '\\':
			i++
		case '"':
			return i + 1
		}
	}
	return len(src)
}

func dropTrailingCommas(src []byte) []byte {
	out := make([]byte, 0, len(src))
	for i := 0; i < len(src); {
		switch {
		case src[i] == '"':
			end := stringEnd(src, i)
			out = append(out, src[i:end]...)
			i = end
		case src[i] == ',' && closesNext(src[i+1:]):
			i++
		default:
			out = append(out, src[i])
			i++
		}
	}
	return out
}

// closesNext reports whether the next non-space byte closes an object or array.
func closesNext(rest []byte) bool {
	rest = bytes.TrimLeft(rest, " \t\r\n")
	return len(rest) > 0 && (rest[0] == '}' || rest[0] == ']')
}
