package inventory

import "errors"

// ParseSelection decodes SelectionTemplate's output, keeping only the package
// managers that apply on its OS.
func ParseSelection(b []byte) (Selection, error) {
	_ = b
	return Selection{}, errors.New("ParseSelection: not implemented")
}
