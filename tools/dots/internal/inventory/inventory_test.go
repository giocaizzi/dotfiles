package inventory

import "testing"

func TestToolState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		src  Source
		tool Tool
		want State
	}{
		{"missing", Brew, Tool{Missing: true}, Missing},
		{"missing beats drift", Release, Tool{Missing: true, Installed: "1", Pinned: "2"}, Missing},
		{"drifted release", Release, Tool{Installed: "1.0.0", Pinned: "1.1.0", Latest: "1.1.0"}, Drifted},
		{"drifted npm pin", Npm, Tool{Installed: "7.0.2", Pinned: "5", Latest: "7.0.2"}, Drifted},
		{"npm pin prefix is whole components", Npm, Tool{Installed: "50.1.0", Pinned: "5", Latest: "50.1.0"}, Drifted},
		{"release outdated", Release, Tool{Installed: "1.0.0", Pinned: "1.0.0", Latest: "1.2.0"}, Outdated},
		{"release up to date", Release, Tool{Installed: "1.0.0", Pinned: "1.0.0", Latest: "1.0.0"}, UpToDate},
		{"release installed unknown", Release, Tool{Pinned: "1.0.0", Latest: "1.0.0"}, Unknown},
		{"release outdated needs no installed version", Release, Tool{Pinned: "1.0.0", Latest: "1.1.0"}, Outdated},
		{"release offline", Release, Tool{Installed: "1.0.0", Pinned: "1.0.0"}, Unknown},
		{"package outdated", Brew, Tool{Installed: "1.0", Latest: "1.1"}, Outdated},
		{"package up to date", Brew, Tool{Installed: "1.0", Latest: "1.0"}, UpToDate},
		{"package latest beyond its pin", Npm, Tool{Installed: "5.1.0", Pinned: "5", Latest: "6.0.0"}, UpToDate},
		{"package outdated within its pin", Npm, Tool{Installed: "5.1.0", Pinned: "5", Latest: "5.2.0"}, Outdated},
		{"package offline", UV, Tool{Installed: "1.0"}, Unknown},
		{"package installed unknown", Apt, Tool{Latest: "1.0"}, Unknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.tool.State(tt.src); got != tt.want {
				t.Errorf("State() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		version, pin string
		want         bool
	}{
		{"5.9.2", "5", true},
		{"5", "5", true},
		{"50.1", "5", false},
		{"5.9.2", "5.9", true},
		{"5.10", "5.1", false},
		{"7.0.2", "5", false},
		{"1.0", "1.0.0", false},
	}
	for _, tt := range tests {
		t.Run(tt.version+"/"+tt.pin, func(t *testing.T) {
			t.Parallel()
			if got := Matches(tt.version, tt.pin); got != tt.want {
				t.Errorf("Matches(%q, %q) = %v, want %v", tt.version, tt.pin, got, tt.want)
			}
		})
	}
}
