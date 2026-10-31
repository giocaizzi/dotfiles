package edit

import (
	"fmt"
	"os"
)

type scratch struct{ dir string }

func mkScratch() (scratch, error) {
	d, err := os.MkdirTemp("", "dots-")
	if err != nil {
		return scratch{}, fmt.Errorf("make scratch dir: %w", err)
	}
	return scratch{dir: d}, nil
}

func (s scratch) remove() { _ = os.RemoveAll(s.dir) }
