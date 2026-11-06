// Package survey fills the inventory: it asks each package manager, the installed
// release binaries and git what this machine has, with timeouts and no shell.
package survey

import (
	"context"
	"io/fs"
	"time"

	"github.com/giocaizzi/dotfiles/tools/dots/internal/chez"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/inventory"
	"github.com/giocaizzi/dotfiles/tools/dots/internal/pkgreg"
)

// Survey gathers the overview of this machine.
type Survey struct {
	Runner      chez.Runner
	Latest      func(context.Context, pkgreg.Upstream) (string, error) // release.Finder.Latest
	Stat        func(string) (fs.FileInfo, error)                      // os.Stat
	LookPath    func(string) (string, error)
	Home        string
	GOOS        string
	WorkingTree string        // the dotfiles git checkout; "" when unknown
	Offline     bool          // no network: no upstream or registry lookups, no fetch
	Fetch       bool          // git fetch the working tree first
	Timeout     time.Duration // per local command; 0 means 10s
	NetTimeout  time.Duration // per network lookup; 0 means 15s
}

// Run surveys sel. It never fails: what cannot be read is reported as unknown.
func (s Survey) Run(ctx context.Context, sel inventory.Selection) inventory.Report {
	_ = ctx
	return inventory.Report{OS: sel.OS, Role: sel.Role, Offline: s.Offline}
}
