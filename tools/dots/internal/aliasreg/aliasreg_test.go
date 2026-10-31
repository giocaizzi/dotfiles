package aliasreg

import (
	"strings"
	"testing"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/finding"
)

func alias(a map[string]any) map[string]any {
	return map[string]any{"alias": map[string]any{"x": a}}
}

func TestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		data map[string]any
		want string // substring of the findings; empty means valid
	}{
		{"cmd", alias(map[string]any{"cmd": "ls"}), ""},
		{"both shells", alias(map[string]any{"posix": "ls", "powershell": "dir"}), ""},
		{"limited with why", alias(map[string]any{"cmd": "ls", "shells": []any{"posix"}, "why": "no pwsh"}), ""},
		{"os and machines", alias(map[string]any{"cmd": "ls", "os": []any{"linux"}, "machines": []any{"work"}}), ""},
		{"cmd and posix", alias(map[string]any{"cmd": "ls", "posix": "ls"}), "use cmd or posix/powershell, not both"},
		{"unknown key", alias(map[string]any{"cmd": "ls", "bogus": 1}), "unknown key bogus"},
		{"missing powershell", alias(map[string]any{"posix": "ls"}), "no powershell command"},
		{"limited without why", alias(map[string]any{"cmd": "ls", "shells": []any{"posix"}}), "limited to [posix] without why"},
		{"empty shells", alias(map[string]any{"cmd": "ls", "shells": []any{}}), "shells must be a non-empty subset"},
		{"command for an excluded shell", alias(map[string]any{"cmd": "ls", "powershell": "dir", "shells": []any{"posix"}, "why": "w"}), "powershell command set but powershell not in shells"},
		{"bad os", alias(map[string]any{"cmd": "ls", "os": []any{"beos"}}), "os must be a subset"},
		{"bad machine", alias(map[string]any{"cmd": "ls", "machines": []any{"boss"}}), "machines must be a subset"},
		{"bad name", map[string]any{"alias": map[string]any{"-x": map[string]any{"cmd": "ls"}}}, "invalid name"},
		{"block ok", map[string]any{"aliasBlock": map[string]any{"pip": map[string]any{"message": "no"}}}, ""},
		{"block extra key", map[string]any{"aliasBlock": map[string]any{"pip": map[string]any{"message": "no", "x": 1}}}, "needs exactly a message string"},
		{"external ok", map[string]any{"aliasExternal": map[string]any{"zsh": map[string]any{"git": true}}}, ""},
		{"external not bool", map[string]any{"aliasExternal": map[string]any{"zsh": map[string]any{"git": "yes"}}}, "must be true or false"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := finding.Lines(Validate(tt.data))
			switch {
			case tt.want == "" && got != "":
				t.Errorf("Validate() = %q, want valid", got)
			case !strings.Contains(got, tt.want):
				t.Errorf("Validate() = %q, want it to contain %q", got, tt.want)
			}
		})
	}
}
