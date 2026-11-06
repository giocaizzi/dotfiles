package catalog

import (
	"io"

	"github.com/giocaizzi/dotfiles/tools/agents/internal/term"
)

// Markdown renders docs/MODELS.md.
func Markdown(c Catalog) []byte {
	_ = c
	return nil // TODO(unit 2)
}

// WriteTerminal prints the catalog as coloured tables: one per tier, then
// untiered models, then the best model per arena category.
func WriteTerminal(w io.Writer, st term.Style, c Catalog) {
	_, _, _ = w, st, c // TODO(unit 2)
}
